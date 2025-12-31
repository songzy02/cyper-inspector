package auth

import (
	"time"

	"cyber-inspector/internal/config"
	"cyber-inspector/internal/model"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   uint64         `json:"user_id"`
	Username string         `json:"username"`
	Role     model.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 生成 JWT
func GenerateToken(user *model.User) (string, error) {
	now := time.Now()
	expire := now.Add(time.Duration(config.Conf.JWT.ExpireHours) * time.Hour)

	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expire),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "cyber-inspector",
		},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(config.Conf.JWT.Secret)) // 这里直接读配置
}

// ParseToken 解析 JWT
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.Conf.JWT.Secret), nil // 这里直接读配置
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}
