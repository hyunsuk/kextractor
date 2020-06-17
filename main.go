package main

import (
	"container/heap"
	"fmt"
	"log"
	"os"

	"github.com/loganstone/kpick/ask"
	"github.com/loganstone/kpick/conf"
	"github.com/loganstone/kpick/dir"
	"github.com/loganstone/kpick/file"
	"github.com/loganstone/kpick/profile"
)

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

	finder, err := dir.NewFinder(opts.DirToFind, opts.FileExtToScan, skipPaths)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("find [*.%s] files in [%s] directory\n", opts.FileExtToScan, opts.DirToFind)
	if finder.Find() != nil {
		log.Fatal(err)
	}

	totalCnt := len(finder.Result)
	if totalCnt == 0 {
		fmt.Printf("[*.%s] file not found in [%s] directory\n", opts.FileExtToScan, opts.DirToFind)
		os.Exit(0)
	}

	if opts.Interactive {
		q := fmt.Sprintf("found [%d] files. do you want to scan it? (y/n): ", totalCnt)
		ok, err := ask.Confirm(q, "y", "n")
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

	filesContainingKorean := &file.Heap{}
	heap.Init(filesContainingKorean)
	var scanErrorsCnt int
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
				heap.Push(filesContainingKorean, f)
			}
		}
	}

	if !opts.ErrorOnly {
		file.PrintFiles(filesContainingKorean)
	}

	summary(totalCnt, scanErrorsCnt, filesContainingKorean.Len())

	profile.Mem(opts.Memprofile)
}
