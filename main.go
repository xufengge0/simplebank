package main

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"github.com/rs/cors"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq" // 数据库引擎的包没有显式使用要加下划线
	"github.com/rakyll/statik/fs"
	"github.com/rs/zerolog"
	"github.com/techschool/simplebank/api"
	db "github.com/techschool/simplebank/db/sqlc"
	_ "github.com/techschool/simplebank/doc/statik" // swagger二进制资源
	"github.com/techschool/simplebank/gapi"
	"github.com/techschool/simplebank/mail"
	"github.com/techschool/simplebank/pb"
	"github.com/techschool/simplebank/util"
	"github.com/techschool/simplebank/worker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

// 定义要监听的信号：Ctrl+C 或 SIGTERM
var interruptSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGINT,
	syscall.SIGTERM,
}

func main() {

	// 从文件读取配置
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config:")
	}
	// 如果是开发环境则设置可视化日志
	if config.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// 创建一个带有取消功能的 context，当收到信号时，该 context 会被取消
	ctx, stop := signal.NotifyContext(context.Background(), interruptSignals...)
	defer stop() // 程序结束时调用 stop 来释放相关资源

	// 连接数据库
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db:")
	}

	// 创建server
	store := db.NewStore(conn)

	redisOpt := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}
	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)

	// 创建一个 errgroup，用于管理多个 goroutine 的执行和错误处理
	waitGroup, ctx := errgroup.WithContext(ctx)

	runGatewayServer(ctx, waitGroup, config, store, taskDistributor)
	runGrpcServer(ctx, waitGroup, config, store, taskDistributor)
	runTaskProcessor(ctx, waitGroup, redisOpt, store, config)

	// 等待所有 goroutine 完成
	err = waitGroup.Wait()
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		log.Fatal().Err(err).Msg("failed to wait for waitGroup")
	}

}

// 处理队列中的任务
func runTaskProcessor(
	ctx context.Context,
	waitGroup *errgroup.Group,
	redisOpt asynq.RedisClientOpt,
	store db.Store,
	config util.Config) {

	mailer := mail.NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)

	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, store, mailer)

	log.Info().Msg("start task processor")

	err := taskProcessor.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start task processor")
	}

	// 在上下文被取消时关闭 taskProcessor
	waitGroup.Go(func() error {
		<-ctx.Done() // 等待上下文的取消信号
		log.Info().Msg("task processor is shutting down...")

		taskProcessor.Shutdown()
		log.Info().Msg("task processor has been gracefully stopped")
		return nil
	})

}

// 运行GinServer
func runGinServer(config util.Config, store db.Store) {
	// 创建server
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server:")
	}

	// 运行server
	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start server:")
	}
}

// 运行GrpcServer
func runGrpcServer(
	ctx context.Context,
	waitGroup *errgroup.Group,
	config util.Config,
	store db.Store,
	taskDistributor worker.TaskDistributor) {

	server, err := gapi.NewServer(config, store, taskDistributor)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server:")
	}

	grpclogger := grpc.UnaryInterceptor(gapi.GrpcLogger) // 注册一元拦截器(日志中间件)

	grpcServer := grpc.NewServer(grpclogger)        // creates a gRPC server which has no service registered
	pb.RegisterSimpleBankServer(grpcServer, server) // grpcServer注册server服务

	reflection.Register(grpcServer) // 注册反射服务

	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create listener:")
	}

	// 启动一个新的 goroutine 来运行 gRPC 服务器
	waitGroup.Go(func() error {
		log.Info().Msgf("start gRPC server at %s", listener.Addr().String())

		err = grpcServer.Serve(listener)
		if err != nil {
			if errors.Is(err, grpc.ErrServerStopped) {
				return nil
			}
			log.Error().Err(err).Msg("cannot start gRPC server")
			return err
		}
		return nil
	})

	// 在上下文被取消时关闭 gRPC 服务器
	waitGroup.Go(func() error {
		<-ctx.Done() // 等待上下文的取消信号
		log.Info().Msg("gRPC server is shutting down...")

		grpcServer.GracefulStop()
		log.Info().Msg("gRPC server has been gracefully stopped")
		return ctx.Err()
	})

}

// runGatewayServer 启动 gRPC-HTTP 网关服务器，
// 将 HTTP 请求转发给 gRPC 服务。
func runGatewayServer(
	ctx context.Context,
	waitGroup *errgroup.Group,
	config util.Config,
	store db.Store,
	taskDistributor worker.TaskDistributor) {

	// 创建gRPC服务器实例
	server, err := gapi.NewServer(config, store, taskDistributor)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server:")
	}

	// 设置JSON序列化和反序列化选项（蛇盒命名）
	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})

	// 创建一个 HTTP-gRPC 转换器（ServeMux）来处理 HTTP 请求
	grpcMux := runtime.NewServeMux(jsonOption)

	// 注册 gRPC 服务器的 HTTP 处理程序，将请求通过 gRPC 转发到服务器
	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot register handler server:")
	}

	
	// 创建一个 HTTP 请求路由器（ServeMux）
	mux := http.NewServeMux()
	// 将所有请求传递给 grpcMux 进行处理
	mux.Handle("/", grpcMux)

	// 创建一个虚拟文件系统，可以通过这个文件系统访问打包进来的静态文件
	statikFS, err := fs.New()
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create statikFS:")
	}

	// 创建一个 HTTP 处理器，用于处理以 "/swagger/" 开头的请求，并将请求转发到 statikFS 中
	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFS))
	// 将以 "/swagger/" 开头的请求给swaggerHandler处理
	mux.Handle("/swagger/", swaggerHandler)

	// 创建一个 CORS 中间件，允许来自前端域的请求
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000"},
		AllowedMethods: []string{
			http.MethodGet, 
			http.MethodPost, 
			http.MethodPut, 
			http.MethodDelete, 
			http.MethodPatch},
		AllowedHeaders: []string{
			"Authidrization",
			"Content-type",
		},
	})
	handler := c.Handler(gapi.HttpLogger(mux))
	httpServer := &http.Server{
		Handler: handler,
		Addr:    config.HTTPServerAddress,
	}

	// 启动一个新的 goroutine 来运行 HTTP 服务器
	waitGroup.Go(func() error {
		log.Info().Msgf("start HTTP gateway server at %s", httpServer.Addr)

		// 启动 HTTP 服务器
		err = httpServer.ListenAndServe()
		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}
			log.Error().Err(err).Msg("cannot start HTTP gateway server:")
			return err
		}
		return nil
	})

	// 在上下文被取消时关闭 gRPC 服务器
	waitGroup.Go(func() error {
		<-ctx.Done() // 等待上下文的取消信号
		log.Info().Msg("gRPC server is shutting down...")

		err := httpServer.Shutdown(context.Background())
		if err != nil {
			log.Error().Err(err).Msg("cannot shutdown HTTP gateway server:")
			return err
		}
		log.Info().Msg("gRPC server has been gracefully stopped")
		return nil
	})

}
