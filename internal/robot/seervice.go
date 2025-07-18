package robot

import (
	"context"
	"log"
)

const (
	// Default warehouse size, can be adjusted as needed
	warehouseSize = 10 // Size of the warehouse grid (10x10)
)

// RobotService defines the interface for the robot service.
type RobotService interface {
	EnqueueTask(commands string) (taskID string, err error)

	CancelTask(taskID string) error

	CurrentState() ServiceState
}

type Service struct {
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
	return s.state
}

func (s *Service) EnqueueTask(commands string) (string, error) {
	task, err := NewTask(commands)
	if err != nil {
		return "", err
	}

	s.state.Mu.Lock()
	defer s.state.Mu.Unlock()
	s.taskQueue <- *task // Send the task to the queue

	// Update the service state with the new task
	s.state.Tasks[task.ID] = *task

	log.Printf("Task %s enqueued with commands: %s", task.ID, commands)

	return task.ID, nil
}

func (s *Service) CancelTask(taskID string) error {

	return nil // Placeholder for task cancellation logic
}

func (s *Service) HandleTask(task RobotTask) {
	log.Println("Handling task:", task.ID)

	// Check if task can be processed, robot must not be crossing the warehouse boundaries

	// Run the task processing logic here
	// Keep on updating the robot state based on the commands in the task

	// Update the task state to Completed
}
