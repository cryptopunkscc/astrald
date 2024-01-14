package relay

import "errors"

var ErrCertAlreadyIndexed = errors.New("certificate already indexed")
var ErrCertNotFound = errors.New("certificate not found")
