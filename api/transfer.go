package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/techschool/simplebank/db/sqlc"
	"github.com/techschool/simplebank/token"
)

// createAccount的请求结构体
type transferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`       // "greater than"大于0
	Currency      string `json:"currency" binding:"required,currency"` // 使用了自定义currency验证器
}

func (server *Server) createTransfer(ctx *gin.Context) {
	var req transferRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err)) // 返回请求参数错误
		return
	}

	// 检查转出账户有效性
	fromAccount, valid := server.vaildAccount(ctx, req.FromAccountID, req.Currency)
	if !valid {
		return
	}
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if fromAccount.Owner!= authPayload.Username {
		err:=errors.New("account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized,errorResponse(err))
		return
	}

	// 检查转入账户有效性
	_, valid = server.vaildAccount(ctx, req.ToAccountID, req.Currency)
	if !valid {
		return
	}

	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}
	// 进行转账（操作三个表）
	result, err := server.store.TransferTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err)) // 返回服务器内部错误
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// 判断账户是否存在/货币类型是否相同
func (server *Server) vaildAccount(ctx *gin.Context, accountID int64, currency string) (db.Account, bool) {
	// 判断账户是否存在
	account, err := server.store.GetAccount(ctx, accountID)
	if err != nil {
		if err == sql.ErrNoRows { // 如果id不存在表中
			ctx.JSON(http.StatusNotFound, errorResponse(err)) // 返回404
			return account, false
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err)) // 返回服务器内部错误
		return account, false
	}
	// 判断账户货币类型是否相同
	if account.Currency != currency {
		err := fmt.Errorf("account[%d] currency mismatch:%s vs %s", accountID, account.Currency, currency)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return account, false
	}
	return account, true
}
