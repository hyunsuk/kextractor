package file

import (
	"container/heap"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

var beforeFn BeforeScanFunc = func(path string) {}
var afterFn AfterScanFunc = func(path string) {}

func TestScanFilesWithMatch(t *testing.T) {
	paths := []string{"test1.md", "test2.md"}

	match, err := regexp.Compile("\\p{Hangul}")
	assert.NoError(t, err)

	for f := range ScanFiles(paths, match, nil, beforeFn, afterFn) {
		assert.NoError(t, f.Error())
		assert.Equal(t, 1, len(f.MatchedLines()))
	}
}

func TestScanFilesWithMatchAndIgnore(t *testing.T) {
	paths := []string{"test1.md", "test2.md"}

	match, err := regexp.Compile("\\p{Hangul}")
	assert.NoError(t, err)

	ignore, err := regexp.Compile("#")
	assert.NoError(t, err)

	files := &Heap{}
	heap.Init(files)
	for f := range ScanFiles(paths, match, ignore, beforeFn, afterFn) {
		assert.NoError(t, f.Error())
		heap.Push(files, f)
	}

	test1File, ok := heap.Pop(files).(*File)
	assert.True(t, ok)
	assert.True(t, len(test1File.MatchedLines()) == 1)

	test2File, ok := heap.Pop(files).(*File)
	assert.True(t, ok)
	assert.True(t, len(test2File.MatchedLines()) == 0)
}

func TestHeap(t *testing.T) {
	paths := []string{"c-test.md", "b-test.md", "a-test.md"}

	match, err := regexp.Compile("\\p{Hangul}")
	assert.NoError(t, err)

	ignore, err := regexp.Compile("#")
	assert.NoError(t, err)

	files := &Heap{}
	heap.Init(files)

	for _, path := range paths {
		f := &File{path, match, ignore, map[int][]byte{}, nil}
		heap.Push(files, f)
	}

	file, ok := heap.Pop(files).(*File)
	assert.True(t, ok)
	assert.True(t, file.Path() == paths[2])
	file, ok = heap.Pop(files).(*File)
	assert.True(t, ok)
	assert.True(t, file.Path() == paths[1])
	file, ok = heap.Pop(files).(*File)
	assert.True(t, ok)
	assert.True(t, file.Path() == paths[0])
}

func TestScan(t *testing.T) {
	paths := []string{"test1.md", "test2.md"}

	match, err := regexp.Compile("\\p{Hangul}")
	assert.NoError(t, err)

	ignore, err := regexp.Compile("#")
	assert.NoError(t, err)

	files := &Heap{}
	heap.Init(files)

	f := &File{paths[0], match, ignore, map[int][]byte{}, nil}
	f.Scan()
	assert.NoError(t, f.scanError)
	assert.Equal(t, 1, len(f.MatchedLines()))

	f = &File{paths[1], match, ignore, map[int][]byte{}, nil}
	f.Scan()
	assert.NoError(t, f.scanError)
	assert.Equal(t, 0, len(f.MatchedLines()))
}
