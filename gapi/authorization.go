package gapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/techschool/simplebank/token"
	"google.golang.org/grpc/metadata"
)

var (
	authorizationHeader     = "authorization"
	authorizationTypeBearer = "bearer"
)

// 身份认证，从metadata中获取authorization信息，返回payload
func (server *Server) authorizeUser(ctx context.Context) (*token.Payload, error) {
	// 获取metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("metadata is not provided")
	}

	// 获取authorization header
	values := md.Get(authorizationHeader)
	if len(values) == 0 {
		return nil, fmt.Errorf("authorization header is not provided")
	}

	// 获取authorization header 格式为：Bearer accessToken
	authHeader := values[0]
	fields := strings.Fields(authHeader)
	if len(fields) < 2 {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	// 检查是否为bearer类型
	authType := strings.ToLower(fields[0])
	if authType != authorizationTypeBearer {
		return nil, fmt.Errorf("unsupported authorization type %s", authType)
	}

	// 检查accessToken是否合法，返回payload
	accessToken := fields[1]
	payload, err := server.tokenMaker.VerifyToken(accessToken)
	if err != nil {
		return nil, fmt.Errorf("token is invalid:%w", err)
	}
	return payload, nil
}
