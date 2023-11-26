package logger

import (
	"bytes"
	"log/slog"
	"sync"
)

type Handler struct {
	h slog.Handler
	r func([]string, slog.Attr) slog.Attr
	b *bytes.Buffer
	m *sync.Mutex
}
