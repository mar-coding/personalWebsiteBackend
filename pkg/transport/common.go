package transport

import (
	"context"
	"time"
)

const (
	_defaultReadTimeout     = 5 * time.Second
	_defaultWriteTimeout    = 5 * time.Second
	_defaultShutdownTimeout = 3 * time.Second
)

type BaseTransporter interface {
	Start()
	Notify() <-chan error
	Shutdown(ctx context.Context) error
}
