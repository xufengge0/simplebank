package db

import (
	"context"
	"database/sql"
	"fmt"
)

type VerifyEmailTXParams struct {
	ID         int64
	SecretCode string
}
type VerifyEmailTXResult struct {
	User        User
	VerifyEmail VerifyEmail
}

// 验证邮件链接的处理函数（一次transaction）
func (store *SqlStore) VerifyEmailTX(ctx context.Context, arg VerifyEmailTXParams) (VerifyEmailTXResult, error) {
	var res VerifyEmailTXResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		// 验证邮件链接是否存在，并将is_used字段设为true
		res.VerifyEmail, err = q.UpdateVerifyEmail(ctx, UpdateVerifyEmailParams{
			ID:         arg.ID,
			SecretCode: arg.SecretCode,
		})
		if err != nil {
			return fmt.Errorf("update verify email failed: %v", err)
		}

		// 验证成功后，将用户的邮箱验证状态设置为true
		res.User, err = q.UpdateUser(ctx, UpdateUserParams{
			IsEmailVerified: sql.NullBool{Bool: true, Valid: true},
			Username:        res.VerifyEmail.Username,
		})
		if err != nil {
			return fmt.Errorf("update user failed: %v", err)
		}

		return nil
	})

	return res, err
}
