package binarysearch

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBinarySearch_Down_Odd(t *testing.T) {
	b := New(1, 3)
	assert.Equal(t, 2, b.GetMidPoint())

	assert.True(t, b.Down())
	assert.Equal(t, 1, b.GetMidPoint())

	assert.False(t, b.Down())
	assert.False(t, b.Up())

	b = New(0, 8)
	assert.Equal(t, 4, b.GetMidPoint())

	assert.True(t, b.Down())
	assert.Equal(t, 2, b.GetMidPoint())

	assert.True(t, b.Down())
	assert.Equal(t, 1, b.GetMidPoint())

	assert.True(t, b.Down())
	assert.Equal(t, 0, b.GetMidPoint())

	assert.False(t, b.Up())
	assert.False(t, b.Down())
}

func TestBinarySearch_Up(t *testing.T) {
	b := New(1, 3)
	assert.Equal(t, 2, b.GetMidPoint())

	assert.True(t, b.Up())
	assert.Equal(t, 3, b.GetMidPoint())

	assert.False(t, b.Up())
	assert.False(t, b.Down())

	b = New(0, 8)
	assert.Equal(t, 4, b.GetMidPoint())

	assert.True(t, b.Up())
	assert.Equal(t, 6, b.GetMidPoint())

	assert.True(t, b.Up())
	assert.Equal(t, 7, b.GetMidPoint())

	assert.True(t, b.Up())
	assert.Equal(t, 8, b.GetMidPoint())

	assert.False(t, b.Down())
	assert.False(t, b.Up())
}
