package styles

type Renderer interface {
	Render(...string) string
}
