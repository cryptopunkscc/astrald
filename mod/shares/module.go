package shares

import (
	"context"
	"errors"
	"time"
)

const ModuleName = "shares"
const DBPrefix = "shares__"
const RemoteSetType = "remote"
const DescribeAction = "shares.describe"

type Module interface {
}

type RemoteShare interface {
	Sync(context.Context) error
	Unsync() error
	LastUpdate() time.Time
}

var ErrDenied = errors.New("access denied")
