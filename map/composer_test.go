package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocation(t *testing.T) {
	// TODO: parametrised test
	// https://semaphoreci.com/blog/table-driven-unit-tests-go
	// https://gist.github.com/zaky/8304b5abbdc03600c29e
	c := newComposer("Johann Sebastian Bach")
	// assert.Equal(t, c.getLocation(), "Eisenach")
	assert.Equal(t, c.Place, "Eisenach")
}
