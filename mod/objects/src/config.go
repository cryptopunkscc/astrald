package objects

const (
	methodPut      = "objects.put"
	methodRead     = "objects.read"
	methodDescribe = "objects.describe"
	methodSearch   = "objects.search"
	methodPush     = "objects.push"
)

const (
	// MaxAlloc is the maximum allocatable storage space for an object
	MaxAlloc int64 = 1 << 40 //1TB; gomobile requires explicit int64 type.

	// MaxObjectSize is the maximum size of an object that can be loaded into memory
	MaxObjectSize int64 = 32 << 20 // 32 MB
)

type Config struct{}

var defaultConfig = Config{}
