package api

import (
	"github.com/gin-gonic/gin"
	"github.com/prasnitt/robot-challenge-prasnitt/internal/robot"
)

func SetupRouter(router *gin.Engine, robot robot.RobotService) {

	apiGroup := router.Group("/api/v1")
	{
		// Experimental endpoint for testing
		apiGroup.GET("/hello", HelloWorld)

		// API endpoints for robot tasks
		apiGroup.POST("/tasks", AddTask(robot))
		apiGroup.PUT("/tasks/:id/cancel", CancelTask(robot))
		apiGroup.GET("/state", GetState(robot))
	}
}
