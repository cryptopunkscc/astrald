package request

type Handler func(
	rc Context,
) error

type Handlers map[byte]Handler
