package rpc

import (
	"context"
	webPB "github.com/mar-coding/personalWebsiteBackend/APIs/proto-gen/services/website/v1"
)

func (w WebsiteService) TestPostAPI(ctx context.Context, in *webPB.TestPostAPIRequest) (*webPB.TestPostAPIResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (w WebsiteService) TestGetAPI(ctx context.Context, in *webPB.TestGetAPIRequest) (*webPB.TestGetAPIResponse, error) {
	//TODO implement me
	panic("implement me")
}
