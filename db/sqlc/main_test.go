package db

import (
	"database/sql"
	"log"
	"os"
	"testing"
	_"github.com/lib/pq" // 数据库引擎的包没有显式使用要加下划线
)

const (
	dbDriver = "postgres"
	//dbSource = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
	dbSource = "host=postgres12 port=5432 user=root password=secret dbname=simple_bank sslmode=disable"
)

var testQueries *Queries
var testDB *sql.DB

func TestMain(m *testing.M) {
	var err error
	testDB, err = sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	testQueries = New(testDB)
	os.Exit(m.Run())
}
