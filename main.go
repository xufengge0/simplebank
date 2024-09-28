package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq" // 数据库引擎的包没有显式使用要加下划线
	"github.com/techschool/simplebank/api"
	db "github.com/techschool/simplebank/db/sqlc"
	"github.com/techschool/simplebank/util"
)

/* const (
	dbDriver = "postgres"
	// github action测试启用
	//dbSource = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
	// 容器中测试
	dbSource      = "host=postgres12 port=5432 user=root password=secret dbname=simple_bank sslmode=disable"
	serverAddress = "0.0.0.0:8080"
) */

func main() {
	// 从文件/环境变量加载配置
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	// 连接数据库
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	// 创建server
	store := db.NewStore(conn)
	server,err := api.NewServer(config,store)
	if err!= nil {
		log.Fatal("cannot create server:", err)
	}

	// 运行server
	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
	
}
