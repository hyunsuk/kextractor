package file

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type isComment func(s string) bool
type isSkipPath func(s string) bool

// Data ...
type Data struct {
	path                       string
	matchString                string
	linesContainingMatchString map[int]string
	isScanned                  bool
	ScanError                  error
}

// New ...
func New(path string, matchString string) *Data {
	return &Data{path, matchString, map[int]string{}, false, nil}
}

// Scan ...
func (d *Data) Scan(fn isComment) {
	f, err := os.Open(d.path)
	if err != nil {
		d.ScanError = err
		return
	}

	defer f.Close()

	d.isScanned = true
	lineNumber := 1
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lineText := scanner.Text()
		if fn(lineText) {
			lineNumber++
			continue
		}

		matched, err := regexp.MatchString(d.matchString, lineText)
		if err != nil {
			d.ScanError = err
			return
		}
		if matched {
			d.linesContainingMatchString[lineNumber] = lineText
		}
		lineNumber++
	}

	if err := scanner.Err(); err != nil {
		d.ScanError = err
		return
	}
}

// HasMatchedString ...
func (d *Data) HasMatchedString() bool {
	if !d.isScanned {
		return false
	}
	return len(d.linesContainingMatchString) > 0
}

// Path ...
func (d *Data) Path() string {
	return d.path
}

// MatchedLine ...
func (d *Data) MatchedLine() *map[int]string {
	return &d.linesContainingMatchString
}

// Search ...
func Search(dir string, filterByFileExt string, fn isSkipPath) (*[]string, error) {
	var resultPaths []string
	err := filepath.Walk(dir,
		func(path string, f os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if f.IsDir() || fn(path) {
				return nil
			}

			if filterByFileExt != "" {
				v := strings.Split(f.Name(), ".")
				if v[len(v)-1] == filterByFileExt {
					resultPaths = append(resultPaths, path)
				}
				return nil
			}
			resultPaths = append(resultPaths, path)
			return nil
		})
	return &resultPaths, err
}
