package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
	"runtime/pprof"
	"strings"

	"github.com/loganstone/kpick/conf"
	"github.com/loganstone/kpick/file"
)

var (
	comments map[string]string

	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
	memprofile = flag.String("memprofile", "", "write memory profile to `file`")

	dirToSearch   = flag.String("d", conf.DefaultDir, "Directory to search.")
	fileExtToScan = flag.String("f", conf.DefaultFileExt, "File extension to scan.")
	skipPaths     = flag.String("s", conf.DefaultSkipPaths, "Directories to skip from search.(delimiter ',')")
	verbose       = flag.Bool("v", false, "Make some output more verbose.")
	interactive   = flag.Bool("i", false, "Interactive scanning.")
	errorOnly     = flag.Bool("e", false, "Make output error only.")
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

func getSkipPathRegexp(skipPaths *string) (*regexp.Regexp, error) {
	if (*skipPaths) != conf.DefaultSkipPaths {
		(*skipPaths) += "," + conf.DefaultSkipPaths
	}
	paths := strings.Split(*skipPaths, ",")
	(*skipPaths) = strings.Join(paths, "|")
	skipPathRegexp, err := regexp.Compile(*skipPaths)
	if err != nil {
		return nil, err
	}
	return skipPathRegexp, nil
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
}

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	skipPathRegexp, err := getSkipPathRegexp(skipPaths)
	if err != nil {
		log.Fatal(err)
	}

	if (*dirToSearch) == "" || (*dirToSearch) == conf.DefaultDir {
		currentDir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		(*dirToSearch) = currentDir
	}

	dirInfo, err := os.Stat((*dirToSearch))
	if err != nil {
		log.Fatal(err)
	}

	if !dirInfo.IsDir() {
		log.Fatalf("'%s' must be directory", (*dirToSearch))
	}

	foundFiles, err := file.Search((*dirToSearch), (*fileExtToScan), skipPathRegexp)
	if err != nil {
		log.Fatal(err)
	}

	foundFilesCnt := uint64(len(*foundFiles))
	if foundFilesCnt == 0 {
		fmt.Printf("[*.%s] file not found in [%s] directory\n", (*fileExtToScan), (*dirToSearch))
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

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close()
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}
