package robot

import (
	"sync"
)

type RobotState struct {
	X uint `json:"x"` // Current X coordinate of the robot
	Y uint `json:"y"` // Current Y coordinate of the robot
}

type ServiceState struct {
	Mu         sync.RWMutex         `json:"-"`           // Mutex for concurrent access
	RobotState RobotState           `json:"robot_state"` // Current state of the robot
	Tasks      map[string]RobotTask `json:"tasks"`       // Map of task IDs to RobotTask objects
}

func NewServiceState() ServiceState {
	return ServiceState{
		Mu:         sync.RWMutex{},
		RobotState: RobotState{X: 0, Y: 0}, // Initialize robot at origin
		Tasks:      make(map[string]RobotTask),
	}
}
