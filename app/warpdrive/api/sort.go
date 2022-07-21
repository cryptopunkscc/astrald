package api

type OffersByCreate []Offer

func (a OffersByCreate) Len() int           { return len(a) }
func (a OffersByCreate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a OffersByCreate) Less(i, j int) bool { return a[i].Create < a[j].Create }

type OfferUpdatesByUpdate []OfferUpdate

func (a OfferUpdatesByUpdate) Len() int           { return len(a) }
func (a OfferUpdatesByUpdate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a OfferUpdatesByUpdate) Less(i, j int) bool { return a[i].Stat().Update < a[j].Stat().Update }
