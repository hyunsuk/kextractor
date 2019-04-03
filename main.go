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

var comments map[string]string
var skipPaths map[string]string

func scanKorean(path string) {
	fmt.Printf("[%s] scanning Korean character in file\n", path)
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	scannedLine := 1
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if isComment(line) {
			scannedLine++
			continue
		}

		matched, err := regexp.MatchString("\\p{Hangul}", line)
		if err != nil {
			log.Fatal(err)
		}
		if matched {
			fmt.Printf("%d :  %v\n", scannedLine, line)
		}
		scannedLine++
	}

	if err := scanner.Err(); err != nil {
		log.Printf("[%s] scanning error - %s\n", path, err)
	}
}

func search(dir string, filterByFileExt string) ([]string, error) {
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
	return resultPaths, err
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
		"test_path": "functional_test",
		"git":       ".git",
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

	for _, filePath := range resultPaths {
		scanKorean(filePath)
	}
}
