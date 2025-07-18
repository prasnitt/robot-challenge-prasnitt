package robot

import (
	"context"
	"log"
)

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
			s.handleTask(task) // Process incoming tasks
		}
	}
}

func (s *Service) handleTask(task RobotTask) {

}

// GetState returns the current state of the robot service.
func (s *Service) GetState() ServiceState {
	return s.state
}
