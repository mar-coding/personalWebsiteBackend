package middlewares

import (
	"context"
	"encoding/hex"
	"github.com/getsentry/sentry-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"regexp"
	"time"
)

func recoverWithSentry(hub *sentry.Hub, ctx context.Context, o *options) {
	if err := recover(); err != nil {
		eventID := hub.RecoverWithContext(ctx, err)
		if eventID != nil && o.WaitForDelivery {
			hub.Flush(o.Timeout)
		}
	}
}

// continueFromGrpcMetadata returns a span option that updates the span to continue
// an existing trace. If it cannot detect an existing trace in the request, the
// span will be left unchanged.
func continueFromGrpcMetadata(md metadata.MD) sentry.SpanOption {
	return func(s *sentry.Span) {
		if md == nil {
			return
		}

		trace, ok := md["sentry-trace"]
		if !ok {
			return
		}
		if len(trace) != 1 {
			return
		}
		if trace[0] == "" {
			return
		}
		updateFromSentryTrace(s, []byte(trace[0]))
	}
}

// Re-export of functions from tracing.go of sentry-go
var sentryTracePattern = regexp.MustCompile(`^([[:xdigit:]]{32})-([[:xdigit:]]{16})(?:-([01]))?$`)

func updateFromSentryTrace(s *sentry.Span, header []byte) {
	m := sentryTracePattern.FindSubmatch(header)
	if m == nil {
		// no match
		return
	}
	_, _ = hex.Decode(s.TraceID[:], m[1])
	_, _ = hex.Decode(s.ParentSpanID[:], m[2])
	if len(m[3]) != 0 {
		switch m[3][0] {
		case '0':
			s.Sampled = sentry.SampledFalse
		case '1':
			s.Sampled = sentry.SampledTrue
		}
	}
}

func toSpanStatus(code codes.Code) sentry.SpanStatus {
	switch code {
	case codes.OK:
		return sentry.SpanStatusOK
	case codes.Canceled:
		return sentry.SpanStatusCanceled
	case codes.Unknown:
		return sentry.SpanStatusUnknown
	case codes.InvalidArgument:
		return sentry.SpanStatusInvalidArgument
	case codes.DeadlineExceeded:
		return sentry.SpanStatusDeadlineExceeded
	case codes.NotFound:
		return sentry.SpanStatusNotFound
	case codes.AlreadyExists:
		return sentry.SpanStatusAlreadyExists
	case codes.PermissionDenied:
		return sentry.SpanStatusPermissionDenied
	case codes.ResourceExhausted:
		return sentry.SpanStatusResourceExhausted
	case codes.FailedPrecondition:
		return sentry.SpanStatusFailedPrecondition
	case codes.Aborted:
		return sentry.SpanStatusAborted
	case codes.OutOfRange:
		return sentry.SpanStatusOutOfRange
	case codes.Unimplemented:
		return sentry.SpanStatusUnimplemented
	case codes.Internal:
		return sentry.SpanStatusInternalError
	case codes.Unavailable:
		return sentry.SpanStatusUnavailable
	case codes.DataLoss:
		return sentry.SpanStatusDataLoss
	case codes.Unauthenticated:
		return sentry.SpanStatusUnauthenticated
	default:
		return sentry.SpanStatusUndefined
	}
}

type Option interface {
	Apply(*options)
}

// newConfig returns a config configured with all the passed Options.
func newConfig(opts []Option) *options {
	optsCopy := *defaultOptions
	c := &optsCopy
	for _, o := range opts {
		o.Apply(c)
	}
	return c
}

var defaultOptions = &options{
	Repanic:         false,
	WaitForDelivery: false,
	ReportOn:        ReportAlways,
	Timeout:         1 * time.Second,
}

type options struct {
	// Repanic configures whether Sentry should repanic after recovery.
	Repanic bool

	// WaitForDelivery configures whether you want to block the request before moving forward with the response.
	WaitForDelivery bool

	// Timeout for the event delivery requests.
	Timeout time.Duration

	ReportOn func(error) bool
}

func ReportAlways(error) bool {
	return true
}

func ReportOnCodes(cc ...codes.Code) Reporter {
	return func(err error) bool {
		for i := range cc {
			if status.Code(err) == cc[i] {
				return true
			}
		}
		return false
	}
}

type repanicOption struct {
	Repanic bool
}

func (r *repanicOption) Apply(o *options) {
	o.Repanic = r.Repanic
}

func WithRepanicOption(b bool) Option {
	return &repanicOption{Repanic: b}
}

type waitForDeliveryOption struct {
	WaitForDelivery bool
}

func (w *waitForDeliveryOption) Apply(o *options) {
	o.WaitForDelivery = w.WaitForDelivery
}

func WithWaitForDelivery(b bool) Option {
	return &waitForDeliveryOption{WaitForDelivery: b}
}

type timeoutOption struct {
	Timeout time.Duration
}

func (t *timeoutOption) Apply(o *options) {
	o.Timeout = t.Timeout
}

func WithTimeout(t time.Duration) Option {
	return &timeoutOption{Timeout: t}
}

type Reporter func(error) bool

type reportOnOption struct {
	ReportOn Reporter
}

func (r *reportOnOption) Apply(o *options) {
	o.ReportOn = r.ReportOn
}

func WithReportOn(r Reporter) Option {
	return &reportOnOption{ReportOn: r}
}
