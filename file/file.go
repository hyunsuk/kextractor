package file

import (
	"container/heap"
	"fmt"
	"math"
	"regexp"
	"sync"
	"syscall"
)

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
