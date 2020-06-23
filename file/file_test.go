package file

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScanFiles(t *testing.T) {
	paths := []string{"test1.md", "test2.md"}

	match, err := regexp.Compile("\\p{Hangul}")
	assert.NoError(t, err)

	ignore, err := regexp.Compile("#")
	assert.NoError(t, err)

	beforeFn := func(filePath string) {}
	afterFn := func(filePath string) {}

	for f := range ScanFiles(paths, match, ignore, beforeFn, afterFn) {
		assert.NoError(t, f.Error())
		assert.Equal(t, 1, len(f.MatchedLines()))
	}
}
