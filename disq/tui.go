package main

import (
	"strconv"

	"github.com/charmbracelet/bubbles/table" // NOT lipgloss/table!
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	table table.Model

	searching bool
	input     string
}

func sqlToTable(query string, args ...any) table.Model {
	sqlRows := []JoinedRow{}
	if err := s.db.Select(&sqlRows, query, args...); err != nil {
		panic(err)
	}

	// fmt.Println(len(sqlRows), sqlRows[0])

	// table := sqlToTable(sqlRows[:50])
	var rows []table.Row
	for _, row := range sqlRows {
		rows = append(rows, table.Row{
			row.Artist,
			row.Title,
			strconv.Itoa(row.Year),
			strconv.Itoa(row.Rating),
		})
	}

	cols := []table.Column{
		{Title: "Artist", Width: 40},
		{Title: "Album", Width: 40},
		{Title: "Year", Width: 4},
		{Title: "Rating", Width: 6},
	}

	return table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
	)
}

func (m *model) filter() {
	// TODO: textinput for each column
	// TODO: buttons for sort field
	m.table = sqlToTable(
		`
		SELECT * FROM collection WHERE artist LIKE '%`+m.input+`%'
		ORDER BY year
		`,
		m.input,
	)
	m.table.SetHeight(10)
}

// Init is the first function that will be called. It returns an optional
// initial command. To not perform an initial command return nil.
func (m *model) Init() tea.Cmd {
	m.table.SetHeight(10)
	return nil
}

// Update is called when a message is received. Use it to inspect messages
// and, in response, update the model and/or send a command.
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:

		if m.searching {
			switch msg.String() {
			case "enter":
				m.filter()
				m.searching = false
				m.input = ""
			case "backspace":
				m.input = m.input[:len(m.input)-1]
			default:
				m.input += string(msg.Runes[0])
			}
			break
		}

		switch msg.String() {
		case "q":
			return m, tea.Quit

		case "/":
			m.searching = true

		case "j":
			m.table.MoveDown(1)
		case "k":
			m.table.MoveUp(1)

		case "pgdown":
			m.table.MoveDown(10)
		case "pgup":
			m.table.MoveUp(10)

		}
	}
	return m, nil
}

// View renders the program's UI, which is just a string. The view is
// rendered after every Update.
func (m *model) View() string {
	var header string
	switch m.searching {
	case true:
		header = "/" + m.input
	case false:
		header = "Searched: " + m.input
	}
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		m.table.View(),
	)
}
