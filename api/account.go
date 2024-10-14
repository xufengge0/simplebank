package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	db "github.com/techschool/simplebank/db/sqlc"
	"github.com/techschool/simplebank/token"
)

// createAccount的请求结构体
type createAccountRequest struct {
	Currency string `json:"currency" binding:"required,currency"` // 使用了自定义currency验证器
}

func (server *Server) createAccount(ctx *gin.Context) {
	var req createAccountRequest
	if err := ctx.ShouldBind(&req); err != nil { // body 绑定到结构体
		ctx.JSON(http.StatusBadRequest, errorResponse(err)) // 返回请求参数错误
		return
	}

	// 从上下文中获取token中的payload
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	arg := db.CreateAccountParams{
		Owner:    authPayload.Username,
		Currency: req.Currency,
		Balance:  0,
	}

	account, err := server.store.CreateAccount(ctx, arg)
	if err != nil {
		// 将err断言
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "foreign_key_violation", "unique_violation": // 唯一、外键约束
				ctx.JSON(http.StatusForbidden, errorResponse(err)) // 返回403权限错误
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err)) // 返回服务器内部错误
		return
	}

	ctx.JSON(http.StatusOK, account)
}

// getAccount的请求结构体
type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getAccount(ctx *gin.Context) {
	var req getAccountRequest
	if err := ctx.ShouldBindUri(&req); err != nil { // 将 URI 路径中的参数 绑定到结构体
		ctx.JSON(http.StatusBadRequest, errorResponse(err)) // 返回请求参数错误
		return
	}

	account, err := server.store.GetAccount(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows { // 如果id不存在表中
			ctx.JSON(http.StatusNotFound, errorResponse(err)) // 返回404
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err)) // 返回服务器内部错误
		return
	}

	// 比较账户的owner是否与token中的username一致
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if account.Owner != authPayload.Username {
		err := errors.New("account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
	}

	ctx.JSON(http.StatusOK, account)
}

// listAccount的请求结构体
type listAccountRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listAccount(ctx *gin.Context) {
	var req listAccountRequest
	if err := ctx.ShouldBindQuery(&req); err != nil { // 从URL查询字符串中提取参数
		ctx.JSON(http.StatusBadRequest, errorResponse(err)) // 返回请求参数错误
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	arg := db.ListAccountsParams{
		Owner:  authPayload.Username,
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	account, err := server.store.ListAccounts(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err)) // 返回服务器内部错误
		return
	}

	ctx.JSON(http.StatusOK, account)
}
