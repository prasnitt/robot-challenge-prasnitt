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
}

type Service struct {
	mu        sync.RWMutex    // Mutex for concurrent access
	ctx       context.Context // Context for cancellation
	state     ServiceState    // Current state of the robot service
	taskQueue chan RobotTask  // Channel for incoming tasks
}

// NewService initializes a new robot service with an empty state and a task channel.
func NewService(ctx context.Context, taskQueue chan RobotTask) *Service {
	return &Service{
		ctx:       ctx,
		state:     NewServiceState(), // Initialize the service state
		taskQueue: taskQueue,         // Buffered channel for tasks
	}
}

// Start begins processing tasks from the task queue.
func (s *Service) Start() {
	log.Println("Robot Service Started...")

	for {
		select {
		case <-s.ctx.Done():
			log.Println("Robot Service Stopping...")
		case task := <-s.taskQueue:
			s.HandleTask(task) // Process incoming tasks
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
	s.taskQueue <- *task // Send the task to the queue

	// Update the service state with the new task
	s.state.Tasks[task.ID] = *task

	log.Printf("Task %s enqueued with commands: '%s' delay between commands: '%s' ", task.ID, commands, task.DelayBetweenCommands)

	return task.ID, nil
}

func (s *Service) CancelTask(taskID string) error {
	taskState, err := s.GetTaskState(taskID)
	if err != nil {
		return err
	}

	if taskState != InProgress {
		return fmt.Errorf("task %s is '%s' state and cannot be cancelled", taskID, taskState)
	}

	s.UpdateTaskState(taskID, RequestCancellation) // Update the task state to RequestCancellation

	return nil // Placeholder for task cancellation logic
}

func (s *Service) HandleTask(task RobotTask) {
	log.Println("Started task:", task.ID)
	s.UpdateTaskState(task.ID, InProgress)

	// Check if task can be processed, robot must not be crossing the warehouse boundaries
	if !s.IsTaskValid(task) {
		s.UpdateTaskState(task.ID, Aborted)
		log.Printf("Task %s is invalid and cannot be processed", task.ID)
		return
	}

	// Run the task processing logic here
	// Keep on updating the robot state based on the commands in the task
	log.Printf("Processing task %s with commands: %s", task.ID, task.Commands)
	for _, cmd := range task.Commands {

		// Make sure if the task is requested for cancellation, we stop processing
		state, err := s.GetTaskState(task.ID)
		if err != nil {
			log.Printf("Error getting task state for %s: %v", task.ID, err)
			return // Stop processing if we can't get the task state
		}

		if state == RequestCancellation {
			log.Printf("Task %s has been requested for cancellation", task.ID)
			s.UpdateTaskState(task.ID, Canceled)
			return // Stop processing the task
		}

		time.Sleep(time.Duration(task.DelayBetweenCommands)) // Simulate delay between commands

		// Execute each command in the task
		err = s.ExecuteRobotCommand(cmd, task.DelayBetweenCommands) // Execute each command in the task

		if err != nil {
			log.Printf("Error executing command '%s' for task %s: %v", cmd, task.ID, err)
			s.UpdateTaskState(task.ID, Aborted) // Update the task state to Aborted
			return                              // Stop processing the task on error
		}

		robotState := s.GetRobotState() // Get the current robot state after executing the command
		log.Printf("Command '%s' Executed Robot moved to position: (%d, %d)", cmd, robotState.X, robotState.Y)
	}

	// Update the task state to Completed
	s.UpdateTaskState(task.ID, Completed)
	log.Printf("Task %s completed successfully", task.ID)
}

// Execute a robot command and update the robot's position
func (s *Service) ExecuteRobotCommand(cmd RobotCommand, durationBetweenCmds CommandDuration) error {

	robotState := s.GetRobotState() // Get the current robot state
	switch cmd {
	case North:
		if robotState.Y >= warehouseSize {
			return fmt.Errorf("robot cannot move north, out of warehouse boundaries")
		}
		robotState.Y++
	case South:
		if robotState.Y <= 0 {
			return fmt.Errorf("robot cannot move south, out of warehouse boundaries")
		}
		robotState.Y--
	case East:
		if robotState.X >= warehouseSize {
			return fmt.Errorf("robot cannot move east, out of warehouse boundaries")
		}
		robotState.X++
	case West:
		if robotState.X <= 0 {
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
	} else {
		log.Printf("Task %s not found for state update", taskID)
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

	destinationX := (int)(robotState.X) + task.DeltaX
	destinationY := (int)(robotState.Y) + task.DeltaY

	if destinationX < 0 || destinationX >= warehouseSize || destinationY < 0 || destinationY >= warehouseSize {
		log.Printf("Task %s is invalid: out of warehouse boundaries", task.ID)
		return false
	}

	return true
}
