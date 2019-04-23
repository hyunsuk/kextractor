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
func New(path string, matchRegex *regexp.Regexp, ignoreRegex *regexp.Regexp) *Source {
	return &Source{
		path,
		matchRegex,
		ignoreRegex,
		map[int][]byte{},
		false,
		nil,
	}
}

// Scan ...
func (d *Source) Scan() {
	f, err := os.Open(d.path)
	if err != nil {
		d.scanError = err
		return
	}

	defer f.Close()

	reader := bufio.NewReader(f)
	lineNumber := startLineNumber
	newLine := []byte{}
	for {
		lineChunk, isPrefix, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			d.scanError = err
			break
		}

		newLine = append(newLine, lineChunk...)
		if isPrefix {
			continue
		}

		if d.ignoreRegex != nil && d.ignoreRegex.Match(newLine) {
			newLine = []byte{}
			lineNumber++
			continue
		}

		if d.matchRegex != nil && d.matchRegex.Match(newLine) {
			d.foundLines[lineNumber] = newLine
		}

		newLine = []byte{}
		lineNumber++
	}
	d.isScanned = true
}

// Path ...
func (d *Source) Path() string {
	return d.path
}

// Error ...
func (d *Source) Error() error {
	return d.scanError
}

// FoundLines ...
func (d *Source) FoundLines() *map[int][]byte {
	return &d.foundLines
}

// PrintFoundLines ...
func (d *Source) PrintFoundLines() {
	keys := make([]int, len(d.foundLines))
	i := 0
	for k := range d.foundLines {
		keys[i] = k
		i++
	}

	sort.Ints(keys)
	for _, k := range keys {
		v, _ := d.foundLines[k]
		fmt.Printf("%d: %s\n", k, v)
	}
}

// MakeRegexForScan ...
func MakeRegexForScan(match string, ignore string) (*regexp.Regexp, *regexp.Regexp, error) {
	var matchRegex *regexp.Regexp
	var ignoreRegex *regexp.Regexp
	if match != "" {
		regex, err := regexp.Compile(match)
		if err != nil {
			return matchRegex, ignoreRegex, err
		}
		matchRegex = regex
	}

	if ignore != "" {
		regex, err := regexp.Compile(ignore)
		if err != nil {
			return matchRegex, ignoreRegex, err
		}
		ignoreRegex = regex
	}

	return matchRegex, ignoreRegex, nil
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
func ScanFiles(filePaths *[]string, verbose bool, matchRegex *regexp.Regexp, ignoreRegex *regexp.Regexp) <-chan *Source {
	cp := make(chan *Source)

	var wg sync.WaitGroup
	wg.Add(len(*filePaths))

	for _, filePath := range *filePaths {
		go func(filePath string) {
			defer wg.Done()
			if verbose {
				fmt.Printf("[%s] scanning for \"%s\" \n", filePath, matchRegex.String())
			}

			source := New(filePath, matchRegex, ignoreRegex)
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
