package util
/* 
	bcrypt加密密码，存储在数据库
*/
import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// 将密码转为bcrypt密码
func HashPassword(password string) (string, error) {
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password:%w", err)
	}
	return string(hashPassword), nil
}

// 检查输入密码是否正确(比较两次密码的哈希值)
func CheckPassword(password, hashPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashPassword),[]byte(password))
}
