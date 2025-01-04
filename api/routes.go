package api

import (
	"GoBlast/api/handlers"
	"GoBlast/api/middleware"
	"GoBlast/internal/routes"
	"GoBlast/internal/tasks"
	"GoBlast/internal/users"
	"GoBlast/pkg/queue"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

func SetupRouter(database *gorm.DB, natsClient *queue.NATSClient) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Middleware
	router.Use(middleware.RequestLogger())
	router.Use(middleware.CORSMiddleware())

	// Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Base Routes
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// Repositories
	authRepo := users.NewAuthUserRepository(database)
	taskRepo := tasks.NewTasksRepository(database) // БЕЗ второго аргумента

	// Handlers
	authHandler := handlers.NewAuthHandler(authRepo)
	taskHandler := handlers.NewTaskHandler(taskRepo, natsClient) // ДВА аргумента

	api := router.Group("/api")
	{
		routes.SetupAuthRoutes(api, authHandler)
	}

	protected := router.Group("/api")
	protected.Use(middleware.JWTMiddleware())
	{
		routes.SetupTaskRoutes(protected, taskHandler)
	}

	return router
}
