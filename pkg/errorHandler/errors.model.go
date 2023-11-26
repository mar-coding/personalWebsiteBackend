package errorHandler

import (
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
)

type Error struct {
	grpcStatusCode codes.Code
	message        string
	errorDetails   []proto.Message
	params         []any
}
