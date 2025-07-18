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

func GetState(service robot.RobotService) gin.HandlerFunc {
	return func(c *gin.Context) {
		state := service.CurrentState()
		c.JSON(http.StatusOK, state)
	}
}

func CancelTask(service robot.RobotService) gin.HandlerFunc {
	return func(c *gin.Context) {
		taskID := c.Param("id")
		if taskID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "task ID is required"})
			return
		}

		err := service.CancelTask(taskID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{"message": "Task cancellation requested successfully"})
	}
}

// TODO: Add unit tests for the API handlers
//  Case 1: Call state endpoint and check if the state is returned correctly wit initial values
//  Case 2: Call add task endpoint with valid commands and check if the task is added
//  Case 3: Call add task endpoint with invalid commands and check if the error is returned
//  Case 4: Call add task endpoint with empty commands and check if the error is returned
//  Case 5: Call add task endpoint with valid commands and check if the task ID is returned
