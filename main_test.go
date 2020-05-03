package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGo(t *testing.T) {
	a := example()
	assert.Equal(t, a, "Hello, World", "The two should be the same.")
}
