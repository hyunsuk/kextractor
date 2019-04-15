package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/loganstone/kpick/file"
)

var (
	comments    map[string]string
	skipPaths   map[string]string
	verbose     = flag.Bool("v", false, "Make some output more verbose.")
	interactive = flag.Bool("i", false, "Interactive scanning.")
	errorOnly   = flag.Bool("e", false, "Make output error only.")
)

func report(filesCnt uint64, errorCnt uint64, containingKorean *[]file.Data) {
	if !(*errorOnly) {
		for _, f := range *containingKorean {
			fmt.Println(f.Path())
			f.PrintMatchedLine()
		}
	}
	fmt.Printf("[%d] scanning files\n", filesCnt)
	fmt.Printf("[%d] error \n", errorCnt)
	fmt.Printf("[%d] success \n", filesCnt-errorCnt)
	fmt.Printf("[%d] files containing korean\n", len(*containingKorean))
}

func isComment(s string) bool {
	for _, v := range comments {
		if matched, err := regexp.MatchString(v, s); err != nil {
			log.Fatal(err)
		} else if matched {
			return matched
		}
	}
	return false
}

func isSkipPath(s string) bool {
	for _, v := range skipPaths {
		if matched, err := regexp.MatchString(v, s); err != nil {
			log.Fatal(err)
		} else if matched {
			return matched
		}
	}
	return false
}

func shouldScan(foundFilesCnt uint64) bool {
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

func init() {
	comments = map[string]string{
		"python":     "\\s*#\\s*",
		"html":       "\\s*<!--\\s*|.*-->$",
		"javascript": "\\s*[//|/*]\\s*",
	}
	skipPaths = map[string]string{
		"test": "test",
		"git":  ".git",
	}
}

func main() {
	flag.Parse()
	filterByFileExt := flag.Arg(0)
	dirPathToSearch := flag.Arg(1)

	if dirPathToSearch == "" {
		currentDir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		dirPathToSearch = currentDir
	}

	dirInfo, err := os.Stat(dirPathToSearch)
	if err != nil {
		log.Fatal(err)
	}

	if !dirInfo.IsDir() {
		log.Fatalf("'%s' must be directory", dirPathToSearch)
	}

	foundFiles, err := file.Search(dirPathToSearch, filterByFileExt, isSkipPath)
	if err != nil {
		log.Fatal(err)
	}

	foundFilesCnt := uint64(len(*foundFiles))
	if foundFilesCnt == 0 {
		fmt.Printf("[*.%s] file not found in [%s] directory\n", filterByFileExt, dirPathToSearch)
		os.Exit(0)
	}
	if *interactive {
		if !shouldScan(foundFilesCnt) {
			os.Exit(0)
		}
	}

	containingKorean := []file.Data{}
	var scanErrorCnt uint64
	scanErrorCnt = 0
	for _, paths := range file.Chunks(foundFiles) {
		for fileData := range file.ScanKorean(&paths, *verbose, isComment) {
			if fileData.ScanError != nil {
				scanErrorCnt++
				if *verbose || *errorOnly {
					fmt.Printf("[%s] scanning error - %s\n", fileData.Path(), fileData.ScanError)
				}
			}
			if fileData.HasMatchedString() {
				containingKorean = append(containingKorean, (*fileData))
			}
		}
	}

	report(foundFilesCnt, scanErrorCnt, &containingKorean)
}
