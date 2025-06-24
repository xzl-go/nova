package auth

import (
	"time"

	"github.com/xzl-go/nova/config"
	"github.com/xzl-go/nova/logger"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 自定义JWT声明
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT令牌
func GenerateToken(userID uint, username string) (string, error) {
	conf := config.Get()

	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(*conf.JWT.ExpireTime) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(conf.JWT.Secret)
}

// ParseToken 解析JWT令牌
func ParseToken(tokenString string) (*Claims, error) {
	conf := config.Get()

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return conf.JWT.Secret, nil
	})

	if err != nil {
		logger.Error("Failed to parse token", logger.Field("error", err))
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

// ValidateToken 验证JWT令牌
func ValidateToken(tokenString string) bool {
	_, err := ParseToken(tokenString)
	return err == nil
}

// GetUserID 从令牌中获取用户ID
func GetUserID(tokenString string) (uint, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}

// GetUsername 从令牌中获取用户名
func GetUsername(tokenString string) (string, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.Username, nil
}
