package handler

func main() {

}

//// file: internal/handler/jwt.go
//package handler
//
//import (
//	"time"
//
//	"cyber-inspector/internal/config"
//	"cyber-inspector/internal/model"
//
//	"github.com/golang-jwt/jwt/v4"
//)
//
//// 自定义 claims
//type Claims struct {
//	UserID uint64         `json:"user_id"`
//	Role   model.UserRole `json:"role"`
//	jwt.RegisteredClaims
//}
//
//// GenerateToken 根据用户信息生成 JWT
//func GenerateToken(user *model.User) (string, error) {
//	now := time.Now()
//	expire := now.Add(time.Duration(config.Conf.JWT.ExpireHours) * time.Hour)
//
//	claims := Claims{
//		UserID: user.ID,
//		Role:   user.Role,
//		RegisteredClaims: jwt.RegisteredClaims{
//			ExpiresAt: jwt.NewNumericDate(expire),
//			IssuedAt:  jwt.NewNumericDate(now),
//			NotBefore: jwt.NewNumericDate(now),
//			Issuer:    "cyber-inspector",
//		},
//	}
//
//	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
//	return token.SignedString([]byte(config.Conf.JWT.Secret))
//}
