package main

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	pkgs := findPackage(os.Args[1])

	_, err := tea.NewProgram(
		Model{Pkgs: pkgs},
		tea.WithAltScreen(),
	).Run()
	if err != nil {
		panic(err)
	}

	// pkgs := findIndexPackage("mux")
	// fmt.Println(pkgs) // 2000
}
