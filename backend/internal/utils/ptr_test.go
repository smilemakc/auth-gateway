package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPtr(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		val := "test"
		ptr := Ptr(val)
		assert.NotNil(t, ptr)
		assert.Equal(t, val, *ptr)
	})

	t.Run("int", func(t *testing.T) {
		val := 123
		ptr := Ptr(val)
		assert.NotNil(t, ptr)
		assert.Equal(t, val, *ptr)
	})

	t.Run("bool", func(t *testing.T) {
		val := true
		ptr := Ptr(val)
		assert.NotNil(t, ptr)
		assert.Equal(t, val, *ptr)
	})
}
