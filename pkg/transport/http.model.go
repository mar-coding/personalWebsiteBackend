package transport

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"net/http"
	"time"
)

type HTTPServer struct {
	ctx             context.Context
	server          *http.Server
	notify          chan error
	shutdownTimeout time.Duration
	rMux            *runtime.ServeMux
	mux             *http.ServeMux
	grpcAddress     string
}
