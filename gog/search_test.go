package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearch(t *testing.T) {
	// note: these numbers are not always stable
	assert.Len(t, findPackage("mux"), 14)
	assert.Len(t, findIndexPackage("mux"), 4)

	assert.Len(t, findPackage("sixel"), 5)
	assert.Len(t, findIndexPackage("sixel"), 0)

	assert.Len(t, findPackage("bubbletea"), 7)
	assert.Len(t, findIndexPackage("bubbletea"), 0)

	assert.Equal(t, findPackage("progressbar")[0].Path, "github.com/schollz/progressbar/v3")

	assert.True(t, isInstalled("github.com/charmbracelet/bubbletea"))
	assert.False(t, isInstalled("github.com/charmbracelet/foo"))
}
