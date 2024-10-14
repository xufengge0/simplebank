package gapi

import (
	"context"

	db "github.com/techschool/simplebank/db/sqlc"
	"github.com/techschool/simplebank/pb"
	"github.com/techschool/simplebank/val"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {
	// 检查参数格式是否正确，返回详细的错误说明
	violations := validateVerifyEmailRequest(req)
	if violations != nil {
		return nil, invalidArgumentErr(violations)
	}

	txResult, err := server.store.VerifyEmailTX(ctx, db.VerifyEmailTXParams{
		ID:         req.GetEmailId(),
		SecretCode: req.GetSecretCode(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to verifyemail:%v",err)
	}

	res:= &pb.VerifyEmailResponse{
		IsVerified: txResult.User.IsEmailVerified,
	}
	return res, nil
}
func validateVerifyEmailRequest(req *pb.VerifyEmailRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateEmailID(req.GetEmailId()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}
	if err := val.ValidateSecretCode(req.GetSecretCode()); err != nil {
		violations = append(violations, fieldViolation("password", err))
	}

	return
}
