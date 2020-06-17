package dir

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const findAllFileExt = "*"

// Finder .
type Finder struct {
	path      string
	fileExt   string
	skipPaths *regexp.Regexp
	Result    []string
}

// NewFinder .
func NewFinder(path, fileExt string, skipPaths *regexp.Regexp) (*Finder, error) {
	if err := check(path); err != nil {
		return nil, err
	}
	return &Finder{
		path:      path,
		fileExt:   fileExt,
		skipPaths: skipPaths,
	}, nil
}

// Find .
func (f *Finder) Find() error {
	return filepath.Walk(f.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || (*f.skipPaths).MatchString(path) {
			return nil
		}

		if f.fileExt == "" || f.fileExt == findAllFileExt {
			f.Result = append(f.Result, path)
			return nil
		}

		v := strings.Split(info.Name(), ".")
		if v[len(v)-1] == f.fileExt {
			f.Result = append(f.Result, path)
		}
		return nil
	})
}

// Check ...
func check(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		// TODO(logan): Should test.
		if os.IsPermission(err) {
			mode := info.Mode()
			log.Printf("'%s' permission is '%s'\n", path, mode.Perm())
		}
		return err
	}

	mode := info.Mode()

	if mode.IsRegular() {
		return fmt.Errorf("'%s' is regular file", path)
	}

	if !mode.IsDir() {
		return fmt.Errorf("'%s' is not directory", path)
	}
	return nil
}
