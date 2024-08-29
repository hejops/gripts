package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearch(t *testing.T) {
	assert.Len(t, findPackage("mux"), 13)
	assert.Len(t, findIndexPackage("mux"), 4)

	assert.Len(t, findPackage("sixel"), 5)
	assert.Len(t, findIndexPackage("sixel"), 0)

	assert.Len(t, findPackage("bubbletea"), 7)
	assert.Len(t, findIndexPackage("bubbletea"), 0)

	assert.True(t, isInstalled("github.com/charmbracelet/bubbletea"))
	assert.False(t, isInstalled("github.com/charmbracelet/foo"))
}
