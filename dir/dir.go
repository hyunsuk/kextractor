package dir

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const findAllFileExt = "*"

// Find ...
func Find(rootPath, matchFileExt string, skip *regexp.Regexp) ([]string, error) {
	var paths []string
	err := filepath.Walk(rootPath,
		func(path string, f os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if f.IsDir() || (*skip).MatchString(path) {
				return nil
			}

			if matchFileExt != "" && matchFileExt != findAllFileExt {
				v := strings.Split(f.Name(), ".")
				if v[len(v)-1] == matchFileExt {
					paths = append(paths, path)
				}
				return nil
			}
			paths = append(paths, path)
			return nil
		})
	return paths, err
}

// MakeSkipPathRegex ...
func MakeSkipPathRegex(skipPaths, delimiter, separator string) (*regexp.Regexp, error) {
	if skipPaths == "" {
		return nil, errors.New("'skipPaths' is required")
	}

	paths := strings.Split(skipPaths, delimiter)
	return regexp.Compile(strings.Join(paths, separator))
}

// Check ...
func Check(path string) error {
	dirInfo, err := os.Stat(path)
	if err != nil {
		// TODO(logan): Should test.
		if os.IsPermission(err) {
			mode := dirInfo.Mode()
			log.Printf("'%s' permission is '%s'\n", path, mode.Perm())
		}
		return err
	}

	mode := dirInfo.Mode()

	if mode.IsRegular() {
		return fmt.Errorf("'%s' is regular file", path)
	}

	if !mode.IsDir() {
		return fmt.Errorf("'%s' is not directory", path)
	}
	return nil
}
