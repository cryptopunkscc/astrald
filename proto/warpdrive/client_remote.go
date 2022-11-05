package warpdrive

import "encoding/json"

func (c Client) SendOffer(files []Info) (offerId OfferId, code uint8, err error) {
	// Request send offer
	offerId = newOfferId()
	err = c.cslq.Encode("c [c]c", remoteSend, offerId)
	if err != nil {
		err = Error(err, "Cannot send offer with id", offerId)
		return
	}
	shrunken := shrinkPaths(files)
	err = json.NewEncoder(c.conn).Encode(shrunken)
	if err != nil {
		err = Error(err, "Cannot send offer info", offerId)
		return
	}
	// Read result code
	err = c.cslq.Decode("c", &code)
	if err != nil {
		err = Error(err, "Cannot read result code")
		return
	}
	return
}

func (c Client) Download(
	offerId OfferId,
	index int,
	offset int64,
) (err error) {
	// Request download
	if err = c.cslq.Encode("c [c]c q q", remoteDownload, offerId, index, offset); err != nil {
		err = Error(err, "Cannot request download")
	}

	// Read confirmation
	var code byte
	err = c.cslq.Decode("c", &code)
	if err != nil {
		err = Error(err, "Cannot read confirmation")
		return
	}

	return
}
