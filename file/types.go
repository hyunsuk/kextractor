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
func (f *File) Scan() {
	file, err := os.Open(f.path)
	if err != nil {
		f.scanError = err
		return
	}

	defer file.Close()

	reader := bufio.NewReader(file)
	line := []byte{}
	var lineNumber int

	for {
		chunk, isPrefix, err := reader.ReadLine()
		if err != nil {
			if err != io.EOF {
				f.scanError = err
			}
			break
		}

		line = append(line, chunk...)
		if isPrefix {
			continue
		}

		lineNumber++
		ignore := f.ignoreRegex != nil && f.ignoreRegex.Match(line)
		if !ignore && f.matchRegex != nil && f.matchRegex.Match(line) {
			f.foundLines[lineNumber] = line
		}

		line = []byte{}
	}
}

// Path ...
func (f *File) Path() string {
	return f.path
}

// Error ...
func (f *File) Error() error {
	return f.scanError
}

// FoundLines ...
func (f *File) FoundLines() map[int][]byte {
	return f.foundLines
}

func (f *File) printFoundLines() {
	lineNumbers := make([]int, len(f.foundLines))
	var i int
	for lineNumber := range f.foundLines {
		lineNumbers[i] = lineNumber
		i++
	}

	sort.Ints(lineNumbers)
	for _, lineNumber := range lineNumbers {
		lineText, _ := f.foundLines[lineNumber]
		fmt.Printf("%d: %s\n", lineNumber, lineText)
	}
}

// Heap .
type Heap []*File

func (h Heap) Len() int {
	return len(h)
}

func (h Heap) Less(i, j int) bool {
	return h[i].path < h[j].path
}

func (h Heap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

// Push .
func (h *Heap) Push(x interface{}) {
	*h = append(*h, x.(*File))
}

// Pop .
func (h *Heap) Pop() interface{} {
	old := *h
	n := len(old)
	element := old[n-1]
	*h = old[0 : n-1]
	return element
}
