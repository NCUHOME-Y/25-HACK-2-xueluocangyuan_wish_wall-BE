package util

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type MyCustomClaims struct {
	UserID uint `json:"userId"`
	jwt.RegisteredClaims
}

func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		//防止崩溃整了个默认的
		return []byte("default_secret_key")
	}
	return []byte(secret)
}

func GenerateToken(userID uint) (string, error) {
	claims := MyCustomClaims{
		userID,
		jwt.RegisteredClaims{
			//过期时间设置为10天
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * 10 * time.Hour)),
			Issuer:    "wish_wall_app",
			//签发时间
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}
	//使用HS256算法生成token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	//用密钥签名，获取完整token字符串
	tokenString, err := token.SignedString(getJWTSecret())
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// ParseToken解析并验证一个token字符串
func ParseToken(tokenString string) (*MyCustomClaims, error) {
	//解析token
	token, err := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return getJWTSecret(), nil
	})

	if err != nil {
		return nil, err //可能是token过期，签名无效
	}
	//验证token是否有效，把claims转换为自定义的MyCustomClaims类型
	if claims, ok := token.Claims.(*MyCustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token claims")

}
