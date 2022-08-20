package content

const (
	Port     = "android/content"
	PortInfo = "android/content/info"
)

type Info struct {
	Uri  string
	Size int64
	Mime string
	Name string
}
