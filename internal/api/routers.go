package api

import (
	"github.com/gin-gonic/gin"
	"github.com/prasnitt/robot-challenge-prasnitt/internal/robot"
)

func SetupRouter(router *gin.Engine, robot robot.RobotService) {

	v1 := router.Group("/api/v1")

	robotGroup := v1.Group("/robot")
	{
		// API endpoints for robot tasks
		robotGroup.POST("/tasks", AddTask(robot))
		robotGroup.PUT("/tasks/:id/cancel", CancelTask(robot))
		robotGroup.GET("/state", GetState(robot))
	}
}
