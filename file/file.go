package file

import (
	"bufio"
	"container/heap"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
	"sort"
	"sync"
	"syscall"
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

func newFile(path string, m, ig *regexp.Regexp) *File {
	return &File{path, m, ig, map[int][]byte{}, nil}
}

func limitNumber() int {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return 2048
	}
	if rLimit.Cur > math.MaxInt32 {
		return math.MaxInt32
	}
	return int(rLimit.Cur)
}

// ScanFiles ...
func ScanFiles(filePaths []string, m, ig *regexp.Regexp,
	beforeFn beforeScanFunc, afterFn afterScanFunc) <-chan *File {
	cp := make(chan *File)

	var wg sync.WaitGroup
	wg.Add(len(filePaths))

	for _, filePath := range filePaths {
		go func(filePath string) {
			defer wg.Done()
			beforeFn(filePath)
			f := newFile(filePath, m, ig)
			f.Scan()
			afterFn(filePath)
			cp <- f
		}(filePath)
	}

	go func() {
		wg.Wait()
		close(cp)
	}()
	return cp
}

// Chunks ...
func Chunks(foundFiles []string) [][]string {
	foundFilesCnt := len(foundFiles)
	chunkSize := limitNumber()

	// NOTE: "too many open files" io error 회피
	// 현재 열려있는 파일 수를 확인하는 것보다 더 간단하고,
	// 프로세스당 파일 제한 값의 반만 사용하더라도
	// 속도에는 크게 차이가 없다.
	chunkSize = chunkSize >> 1

	var chunks [][]string
	var i int
	for i = 0; i < foundFilesCnt; i += chunkSize {
		end := i + chunkSize

		if end > foundFilesCnt {
			end = foundFilesCnt
		}

		chunks = append(chunks, foundFiles[i:end])
	}
	return chunks
}

// PrintFiles .
func PrintFiles(files *SortedFiles) {
	for files.Len() > 0 {
		f, ok := heap.Pop(files).(*File)
		if ok {
			fmt.Println(f.Path())
			f.printFoundLines()
		}
	}
}
