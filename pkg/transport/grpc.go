package transport

import (
	"context"
	"github.com/mar-coding/personalWebsiteBackend/pkg/middlewares"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

type GRPCBootstrapper interface {
	BaseTransporter
	GetGRPCServer() *grpc.Server
}

// NewGRPCServer create grpc server transport
func NewGRPCServer(
	grpcAddress string,
	development bool,
	grpcMiddlewares ...grpc.UnaryServerInterceptor,
) (GRPCBootstrapper, error) {
	grpcServer := new(GRPCServer)

	listener, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		return nil, err
	}

	srv := grpc.NewServer(middlewares.New(grpcMiddlewares...))

	if development {
		reflection.Register(srv)
	}

	grpcServer.listener = listener
	grpcServer.server = srv
	grpcServer.notify = make(chan error)

	return grpcServer, nil

}

func (g *GRPCServer) Start() {
	go func() {
		g.notify <- g.server.Serve(g.listener)
		close(g.notify)
	}()
}

func (g *GRPCServer) Notify() <-chan error {
	return g.notify
}

func (g *GRPCServer) Shutdown(_ context.Context) error {
	g.server.GracefulStop()
	return nil
}

func (g *GRPCServer) GetGRPCServer() *grpc.Server {
	return g.server
}
