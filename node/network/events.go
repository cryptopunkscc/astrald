package network

type EventLinkAdded struct {
	Link *ActiveLink
}

type EventLinkRemoved struct {
	Link *ActiveLink
}
