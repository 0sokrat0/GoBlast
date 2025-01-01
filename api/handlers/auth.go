package handlers

import (
	"GoBlast/api/middleware"
	"GoBlast/internal/users"
	"GoBlast/pkg/response"
	"GoBlast/pkg/storage/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// LoginInput описывает входные данные для входа пользователя
type LoginInput struct {
	Username string `json:"username" binding:"required"`
	Token    string `json:"token" binding:"required"`
}

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

// LoginHandler handles user login
// @Summary User login
// @Description Authenticates the user and returns a JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param input body LoginInput true "User credentials (username and token)"
// @Success 200 {object} map[string]interface{} "JWT token"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 401 {object} map[string]interface{} "Invalid credentials"
// @Router /auth/login [post]
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

	if !middleware.CheckPassword(user.Token, input.Token) {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse("Invalid password"))
		return
	}

	token, err := middleware.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse("Failed to generate token"))
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse(map[string]interface{}{
		"token": token,
	}))
}

// RegisterHandler handles user registration
// @Summary Register a new user
// @Description Creates a new user account
// @Tags Authentication
// @Accept json
// @Produce json
// @Param input body RegisterInput true "User registration data (username and token)"
// @Success 201 {string} string "User registered successfully"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 409 {object} map[string]interface{} "Username already exists"
// @Failure 500 {object} map[string]interface{} "Failed to register user"
// @Router /auth/register [post]
func (h *AuthHandler) RegisterHandler(c *gin.Context) {
	var input RegisterInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid input"))
		return
	}

	_, err := h.repo.FindByUsername(input.Username)
	if err == nil {
		c.JSON(http.StatusConflict, response.ErrorResponse("Username already exists"))
		return
	}

	hashedToken, err := middleware.HashPassword(input.Token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse("Failed to hash token"))
		return
	}

	newUser := &models.AuthUser{
		Username: input.Username,
		Token:    hashedToken,
	}

	if err := h.repo.Create(newUser); err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse("Failed to register user"))
		return
	}

	c.JSON(http.StatusCreated, response.SuccessResponse("User registered successfully"))
}
