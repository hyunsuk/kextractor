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

var fileExtenxion string
var filesList []string
var filePathWatchFlag string

func makeFilesList(path string, f os.FileInfo, err error) error {
	if !f.IsDir() && !isTestPath(path) {
		if fileExtenxion == "" {
			filesList = append(filesList, path)
		} else {
			paths := strings.Split(f.Name(), ".")
			if paths[len(paths)-1] == fileExtenxion {
				filesList = append(filesList, path)
			}
		}
	}
	return nil
}

func fileReader(path string) {
	if file, err := os.Open(path); err == nil {
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNumber := 1

		for scanner.Scan() {
			lineText := scanner.Text()
			matched, err := regexp.MatchString("\\p{Hangul}", lineText)

			if err != nil {
				log.Fatal(err)
			}

			if matched {
				if !isPythonComment(lineText) && !isHTMLComment(lineText) && !isJavascriptComment(lineText) {
					if filePathWatchFlag != path {
						if filePathWatchFlag != "" {
							fmt.Println()
						}
						filePathWatchFlag = path
						fmt.Printf("path : %v\n", filePathWatchFlag)
					}

					fmt.Printf("%d :  %v\n", lineNumber, lineText)
				}
			}
			lineNumber++
		}

		if err = scanner.Err(); err != nil {
			log.Println(path)
			log.Fatal(err)
		}

	} else {
		log.Fatal(err)
	}
}

func isPythonComment(s string) bool {
	matched, err := regexp.MatchString("\\s*#\\s*", s)

	if err != nil {
		log.Fatal(err)
	}
	return matched
}

func isHTMLComment(s string) bool {
	matched, err := regexp.MatchString("\\s*<!--\\s*|.*-->$", s)

	if err != nil {
		log.Fatal(err)
	}
	return matched
}

func isJavascriptComment(s string) bool {
	matched, err := regexp.MatchString("\\s*[//|/*]\\s*", s)

	if err != nil {
		log.Fatal(err)
	}
	return matched
}

func isTestPath(path string) bool {
	matched, err := regexp.MatchString("functional_test", path)

	if err != nil {
		log.Fatal(err)
	}
	return matched
}

func main() {
	flag.Parse()
	root := flag.Arg(1)
	fileExtenxion = flag.Arg(0)

	if root == "" {
		currentDir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		root = currentDir
	}

	filepath.Walk(root, makeFilesList)

	for _, filePath := range filesList {
		fileReader(filePath)
	}
}
