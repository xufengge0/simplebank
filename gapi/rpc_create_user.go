package gapi

import (
	"context"
	"time"

	"github.com/hibiken/asynq"

	db "github.com/techschool/simplebank/db/sqlc"
	"github.com/techschool/simplebank/pb"
	"github.com/techschool/simplebank/util"
	"github.com/techschool/simplebank/val"
	"github.com/techschool/simplebank/worker"
	"google.golang.org/genproto/googleapis/rpc/errdetails"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	// 检查参数格式是否正确，返回详细的错误说明
	violations := validateCreateUserRequest(req)
	if violations != nil {
		return nil, invalidArgumentErr(violations)
	}

	hashPassword, err := util.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password:%s", err)
	}

	arg := db.CreateUserTXParams{
		CreateUserParams: db.CreateUserParams{
			Username:       req.GetUsername(),
			HashedPassword: hashPassword,
			FullName:       req.GetFullName(),
			Email:          req.GetEmail(),
		},

		AfterCreate: func(user db.User) error {
			opts := []asynq.Option{
				asynq.MaxRetry(10),                // 最大重试次数
				asynq.ProcessIn(10 * time.Second), // 10秒后再处理任务
				asynq.Queue(worker.QueueCritical), // 队列名称
			}

			// 发送验证邮件
			payload := &worker.PayloadSendVerifyEmail{
				Username: user.Username,
			}
			err = server.taskDistributor.DistributeTaskSendVerifyEmail(ctx, payload, opts...)
			if err != nil {
				return status.Errorf(codes.Internal, "failed to send verification email task:%s", err)
			}
			return nil
		},
	}
	
	// 创建用户，并发送验证邮件（一次transaction）
	txResult, err := server.store.CreateUserTX(ctx, arg)
	if err!= nil {
		return nil, status.Errorf(codes.Internal, "failed to create user:%s", err)
	}

	res := &pb.CreateUserResponse{
		User: convertUser(txResult.User),
	}

	return res, nil
}
func validateCreateUserRequest(req *pb.CreateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}
	if err := val.ValidatePassword(req.GetPassword()); err != nil {
		violations = append(violations, fieldViolation("password", err))
	}
	if err := val.ValidateFullname(req.GetFullName()); err != nil {
		violations = append(violations, fieldViolation("full_name", err))
	}
	if err := val.ValidateEmail(req.GetEmail()); err != nil {
		violations = append(violations, fieldViolation("email", err))
	}
	return
}
