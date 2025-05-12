package jwts

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
)

// JWT载荷结构体（Claims）
type CustomClaims struct {
	Uid string `json:"uid"`
	jwt.RegisteredClaims
}

// GenToken 生成一个带签名的JWT字符串
func GenToken(claims *CustomClaims, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims) //使用HMAC-SHA256作为签名算法创建一个新的JWT
	return token.SignedString([]byte(secret))
}

// ParseToken 解析和验证JWT字符串
func ParseToken(token, secret string) (string, error) {
	t, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok { //如果不是HMAC类型的签名方法，拒绝解析
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := t.Claims.(jwt.MapClaims); ok && t.Valid { //检查Token是否有效
		return fmt.Sprintf("%v", claims["uid"]), nil
	}
	return "", errors.New("token not valid")
}
