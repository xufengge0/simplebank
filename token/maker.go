package token

import "time"

type Maker interface {
	// 创建token
	CreateToken(username string, duration time.Duration) (string, error)
	// 验证token是否有效
	VerifyToken(token string) (*Payload, error)
}