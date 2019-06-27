package dir

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/loganstone/kpick/conf"
)

// Search ...
func Search(dir, filterByFileExt string, skip *regexp.Regexp) (*[]string, error) {
	var paths []string
	err := filepath.Walk(dir,
		func(path string, f os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if f.IsDir() || (*skip).MatchString(path) {
				return nil
			}

			if filterByFileExt != "" && filterByFileExt != conf.DefaultFileExt {
				v := strings.Split(f.Name(), ".")
				if v[len(v)-1] == filterByFileExt {
					paths = append(paths, path)
				}
				return nil
			}
			paths = append(paths, path)
			return nil
		})
	return &paths, err
}

// MakeSkipPathRegex ...
func MakeSkipPathRegex(skipPaths *string) (*regexp.Regexp, error) {
	paths := strings.Split(*skipPaths, ",")
	(*skipPaths) = strings.Join(paths, "|")
	return regexp.Compile(*skipPaths)
}

// Check ...
func Check(path *string) error {
	dirInfo, err := os.Stat((*path))
	if err != nil {
		return err
	}

	if !dirInfo.IsDir() {
		return fmt.Errorf("'%s' is not directory", (*path))
	}
	return nil
}
