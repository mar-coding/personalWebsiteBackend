package logger

import (
	"github.com/getsentry/sentry-go"
	"log/slog"
)

type Log struct {
	skipCaller int
	slog       *slog.Logger
	sentry     *sentry.Client
}

type SentryConfig struct {
	DSN                string
	ServerName         string
	Environment        Environment
	AttachStacktrace   bool
	SampleRate         float64
	EnableTracing      bool
	TracesSampleRate   float64
	ProfilesSampleRate float64
	MaxBreadcrumbs     int
	Release            string
	Dist               string
	HTTPProxy          string
	HTTPSProxy         string
	MaxErrorDepth      int
	Debug              bool
}

type Options struct {
	Development  bool          // Development add development details of machine
	Debug        bool          // Debug show debug devel message
	EnableCaller bool          // EnableCaller show caller in line code
	SkipCaller   int           // SkipCaller skip caller level of CallerFrames https://github.com/golang/go/issues/59145#issuecomment-1481920720
	Sentry       *SentryConfig // Sentry enable sentry with specific configuration
}
