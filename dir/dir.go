package dir

import (
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
)

const findAllFileExt = "*"

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

func check(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		// TODO(logan): Should test.
		if os.IsPermission(err) {
			mode := info.Mode()
			log.Printf("'%s' permission is '%s'\n", path, mode.Perm())
		}
		return err
	}

	mode := info.Mode()

	if mode.IsRegular() {
		return fmt.Errorf("'%s' is regular file", path)
	}

	if !mode.IsDir() {
		return fmt.Errorf("'%s' is not directory", path)
	}
	return nil
}

// Finder saves the necessary information and
// the found file when searching for a file containing Korean.
type Finder struct {
	path      string
	fileExt   string
	skipPaths *regexp.Regexp
	result    []string
}

// NewFinder returns a new Finder object.
// Check the "path" is correct before returning.
func NewFinder(path, fileExt string, skipPaths *regexp.Regexp) (*Finder, error) {
	if err := check(path); err != nil {
		return nil, err
	}
	return &Finder{
		path:      path,
		fileExt:   fileExt,
		skipPaths: skipPaths,
	}, nil
}

// Find finds and saves the same file
// with the specified file extension in the specified path and all sub paths.
func (f *Finder) Find() error {
	return filepath.Walk(f.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || (*f.skipPaths).MatchString(path) {
			return nil
		}

		if f.fileExt == "" || f.fileExt == findAllFileExt {
			f.result = append(f.result, path)
			return nil
		}

		v := strings.Split(info.Name(), ".")
		if v[len(v)-1] == f.fileExt {
			f.result = append(f.result, path)
		}
		return nil
	})
}

// ResultCount .
func (f *Finder) ResultCount() int {
	return len(f.result)
}

// Chunk .
func (f *Finder) Chunk() [][]string {
	filePathsCnt := f.ResultCount()
	chunkSize := limitNumber()

	// NOTE: "too many open files" io error 회피
	// 현재 열려있는 파일 수를 확인하는 것보다 더 간단하고,
	// 프로세스당 파일 제한 값의 반만 사용하더라도
	// 속도에는 크게 차이가 없다.
	chunkSize = chunkSize >> 1

	var chunk [][]string
	var i int
	for i = 0; i < filePathsCnt; i += chunkSize {
		end := i + chunkSize

		if end > filePathsCnt {
			end = filePathsCnt
		}

		chunk = append(chunk, f.result[i:end])
	}

	return chunk
}
