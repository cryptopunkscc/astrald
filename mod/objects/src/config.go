package objects

const (
	methodPut      = "objects.put"
	methodRead     = "objects.read"
	methodDescribe = "objects.describe"
	methodSearch   = "objects.search"
	methodPush     = "objects.push"
)

const DefaultRepoName = "default"

const (
	// MaxAlloc is the maximum allocatable storage space for an object
	MaxAlloc int64 = 1 << 40 //1TB; gomobile requires explicit int64 type.
)

type Config struct{}

var defaultConfig = Config{}
