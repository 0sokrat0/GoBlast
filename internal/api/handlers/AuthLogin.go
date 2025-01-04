// api/handlers/AuthLogin.go

package handlers

import (
	"GoBlast/internal/api/middleware"
	"encoding/base64"
	"log"
	"net/http"

	"GoBlast/pkg/encryption"
	"GoBlast/pkg/response"
	"github.com/gin-gonic/gin"
)

// LoginInput описывает входные данные для входа пользователя
type LoginInput struct {
	Username string `json:"username" binding:"required"`
	Token    string `json:"token" binding:"required"`
}

// LoginHandler @Summary User login
// @Description Authenticates the user and returns a JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param input body LoginInput true "User credentials (username and token)"
// @Success 200 {object} response.APIResponse{data=map[string]string} "JWT token"
// @Failure 400 {object} response.APIResponse "Invalid input"
// @Failure 401 {object} response.APIResponse "Invalid credentials"
// @Failure 500 {object} response.APIResponse "Failed to generate token"
// @Router /auth/login [post]
// @example Request:
//
//	{
//	  "username": "sokrat",
//	  "token": "your_telegram_bot_token_here"
//	}
//
// @example Response (200):
//
//	{
//	  "success": true,
//	  "data": {
//	    "token": "your_jwt_token_here"
//	  }
//	}
func (h *AuthHandler) LoginHandler(c *gin.Context) {
	var input LoginInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid input"))
		return
	}

	user, err := h.repo.FindByUsername(input.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse("User not found"))
		return
	}

	// Декодирование base64 строки обратно в []byte
	encryptedToken, err := base64.StdEncoding.DecodeString(user.Token)
	if err != nil {
		log.Printf("Error decoding token for user %s: %v", input.Username, err)
		c.JSON(http.StatusInternalServerError, response.ErrorResponse("Failed to decode token"))
		return
	}

	// Дешифрование токена
	decryptedTokenBytes, err := encryption.Decrypt(encryptedToken, []byte(middleware.EncryptionKey))
	if err != nil {
		log.Printf("Error decrypting token for user %s: %v", input.Username, err)
		c.JSON(http.StatusInternalServerError, response.ErrorResponse("Failed to decrypt token"))
		return
	}
	decryptedToken := string(decryptedTokenBytes)

	// Сравнение введённого токена с хранимым
	if decryptedToken != input.Token {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse("Invalid token"))
		return
	}

	// Генерация JWT-токена
	jwtToken, err := middleware.GenerateToken(user.ID)
	if err != nil {
		log.Printf("Error generating JWT token for user %s: %v", input.Username, err)
		c.JSON(http.StatusInternalServerError, response.ErrorResponse("Failed to generate token"))
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse(map[string]string{
		"token": jwtToken,
	}))
}
