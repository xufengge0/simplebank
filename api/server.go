package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "github.com/techschool/simplebank/db/sqlc"
	"github.com/techschool/simplebank/token"
	"github.com/techschool/simplebank/util"
)

type Server struct {
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	// 使用配置文件的对称密钥创建令牌
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey) // 可实现JWT和paseto任意切换
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}

	// 尝试将 Engine() 返回的值断言为 *validator.Validate 类型
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 在验证引擎 v 上为 "currency" 标签注册了 validCurrency 自定义验证函数
		v.RegisterValidation("currency", vaildCurrency)
	}

	server.setupRouter()
	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()
	
	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker))
	authRoutes.POST("/accounts", server.createAccount)   // 创建账户
	authRoutes.GET("/accounts/:id", server.getAccount)   // 根据id获取账户
	authRoutes.GET("/accounts", server.listAccount)      // 根据page_id、page_size获取账户
	authRoutes.POST("/transfers", server.createTransfer) // 转账

	router.POST("/users", server.createUser)      // 用户注册
	router.POST("/users/login", server.loginUser) // 用户登录，并返回令牌

	server.router = router
}

// 在指定IP地址运行服务
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

// 将err转为map
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
