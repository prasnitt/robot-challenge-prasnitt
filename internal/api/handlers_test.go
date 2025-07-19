package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prasnitt/robot-challenge-prasnitt/internal/robot"
)

// MockRobotService implements the RobotService interface for testing
type MockRobotService struct {
	state             robot.ServiceState
	enqueuedTasks     []mockTask
	enqueueError      error
	cancelError       error
	shouldFailEnqueue bool
	shouldFailCancel  bool
}

type mockTask struct {
	commands             string
	delayBetweenCommands string
	taskID               string
}

func NewMockRobotService() *MockRobotService {
	return &MockRobotService{
		state: robot.ServiceState{
			RobotState:   robot.RobotState{X: 0, Y: 0},
			Tasks:        make(map[string]robot.RobotTask),
			CurTaskCount: 0,
		},
		enqueuedTasks: make([]mockTask, 0),
	}
}

func (m *MockRobotService) EnqueueTask(commands string, delayBetweenCommands string) (string, error) {
	if m.shouldFailEnqueue {
		return "", m.enqueueError
	}

	taskID := "test-task-id-123"
	m.enqueuedTasks = append(m.enqueuedTasks, mockTask{
		commands:             commands,
		delayBetweenCommands: delayBetweenCommands,
		taskID:               taskID,
	})

	// Update state to reflect the new task
	m.state.CurTaskCount++
	task := robot.RobotTask{
		ID:          taskID,
		SequenceNum: m.state.CurTaskCount,
		State:       robot.Pending,
		Error:       "",
	}
	m.state.Tasks[taskID] = task

	return taskID, nil
}

func (m *MockRobotService) CancelTask(taskID string) error {
	if m.shouldFailCancel {
		return m.cancelError
	}
	return nil
}

func (m *MockRobotService) CurrentState() robot.ServiceState {
	return m.state
}

// Helper function to set up Gin router for testing
func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

// Test Case 1: Call state endpoint and check if the state is returned correctly with initial values
func TestGetState_InitialValues(t *testing.T) {
	mockService := NewMockRobotService()
	router := setupRouter()

	router.GET("/robot/state", GetState(mockService))

	req, _ := http.NewRequest("GET", "/robot/state", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Check content type
	expectedContentType := "application/json; charset=utf-8"
	if w.Header().Get("Content-Type") != expectedContentType {
		t.Errorf("Expected content type %s, got %s", expectedContentType, w.Header().Get("Content-Type"))
	}

	// Parse response body
	var state robot.ServiceState
	err := json.Unmarshal(w.Body.Bytes(), &state)
	if err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	// Verify initial values
	if state.RobotState.X != 0 {
		t.Errorf("Expected robot X position to be 0, got %d", state.RobotState.X)
	}
	if state.RobotState.Y != 0 {
		t.Errorf("Expected robot Y position to be 0, got %d", state.RobotState.Y)
	}
	if state.CurTaskCount != 0 {
		t.Errorf("Expected task count to be 0, got %d", state.CurTaskCount)
	}
	if len(state.Tasks) != 0 {
		t.Errorf("Expected empty tasks map, got %d tasks", len(state.Tasks))
	}
}

// Test Case 2: Call add task endpoint with valid commands and check if the task is added
func TestAddTask_ValidCommands(t *testing.T) {
	mockService := NewMockRobotService()
	router := setupRouter()

	router.POST("/robot/tasks", AddTask(mockService))

	// Create request body
	requestBody := AddTaskRequest{
		Commands:             "N E S W",
		DelayBetweenCommands: "1s",
	}
	jsonBody, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/robot/tasks", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusAccepted {
		t.Errorf("Expected status code %d, got %d", http.StatusAccepted, w.Code)
	}

	// Parse response body
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	// Check if task ID is returned
	taskID, exists := response["task_id"]
	if !exists {
		t.Error("Expected task_id in response")
	}
	if taskID == "" {
		t.Error("Expected non-empty task ID")
	}

	// Verify task was enqueued
	if len(mockService.enqueuedTasks) != 1 {
		t.Errorf("Expected 1 enqueued task, got %d", len(mockService.enqueuedTasks))
	} else {
		enqueuedTask := mockService.enqueuedTasks[0]
		if enqueuedTask.commands != "N E S W" {
			t.Errorf("Expected commands 'N E S W', got '%s'", enqueuedTask.commands)
		}
		if enqueuedTask.delayBetweenCommands != "1s" {
			t.Errorf("Expected delay '1s', got '%s'", enqueuedTask.delayBetweenCommands)
		}
	}
}

// Test Case 3: Call add task endpoint with invalid commands and check if the error is returned
func TestAddTask_InvalidCommands(t *testing.T) {
	mockService := NewMockRobotService()
	mockService.shouldFailEnqueue = true
	mockService.enqueueError = fmt.Errorf("invalid command: X")

	router := setupRouter()
	router.POST("/robot/tasks", AddTask(mockService))

	// Create request body with invalid commands
	requestBody := AddTaskRequest{
		Commands:             "N X E",
		DelayBetweenCommands: "1s",
	}
	jsonBody, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/robot/tasks", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	// Parse response body
	var errorResponse ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	if err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	// Check if error message is returned
	if errorResponse.Error == "" {
		t.Error("Expected error message in response")
	}
	if !strings.Contains(errorResponse.Error, "invalid command") {
		t.Errorf("Expected error to contain 'invalid command', got '%s'", errorResponse.Error)
	}
}

// Test Case 4: Call add task endpoint with empty commands and check if the error is returned
func TestAddTask_EmptyCommands(t *testing.T) {
	mockService := NewMockRobotService()
	router := setupRouter()

	router.POST("/robot/tasks", AddTask(mockService))

	// Create request body with empty commands
	requestBody := AddTaskRequest{
		Commands:             "",
		DelayBetweenCommands: "1s",
	}
	jsonBody, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/robot/tasks", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	// Parse response body
	var errorResponse ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	if err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	// Check if error message is returned
	if errorResponse.Error == "" {
		t.Error("Expected error message in response")
	}
}

// Test Case 5: Call add task endpoint with valid commands and check if the task ID is returned
func TestAddTask_ValidCommands_TaskIDReturned(t *testing.T) {
	mockService := NewMockRobotService()
	router := setupRouter()

	router.POST("/robot/tasks", AddTask(mockService))

	// Create request body
	requestBody := AddTaskRequest{
		Commands:             "N N E E",
		DelayBetweenCommands: "500ms",
	}
	jsonBody, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/robot/tasks", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusAccepted {
		t.Errorf("Expected status code %d, got %d", http.StatusAccepted, w.Code)
	}

	// Parse response body
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	// Check if task ID is returned and valid
	taskID, exists := response["task_id"]
	if !exists {
		t.Error("Expected task_id in response")
	}
	if taskID == "" {
		t.Error("Expected non-empty task ID")
	}
	if taskID != "test-task-id-123" {
		t.Errorf("Expected task ID 'test-task-id-123', got '%s'", taskID)
	}
}

// Test CancelTask endpoint with valid task ID
func TestCancelTask_ValidTaskID(t *testing.T) {
	mockService := NewMockRobotService()
	router := setupRouter()

	router.PUT("/robot/tasks/:id/cancel", CancelTask(mockService))

	req, _ := http.NewRequest("PUT", "/robot/tasks/test-task-123/cancel", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusAccepted {
		t.Errorf("Expected status code %d, got %d", http.StatusAccepted, w.Code)
	}

	// Parse response body
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	// Check if success message is returned
	message, exists := response["message"]
	if !exists {
		t.Error("Expected message in response")
	}
	if !strings.Contains(message, "successfully") {
		t.Errorf("Expected success message, got '%s'", message)
	}
}

// Test CancelTask endpoint with empty task ID
func TestCancelTask_EmptyTaskID(t *testing.T) {
	mockService := NewMockRobotService()
	router := setupRouter()

	router.PUT("/robot/tasks/:id/cancel", CancelTask(mockService))

	req, _ := http.NewRequest("PUT", "/robot/tasks//cancel", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	// Parse response body
	var errorResponse ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	if err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	// Check if error message is returned
	if errorResponse.Error == "" {
		t.Error("Expected error message in response")
	}
	if !strings.Contains(errorResponse.Error, "required") {
		t.Errorf("Expected error to mention 'required', got '%s'", errorResponse.Error)
	}
}

// Test CancelTask endpoint with service error
func TestCancelTask_ServiceError(t *testing.T) {
	mockService := NewMockRobotService()
	mockService.shouldFailCancel = true
	mockService.cancelError = fmt.Errorf("task not found")

	router := setupRouter()
	router.PUT("/robot/tasks/:id/cancel", CancelTask(mockService))

	req, _ := http.NewRequest("PUT", "/robot/tasks/non-existent-task/cancel", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	// Parse response body
	var errorResponse ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	if err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	// Check if error message is returned
	if errorResponse.Error == "" {
		t.Error("Expected error message in response")
	}
	if !strings.Contains(errorResponse.Error, "task not found") {
		t.Errorf("Expected error to contain 'task not found', got '%s'", errorResponse.Error)
	}
}

// Test AddTask with missing required fields
func TestAddTask_MissingRequiredField(t *testing.T) {
	mockService := NewMockRobotService()
	router := setupRouter()

	router.POST("/robot/tasks", AddTask(mockService))

	// Create request body without required commands field
	requestBody := map[string]string{
		"delay_between_commands": "1s",
	}
	jsonBody, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/robot/tasks", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	// Parse response body
	var errorResponse ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	if err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	// Check if error message is returned
	if errorResponse.Error == "" {
		t.Error("Expected error message in response")
	}
}

// Test AddTask with invalid JSON
func TestAddTask_InvalidJSON(t *testing.T) {
	mockService := NewMockRobotService()
	router := setupRouter()

	router.POST("/robot/tasks", AddTask(mockService))

	// Create invalid JSON
	invalidJSON := `Invalid JSON`

	req, _ := http.NewRequest("POST", "/robot/tasks", strings.NewReader(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	// Parse response body
	var errorResponse ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	if err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	// Check if error message is returned
	if errorResponse.Error == "" {
		t.Error("Expected error message in response")
	}
}
