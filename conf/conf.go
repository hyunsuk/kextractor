package conf

const (
	// DefaultDir used when is no value of '-d' option
	DefaultDir = "."
	// DefaultFileExt used when is no value od '-f' option
	DefaultFileExt = "*"
	// MustIncludeSkipPaths add directories that not need to be find,
	// typing '-s .git' every time is annoying.
	MustIncludeSkipPaths = ".git,tmp"
	// RegexStrToKorean is a regular expression to look up Korean(Hangul)
	RegexStrToKorean = "\\p{Hangul}"
)
