// TUI wrapper for `go get`

package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("An argument is required")
		os.Exit(1)
	}

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
