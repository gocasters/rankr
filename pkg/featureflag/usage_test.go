package featureflag

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/gocasters/rankr/pkg/path"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// flagMethods are the *Flags methods whose first argument is a flag name.
// Call sites must pass a string literal so the scan below can read it.
var flagMethods = map[string]struct{}{
	"State":   {},
	"Enabled": {},
	"Shadow":  {},
}

// usage is one read of a flag that is not in the registry.
type usage struct {
	file string
	line int
	name string
}

func (u usage) String() string {
	return fmt.Sprintf("%s:%d: %q", u.file, u.line, u.name)
}

// scanSources parses the given Go sources and reports every flag name passed
// as a string literal to a flag method that is missing from the registry.
// Names passed as anything other than a literal cannot be checked statically
// and are skipped.
func scanSources(files map[string]string) ([]usage, error) {
	fset := token.NewFileSet()

	var usages []usage
	for _, filename := range slices.Sorted(maps.Keys(files)) {
		tree, err := parser.ParseFile(fset, filename, files[filename], 0)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", filename, err)
		}

		ast.Inspect(tree, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || len(call.Args) == 0 {
				return true
			}
			if _, ok := flagMethods[sel.Sel.Name]; !ok {
				return true
			}
			lit, ok := call.Args[0].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			name, err := strconv.Unquote(lit.Value)
			if err != nil {
				return true
			}
			if !IsRegistered(name) {
				pos := fset.Position(lit.Pos())
				usages = append(usages, usage{file: pos.Filename, line: pos.Line, name: name})
			}
			return true
		})
	}

	return usages, nil
}

// TestScanDetectsUnregisteredFlag proves the scan has teeth: if code reads a
// flag that is not in the registry, the scan must report it.
func TestScanDetectsUnregisteredFlag(t *testing.T) {
	files := map[string]string{
		"ok.go": `package demo

func f(fl flagReader) bool { return fl.Enabled("scoring_v1") }
`,
		"bad.go": `package demo

func g(fl flagReader) bool { return fl.Shadow("ghost_flag") && fl.State("other_ghost") == 1 }
`,
		"dynamic.go": `package demo

func h(fl flagReader, name string) bool { return fl.Enabled(name) }
`,
	}

	usages, err := scanSources(files)
	require.NoError(t, err)
	require.Len(t, usages, 2, "expected exactly the two unregistered literals, got: %v", usages)

	var names []string
	for _, u := range usages {
		names = append(names, u.name)
	}
	assert.ElementsMatch(t, []string{"ghost_flag", "other_ghost"}, names)
}

// TestCodeOnlyReadsRegisteredFlags is the guard for the acceptance criterion:
// it fails when any non-test Go file in the repo reads a flag that is missing
// from the registry.
func TestCodeOnlyReadsRegisteredFlags(t *testing.T) {
	root, err := path.PathProjectRoot()
	require.NoError(t, err)

	files := map[string]string{}
	walkErr := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() {
			if p != root && strings.HasPrefix(name, ".") {
				return fs.SkipDir
			}
			if name == "vendor" || p == filepath.Join(root, "protobuf", "golang") {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			return nil
		}
		content, readErr := os.ReadFile(p)
		if readErr != nil {
			return readErr
		}
		files[filepath.ToSlash(strings.TrimPrefix(p, root+string(filepath.Separator)))] = string(content)
		return nil
	})
	require.NoError(t, walkErr)
	require.NotEmpty(t, files, "no Go sources found under %s", root)

	usages, err := scanSources(files)
	require.NoError(t, err)
	assert.Empty(t, usages, "code reads flags missing from the registry: %v", usages)
}
