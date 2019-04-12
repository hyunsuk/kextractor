package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/loganstone/kextractor/file"
)

var (
	comments    map[string]string
	skipPaths   map[string]string
	verbose     = flag.Bool("v", false, "Make some output more verbose.")
	interactive = flag.Bool("i", false, "Interactive scanning.")
	errorOnly   = flag.Bool("e", false, "Make output error only.")
)

func report(filteredFilesCount uint64, scanErrorCount uint64, files *[]file.Data) {
	if !(*errorOnly) {
		for _, f := range *files {
			fmt.Println(f.Path())
			for n, t := range *f.MatchedLine() {
				fmt.Printf("%d: %s\n", n, t)
			}
		}
	}
	fmt.Printf("[%d] scanning files\n", filteredFilesCount)
	fmt.Printf("[%d] error \n", scanErrorCount)
	fmt.Printf("[%d] success \n", filteredFilesCount-scanErrorCount)
	fmt.Printf("[%d] files containing korean\n", len(*files))
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

func shouldScan(foundFilesCount uint64) bool {
	var response string
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("found files [%d]. do you want to scan it? (y/n): ", foundFilesCount)
	response, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	response = strings.Trim(response, " \n")
	if response != "y" && response != "n" {
		return shouldScan(foundFilesCount)
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

	resultPaths, err := file.Search(dirPathToSearch, filterByFileExt, isSkipPath)
	if err != nil {
		log.Fatal(err)
	}

	resultPathsLen := uint64(len(*resultPaths))
	if resultPathsLen == 0 {
		fmt.Printf("[*.%s] file not found in [%s] directory\n", filterByFileExt, dirPathToSearch)
		os.Exit(0)
	}
	if *interactive {
		if !shouldScan(resultPathsLen) {
			os.Exit(0)
		}
	}

	chunkSize, err := file.Limit()
	if err != nil {
		chunkSize = 1024
	}

	var chunks [][]string
	var i uint64
	for i = 0; i < resultPathsLen; i += chunkSize {
		end := i + chunkSize

		if end > resultPathsLen {
			end = resultPathsLen
		}

		chunks = append(chunks, (*resultPaths)[i:end])
	}

	filesContainingKorean := []file.Data{}
	var scanErrorCount uint64
	scanErrorCount = 0
	for _, paths := range chunks {
		for fileData := range file.ScanKorean(&paths, *verbose, isComment) {
			if fileData.ScanError != nil {
				scanErrorCount++
				if *verbose || *errorOnly {
					fmt.Printf("[%s] scanning error - %s\n", fileData.Path(), fileData.ScanError)
				}
			}
			if fileData.HasMatchedString() {
				filesContainingKorean = append(filesContainingKorean, (*fileData))
			}
		}
	}

	report(resultPathsLen, scanErrorCount, &filesContainingKorean)
}
