package utility

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsNil(t *testing.T) {
	val1 := 0
	assert.Equal(t, false, IsNil(val1), "0 should not be nil")
	val2 := []string{}
	assert.Equal(t, true, IsNil(val2), "empty array should be nil")
	val3 := make(map[string]string)
	assert.Equal(t, true, IsNil(val3), "empty map should be nil")
	assert.Equal(t, true, IsNil(nil), "nil should be nil")
	assert.Equal(t, false, IsNil(val1, nil), "int and nil array should not be nil")
	assert.Equal(t, true, IsNil(val2, nil), "empty array and nil array should be nil")
}

func TestGetError(t *testing.T) {
	assert.Equal(t, "testing error", GetError(errors.New("testing error")), "It should return the error with error message")
	assert.Equal(t, "", GetError(nil), "It should return the empty error message")
}
