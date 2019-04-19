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
	startLineNumber  = 1
	regexStrToKorean = "\\p{Hangul}"
	regexStrComments = "\\s*[#|//|/*|<!--]\\s*|.[*-->|\\*/]$"
)

var comments *regexp.Regexp

// Source ...
type Source struct {
	path           string
	lineScanner    *regexp.Regexp
	commentChecker *regexp.Regexp
	foundLines     map[int]string
	isScanned      bool
	scanError      error
}

// New ...
func New(path string, lineScanner *regexp.Regexp) *Source {
	return &Source{
		path,
		lineScanner,
		regexp.MustCompile(regexStrComments),
		map[int]string{},
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
	pre := []byte{}
	for {
		line, isPrefix, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			d.scanError = err
			break
		}

		pre = append(pre, line...)
		if isPrefix {
			continue
		}

		texts := string(pre)
		pre = []byte{}
		if d.commentChecker.MatchString(texts) {
			lineNumber++
			continue
		}

		if d.lineScanner.MatchString(texts) {
			d.foundLines[lineNumber] = texts
		}
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
func (d *Source) FoundLines() *map[int]string {
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

// LimitNumberOfFiles ...
func LimitNumberOfFiles() (uint64, error) {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return 0, err
	}
	return rLimit.Cur, nil
}

// ScanKorean ...
func ScanKorean(filePaths *[]string, verbose bool) <-chan *Source {
	cp := make(chan *Source)
	lineScanner := regexp.MustCompile(regexStrToKorean)

	var wg sync.WaitGroup
	wg.Add(len(*filePaths))

	for _, filePath := range *filePaths {
		go func(filePath string) {
			defer wg.Done()
			if verbose {
				fmt.Printf("[%s] scanning Korean character in file\n", filePath)
			}

			fileData := New(filePath, lineScanner)
			fileData.Scan()
			cp <- fileData
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
