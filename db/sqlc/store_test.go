package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)
	account1 := creatRandomAccount(t)
	account2 := creatRandomAccount(t)

	n := 5              // 协程数量
	amount := int64(10) // 转账金额

	errs := make(chan error)
	results := make(chan TransferTxResult)

	// 多协程测试
	for i := 0; i < n; i++ {
		go func() {
			result, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})
			// 传回管道
			errs <- err
			results <- result
		}()
	}

	// 检查各协程的返回值
	existed := make(map[int]bool)
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		// 检查transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, transfer.FromAccountID, account1.ID)
		require.Equal(t, transfer.ToAccountID, account2.ID)
		require.Equal(t, transfer.Amount, amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// 检查entry
		fromentry := result.FromEntry
		require.NotEmpty(t, fromentry)
		require.Equal(t, fromentry.AccountID, account1.ID)
		require.Equal(t, fromentry.Amount, -amount)
		require.NotZero(t, fromentry.ID)
		require.NotZero(t, fromentry.CreatedAt)

		_, err = store.GetEntry(context.Background(), fromentry.ID)
		require.NoError(t, err)

		toentry := result.ToEntry
		require.NotEmpty(t, toentry)
		require.Equal(t, toentry.AccountID, account2.ID)
		require.Equal(t, toentry.Amount, amount)
		require.NotZero(t, toentry.ID)
		require.NotZero(t, toentry.CreatedAt)

		_, err = store.GetEntry(context.Background(), toentry.ID)
		require.NoError(t, err)

		// 检查account
		fromaccount := result.FromAccount
		require.NotEmpty(t, fromaccount)
		require.Equal(t, fromaccount.ID, account1.ID)

		toaccount := result.ToAccount
		require.NotEmpty(t, toaccount)
		require.Equal(t, toaccount.ID, account2.ID)

		diff1 := account1.Balance - fromaccount.Balance
		diff2 := toaccount.Balance - account2.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0)

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k) // 检查map不包含k
		existed[k] = true
	}
	// 检查最终的账户余额
	updateAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	updateAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	require.Equal(t, account1.Balance-amount*int64(n), updateAccount1.Balance)
	require.Equal(t, account2.Balance+amount*int64(n), updateAccount2.Balance)
}

func TestTransferTxDeadLock(t *testing.T) {
	store := NewStore(testDB)
	account1 := creatRandomAccount(t)
	account2 := creatRandomAccount(t)

	n := 10             // 协程数量
	amount := int64(10) // 转账金额
	errs := make(chan error)

	// 多协程测试
	for i := 0; i < n; i++ { // 正向、反向转账5次检查死锁
		FromAccountID := account1.ID
		ToAccountID := account2.ID

		if i%2 == 1 {
			FromAccountID = account2.ID
			ToAccountID = account1.ID
		}

		go func() {
			_, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: FromAccountID,
				ToAccountID:   ToAccountID,
				Amount:        amount,
			})
			// 传回管道
			errs <- err
		}()
	}

	// 检查各协程的返回值
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
	}
	// 检查最终的账户余额
	updateAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	updateAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	require.Equal(t, account1.Balance, updateAccount1.Balance)
	require.Equal(t, account2.Balance, updateAccount2.Balance)
}
