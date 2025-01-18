package main

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/list"
)

type model struct {
	dir    string
	files  []string
	marked map[int]bool
	cursor int
}

// Init is the first function that will be called. It returns an optional
// initial command. To not perform an initial command return nil.
func (m *model) Init() tea.Cmd {
	// go play(m.files)
	// go play(m.dir)
	go play(m.files[m.cursor])
	marked := make(map[int]bool)
	for i := range m.files {
		marked[i] = false
	}
	m.marked = marked
	return nil
}

func play(file string) {
	_ = exec.Command("pkill", "mpv").Run()
	// [ext=webm] overrides profile
	args := []string{"--config=no", "--audio-display=no", "--video=no", "--start=30%"}
	// args = append(args, files...)
	args = append(args, file)
	_ = exec.Command("mpv", args...).Run()
}

// // if config=no, playerctl doesn't detect mpv (only goes to show how brittle playerctl is)
// func next() { _ = exec.Command("playerctl", "next").Run() }
// func prev() { _ = exec.Command("playerctl", "previous").Run() }

// func next() { _ = exec.Command("playerctl", "next").Run() }
// func prev() { _ = exec.Command("playerctl", "previous").Run() }

// Update is called when a message is received. Use it to inspect messages
// and, in response, update the model and/or send a command.
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m, tea.ClearScreen

	case tea.KeyMsg:
		switch msg.String() {
		case " ":
			m.marked[m.cursor] = !m.marked[m.cursor]
		// TODO: slsk js client (if ambitious)

		case "j":
			if m.cursor < len(m.files)-1 {
				m.cursor++
				go play(m.files[m.cursor])
			}
			return m, nil
		case "k":
			if m.cursor > 0 {
				m.cursor--
				go play(m.files[m.cursor])
			}
			return m, nil
		case "ctrl+c", "esc", "q":
			_ = exec.Command("pkill", "mpv").Run()
			var sel []string
			for f, marked := range m.marked {
				if marked {
					sel = append(sel, m.files[f])
				}
			}
			_ = os.WriteFile("selected", []byte(strings.Join(sel, "\n")), 0644)
			// os.RemoveAll(Cfg.Dest)
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the program's UI, which is just a string. The view is
// rendered after every Update.
func (m *model) View() string {
	// TODO: windowing
	l := list.New(m.files).Enumerator(func(items list.Items, index int) (arrow string) {
		switch index == m.cursor {
		case true:
			arrow += "> "
		case false:
			arrow += "  "
		}

		switch m.marked[index] {
		case true:
			arrow += "/ "
		case false:
			arrow += "  "
		}

		return arrow
	})
	return lipgloss.NewStyle().MaxHeight(50).Render(l.String())
}

func browseFiles(dir string) []string {
	files := []string{}
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})

	m, err := tea.NewProgram(&model{
		dir:   dir,
		files: files,
	}, tea.WithAltScreen()).Run()
	if err != nil {
		panic(err)
	}
	// https://github.com/PS6/huh/blob/4bd4657a36ac2db392a0fa2a878773660b5e1e5b/form.go#L415
	return m.(*model).files
}
