package robot

import (
	"context"
	"testing"
)

// TestServiceImplementsRobotService verifies that the Service struct implements the RobotService interface.
func TestServiceImplementsRobotService(t *testing.T) {
	// This test ensures that Service implements RobotService interface
	// If Service doesn't implement RobotService, this will fail at compile time
	var _ RobotService = (*Service)(nil)
}

// TestNewService tests the NewService constructor.
func TestNewService(t *testing.T) {
	ctx := context.Background()
	taskQueue := make(chan RobotTask, 10)

	service := NewService(ctx, taskQueue)

	if service == nil {
		t.Error("NewService() should return a non-nil service")
	}

	if service.ctx != ctx {
		t.Error("NewService() should set the context correctly")
	}

	if service.taskQueue != taskQueue {
		t.Error("NewService() should set the task queue correctly")
	}

	// Verify service implements RobotService
	var _ RobotService = service

	// Verify initial state
	state := service.CurrentState()
	if state.Tasks == nil {
		t.Error("NewService() should initialize service state with Tasks map")
	}
	if len(state.Tasks) != 0 {
		t.Error("NewService() should initialize empty tasks map")
	}
}

// TODO: Add more tests for the Service methods
// - TestEnqueueTask
// - TestCancelTask
// - TestCurrentState
