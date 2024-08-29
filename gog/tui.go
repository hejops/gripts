package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	Pkgs      []SearchPackage
	IndexPkgs []IndexPackage

	cursor int
}

// Init is the first function that will be called. It returns an optional
// initial command. To not perform an initial command return nil.
func (m Model) Init() tea.Cmd {
	return nil
	// return tea.EnterAltScreen
}

// Update is called when a message is received. Use it to inspect messages
// and, in response, update the model and/or send a command.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j":
			m.cursor++
		case "k":
			m.cursor--
		case "esc":
			return m, tea.Quit
		case "enter":
			return m, tea.Sequence(
				// tea.EnterAltScreen,
				// tea.ExitAltScreen,
				tea.Exec(&Install{pkg: m.Pkgs[m.cursor]}, nil),
				tea.Quit,
			)
		}
	}
	return m, nil
}

// View renders the program's UI, which is just a string. The view is
// rendered after every Update.
func (m Model) View() string {
	var items []string
	for i, p := range m.Pkgs {
		item := strconv.Itoa(i+1) + " " + p.Path
		switch m.cursor == i {
		case true:
			item = "> " + item
		case false:
			item = "  " + item
		}

		items = append(items, item)
	}
	sel := m.Pkgs[m.cursor]
	desc := sel.Synopsis
	if desc == "" {
		desc = "(No description)"
	}
	if isInstalled(sel.Path) {
		desc = "[installed] " + desc
	}
	return lipgloss.JoinVertical(
		lipgloss.Top,
		desc,
		strings.Join(items, "\n"),
	)
}

// Check if installed in ~/go/pkg/mod/
//
// ~/go/pkg/mod/github.com/user/package@v1.2.3
func isInstalled(path string) bool {
	base := filepath.Base(path)
	fullpath := filepath.Join(GO_PKG_BASE, path)
	parent, _ := filepath.Split(fullpath) // ~/go/.../github.com/user
	entries, err := os.ReadDir(parent)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if strings.Split(e.Name(), "@")[0] == base {
			return true
		}
	}
	return false
}

type Install struct {
	pkg SearchPackage
}

func (i *Install) Run() error {
	if isInstalled(i.pkg.Path) {
		fmt.Println("already installed:", i.pkg.Path)
		return nil
	}
	fmt.Println("installing", i.pkg.Path)
	err := exec.Command("go", "get", i.pkg.Path).Run()
	if err != nil {
		panic(err)
	}
	fmt.Println("ok")
	return nil
}

func (i Install) SetStdin(_ io.Reader)  {}
func (i Install) SetStdout(_ io.Writer) {}
func (i Install) SetStderr(_ io.Writer) {}
