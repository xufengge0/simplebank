package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	db "github.com/techschool/simplebank/db/sqlc"
	"github.com/techschool/simplebank/util"
)

// createUser的请求结构体
type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"` // 限制ASCII码
	Password string `json:"password" binding:"required,min=6"`
	Fullname string `json:"fullname" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

// 回复的结构体
type createUserResponse struct {
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

// 用户回复结构体(不包含密码)
func newUserResponse(user db.User) createUserResponse {
	return createUserResponse{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}
}

// 用户注册
func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err)) // 返回请求参数错误
		return
	}

	hashPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err)) // 返回请求参数错误
		return
	}

	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashPassword,
		FullName:       req.Fullname,
		Email:          req.Email,
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		// 将err断言
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation": // 唯一约束(username、email)
				ctx.JSON(http.StatusForbidden, errorResponse(err)) // 返回403权限错误
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err)) // 返回服务器内部错误
		return
	}

	res := newUserResponse(user)

	ctx.JSON(http.StatusOK, res)
}

// 登录请求结构体
type LoginUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"` // 限制ASCII码
	Password string `json:"password" binding:"required,min=6"`
}

// 登录回复结构体
type LoginUserResponse struct {
	AccessToken string             `json:"access_token"`
	User        createUserResponse `json:"user"`
}

func (server *Server) loginUser(ctx *gin.Context) {
	var req LoginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// 从数据库查询用户是否存在
	user, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// 检查密码是否正确
	err = util.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err)) // 返回401未授权错误
		return
	}

	// 创建token
	accessToken, err := server.tokenMaker.CreateToken(user.Username, server.config.AccessTokenDurationMinutes)
	if err!= nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	res := LoginUserResponse{
		AccessToken: accessToken,
		User:        newUserResponse(user),
	}

	ctx.JSON(http.StatusOK, res)
	
}