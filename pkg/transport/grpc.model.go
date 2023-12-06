package transport

import (
	"google.golang.org/grpc"
	"net"
)

type GRPCServer struct {
	server   *grpc.Server
	listener net.Listener
	notify   chan error
}
