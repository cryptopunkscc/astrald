package warpdrive

import (
	"encoding/json"
)

func (c RemoteClient) Download(offerId OfferId) (err error) {
	// Request download
	if err = c.Encode("c", remoteDownload); err != nil {
		err = Error(err, "Cannot request download")
	}

	// Send file request id
	if err = c.Encode("[c]c", offerId); err != nil {
		err = Error(err, "Cannot send request id")
		return
	}

	// Read confirmation
	var code byte
	err = c.Decode("c", &code)
	if err != nil {
		err = Error(err, "Cannot read confirmation")
		return
	}

	return
}

func (c RemoteClient) send(files []Info) (offerId OfferId, code uint8, err error) {
	// Request send
	err = c.Encode("c", remoteSend)
	if err != nil {
		err = Error(err, "Cannot request send")
		return
	}

	// Send file request
	offerId = newOfferId()
	err = c.Encode("[c]c", offerId)
	if err != nil {
		err = Error(err, "Cannot send offer id", offerId)
		return
	}
	shrunken := shrinkPaths(files)
	err = json.NewEncoder(c.Conn).Encode(shrunken)
	if err != nil {
		err = Error(err, "Cannot send offer info", offerId)
		return
	}
	// Read result code
	err = c.Decode("c", &code)
	if err != nil {
		err = Error(err, "Cannot read result code")
		return
	}
	return
}
