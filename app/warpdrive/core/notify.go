package core

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
)

func (c *core) notifyListeners(data interface{}, listeners *api.Subscriptions) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		c.Println("Cannot create json from data", data, err)
		return
	}
	listeners.Lock()
	defer listeners.Unlock()
	for listener := range listeners.Set {
		_, err := listener.Write(jsonData)
		if err != nil {
			c.Println("Error while sending files to listener", err)
			return
		}
	}
}
