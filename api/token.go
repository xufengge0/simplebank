package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type renewAccessTockenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type renewAccessTockenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

// 根据刷新令牌返回新的访问令牌
func (server *Server) renewAccessToken(ctx *gin.Context) {
	var req renewAccessTockenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// 验证refreshToken是否有效、未过期
	refreshPayload, err := server.tokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	// 从数据库查询session是否存在
	session, err := server.store.GetSession(ctx, refreshPayload.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// 检查session是否被阻塞
	if session.IsBlocked {
		err = fmt.Errorf("blocked session")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}
	if refreshPayload.Username != session.Username {
		err = fmt.Errorf("incorrect session user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}
	if time.Now().After(session.ExpiresAt) {
		err := fmt.Errorf("expired  session")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}
	
	// 创建新的accessToken
	accessToken, accessPayload, err := server.tokenMaker.CreateToken(session.Username, server.config.AccessTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// 返回新的accessToken
	res := renewAccessTockenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessPayload.ExpiredAt,
	}

	ctx.JSON(http.StatusOK, res)
}
