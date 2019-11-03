package regex

import (
	"errors"
	"regexp"
	"strings"
)

// FileScan .
func FileScan(match, ignore string) (m, ig *regexp.Regexp, err error) {
	if match != "" {
		m, err = regexp.Compile(match)
		if err != nil {
			return
		}
	}

	if ignore != "" {
		ig, err = regexp.Compile(ignore)
		if err != nil {
			return
		}
	}
	return
}

// SkipPaths .
func SkipPaths(skipPaths, delimiter, separator string) (*regexp.Regexp, error) {
	if skipPaths == "" {
		return nil, errors.New("'skipPaths' is required")
	}

	paths := strings.Split(skipPaths, delimiter)
	return regexp.Compile(strings.Join(paths, separator))
}
