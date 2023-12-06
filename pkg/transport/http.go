package transport

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/mar-coding/personalWebsiteBackend/pkg/middlewares"
	"github.com/mar-coding/swaggerui"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
)

type HTTPBootstrapper interface {
	BaseTransporter
	AddHandler(routerPath string, handler http.Handler)
	AddHandlerFunc(routerPath string, handlerFunc http.HandlerFunc)
	// RegisterServiceEndpoint register grpc gateway endpoint
	RegisterServiceEndpoint(endpoint func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error) error
}

// NewHTTPServer create http server transport
func NewHTTPServer(
	ctx context.Context,
	httpAddress, grpcAddress string,
	development bool,
	swagger []byte,
	customHeaders []string,
	origins []string,
	middleware func(handler http.Handler) http.Handler,
	muxOpts ...runtime.ServeMuxOption,
) HTTPBootstrapper {

	httpServer := new(HTTPServer)

	if len(muxOpts) == 0 {
		muxOpts = make([]runtime.ServeMuxOption, 0)
		muxOpts = append(muxOpts, runtime.WithErrorHandler(middlewares.ErrorHandler))
	}

	rMux := runtime.NewServeMux(muxOpts...)

	muxHandlers := http.NewServeMux()
	muxHandlers = middlewares.SetRuntimeAsRootHandler(muxHandlers, rMux)

	if development {
		muxHandlers = middlewares.DebuggerHandler(muxHandlers)
		muxHandlers = middlewares.SwaggerHandler(muxHandlers, "swagger.json", swagger)
		muxHandlers.Handle("/api-docs/", http.StripPrefix("/api-docs", swaggerui.Handler(swagger)))
	}

	srv := &http.Server{
		Handler:      middleware(middlewares.AllowCORS(muxHandlers, origins, customHeaders...)),
		Addr:         httpAddress,
		ReadTimeout:  _defaultReadTimeout,
		WriteTimeout: _defaultWriteTimeout,
	}

	httpServer.server = srv
	httpServer.notify = make(chan error)
	httpServer.ctx = ctx
	httpServer.shutdownTimeout = _defaultShutdownTimeout
	httpServer.grpcAddress = grpcAddress
	httpServer.rMux = rMux
	httpServer.mux = muxHandlers

	return httpServer
}

func (s *HTTPServer) Start() {
	go func() {
		s.notify <- s.server.ListenAndServe()
		close(s.notify)
	}()
}

func (s *HTTPServer) Notify() <-chan error {
	return s.notify
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *HTTPServer) AddHandler(routerPath string, handler http.Handler) {
	s.mux.Handle(routerPath, handler)
}

func (s *HTTPServer) AddHandlerFunc(routerPath string, handlerFunc http.HandlerFunc) {
	s.mux.HandleFunc(routerPath, handlerFunc)
}

func (s *HTTPServer) RegisterServiceEndpoint(endpoint func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error) error {
	return endpoint(s.ctx, s.rMux, s.grpcAddress, []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())})
}
