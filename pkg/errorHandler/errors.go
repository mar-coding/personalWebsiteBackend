package errorHandler

import (
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	errPB "github.com/mar-coding/personalWebsiteBackend/APIs/proto-gen/components/errors/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ Handler = (*Error)(nil)

type Handler interface {
	New(grpcCode codes.Code, grpcErrorDetails []proto.Message, message string, params ...any) *Error
	Error() string
	GRPCStatus() *status.Status
}

// NewError create error handler object
func NewError(serviceCode uint32, serviceName, serviceVersion, domain string) (Handler, error) {
	e := new(Error)

	if serviceCode < 1 {
		return nil, errors.New("errors: can't use zero for service code")
	}

	if len(serviceName) == 0 {
		return nil, errors.New("errors: service name is empty")
	}

	e.errorDetails = make([]proto.Message, 0)

	e.errorDetails = append(e.errorDetails, &errPB.ErrorServiceDetails{
		ServiceCode:    serviceCode,
		ServiceName:    serviceName,
		ServiceVersion: serviceVersion,
		Domain:         domain,
	})
	return e, nil
}

// New create error with custom message and params
func (e *Error) New(grpcCode codes.Code, grpcErrorDetails []proto.Message, message string, params ...any) *Error {
	e.message = message
	e.params = params
	e.errorDetails = append(e.errorDetails, grpcErrorDetails...)
	e.grpcStatusCode = grpcCode

	return e
}

// Error show error message with appended parameters
func (e *Error) Error() string {
	if len(e.params) != 0 {
		return fmt.Sprintf(e.message, e.params)
	}
	return e.message
}

// GRPCStatus return grpc status
func (e *Error) GRPCStatus() *status.Status {
	s := status.Newf(e.grpcStatusCode, e.message, e.params...)

	statusWithDetails, _ := s.WithDetails(e.errorDetails...)

	return statusWithDetails
}
