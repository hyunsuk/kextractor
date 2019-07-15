package file

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
)

type beforeScanFunc func(path string)
type afterScanFunc func(path string)

// File ...
type File struct {
	path        string
	matchRegex  *regexp.Regexp
	ignoreRegex *regexp.Regexp
	foundLines  map[int][]byte
	scanError   error
}

// Scan ...
func (s *File) Scan() {
	f, err := os.Open(s.path)
	if err != nil {
		s.scanError = err
		return
	}

	defer f.Close()

	reader := bufio.NewReader(f)
	line := []byte{}
	var lineNumber int

	for {
		chunk, isPrefix, err := reader.ReadLine()
		if err != nil {
			if err != io.EOF {
				s.scanError = err
			}
			break
		}

		line = append(line, chunk...)
		if isPrefix {
			continue
		}

		lineNumber++
		ignore := s.ignoreRegex != nil && s.ignoreRegex.Match(line)
		if !ignore && s.matchRegex != nil && s.matchRegex.Match(line) {
			s.foundLines[lineNumber] = line
		}

		line = []byte{}
	}
}

// Path ...
func (s *File) Path() string {
	return s.path
}

// Error ...
func (s *File) Error() error {
	return s.scanError
}

// FoundLines ...
func (s *File) FoundLines() map[int][]byte {
	return s.foundLines
}

func (s *File) printFoundLines() {
	lineNumbers := make([]int, len(s.foundLines))
	var i int
	for lineNumber := range s.foundLines {
		lineNumbers[i] = lineNumber
		i++
	}

	sort.Ints(lineNumbers)
	for _, lineNumber := range lineNumbers {
		lineText, _ := s.foundLines[lineNumber]
		fmt.Printf("%d: %s\n", lineNumber, lineText)
	}
}

// SortedFiles .
type SortedFiles []*File

func (s SortedFiles) Len() int {
	return len(s)
}

func (s SortedFiles) Less(i, j int) bool {
	return s[i].path < s[j].path
}

func (s SortedFiles) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Push .
func (s *SortedFiles) Push(x interface{}) {
	*s = append(*s, x.(*File))
}

// Pop .
func (s *SortedFiles) Pop() interface{} {
	old := *s
	n := len(old)
	element := old[n-1]
	*s = old[0 : n-1]
	return element
}
