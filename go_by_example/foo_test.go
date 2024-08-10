package main

import (
	"testing"

	// go get = go get -d (i.e. download) + go install
	// https://stackoverflow.com/a/46090493
	// https://go.dev/doc/go-get-install-deprecation

	// lsp offers autocomplete for pkgs, but only those that have been
	// installed via go get

	// Go lacks a native assert keyword/function, instead requiring t.Error
	// or t.Fail
	"github.com/stretchr/testify/assert"
)

// https://go.dev/doc/code#Testing

// https://www.digitalocean.com/community/tutorials/importing-packages-in-go#step-2-using-third-party-packages
// https://pkg.go.dev/github.com/stretchr/testify@v1.9.0/assert#hdr-Example_Usage
// https://stackoverflow.com/a/28240537
// test files -must- end in `_test.go` (ugh)

func TestSomething(t *testing.T) {
	assert.Equal(t, 1, 1)
	assert.NotEqual(t, 1, 2)
	// assert.Equal(t, 1, 2)

	tests := []struct {
		a int
		b int
	}{
		{1, 1},
		{2, 2},
		{3, 3},
	}
	for _, test := range tests {
		assert.Equal(t, test.a, test.b)
	}
}
