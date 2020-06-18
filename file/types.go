package file

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
)

// BeforeScanFunc .
type BeforeScanFunc func(path string)

// AfterScanFunc .
type AfterScanFunc func(path string)

// File ...
type File struct {
	path         string
	matchRegex   *regexp.Regexp
	ignoreRegex  *regexp.Regexp
	matchedLines map[int][]byte
	scanError    error
}

// Scan ...
func (f *File) Scan() {
	if f.matchRegex == nil {
		return
	}

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
			// NOTE(hs.lee): 줄 읽기가 다 끝나지 않았음. line 유지
			continue
		}

		// NOTE(hs.lee): 줄 읽기가 끝남
		lineNumber++
		if f.ignoreRegex != nil && f.ignoreRegex.Match(line) {
			line = []byte{}
			continue
		}

		if f.matchRegex.Match(line) {
			f.matchedLines[lineNumber] = line
		}

		line = []byte{}
	}
}

// Path returns a file path.
func (f *File) Path() string {
	return f.path
}

// Error returns an error scanned file.
func (f *File) Error() error {
	return f.scanError
}

// MatchedLines ...
func (f *File) MatchedLines() map[int][]byte {
	return f.matchedLines
}

func (f *File) printMatchedLines() {
	lineNumbers := make([]int, len(f.matchedLines))
	var i int
	for lineNumber := range f.matchedLines {
		lineNumbers[i] = lineNumber
		i++
	}

	sort.Ints(lineNumbers)
	for _, lineNumber := range lineNumbers {
		lineText, _ := f.matchedLines[lineNumber]
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
