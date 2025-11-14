package kcp

import "errors"

var ErrEphemeralListenerExists = errors.New("ephemeral listener already exists")
var ErrEndpointLocalSocketExists = errors.New("endpoint local socket mapping already exists")
var ErrEphemeralListenerNotExist = errors.New("ephemeral listener not exists")
