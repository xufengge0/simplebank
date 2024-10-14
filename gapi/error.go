package gapi

import (
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func fieldViolation(field string, err error) *errdetails.BadRequest_FieldViolation {
	return &errdetails.BadRequest_FieldViolation{
		Field:       field,
		Description: err.Error(),
	}
}

// 将violations转为error
func invalidArgumentErr(violations []*errdetails.BadRequest_FieldViolation) error {
	badRequest := &errdetails.BadRequest{FieldViolations: violations}

	// 创建了一个gRPC错误状态对象
	statusInvalid := status.New(codes.InvalidArgument, "invalid paremeters")

	// 将badRequest添加到statusInvalid中
	statusDetails, err := statusInvalid.WithDetails(badRequest)
	if err != nil {
		return statusInvalid.Err()
	}
	return statusDetails.Err()
}

// 未认证化错误
func unauthenticatedErr(err error) error {
	return status.Error(codes.Unauthenticated, err.Error())
}
