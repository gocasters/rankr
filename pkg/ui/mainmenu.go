package ui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type MenuItem string

func (m MenuItem) Title() string {
	return string(m)
}

func (m MenuItem) Description() string {
	return ""
}

func (m MenuItem) FilterValue() string {
	return string(m)
}

type model struct {
	list list.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {

		case "q", "esc":
			return m, tea.Quit

		case "enter":
			selected := m.list.SelectedItem().(MenuItem)
			switch selected {
			case "🏆 Scoreboard":
				return m, nil

			case "🔔 Notifications":
				return m, nil

			case "❌ Exit":
				return m, tea.Quit
			}
		}

	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return m.list.View()
}

func NewMainMenuModel() tea.Model {
	items := []list.Item{
		MenuItem("🏆 Scoreboard"),
		MenuItem("🔔 Notifications"),
		MenuItem("❌ Exit"),
	}

	l := list.New(items, list.NewDefaultDelegate(), 30, 10)
	l.Title = "🔥 Ranker App"

	return model{list: l}
}
