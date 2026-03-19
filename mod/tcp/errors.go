package tcp

import "errors"

var ErrEphemeralListenerExists = errors.New("ephemeral listener already exists")
var ErrEphemeralListenerNotExist = errors.New("ephemeral listener not exists")
