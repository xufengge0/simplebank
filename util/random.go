package util

import (
	"math/rand"
	"strings"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

func init() {
	rand.Seed(time.Now().UnixNano())
}

// 随机生成一个位于范围[min,max]内的整数
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

// 随机生成一个由n个字符组成的字符串
func RandomString(n int) string {
	var sb strings.Builder // 频繁进行+号字符串拼接时 优化性能
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)] // 随机生成一个字符
		sb.WriteByte(c)             // 添加字符
	}
	return sb.String()
}

// 随机生成owner
func RandomOwner() string {
	return RandomString(6)
}

// 随机生成blance
func RandomBlance() int64 {
	return RandomInt(0, 1000)
}

// 随机生成currency
func RandomCurrency() string {
	currencies := []string{"USD", "EUR", "CAD"}
	k := len(currencies)
	return currencies[rand.Intn(k)]
}
