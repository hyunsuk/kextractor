package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	comments  map[string]string
	skipPaths map[string]string
	verbose   = flag.Bool("v", false, "Make some output more verbose.")
)

type scannedFileData struct {
	path                  string
	linesContainingKorean map[int]string
}

func scanKorean(path string) (scannedFileData, error) {
	fileData := scannedFileData{path, map[int]string{}}

	f, err := os.Open(path)
	if err != nil {
		return fileData, err
	}

	defer f.Close()

	lineNumber := 1
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lineText := scanner.Text()
		if isComment(lineText) {
			lineNumber++
			continue
		}

		matched, err := regexp.MatchString("\\p{Hangul}", lineText)
		if err != nil {
			return fileData, err
		}
		if matched {
			fileData.linesContainingKorean[lineNumber] = lineText
		}
		lineNumber++
	}

	if err := scanner.Err(); err != nil {
		return fileData, err
	}

	return fileData, nil
}

func search(dir string, filterByFileExt string) (*[]string, error) {
	var resultPaths []string
	err := filepath.Walk(dir,
		func(path string, f os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if f.IsDir() || isSkipPath(path) {
				return nil
			}

			if filterByFileExt != "" {
				v := strings.Split(f.Name(), ".")
				if v[len(v)-1] == filterByFileExt {
					resultPaths = append(resultPaths, path)
				}
				return nil
			}

			resultPaths = append(resultPaths, path)
			return nil
		})
	return &resultPaths, err
}

func report(FilteredFilesCount int, scanErrorCount int, files *[]scannedFileData) {
	for _, f := range *files {
		fmt.Println(f.path)
		for n, t := range f.linesContainingKorean {
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

	resultPaths, err := search(dirPathToSearch, filterByFileExt)
	if err != nil {
		log.Fatal(err)
	}

	filesContainingKorean := []scannedFileData{}
	scanError := 0
	for _, filePath := range *resultPaths {
		if *verbose {
			fmt.Printf("[%s] scanning Korean character in file\n", filePath)
		}
		data, err := scanKorean(filePath)
		if len(data.linesContainingKorean) > 0 {
			filesContainingKorean = append(filesContainingKorean, data)
		}

		if err != nil {
			scanError++
			if *verbose {
				fmt.Printf("[%s] scanning error - %s\n", filePath, err)
			}
		}
	}

	report(len(*resultPaths), scanError, &filesContainingKorean)
}
