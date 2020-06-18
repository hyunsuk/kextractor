package main

import (
	"bufio"
	"container/heap"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/loganstone/kpick/conf"
	"github.com/loganstone/kpick/dir"
	"github.com/loganstone/kpick/file"
	"github.com/loganstone/kpick/profile"
)

func confirm(question, ok, cancel string) (bool, error) {
	var input string
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf(question)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	input = strings.Trim(input, " \n")
	if input != ok && input != cancel {
		return confirm(question, ok, cancel)
	}
	if input == cancel {
		return false, nil
	}
	return true, nil
}

func summary(totalCnt, errorsCnt, containedFilesCnt int) {
	fmt.Printf("[%d] scanning files\n", totalCnt)
	fmt.Printf("[%d] error \n", errorsCnt)
	fmt.Printf("[%d] success \n", totalCnt-errorsCnt)
	fmt.Printf("[%d] files containing Korean\n", containedFilesCnt)
}

func main() {
	opts := conf.Opts()

	profile.CPU(opts.Cpuprofile)

	skipPaths, err := opts.SkipPathsRegex()
	if err != nil {
		log.Fatal(err)
	}

	finder, err := dir.NewFinder(opts.DirPathToFind, opts.FileExtToScan, skipPaths)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("find [*.%s] files in [%s] directory\n", opts.FileExtToScan, opts.DirPathToFind)
	if finder.Find() != nil {
		log.Fatal(err)
	}

	totalCnt := len(finder.Result)
	if totalCnt == 0 {
		fmt.Printf("[*.%s] file not found in [%s] directory\n", opts.FileExtToScan, opts.DirPathToFind)
		os.Exit(0)
	}

	if opts.Interactive {
		q := fmt.Sprintf("found [%d] files. scan it? (y/n): ", totalCnt)
		ok, err := confirm(q, "y", "n")
		if err != nil {
			log.Fatal(err)
		}
		if !ok {
			os.Exit(0)
		}
	}

	match, err := opts.Match()
	if err != nil {
		log.Fatal(err)
	}

	ignore, err := opts.Ignore()
	if err != nil {
		log.Fatal(err)
	}

	containKorean := &file.Heap{}
	heap.Init(containKorean)
	beforeFn := func(filePath string) {
		if opts.Verbose {
			fmt.Printf("[%s] scanning \"%s\"\n", filePath, match.String())
		}
	}
	afterFn := func(filePath string) {
		if opts.Verbose {
			fmt.Printf("[%s] scanning done\n", filePath)
		}
	}
	var scanErrorsCnt int
	for _, paths := range file.Chunk(finder.Result) {
		for f := range file.ScanFiles(paths, match, ignore, beforeFn, afterFn) {
			if err := f.Error(); err != nil {
				scanErrorsCnt++
				if opts.Verbose || opts.ErrorOnly {
					fmt.Printf("[%s] scanning error - %s\n", f.Path(), err)
				}
				continue
			}

			if len(f.MatchedLines()) > 0 {
				heap.Push(containKorean, f)
			}
		}
	}

	if !opts.ErrorOnly {
		file.PrintFiles(containKorean)
	}

	summary(totalCnt, scanErrorsCnt, containKorean.Len())

	profile.Mem(opts.Memprofile)
}
