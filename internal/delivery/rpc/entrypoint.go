package rpc

import (
	webPB "github.com/mar-coding/personalWebsiteBackend/APIs/proto-gen/services/website/v1"
	"google.golang.org/grpc"
)

type WebsiteService struct {
	webPB.UnsafeWebServiceServer
}

func NewWebsiteService(server *grpc.Server) {
	websiteService := &WebsiteService{}
	webPB.RegisterWebServiceServer(server, websiteService)
}
