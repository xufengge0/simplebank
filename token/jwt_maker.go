package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const minSecretKeySize = 32

type JWTMaker struct {
	secretKey string 
}
// 创建一个新的JWTMaker, 并返回一个Maker接口
func NewJWTMaker(secretKey string) (Maker, error) {
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeySize)
	}
	return &JWTMaker{secretKey}, nil
}
// 创建token, 并返回token字符串
func (m *JWTMaker) CreateToken(username string, duration time.Duration) (string,*Payload, error) {
	// 创建一个新的payload
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "",payload, err
	}

	// 创建一个新的JWT
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload) // 签名方法为HS256

	// 使用secretKey进行签名
	token,err:= jwtToken.SignedString([]byte(m.secretKey))
	return token,payload,err
}

// 验证token是否有效, 并返回payload
func (m *JWTMaker) VerifyToken(token string) (*Payload, error) {

	// 接收解析后的JWT token作为参数，并根据需要返回正确的签名密钥。
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		// 检查签名方法是否为正确
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, errInvalidToken
		}
		return []byte(m.secretKey), nil
	}

	// 解析JWT字符串并将其声明（Claims）转换为指定的结构体Payload
	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)
	if err != nil {
		// 对err断言
		verr, ok := err.(*jwt.ValidationError)
		// 判断err的类型并返回
		if ok && errors.Is(verr.Inner, errExpiredToken) {
			return nil, errExpiredToken
		}
		return nil, errInvalidToken
	}

	// 将jwtToken的声明断言为Payload结构体
	payload, ok := jwtToken.Claims.(*Payload)
	if !ok {
		return nil, errInvalidToken
	}
	return payload, nil
}
