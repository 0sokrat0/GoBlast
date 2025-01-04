// pkg/middleware/middleware.go

package middleware

import (
	"GoBlast/configs"
	"GoBlast/pkg/response"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

var (
	JWTSecret     string
	EncryptionKey string
)

// Initialize инициализирует ключи из конфигурации
func Initialize(cfg *configs.Config) {
	JWTSecret = cfg.App.JWTSecret
	if JWTSecret == "" {
		panic("JWTSecret must be set")
	}

	EncryptionKey = cfg.Encricrypted.EncryptionKey

}

// Claims представляет структуру JWT-токена
type Claims struct {
	UserID uint `json:"user_id"`
	jwt.StandardClaims
}

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем заголовок Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, response.ErrorResponse("Authorization header required"))
			c.Abort()
			return
		}

		// Извлекаем токен из заголовка
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Парсим токен
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Проверяем метод подписи
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, response.ErrorResponse("Invalid token"))
			c.Abort()
			return
		}

		// Устанавливаем claims в контекст
		c.Set("claims", claims)
		c.Next()
	}
}

// GenerateToken генерирует JWT-токен для заданного userID
func GenerateToken(userID uint) (string, error) {
	claims := &Claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(), // Токен истекает через 24 часа
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(JWTSecret))
}

// ValidateToken валидирует JWT-токен и возвращает claims
func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
