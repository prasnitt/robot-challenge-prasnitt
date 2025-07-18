package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prasnitt/robot-challenge-prasnitt/internal/robot"
)

type AddTaskRequest struct {
	Commands string `json:"commands" binding:"required"`
}

func HelloWorld(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello, World!",
	})
}

func AddTask(service robot.RobotService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AddTaskRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{

				"error": err.Error()})
			return
		}

		taskID, err := service.EnqueueTask(req.Commands)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{"task_id": taskID})
	}
}
