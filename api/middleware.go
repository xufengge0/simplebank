package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/techschool/simplebank/token"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

// 身份认证中间件
func authMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		
		// 从HTTP请求中提取authorization头部
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			err := errors.New("authorization header is not provided")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
		// 按空格拆分authorization头部字符串
		fileds := strings.Fields(authorizationHeader)
		if len(fileds) < 2 {
			err := errors.New("invaild authorization header fomat")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
		// 判断type是否为bearer
		authorizationType := strings.ToLower(fileds[0])
		if authorizationType != authorizationTypeBearer {
			err := errors.New("unsupported authorization type")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
		// 验证token是否有效
		accessToken := fileds[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			err := errors.New("unsupported authorization type")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
		// 将payload中存入ctx中
		ctx.Set(authorizationPayloadKey, payload)
		ctx.Next()
	}
}
