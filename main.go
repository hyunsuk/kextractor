package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/loganstone/kextractor/file"
)

var (
	comments  map[string]string
	skipPaths map[string]string
	verbose   = flag.Bool("v", false, "Make some output more verbose.")
)

func report(FilteredFilesCount int, scanErrorCount int, files *[]file.Data) {
	for _, f := range *files {
		fmt.Println(f.Path())
		for n, t := range *f.MatchedLine() {
			fmt.Printf("%d: %s\n", n, t)
		}
	}
	fmt.Printf("[%d] scanning files\n", FilteredFilesCount)
	fmt.Printf("[%d] scanning error files\n", scanErrorCount)
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

	resultPaths := []string{}
	chann := file.Search(dirPathToSearch, filterByFileExt, isSkipPath)
	for path := range chann {
		resultPaths = append(resultPaths, path)
	}

	filesContainingKorean := []file.Data{}
	ch := make(chan *file.Data)

	scanError := 0
	for _, filePath := range resultPaths {
		if *verbose {
			fmt.Printf("[%s] scanning Korean character in file\n", filePath)
		}

		go func(filePath string) {
			fileData := file.New(filePath, "\\p{Hangul}")
			fileData.Scan(isComment)
			ch <- fileData
		}(filePath)
	}

	for i := 0; i < len(resultPaths); i++ {
		fileData := <-ch
		if fileData.ScanError != nil {
			scanError++
			if *verbose {
				fmt.Printf("[%s] scanning error - %s\n", fileData.Path(), fileData.ScanError)
			}
		}
		if fileData.HasMatchedString() {
			filesContainingKorean = append(filesContainingKorean, (*fileData))
		}
	}

	report(len(resultPaths), scanError, &filesContainingKorean)
}
