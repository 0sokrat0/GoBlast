package api

import (
	handlers2 "GoBlast/internal/api/handlers"
	middleware2 "GoBlast/internal/api/middleware"
	"GoBlast/internal/routes"
	"GoBlast/internal/tasks"
	"GoBlast/internal/users"
	"GoBlast/pkg/metrics"
	"GoBlast/pkg/queue"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

func SetupRouter(database *gorm.DB, natsClient *queue.NATSClient) *gin.Engine {
	metrics.InitMetrics()
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.GET("/metrics", gin.WrapH(metrics.MetricsHandler()))

	// Middleware
	router.Use(middleware2.RequestLogger())
	router.Use(middleware2.CORSMiddleware())

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
	taskRepo := tasks.NewTasksRepository(database)

	// Handlers
	authHandler := handlers2.NewAuthHandler(authRepo)
	taskHandler := handlers2.NewTaskHandler(taskRepo, natsClient)

	api := router.Group("/api")
	{
		routes.SetupAuthRoutes(api, authHandler)
	}

	protected := router.Group("/api")
	protected.Use(middleware2.JWTMiddleware())
	{
		routes.SetupTaskRoutes(protected, taskHandler)
	}

	return router
}
