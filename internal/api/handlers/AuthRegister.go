// api/handlers/AuthRegister.go

package handlers

import (
	"GoBlast/internal/api/middleware"
	"encoding/base64"
	"log"
	"net/http"

	"GoBlast/internal/users"
	"GoBlast/pkg/encryption"
	"GoBlast/pkg/response"
	"GoBlast/pkg/storage/models"

	"github.com/gin-gonic/gin"
)

// RegisterInput описывает входные данные для регистрации пользователя
type RegisterInput struct {
	Username string `json:"username" binding:"required"`
	Token    string `json:"token" binding:"required"`
}

type AuthHandler struct {
	repo *users.AuthUserRepository
}

func NewAuthHandler(repo *users.AuthUserRepository) *AuthHandler {
	return &AuthHandler{repo: repo}
}

// RegisterHandler @Summary Зарегистрировать пользователя
// @Description Создаёт новый аккаунт пользователя с уникальным именем и Telegram Bot Token.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param register body RegisterInput true "User registration data"
// @Success 201 {object} response.APIResponse "Пользователь успешно зарегистрирован"
// @Failure 400 {object} response.APIResponse "Некорректные входные данные"
// @Failure 409 {object} response.APIResponse "Имя пользователя уже существует"
// @Failure 500 {object} response.APIResponse "Внутренняя ошибка сервера"
// @Router /auth/register [post]
// @example Request:
//
//	{
//	  "username": "sokrat",
//	  "token": "your_telegram_bot_token_here"
//	}
//
// @example Response (201):
//
//	{
//	  "success": true,
//	  "data": "Пользователь успешно зарегистрирован"
//	}
func (h *AuthHandler) RegisterHandler(c *gin.Context) {
	var input RegisterInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid input"))
		return
	}

	// Проверка, существует ли уже пользователь
	_, err := h.repo.FindByUsername(input.Username)
	if err == nil {
		c.JSON(http.StatusConflict, response.ErrorResponse("Username already exists"))
		return
	}

	// Шифрование токена
	encryptedToken, err := encryption.Encrypt([]byte(input.Token), []byte(middleware.EncryptionKey))
	if err != nil {
		log.Printf("Error encrypting token for user %s: %v", input.Username, err)
		c.JSON(http.StatusInternalServerError, response.ErrorResponse("Failed to encrypt token"))
		return
	}

	// Кодирование зашифрованного токена в base64 строку
	encodedToken := base64.StdEncoding.EncodeToString(encryptedToken)

	// Создание нового пользователя
	newUser := &models.AuthUser{
		Username: input.Username,
		Token:    encodedToken, // Хранение зашифрованного токена как строки
	}

	// Сохранение пользователя в БД
	if err := h.repo.Create(newUser); err != nil {
		log.Printf("Error saving user %s: %v", input.Username, err)
		c.JSON(http.StatusInternalServerError, response.ErrorResponse("Failed to register user"))
		return
	}

	c.JSON(http.StatusCreated, response.SuccessResponse("User registered successfully"))
}
