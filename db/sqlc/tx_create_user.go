package db

import (
	"context"

	"github.com/lib/pq"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CreateUserTXParams struct {
	CreateUserParams
	AfterCreate func(user User) error // CreateUser执行后的回调函数
}
type CreateUserTXResult struct {
	User User
}
// CreateUserTX 创建用户并发送验证邮件（一次transaction）
func (store *SqlStore) CreateUserTX(ctx context.Context, arg CreateUserTXParams) (CreateUserTXResult, error) {
	var res CreateUserTXResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		res.User, err = q.CreateUser(ctx, arg.CreateUserParams)
		if err != nil {
			// 将err断言
			if pqErr, ok := err.(*pq.Error); ok {
				switch pqErr.Code.Name() {
				case "unique_violation": // 唯一约束(username、email)
					return status.Errorf(codes.AlreadyExists, "username already exists:%s", err)
				}
			}
			return status.Errorf(codes.Internal, "failed to create user:%s", err)
		}

		err = arg.AfterCreate(res.User)
		if err != nil {
			return err
		}

		return nil
	})

	return res, err
}
