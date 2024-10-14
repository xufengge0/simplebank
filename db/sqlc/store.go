package db

import (
	"context"
	"database/sql"
	"fmt"
)
type Store interface{
	Querier
	TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error)
	CreateUserTX(ctx context.Context, arg CreateUserTXParams) (CreateUserTXResult, error)
	VerifyEmailTX(ctx context.Context, arg VerifyEmailTXParams) (VerifyEmailTXResult, error) 
}
// SqlStore provides all functions to execute db queries and transactions
type SqlStore struct {
	*Queries
	db *sql.DB // 连接池
}

func NewStore(db *sql.DB) Store {
	return &SqlStore{
		db:      db,
		Queries: New(db),
	}
}

// execTx executes a function within a database transaction
func (store *SqlStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil) // 启动一个事务
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q) // 执行回调函数
	if err != nil {
		if rberr := tx.Rollback(); rberr != nil {
			return fmt.Errorf("tx err:%v,rberr:%v", err, rberr)
		}
		return err
	}
	return tx.Commit()
}
