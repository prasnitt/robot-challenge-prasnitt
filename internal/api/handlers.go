package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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

// GetState handles the request to get the current state of the robot service.
// @Summary Get the current state of the robot service
// @Description Get the current state of the robot service including robot position, task count and tasks
// @Produce json
// @Success 200 {object} robot.ServiceState "Current state of the robot service"
// @Router /robot/state [get]
// @Tags Robot State
func GetState(service robot.RobotService) gin.HandlerFunc {
	return func(c *gin.Context) {
		state := service.CurrentState()
		c.JSON(http.StatusOK, state)
	}
}

// CancelTask handles the request to cancel a robot task by its ID.
// @Summary Cancel a robot task by ID
// @Description Cancel a robot task by its ID, if the task is in progress or pending
// @Param id path string true "Task ID"
// @Success 202 {object} map[string]string "Cancellation request accepted"
// @Failure 400 {object} ErrorResponse "Error message"
// @Router /robot/tasks/{id}/cancel [put]
// @Tags Robot Tasks
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

// WebSocket upgrader configuration
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin in development
		// TODO: In production, we should validate the origin properly
		return true
	},
}

// TaskStatusWebSocket handles WebSocket connections for real-time task status updates.
// @Summary WebSocket endpoint for real-time task status updates
// @Description Establishes a WebSocket connection to receive real-time task status updates. This endpoint requires a WebSocket client (not accessible via Swagger UI). Use tools like Postman, wscat, or the provided HTML test page.
// @Produce json
// @Success 101 {object} robot.TaskStatusUpdateEvent "WebSocket connection established, events will be sent as JSON"
// @Failure 400 {object} ErrorResponse "Failed to upgrade connection"
// @Router /robot/events [get]
// @Tags Robot Events
func TaskStatusWebSocket(service robot.RobotService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Upgrade HTTP connection to WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("Failed to upgrade connection: %v", err)
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Failed to upgrade to WebSocket"})
			return
		}
		defer conn.Close()
		log.Printf("WebSocket connection established from %s", c.ClientIP())

		// Get the event channel from the service
		eventChannel := service.GetEventChannel()

		// Listen for task status events and send them to the WebSocket client
		for {
			select {
			case event := <-eventChannel:
				// Send the event to the WebSocket client
				if err := conn.WriteJSON(event); err != nil {
					log.Printf("Failed to send event to WebSocket client: %v", err)
					return
				}
				log.Printf("Sent event to WebSocket client: task=%s, state=%s", event.TaskID, event.State)

			case <-c.Request.Context().Done():
				// Client disconnected
				log.Printf("WebSocket client disconnected: %s", c.ClientIP())
				return
			}
		}
	}
}
