package routes

import (
	"GoBlast/internal/api/handlers"

	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(router *gin.RouterGroup, authHandler *handlers.AuthHandler) {
	authRoutes := router.Group("/auth")
	{
		authRoutes.POST("/login", authHandler.LoginHandler)
		authRoutes.POST("/register", authHandler.RegisterHandler)
	}
}
