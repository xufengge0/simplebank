package util

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestPassword(t *testing.T){
	password := RandomString(6)

	hashpassword1,err := HashPassword(password)
	require.NoError(t,err)
	require.NotEmpty(t,hashpassword1)

	// 测试正确密码
	err = CheckPassword(password,hashpassword1)
	require.NoError(t,err)

	// 测试错误密码
	wrongPassword := RandomString(6)
	err = CheckPassword(wrongPassword,hashpassword1)
	require.EqualError(t,err,bcrypt.ErrMismatchedHashAndPassword.Error())

	// 测试相同密码两次哈希结果不同
	hashpassword2,err := HashPassword(password)
	require.NoError(t,err)
	require.NotEmpty(t,hashpassword2)
	require.NotEqual(t,hashpassword1,hashpassword2)

}