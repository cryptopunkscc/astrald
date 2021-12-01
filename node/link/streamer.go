package link

type Streamer interface {
	Links() <-chan *Link
}
