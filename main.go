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

func showFoundFiles(filesContainingKorean *file.SortedFiles) {
	for filesContainingKorean.Len() > 0 {
		f, ok := heap.Pop(filesContainingKorean).(*file.Source)
		if ok {
			fmt.Println(f.Path())
			f.PrintFoundLines()
		}
	}
}

func showNumbers(foundFilesCnt int, scanErrorsCnt int, filesCntContainingKorean int) {
	fmt.Printf("[%d] scanning files\n", foundFilesCnt)
	fmt.Printf("[%d] error \n", scanErrorsCnt)
	fmt.Printf("[%d] success \n", foundFilesCnt-scanErrorsCnt)
	fmt.Printf("[%d] files containing korean\n", filesCntContainingKorean)
}

func main() {
	opts := conf.Opts()

	profile.CPU(opts.Cpuprofile)

	err := dir.Check(opts.DirToFind)
	if err != nil {
		log.Fatal(err)
	}

	skipPathRegex, err := dir.SkipPathRegex(opts.SkipPaths, ",", "|")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("find for files [*.%s] in [%s] directory\n", opts.FileExtToScan, opts.DirToFind)
	foundFiles, err := dir.Find(opts.DirToFind, opts.FileExtToScan, skipPathRegex)
	if err != nil {
		log.Fatal(err)
	}

	foundFilesCnt := len(foundFiles)
	if foundFilesCnt == 0 {
		fmt.Printf("[*.%s] file not found in [%s] directory\n", opts.FileExtToScan, opts.DirToFind)
		os.Exit(0)
	}

	if opts.Interactive {
		q := fmt.Sprintf("found files [%d]. do you want to scan it? (y/n): ", foundFilesCnt)
		ok, err := ask.Confirm(q, "y", "n")
		if err != nil {
			log.Fatal(err)
		}
		if !ok {
			os.Exit(0)
		}
	}

	matchRegex, ignoreRegex, err := file.RegexForScan(conf.RegexStrToKorean, opts.IgnorePattern)
	if err != nil {
		log.Fatal(err)
	}

	filesContainingKorean := &file.SortedFiles{}
	heap.Init(filesContainingKorean)
	var scanErrorsCnt int
	beforeFn := func(filePath string) {
		if opts.Verbose {
			fmt.Printf("[%s] scanning for \"%s\"\n", filePath, matchRegex.String())
		}
	}
	afterFn := func(filePath string) {
		if opts.Verbose {
			fmt.Printf("[%s] scanning done\n", filePath)
		}
	}
	for _, paths := range file.Chunks(foundFiles) {
		for source := range file.ScanFiles(paths, matchRegex, ignoreRegex, beforeFn, afterFn) {
			if err := source.Error(); err != nil {
				scanErrorsCnt++
				if opts.Verbose || opts.ErrorOnly {
					fmt.Printf("[%s] scanning error - %s\n", source.Path(), err)
				}
				continue
			}

			if len(source.FoundLines()) > 0 {
				heap.Push(filesContainingKorean, source)
			}
		}
	}

	if !opts.ErrorOnly {
		showFoundFiles(filesContainingKorean)
	}

	showNumbers(foundFilesCnt, scanErrorsCnt, filesContainingKorean.Len())

	profile.Mem(opts.Memprofile)
}
