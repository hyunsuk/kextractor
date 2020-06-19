package conf

import (
	"errors"
	"flag"
	"log"
	"os"
	"regexp"
	"strings"
)

const (
	// DefaultDir used when is no value of '-d' option
	DefaultDir = "."
	// DefaultFilenameExt used when is no value of '-f' option
	DefaultFilenameExt = "*"
	// MustIncludeSkipPaths add directories that not need to be find,
	// typing '-s .git' every time is annoying.
	MustIncludeSkipPaths = ".git,tmp"
	// KoreanPattern is a regular expression to look up Korean(Hangul)
	KoreanPattern = "\\p{Hangul}"
)

// ErrSkipPathsIsRequired .
var ErrSkipPathsIsRequired = errors.New("'SkipPaths' is required")

// Options .
type Options struct {
	Cpuprofile        string
	Memprofile        string
	DirPathToFind     string
	FileExtToScan     string
	SkipPaths         string
	IgnoreRegexString string
	Verbose           bool
	Interactive       bool
	ErrorOnly         bool
}

func (o *Options) setDefaultValue() {
	if o.DirPathToFind == "" || o.DirPathToFind == DefaultDir {
		currentDir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		o.DirPathToFind = currentDir
	}

	if o.FileExtToScan == "" {
		o.FileExtToScan = DefaultFilenameExt
	}

	if o.SkipPaths == "" {
		o.SkipPaths = MustIncludeSkipPaths
	}
}

// SkipPathsRegex .
func (o *Options) SkipPathsRegex() (*regexp.Regexp, error) {
	if o.SkipPaths == "" {
		return nil, ErrSkipPathsIsRequired
	}
	paths := strings.Split(o.SkipPaths, ",")
	return regexp.Compile(strings.Join(paths, "|"))
}

// Match .
func (o *Options) Match() (*regexp.Regexp, error) {
	return regexp.Compile(KoreanPattern)
}

// Ignore .
func (o *Options) Ignore() (*regexp.Regexp, error) {
	if o.IgnoreRegexString == "" {
		return nil, nil
	}
	return regexp.Compile(o.IgnoreRegexString)
}

var opts Options

func init() {
	flag.StringVar(&opts.Cpuprofile, "cpuprofile", "", "Write cpu profile to `file`.")
	flag.StringVar(&opts.Memprofile, "memprofile", "", "Write memory profile to `file`.")

	flag.StringVar(&opts.DirPathToFind, "d", DefaultDir, "Directory to find files.")
	flag.StringVar(&opts.FileExtToScan, "f", DefaultFilenameExt, "Filename extension to scan.")
	flag.StringVar(&opts.SkipPaths, "s", MustIncludeSkipPaths, "Directories to skip walk.(delimiter ',')")
	flag.StringVar(&opts.IgnoreRegexString, "ignore", "", "Regex for line to ignore when scanning file.")
	flag.BoolVar(&opts.Verbose, "v", false, "Make some output more verbose.")
	flag.BoolVar(&opts.Interactive, "i", false, "Interactive scanning.")
	flag.BoolVar(&opts.ErrorOnly, "e", false, "Make output error only.")
}

// Opts .
func Opts() *Options {
	flag.Parse()
	opts.setDefaultValue()
	return &opts
}
