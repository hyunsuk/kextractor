package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/loganstone/kpick/conf"
	"github.com/loganstone/kpick/dir"
	"github.com/loganstone/kpick/file"
	"github.com/loganstone/kpick/profile"
)

func report(errorOnly bool, foundFilesCnt int, scanErrorsCnt int, filesContainingKorean []file.Source) {
	if !errorOnly {
		for _, f := range filesContainingKorean {
			fmt.Println(f.Path())
			f.PrintFoundLines()
		}
	}
	fmt.Printf("[%d] scanning files\n", foundFilesCnt)
	fmt.Printf("[%d] error \n", scanErrorsCnt)
	fmt.Printf("[%d] success \n", foundFilesCnt-scanErrorsCnt)
	fmt.Printf("[%d] files containing korean\n", len(filesContainingKorean))
}

func shouldScan(foundFilesCnt int) bool {
	var response string
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("found files [%d]. do you want to scan it? (y/n): ", foundFilesCnt)
	response, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	response = strings.Trim(response, " \n")
	if response != "y" && response != "n" {
		return shouldScan(foundFilesCnt)
	}
	if response == "n" {
		return false
	}
	return true
}

func main() {
	opts := conf.Opts()

	profile.CPU(opts.Cpuprofile)

	err := dir.Check(opts.DirToFind)
	if err != nil {
		log.Fatal(err)
	}

	skipPathRegex, err := dir.MakeSkipPathRegex(opts.SkipPaths, ",", "|")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("find for files [opts..%s] in [%s] directory\n", opts.FileExtToScan, opts.DirToFind)
	foundFiles, err := dir.Find(opts.DirToFind, opts.FileExtToScan, skipPathRegex)
	if err != nil {
		log.Fatal(err)
	}

	foundFilesCnt := len(foundFiles)
	if foundFilesCnt == 0 {
		fmt.Printf("[opts..%s] file not found in [%s] directory\n", opts.FileExtToScan, opts.DirToFind)
		os.Exit(0)
	}

	if opts.Interactive {
		if !shouldScan(foundFilesCnt) {
			os.Exit(0)
		}
	}

	matchRegex, ignoreRegex, err := file.MakeRegexForScan(conf.RegexStrToKorean, opts.IgnorePattern)
	if err != nil {
		log.Fatal(err)
	}

	filesContainingKorean := []file.Source{}
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
				filesContainingKorean = append(filesContainingKorean, *source)
			}
		}
	}

	sort.Slice(filesContainingKorean, func(i, j int) bool {
		return filesContainingKorean[i].Path() < filesContainingKorean[j].Path()
	})

	report(opts.ErrorOnly, foundFilesCnt, scanErrorsCnt, filesContainingKorean)

	profile.Mem(opts.Memprofile)
}
