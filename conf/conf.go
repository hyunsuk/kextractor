package conf

const (
	// DefaultDir used when is no value of '-d' option
	DefaultDir = "."
	// DefaultFilenameExt used when is no value of '-f' option
	DefaultFilenameExt = "*"
	// MustIncludeSkipPaths add directories that not need to be find,
	// typing '-s .git' every time is annoying.
	MustIncludeSkipPaths = ".git,tmp"
	// RegexStrToKorean is a regular expression to look up Korean(Hangul)
	RegexStrToKorean = "\\p{Hangul}"
)
