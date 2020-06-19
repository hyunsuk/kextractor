package conf

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpts(t *testing.T) {
	opts := Opts()
	assert.Equal(t, "", opts.Cpuprofile)
	assert.Equal(t, "", opts.Memprofile)

	expected, err := os.Getwd()
	assert.NoError(t, err)
	assert.Equal(t, expected, opts.DirPathToFind)
	assert.Equal(t, DefaultFilenameExt, opts.FileExtToScan)
	assert.Equal(t, MustIncludeSkipPaths, opts.SkipPaths)
	assert.Equal(t, "", opts.IgnoreRegexString)
	assert.Equal(t, false, opts.Verbose)
	assert.Equal(t, false, opts.Interactive)
	assert.Equal(t, false, opts.ErrorOnly)
}

func TestSkipPathsRegex(t *testing.T) {
	opts := Opts()
	skip, err := opts.SkipPathsRegex()
	assert.NoError(t, err)
	assert.True(t, skip.Match([]byte(".git")))
	assert.True(t, skip.Match([]byte(".tmp")))
	assert.False(t, skip.Match([]byte("invalid")))

	opts.SkipPaths = ""
	_, err = opts.SkipPathsRegex()
	assert.Error(t, ErrSkipPathsIsRequired, err)
}

func TestMatch(t *testing.T) {
	opts := Opts()
	match, err := opts.Match()
	assert.NoError(t, err)
	assert.True(t, match.Match([]byte("한글")))
	assert.True(t, match.Match([]byte("한글 Korean")))
	assert.True(t, match.Match([]byte("1,2,3 한글 Korean")))
	assert.False(t, match.Match([]byte("invalid")))
}

func TestIgnore(t *testing.T) {
	opts := Opts()
	ignore, err := opts.Ignore()
	assert.NoError(t, err)
	assert.Nil(t, ignore, err)

	opts.IgnoreRegexString = "//|#|/\\*"
	ignore, err = opts.Ignore()
	assert.NoError(t, err)
	assert.NotNil(t, ignore, err)
	assert.True(t, ignore.Match([]byte("//")))
	assert.True(t, ignore.Match([]byte("#")))
	assert.True(t, ignore.Match([]byte("/*")))
	assert.False(t, ignore.Match([]byte("invalid")))
}
