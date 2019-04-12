package file

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
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

	lineNumber := 1
	reader := bufio.NewReader(f)
	preFix := []byte{}
	for {
		line, isPrefix, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			d.ScanError = err
			break
		}

		preFix = append(preFix, line...)
		if isPrefix {
			continue
		}

		lineText := string(preFix)
		preFix = []byte{}
		if fn(lineText) {
			lineNumber++
			continue
		}

		matched, err := regexp.MatchString(d.matchString, lineText)
		if err != nil {
			d.ScanError = err
			break
		}
		if matched {
			d.linesContainingMatchString[lineNumber] = lineText
		}
		lineNumber++
	}
	d.isScanned = true
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
	fmt.Printf("search for files [*.%s] in [%s] directory\n", filterByFileExt, dir)
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

// Limit ...
func Limit() (uint64, error) {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return 0, err
	}
	return rLimit.Cur, nil
}

// ScanKorean ...
func ScanKorean(filePaths *[]string, verbose bool, fn isComment) <-chan *Data {
	cp := make(chan *Data)
	var wg sync.WaitGroup
	wg.Add(len(*filePaths))

	for _, filePath := range *filePaths {
		go func(filePath string) {
			defer wg.Done()
			if verbose {
				fmt.Printf("[%s] scanning Korean character in file\n", filePath)
			}

			fileData := New(filePath, "\\p{Hangul}")
			fileData.Scan(fn)
			cp <- fileData
		}(filePath)
	}

	go func() {
		wg.Wait()
		close(cp)
	}()
	return cp
}
