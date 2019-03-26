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

var fileExt string
var filePaths []string
var commets map[string]string
var ignorePaths map[string]string

func procFile(path string, f os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if f.IsDir() || isIgnorePath(path) {
		return nil
	}

	if fileExt == "" {
		filePaths = append(filePaths, path)
	} else {
		v := strings.Split(f.Name(), ".")
		if v[len(v)-1] == fileExt {
			filePaths = append(filePaths, path)
		}
	}
	return nil
}

func fileReader(path string) {
	fmt.Printf("[%s] file scanning\n", path)
	if f, err := os.Open(path); err != nil {
		log.Fatal(err)

	} else {
		defer f.Close()

		scanner := bufio.NewScanner(f)
		n := 1

		for scanner.Scan() {
			line := scanner.Text()
			if isComment(line) {
				n++
				continue
			}

			matched, err := regexp.MatchString("\\p{Hangul}", line)
			if err != nil {
				log.Fatal(err)
			}
			if matched {
				fmt.Printf("%d :  %v\n", n, line)
			}
			n++
		}

		if err := scanner.Err(); err != nil {
			log.Printf("[%s] file scan error - %s\n", path, err)
		}

	}
}

func isComment(s string) bool {
	for _, v := range commets {
		if matched, err := regexp.MatchString(v, s); err != nil {
			log.Fatal(err)
		} else if matched {
			return matched
		}
	}
	return false
}

func isIgnorePath(s string) bool {
	for _, v := range ignorePaths {
		if matched, err := regexp.MatchString(v, s); err != nil {
			log.Fatal(err)
		} else if matched {
			return matched
		}
	}
	return false
}

func init() {
	commets = map[string]string{
		"python":     "\\s*#\\s*",
		"html":       "\\s*<!--\\s*|.*-->$",
		"javascript": "\\s*[//|/*]\\s*",
	}
	ignorePaths = map[string]string{
		"test_path": "functional_test",
		"git":       ".git",
	}
}

func main() {
	flag.Parse()
	dirPath := flag.Arg(1)
	fileExt = flag.Arg(0)

	if dirPath == "" {
		currentDir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		dirPath = currentDir
	}

	dirInfo, err := os.Stat(dirPath)
	if err != nil {
		log.Fatal(err)
	}

	if !dirInfo.IsDir() {
		log.Fatalf("'%s' must be directory", dirPath)
	}

	err = filepath.Walk(dirPath, procFile)
	if err != nil {
		log.Fatal(err)
	}

	for _, filePath := range filePaths {
		fileReader(filePath)
	}
}
