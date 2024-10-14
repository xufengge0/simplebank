package gapi

import (
	"context"
	"database/sql"
	"time"

	db "github.com/techschool/simplebank/db/sqlc"
	"github.com/techschool/simplebank/pb"
	"github.com/techschool/simplebank/util"
	"github.com/techschool/simplebank/val"
	"google.golang.org/genproto/googleapis/rpc/errdetails"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	// 鉴权
	payload, err := server.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedErr(err)
	}

	// 检查参数格式是否正确，返回详细的错误说明
	violations := validateUpdateUserRequest(req)
	if violations != nil {
		return nil, invalidArgumentErr(violations)
	}

	// 检查用户是否有权限更新其他用户信息(授权)
	if payload.Username != req.GetUsername() {
		return nil, status.Errorf(codes.PermissionDenied, "cannot update other user info")
	}

	arg := db.UpdateUserParams{
		Username: req.GetUsername(),
		FullName: sql.NullString{String: req.GetFullName(), Valid: req.GetFullName() != ""},
		Email:    sql.NullString{String: req.GetEmail(), Valid: req.GetEmail() != ""},
	}

	// 检查密码是否为空，不为空则进行哈希、更新密码、更新密码修改时间
	if req.GetPassword() != "" {
		hashPassword, err := util.HashPassword(req.GetPassword())
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to hash password:%s", err)
		}

		arg.HashedPassword = sql.NullString{String: hashPassword, Valid: req.GetPassword() != ""}
		arg.PasswordChangedAt = sql.NullTime{Time: time.Now(), Valid: true}
	}

	user, err := server.store.UpdateUser(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to update user:%s", err)
	}

	res := &pb.UpdateUserResponse{
		User: convertUser(user),
	}

	return res, nil
}
func validateUpdateUserRequest(req *pb.UpdateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}
	if req.Password != nil {
		if err := val.ValidatePassword(req.GetPassword()); err != nil {
			violations = append(violations, fieldViolation("password", err))
		}
	}
	if req.FullName != nil {
		if err := val.ValidateFullname(req.GetFullName()); err != nil {
			violations = append(violations, fieldViolation("full_name", err))
		}
	}
	if req.Email != nil {
		if err := val.ValidateEmail(req.GetEmail()); err != nil {
			violations = append(violations, fieldViolation("email", err))
		}
	}
	return
}
