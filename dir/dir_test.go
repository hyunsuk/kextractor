package dir

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFinder(t *testing.T) {
	skipPaths, err := regexp.Compile(".git")
	assert.NoError(t, err)
	finder, err := NewFinder("invalid", "go", skipPaths)
	assert.Error(t, err)
	assert.Nil(t, finder)

	finder, err = NewFinder(".", "go", skipPaths)
	assert.NoError(t, err)
	assert.NotNil(t, finder)
	err = finder.Find()
	assert.NoError(t, err)
	assert.Equal(t, 2, finder.ResultCount())
	expected := [][]string{{"dir.go", "dir_test.go"}}
	assert.Equal(t, expected, finder.Chunk())

	finder, err = NewFinder("../", "*", skipPaths)
	assert.NoError(t, err)
	assert.NotNil(t, finder)
	err = finder.Find()
	assert.NoError(t, err)
	assert.Equal(t, 14, finder.ResultCount())
}
