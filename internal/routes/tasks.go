package routes

import (
	"GoBlast/api/handlers"
	"github.com/gin-gonic/gin"
)

func SetupTaskRoutes(router *gin.RouterGroup, taskHandler *handlers.TaskHandler) {
	router.POST("/tasks", taskHandler.CreateTask)
	router.GET("/tasks/:id", taskHandler.GetTask)
}
