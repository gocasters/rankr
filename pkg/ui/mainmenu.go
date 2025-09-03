package ui

import (
	"bufio"
	"context"
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"io"
	"os/exec"
	"strings"
	"sync"
)

type state int

const (
	selectService state = iota
	selectCommand
	enterFlags
	executing
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFA500")).
			MarginLeft(2)

	docStyle = lipgloss.NewStyle().Margin(1, 2)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF69B4")).
			Padding(0, 1).
			Margin(1, 0)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true).
			Margin(1, 0)
)

// item represents a list item for both services and commands
type item struct {
	title       string
	description string
	command     *cobra.Command
	flags       string
}

func (i item) FilterValue() string { return i.title }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%s", i.title)

	fn := lipgloss.NewStyle().PaddingLeft(4).Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("#FF69B4")).
				Bold(true).
				Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type logMsg struct {
	line  string
	isErr bool
}

func waitForLog(ch chan logMsg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return commandFinishedMsg{success: true, output: "finished"}
		}
		return msg
	}
}

type model struct {
	root      *cobra.Command
	cancel    context.CancelFunc
	state     state
	list      list.Model
	textInput textinput.Model
	selected  *cobra.Command
	width     int
	height    int

	logCh chan logMsg // channel for streaming logs
	logs  strings.Builder
}

func NewModel(root *cobra.Command) *model {
	ti := textinput.New()
	ti.Placeholder = "Enter flags/args (e.g., up, down, --flag=value)"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	items := make([]list.Item, 0, len(root.Commands()))
	for _, cmd := range root.Commands() {
		items = append(items, item{
			title:       cmd.Name(),
			description: cmd.Short,
			command:     cmd,
		})
	}

	const defaultWidth = 20
	const listHeight = 14

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Select a Service"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	l.Styles.HelpStyle = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)

	return &model{
		root:      root,
		state:     selectService,
		list:      l,
		textInput: ti,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

type commandFinishedMsg struct {
	success bool
	output  string
	err     error
}

func (m *model) executeCommand(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		selectedItem, ok := m.list.SelectedItem().(item)
		if !ok {
			return commandFinishedMsg{success: false, err: fmt.Errorf("no selected item")}
		}

		serviceName := m.selected.Name()
		commandName := selectedItem.command.Name()
		args := m.parseFlags(m.textInput.Value())

		cmdArgs := []string{"run", "cmd/ci/main.go", serviceName, commandName}
		cmdArgs = append(cmdArgs, args...)

		cmd := exec.CommandContext(ctx, "go", cmdArgs...)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return commandFinishedMsg{success: false, err: err}
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return commandFinishedMsg{success: false, err: err}
		}

		if err := cmd.Start(); err != nil {
			return commandFinishedMsg{success: false, err: err}
		}

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				m.logCh <- logMsg{line: scanner.Text(), isErr: false}
			}
		}()
		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				m.logCh <- logMsg{line: scanner.Text(), isErr: true}
			}
		}()

		go func() {
			// wait for command to finish
			err := cmd.Wait()
			wg.Wait() // wait until stdout/stderr finished

			if err != nil {
				m.logCh <- logMsg{line: err.Error(), isErr: true}
			}
			close(m.logCh)
		}()

		return commandFinishedMsg{success: true, output: "finished"}
	}
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetWidth(msg.Width)
		return m, nil
	case logMsg:
		_, err := fmt.Fprintln(&m.logs, msg.line)
		if err != nil {
			return m, nil
		}
		return m, waitForLog(m.logCh) // keep listening until channel closes

	case commandFinishedMsg:
		_, err := fmt.Fprintln(&m.logs, msg.output)
		if err != nil {
			return m, nil
		}
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case enterFlags:
			switch msg.String() {
			case "ctrl+c":
				m.state = selectService
				m.selected = nil
				m.setupServiceList()
				return m, nil
			case "enter":
				ctx, cancel := context.WithCancel(context.Background())
				m.cancel = cancel
				m.state = executing
				m.logCh = make(chan logMsg, 10)
				return m, tea.Batch(
					m.executeCommand(ctx),
					waitForLog(m.logCh),
				)
			default:
				var cmd tea.Cmd
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}

		case executing:
			switch msg.String() {
			case "ctrl+c":
				if m.cancel != nil {
					m.cancel()
				}
				m.logs.Reset()
				m.state = selectCommand
				return m, nil
			}
			// ignore other keys while executing
			return m, nil
		default:
			switch keypress := msg.String(); keypress {
			case "ctrl+c":
				if m.state == selectCommand {
					if m.cancel != nil {
						m.cancel()
					}
					// Go back to service selection
					m.state = selectService
					m.selected = nil
					m.setupServiceList()
					return m, nil
				}
				return m, tea.Quit

			case "enter":
				if m.state == selectService {
					i, ok := m.list.SelectedItem().(item)
					if ok {
						m.selected = i.command
						m.state = selectCommand
						m.setupCommandList()
					}
					return m, nil
				} else if m.state == selectCommand {
					i, ok := m.list.SelectedItem().(item)
					if ok {
						cmd := i.command

						if cmd.Flags().HasFlags() {
							m.state = enterFlags
							m.textInput.SetValue(i.flags)
							return m, nil
						} else {
							ctx, cancel := context.WithCancel(context.Background())
							m.cancel = cancel
							m.state = executing
							m.logCh = make(chan logMsg, 10)
							return m, tea.Batch(
								m.executeCommand(ctx), // start command
								waitForLog(m.logCh),   // start listening for logs
							)
						}

					}
					return m, nil
				}

			}
		}
	}

	return m, nil
}

func (m *model) setupServiceList() {
	items := make([]list.Item, 0, len(m.root.Commands()))
	for _, cmd := range m.root.Commands() {
		items = append(items, item{
			title:       cmd.Name(),
			description: cmd.Short,
			command:     cmd,
		})
	}
	m.list.SetItems(items)
	m.list.Title = "Select a Service"
}

func (m *model) setupCommandList() {
	if m.selected == nil {
		return
	}

	items := make([]list.Item, 0, len(m.selected.Commands()))
	for _, cmd := range m.selected.Commands() {
		items = append(items, item{
			title:       cmd.Name(),
			description: cmd.Short,
			command:     cmd,
		})
	}
	m.list.SetItems(items)
	m.list.Title = fmt.Sprintf("Service: %s - Select a Command", m.selected.Name())
}

func (m *model) View() string {
	switch m.state {
	case selectService, selectCommand:
		return docStyle.Render(m.list.View())
	case executing:
		return docStyle.Render(
			titleStyle.Render("Executing...") + "\n\n" +
				m.logs.String() + "\n\n" +
				helpStyle.Render("Press ctrl+c to back"),
		)

	case enterFlags:
		selectedItem, _ := m.list.SelectedItem().(item)
		cmdName := selectedItem.command.Name()

		help := m.getCommandHelp(selectedItem.command)

		return docStyle.Render(
			titleStyle.Render(fmt.Sprintf("Command: %s %s", m.selected.Name(), cmdName)) + "\n\n" +
				helpStyle.Render(help) + "\n" +
				"Enter flags:\n" +
				inputStyle.Render(m.textInput.View()) + "\n" +
				helpStyle.Render("Press Enter to execute, ctrl+c to go back"),
		)
	}
	return ""
}

func (m *model) getCommandHelp(cmd *cobra.Command) string {
	var help strings.Builder

	if cmd.Short != "" {
		help.WriteString(cmd.Short)
		help.WriteString("\n")
	}

	if cmd.Flags().HasFlags() {
		help.WriteString("\nAvailable flags:\n")
		cmd.Flags().VisitAll(func(flag *pflag.Flag) {
			help.WriteString(fmt.Sprintf("  --%s", flag.Name))
			if flag.Shorthand != "" {
				help.WriteString(fmt.Sprintf(", -%s", flag.Shorthand))
			}
			if flag.Usage != "" {
				help.WriteString(fmt.Sprintf(" - %s", flag.Usage))
			}
			help.WriteString("\n")
		})
	}

	return help.String()
}

func (m *model) parseFlags(flagsStr string) []string {
	if strings.TrimSpace(flagsStr) == "" {
		return []string{}
	}

	// Simple parsing - split by spaces but handle quoted strings
	var args []string
	var current strings.Builder
	inQuotes := false

	for i, r := range flagsStr {
		switch r {
		case '"':
			inQuotes = !inQuotes
		case ' ':
			if !inQuotes {
				if current.Len() > 0 {
					args = append(args, current.String())
					current.Reset()
				}
			} else {
				current.WriteRune(r)
			}
		default:
			current.WriteRune(r)
		}

		if i == len(flagsStr)-1 && current.Len() > 0 {
			args = append(args, current.String())
		}
	}

	return args
}
