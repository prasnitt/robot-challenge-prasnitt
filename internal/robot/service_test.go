package robot

import (
	"context"
	"sync"
	"testing"
	"time"
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
	taskIdQueue := make(chan string, 10)

	service := NewService(ctx, taskIdQueue)

	if service == nil {
		t.Error("NewService() should return a non-nil service")
	}

	if service.ctx != ctx {
		t.Error("NewService() should set the context correctly")
	}

	if service.taskIdQueue != taskIdQueue {
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
	if state.RobotState.X != 0 || state.RobotState.Y != 0 {
		t.Error("NewService() should initialize robot at origin (0,0)")
	}
}

// TestEnqueueTask tests task enqueueing functionality.
func TestEnqueueTask(t *testing.T) {
	ctx := context.Background()
	taskIdQueue := make(chan string, 10)
	service := NewService(ctx, taskIdQueue)

	tests := []struct {
		name                 string
		commands             string
		delayBetweenCommands string
		expectError          bool
	}{
		{
			name:                 "Valid task with basic commands",
			commands:             "N E S W",
			delayBetweenCommands: "100ms",
			expectError:          false,
		},
		{
			name:                 "Valid task with default delay",
			commands:             "N N E",
			delayBetweenCommands: "",
			expectError:          false,
		},
		{
			name:                 "Invalid commands",
			commands:             "N X E",
			delayBetweenCommands: "100ms",
			expectError:          true,
		},
		{
			name:                 "Invalid delay format",
			commands:             "N E",
			delayBetweenCommands: "invalid",
			expectError:          true,
		},
		{
			name:                 "Empty commands",
			commands:             "",
			delayBetweenCommands: "100ms",
			expectError:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskID, err := service.EnqueueTask(tt.commands, tt.delayBetweenCommands)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for test case %s, but got none", tt.name)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for test case %s: %v", tt.name, err)
				return
			}

			if taskID == "" {
				t.Error("Expected non-empty task ID")
			}

			// Verify task was added to queue
			select {
			case queuedTaskID := <-taskIdQueue:
				if queuedTaskID != taskID {
					t.Errorf("Expected task ID %s in queue, got %s", taskID, queuedTaskID)
				}
			case <-time.After(100 * time.Millisecond):
				t.Error("Task was not added to queue within timeout")
			}

			// Verify task exists in service state
			state := service.CurrentState()
			task, exists := state.Tasks[taskID]
			if !exists {
				t.Error("Task should exist in service state")
			}
			if task.State != Pending {
				t.Errorf("Expected task state to be Pending, got %s", task.State)
			}
		})
	}
}

// TestCancelTask tests task cancellation functionality.
func TestCancelTask(t *testing.T) {
	ctx := context.Background()
	taskIdQueue := make(chan string, 10)
	service := NewService(ctx, taskIdQueue)

	// Test canceling non-existent task
	t.Run("Cancel non-existent task", func(t *testing.T) {
		err := service.CancelTask("non-existent-id")
		if err == nil {
			t.Error("Expected error when canceling non-existent task")
		}
	})

	// Test canceling pending task
	t.Run("Cancel pending task", func(t *testing.T) {
		taskID, err := service.EnqueueTask("N E", "100ms")
		if err != nil {
			t.Fatalf("Failed to enqueue task: %v", err)
		}

		// Clear the queue
		<-taskIdQueue

		err = service.CancelTask(taskID)
		if err != nil {
			t.Errorf("Unexpected error canceling pending task: %v", err)
		}

		state := service.CurrentState()
		task := state.Tasks[taskID]
		if task.State != Canceled {
			t.Errorf("Expected task state to be Canceled, got %s", task.State)
		}
		if task.Error == "" {
			t.Error("Expected error message for canceled task")
		}
	})

	// Test canceling in-progress task
	t.Run("Cancel in-progress task", func(t *testing.T) {
		taskID, err := service.EnqueueTask("N E", "100ms")
		if err != nil {
			t.Fatalf("Failed to enqueue task: %v", err)
		}

		// Clear the queue
		<-taskIdQueue

		// Manually set task to InProgress
		service.UpdateTaskState(taskID, InProgress)

		err = service.CancelTask(taskID)
		if err != nil {
			t.Errorf("Unexpected error canceling in-progress task: %v", err)
		}

		state := service.CurrentState()
		task := state.Tasks[taskID]
		if task.State != RequestCancellation {
			t.Errorf("Expected task state to be RequestCancellation, got %s", task.State)
		}
	})

	// Test canceling completed task
	t.Run("Cancel completed task", func(t *testing.T) {
		taskID, err := service.EnqueueTask("N E", "100ms")
		if err != nil {
			t.Fatalf("Failed to enqueue task: %v", err)
		}

		// Clear the queue
		<-taskIdQueue

		// Manually set task to Completed
		service.UpdateTaskState(taskID, Completed)

		err = service.CancelTask(taskID)
		if err == nil {
			t.Error("Expected error when canceling completed task")
		}
	})
}

// TestCurrentState tests getting current service state.
func TestCurrentState(t *testing.T) {
	ctx := context.Background()
	taskIdQueue := make(chan string, 10)
	service := NewService(ctx, taskIdQueue)

	// Test initial state
	state := service.CurrentState()
	if state.CurTaskCount != 0 {
		t.Errorf("Expected initial task count to be 0, got %d", state.CurTaskCount)
	}
	if len(state.Tasks) != 0 {
		t.Errorf("Expected initial tasks map to be empty, got %d tasks", len(state.Tasks))
	}

	// Add a task and verify state changes
	taskID, err := service.EnqueueTask("N E", "100ms")
	if err != nil {
		t.Fatalf("Failed to enqueue task: %v", err)
	}

	state = service.CurrentState()
	if state.CurTaskCount != 1 {
		t.Errorf("Expected task count to be 1, got %d", state.CurTaskCount)
	}
	if len(state.Tasks) != 1 {
		t.Errorf("Expected 1 task in map, got %d", len(state.Tasks))
	}
	if _, exists := state.Tasks[taskID]; !exists {
		t.Error("Expected task to exist in state")
	}
}

// TestExecuteRobotCommand tests individual robot command execution.
func TestExecuteRobotCommand(t *testing.T) {
	ctx := context.Background()
	taskIdQueue := make(chan string, 10)
	service := NewService(ctx, taskIdQueue)

	tests := []struct {
		name                 string
		command              RobotCommand
		startX, startY       uint
		expectedX, expectedY uint
		expectError          bool
	}{
		{"Move North", North, 5, 5, 5, 6, false},
		{"Move South", South, 5, 5, 5, 4, false},
		{"Move East", East, 5, 5, 6, 5, false},
		{"Move West", West, 5, 5, 4, 5, false},
		{"Move North at boundary", North, 5, warehouseSize, 5, warehouseSize, true},
		{"Move South at boundary", South, 5, 0, 5, 0, true},
		{"Move East at boundary", East, warehouseSize, 5, warehouseSize, 5, true},
		{"Move West at boundary", West, 0, 5, 0, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set initial robot position
			service.SetRobotState(RobotState{X: tt.startX, Y: tt.startY})

			err := service.ExecuteRobotCommand(tt.command)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			robotState := service.GetRobotState()
			if robotState.X != tt.expectedX || robotState.Y != tt.expectedY {
				t.Errorf("Expected position (%d, %d), got (%d, %d)",
					tt.expectedX, tt.expectedY, robotState.X, robotState.Y)
			}
		})
	}
}

// TestIsTaskValid tests task validation logic.
func TestIsTaskValid(t *testing.T) {
	ctx := context.Background()
	taskIdQueue := make(chan string, 10)
	service := NewService(ctx, taskIdQueue)

	tests := []struct {
		name           string
		startX, startY uint
		deltaX, deltaY int
		expectValid    bool
	}{
		{"Valid task within bounds", 5, 5, 2, 2, true},
		{"Valid task at origin", 0, 0, 5, 5, true},
		{"Invalid task - exceeds X boundary", 8, 5, 5, 0, false},
		{"Invalid task - exceeds Y boundary", 5, 8, 0, 5, false},
		{"Invalid task - negative X", 2, 5, -5, 0, false},
		{"Invalid task - negative Y", 5, 2, 0, -5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service.SetRobotState(RobotState{X: tt.startX, Y: tt.startY})

			task := RobotTask{
				DeltaX: tt.deltaX,
				DeltaY: tt.deltaY,
			}

			isValid := service.IsTaskValid(task)
			if isValid != tt.expectValid {
				t.Errorf("Expected validity %t, got %t", tt.expectValid, isValid)
			}
		})
	}
}

// TestConcurrentAccess tests concurrent access to service methods.
func TestConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	taskIdQueue := make(chan string, 100)
	service := NewService(ctx, taskIdQueue)

	// Test concurrent task enqueueing
	t.Run("Concurrent enqueue", func(t *testing.T) {
		var wg sync.WaitGroup
		numGoroutines := 10
		tasksPerGoroutine := 5

		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func(routineID int) {
				defer wg.Done()
				for j := 0; j < tasksPerGoroutine; j++ {
					_, err := service.EnqueueTask("N E", "10ms")
					if err != nil {
						t.Errorf("Goroutine %d: Failed to enqueue task %d: %v", routineID, j, err)
					}
				}
			}(i)
		}

		wg.Wait()

		// Verify all tasks were enqueued
		state := service.CurrentState()
		expectedTasks := numGoroutines * tasksPerGoroutine
		if state.CurTaskCount != expectedTasks {
			t.Errorf("Expected %d tasks, got %d", expectedTasks, state.CurTaskCount)
		}
		if len(state.Tasks) != expectedTasks {
			t.Errorf("Expected %d tasks in map, got %d", expectedTasks, len(state.Tasks))
		}
	})

	// Test concurrent state access
	t.Run("Concurrent state access", func(t *testing.T) {
		var wg sync.WaitGroup
		numReaders := 20

		wg.Add(numReaders)
		for i := 0; i < numReaders; i++ {
			go func(readerID int) {
				defer wg.Done()
				for j := 0; j < 10; j++ {
					state := service.CurrentState()
					if state.Tasks == nil {
						t.Errorf("Reader %d: Tasks map should not be nil", readerID)
					}
				}
			}(i)
		}

		wg.Wait()
	})
}

// TestHandleTaskCancellation tests task cancellation during processing.
func TestHandleTaskCancellation(t *testing.T) {
	ctx := context.Background()
	taskIdQueue := make(chan string, 10)
	service := NewService(ctx, taskIdQueue)

	// Create a task with longer delay for testing cancellation
	taskID, err := service.EnqueueTask("N N N", "50ms")
	if err != nil {
		t.Fatalf("Failed to enqueue task: %v", err)
	}

	// Start processing the task in a goroutine
	go func() {
		// Simulate starting task processing
		service.UpdateTaskState(taskID, InProgress)

		// Simulate some processing time
		time.Sleep(10 * time.Millisecond)

		// Request cancellation
		service.UpdateTaskState(taskID, RequestCancellation)
	}()

	// Wait for the task to be in RequestCancellation state
	for i := 0; i < 50; i++ {
		state, _ := service.GetTaskState(taskID)
		if state == RequestCancellation {
			break
		}
		time.Sleep(1 * time.Millisecond)
	}

	// Verify the task is in RequestCancellation state
	state, err := service.GetTaskState(taskID)
	if err != nil {
		t.Fatalf("Failed to get task state: %v", err)
	}
	if state != RequestCancellation {
		t.Errorf("Expected task state to be RequestCancellation, got %s", state)
	}
}
