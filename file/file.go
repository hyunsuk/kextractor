package file

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"sync"
	"syscall"

	"github.com/loganstone/kpick/conf"
)

const (
	startLineNumber = 1
)

// Source ...
type Source struct {
	path        string
	matchRegex  *regexp.Regexp
	ignoreRegex *regexp.Regexp
	foundLines  map[int][]byte
	isScanned   bool
	scanError   error
}

// New ...
func New(path string, m, ig *regexp.Regexp) *Source {
	return &Source{
		path,
		m,
		ig,
		map[int][]byte{},
		false,
		nil,
	}
}

// Scan ...
func (s *Source) Scan() {
	f, err := os.Open(s.path)
	if err != nil {
		s.scanError = err
		return
	}

	defer f.Close()

	reader := bufio.NewReader(f)
	newLine := []byte{}
	lineNumber := startLineNumber
	for {
		lineChunk, isPrefix, err := reader.ReadLine()
		if err != nil {
			if err != io.EOF {
				s.scanError = err
			}
			break
		}

		newLine = append(newLine, lineChunk...)
		if isPrefix {
			continue
		}

		if s.ignoreRegex != nil && s.ignoreRegex.Match(newLine) {
			newLine = []byte{}
			lineNumber++
			continue
		}

		if s.matchRegex != nil && s.matchRegex.Match(newLine) {
			s.foundLines[lineNumber] = newLine
		}

		newLine = []byte{}
		lineNumber++
	}
	s.isScanned = true
}

// Path ...
func (s *Source) Path() string {
	return s.path
}

// Error ...
func (s *Source) Error() error {
	return s.scanError
}

// FoundLines ...
func (s *Source) FoundLines() *map[int][]byte {
	return &s.foundLines
}

// PrintFoundLines ...
func (s *Source) PrintFoundLines() {
	keys := make([]int, len(s.foundLines))
	i := 0
	for k := range s.foundLines {
		keys[i] = k
		i++
	}

	sort.Ints(keys)
	for _, k := range keys {
		v, _ := s.foundLines[k]
		fmt.Printf("%d: %s\n", k, v)
	}
}

// MakeRegexForScan ...
func MakeRegexForScan(match, ignore string) (m, ig *regexp.Regexp, err error) {
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

// LimitNumberOfFiles ...
func LimitNumberOfFiles() (uint64, error) {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return 0, err
	}
	return rLimit.Cur, nil
}

// ScanFiles ...
func ScanFiles(filePaths *[]string, verbose bool, m, ig *regexp.Regexp) <-chan *Source {
	cp := make(chan *Source)

	var wg sync.WaitGroup
	wg.Add(len(*filePaths))

	for _, filePath := range *filePaths {
		go func(filePath string) {
			defer wg.Done()
			if verbose {
				fmt.Printf("[%s] scanning for \"%s\" \n", filePath, m.String())
			}

			source := New(filePath, m, ig)
			source.Scan()
			cp <- source
		}(filePath)
	}

	go func() {
		wg.Wait()
		close(cp)
	}()
	return cp
}

// Chunks ...
func Chunks(foundFiles *[]string) [][]string {
	foundFilesCnt := uint64(len(*foundFiles))
	chunkSize, err := LimitNumberOfFiles()
	if err != nil {
		chunkSize = conf.DefaultChunksSizeToScan
	}

	// NOTE: "too many open files" io error 회피
	// 현재 열려있는 파일 수를 확인하는 것보다 더 간단하고,
	// 프로세스당 파일 제한 값의 반만 사용하더라도
	// 속도에는 크게 차이가 없다.
	chunkSize = chunkSize >> 1

	var chunks [][]string
	var i uint64
	for i = 0; i < foundFilesCnt; i += chunkSize {
		end := i + chunkSize

		if end > foundFilesCnt {
			end = foundFilesCnt
		}

		chunks = append(chunks, (*foundFiles)[i:end])
	}
	return chunks
}
