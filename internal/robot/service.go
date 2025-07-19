package robot

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

const (
	// Default warehouse size, can be adjusted as needed
	warehouseSize = 10 // Size of the warehouse grid (10x10)

)

// RobotService defines the interface for the robot service.
type RobotService interface {
	EnqueueTask(commands string, delayBetweenCommands string) (taskID string, err error)

	CancelTask(taskID string) error

	CurrentState() ServiceState

	GetEventChannel() <-chan TaskStatusUpdateEvent
}

// Websocket response for task status updates.
// @Description Websocket response for task status updates.
type TaskStatusUpdateEvent struct {
	TaskID    string    `json:"task_id" example:"12345"`                         // Unique identifier for the task
	State     TaskState `json:"state" swaggertype:"string" example:"InProgress"` // Current state of the task
	Error     string    `json:"error,omitempty" example:""`                      // Error message if any
	Timestamp time.Time `json:"timestamp" example:"2024-01-15T10:30:00Z"`        // Timestamp when the event occurred
}

type Service struct {
	mu          sync.RWMutex               // Mutex for concurrent access
	ctx         context.Context            // Context for cancellation
	state       ServiceState               // Current state of the robot service
	taskIdQueue chan string                // Channel for incoming tasks
	eventChan   chan TaskStatusUpdateEvent // Channel for broadcasting task status updates
}

// NewService initializes a new robot service with an empty state and a task channel.
func NewService(ctx context.Context, taskIdQueue chan string) *Service {
	return &Service{
		ctx:         ctx,
		state:       NewServiceState(),                     // Initialize the service state
		taskIdQueue: taskIdQueue,                           // Buffered channel for tasks
		eventChan:   make(chan TaskStatusUpdateEvent, 100), // Buffered channel for events
	}
}

// Start begins processing tasks from the task queue.
func (s *Service) Start() {
	log.Println("Robot Service Started...")

	for {
		select {
		case <-s.ctx.Done():
			log.Println("Robot Service Stopping...")
			return // Exit if the context is cancelled
		case taskId := <-s.taskIdQueue:
			err := s.ExecuteTask(taskId) // Process incoming tasks
			if err != nil {
				log.Printf("Error handling task %s: %v", taskId, err)
			}
		}
	}
}

// GetState returns the current state of the robot service.
func (s *Service) CurrentState() ServiceState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

func (s *Service) EnqueueTask(commands string, delayBetweenCommands string) (string, error) {
	task, err := NewTask(commands, delayBetweenCommands)
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	// Update the service state with the new task

	s.state.CurTaskCount++                  // Increment the current task count
	task.SequenceNum = s.state.CurTaskCount // Assign a sequence number to the task
	s.state.Tasks[task.ID] = *task
	s.taskIdQueue <- task.ID // Send the task to the queue

	log.Printf("Task %s enqueued with commands: '%s', delay between commands: '%s' ", task.ID, commands, task.DelayBetweenCommands)

	// Publish event for new task creation
	go s.publishEvent(task.ID, task.State, "")

	return task.ID, nil
}

func (s *Service) CancelTask(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.state.Tasks[taskID]
	if !exists {
		return fmt.Errorf("task with ID %s not found", taskID)
	}

	switch task.State {
	case InProgress:
		// Update the task state to RequestCancellation
		log.Printf("Task %s is in progress, requesting cancellation", taskID)
		task.State = RequestCancellation
		s.state.Tasks[taskID] = task // Update the task in the state

		// Publish event for cancellation request
		go s.publishEvent(taskID, RequestCancellation, "")

	case Pending:
		// If the task is pending, we simply mark it as Canceled
		log.Printf("Task %s is pending, marking as Canceled", taskID)
		task.Error = "Pending Task cancelled by user"
		task.State = Canceled
		s.state.Tasks[taskID] = task // Update the task in the state

		// Publish event for immediate cancellation
		go s.publishEvent(taskID, Canceled, task.Error)

	default:
		return fmt.Errorf("task %s is '%s' state and cannot be cancelled", taskID, task.State)
	}

	return nil
}

func (s *Service) ExecuteTask(taskId string) error {
	s.mu.RLock()
	task, exists := s.state.Tasks[taskId]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("Task %s not found in service state", taskId)
	}

	// The task should be in pending state when it is handled
	if task.State != Pending {
		return fmt.Errorf("Task %s is not in Pending state, current state: %s", task.ID, task.State)
	}

	log.Println("Started task:", task.ID)
	s.UpdateTaskState(task.ID, InProgress)

	// Check if task can be processed, robot must not be crossing the warehouse boundaries
	if !s.IsTaskValid(task) {
		s.UpdateTaskState(task.ID, Aborted)
		s.UpdateTaskError(task.ID, "Task is invalid: out of warehouse boundaries, marking as Aborted")
		return fmt.Errorf("Task %s is invalid and cannot be processed", task.ID)
	}

	// Run the task processing logic here
	// Keep on updating the robot state based on the commands in the task
	log.Printf("Processing task %s with commands: %s", task.ID, task.Commands)
	for _, cmd := range task.Commands {

		// Make sure if the task is requested for cancellation, we stop processing
		state, err := s.GetTaskState(task.ID)
		if err != nil {
			return fmt.Errorf("Error getting task state for %s: %v", task.ID, err)
		}

		if state == RequestCancellation {
			log.Printf("Task %s has been requested for cancellation", task.ID)
			s.UpdateTaskError(task.ID, "Task cancellation requested by user")
			s.UpdateTaskState(task.ID, Canceled)
			return nil // Stop processing the task if cancellation is requested
		}

		// Simulate delay between commands
		time.Sleep(time.Duration(task.DelayBetweenCommands))

		// Execute each command in the task
		err = s.ExecuteRobotCommand(cmd)

		if err != nil {
			s.UpdateTaskError(task.ID, fmt.Sprintf("Error executing command '%s': %v", cmd, err))
			s.UpdateTaskState(task.ID, Aborted) // Update the task state to Aborted
			return fmt.Errorf("Error executing command '%s' for task %s: %v", cmd, task.ID, err)
		}

		robotState := s.GetRobotState() // Get the current robot state after executing the command
		log.Printf("Command '%s' Executed Robot moved to position: (%d, %d)", cmd, robotState.X, robotState.Y)
	}

	// Update the task state to Completed
	s.UpdateTaskState(task.ID, Completed)
	log.Printf("Task %s completed successfully", task.ID)

	return nil
}

// Execute a robot command and update the robot's position
func (s *Service) ExecuteRobotCommand(cmd RobotCommand) error {

	robotState := s.GetRobotState() // Get the current robot state
	switch cmd {
	case North:
		if robotState.Y >= warehouseSize {
			return fmt.Errorf("robot cannot move north, out of warehouse boundaries")
		}
		robotState.Y++
	case South:
		if robotState.Y == 0 {
			return fmt.Errorf("robot cannot move south, out of warehouse boundaries")
		}
		robotState.Y--
	case East:
		if robotState.X >= warehouseSize {
			return fmt.Errorf("robot cannot move east, out of warehouse boundaries")
		}
		robotState.X++
	case West:
		if robotState.X == 0 {
			return fmt.Errorf("robot cannot move west, out of warehouse boundaries")
		}
		robotState.X--
	}

	s.SetRobotState(robotState) // Update the robot state in the service
	return nil
}

func (s *Service) GetTaskState(taskID string) (TaskState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if task, exists := s.state.Tasks[taskID]; exists {
		return task.State, nil
	}
	return Invalid, fmt.Errorf("task with ID %s not found", taskID)
}

func (s *Service) UpdateTaskState(taskID string, state TaskState) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task, exists := s.state.Tasks[taskID]; exists {
		task.State = state
		s.state.Tasks[taskID] = task // Update the task in the state
		log.Printf("Task %s updated to state: %s", taskID, state)

		// Publish event for WebSocket clients
		go s.publishEvent(taskID, state, task.Error)
	} else {
		log.Printf("Task %s not found for state update", taskID)
	}
}

// update error in the task
func (s *Service) UpdateTaskError(taskID string, errMsg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if task, exists := s.state.Tasks[taskID]; exists {
		task.Error = errMsg          // Update the error message in the task
		s.state.Tasks[taskID] = task // Update the task in the state
		log.Printf("Task %s updated with error: %s", taskID, errMsg)

		// Publish event for WebSocket clients with error information
		go s.publishEvent(taskID, task.State, errMsg)
	} else {
		log.Printf("Task %s not found for error update", taskID)
	}
}

// Get current robot state from the service state
func (s *Service) GetRobotState() RobotState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.RobotState
}

func (s *Service) SetRobotState(state RobotState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.RobotState = state // Update the robot state in the service state
}

// Check if a task can be processed based on the robot's current position and warehouse boundaries.
// if the task is valid, it will return true, otherwise false.
func (s *Service) IsTaskValid(task RobotTask) bool {
	robotState := s.GetRobotState()

	destinationX := int(robotState.X) + task.DeltaX
	destinationY := int(robotState.Y) + task.DeltaY

	if destinationX < 0 || destinationX >= warehouseSize || destinationY < 0 || destinationY >= warehouseSize {
		log.Printf("Task %s is invalid: out of warehouse boundaries", task.ID)
		return false
	}

	return true
}

// GetEventChannel returns the channel for task status update events.
// This channel can be used by WebSocket handlers to listen for real-time updates.
func (s *Service) GetEventChannel() <-chan TaskStatusUpdateEvent {
	return s.eventChan
}

// publishEvent sends a task status update event to the event channel.
// This method is non-blocking and will drop events if the channel is full.
func (s *Service) publishEvent(taskID string, state TaskState, errorMsg string) {
	event := TaskStatusUpdateEvent{
		TaskID:    taskID,
		State:     state,
		Error:     errorMsg,
		Timestamp: time.Now(),
	}

	// Non-blocking send to avoid deadlocks
	select {
	case s.eventChan <- event:
		log.Printf("Published event for task %s: state=%s at %s", taskID, state, event.Timestamp.Format(time.RFC3339))
	default:
		log.Printf("Event channel full, dropped event for task %s", taskID)
	}
}
