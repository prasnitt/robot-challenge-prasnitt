package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prasnitt/robot-challenge-prasnitt/internal/robot"
)

// AddTaskRequest represents the request body for adding a new robot task.
// @Description Request body for adding a new robot task
type AddTaskRequest struct {
	Commands             string `json:"commands" binding:"required" example:"N E S W"`           // Commands to be executed by the robot
	DelayBetweenCommands string `json:"delay_between_commands" binding:"omitempty" example:"1s"` // Delay between executing commands, optional
}

// ErrorResponse represents a generic error response.
// @Description Generic error response.
type ErrorResponse struct {
	Error string `json:"error" example:"Job not found"`
}

// TODO: remove this function after testing
func HelloWorld(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello, World!",
	})
}

// AddTask handles the request to add a new robot task.
// @Summary Add a new robot task
// @Description Add a new robot task with commands and optional delay
// @Accept json
// @Produce json
// @Param request body AddTaskRequest true "Add Task Request"
// @Success 202 {object} map[string]string "Task ID"
// @Failure 400 {object} ErrorResponse "Error message"
// @Router /robot/tasks [post]
// @Tags Robot Tasks
func AddTask(service robot.RobotService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AddTaskRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}

		taskID, err := service.EnqueueTask(req.Commands, req.DelayBetweenCommands)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
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
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "task ID is required"})
			return
		}

		err := service.CancelTask(taskID)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
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
