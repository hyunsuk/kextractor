package conf

import (
	"flag"
	"log"
	"os"
)

const (
	// DefaultDir used when is no value of '-d' option
	DefaultDir = "."
	// DefaultFilenameExt used when is no value of '-f' option
	DefaultFilenameExt = "*"
	// MustIncludeSkipPaths add directories that not need to be find,
	// typing '-s .git' every time is annoying.
	MustIncludeSkipPaths = ".git,tmp"
	// KoreanPatternForRegex is a regular expression to look up Korean(Hangul)
	KoreanPatternForRegex = "\\p{Hangul}"
)

// Options .
type Options struct {
	Cpuprofile    string
	Memprofile    string
	DirToFind     string
	FileExtToScan string
	SkipPaths     string
	IgnorePattern string
	Verbose       bool
	Interactive   bool
	ErrorOnly     bool
}

var opts Options

func init() {
	flag.StringVar(&opts.Cpuprofile, "cpuprofile", "", "Write cpu profile to `file`.")
	flag.StringVar(&opts.Memprofile, "memprofile", "", "Write memory profile to `file`.")

	flag.StringVar(&opts.DirToFind, "d", DefaultDir, "Directory to find files.")
	flag.StringVar(&opts.FileExtToScan, "f", DefaultFilenameExt, "Filename extension to scan.")
	flag.StringVar(&opts.SkipPaths, "s", MustIncludeSkipPaths, "Directories to skip walk.(delimiter ',')")
	flag.StringVar(&opts.IgnorePattern, "igg", "", "Pattern for line to ignore when scanning file.")
	flag.BoolVar(&opts.Verbose, "v", false, "Make some output more verbose.")
	flag.BoolVar(&opts.Interactive, "i", false, "Interactive scanning.")
	flag.BoolVar(&opts.ErrorOnly, "e", false, "Make output error only.")

	flag.Parse()

	if opts.DirToFind == "" || opts.DirToFind == DefaultDir {
		currentDir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		opts.DirToFind = currentDir
	}

	if opts.FileExtToScan == "" {
		opts.FileExtToScan = DefaultFilenameExt
	}

	if opts.SkipPaths == "" {
		opts.SkipPaths = MustIncludeSkipPaths
	}
}

// Opts .
func Opts() Options {
	return opts
}
