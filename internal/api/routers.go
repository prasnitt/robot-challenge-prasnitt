package api

import (
	"github.com/gin-gonic/gin"
	"github.com/prasnitt/robot-challenge-prasnitt/internal/robot"
)

func SetupRouter(router *gin.Engine, robotService robot.RobotService) {

	v1 := router.Group("/api/v1")

	robotGroup := v1.Group("/robot")
	{
		// API endpoints for robot tasks
		robotGroup.POST("/tasks", AddTask(robotService))
		robotGroup.PUT("/tasks/:id/cancel", CancelTask(robotService))
		robotGroup.GET("/state", GetState(robotService))

		// WebSocket endpoint for real-time task status updates
		robotGroup.GET("/events", TaskStatusWebSocket(robotService))
	}
}
