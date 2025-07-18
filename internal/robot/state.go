package robot

type RobotState struct {
	X uint `json:"x"` // Current X coordinate of the robot
	Y uint `json:"y"` // Current Y coordinate of the robot
}

type ServiceState struct {
	RobotState RobotState           `json:"robot_state"` // Current state of the robot
	Tasks      map[string]RobotTask `json:"tasks"`       // Map of task IDs to RobotTask objects
}

func NewServiceState() ServiceState {
	return ServiceState{
		RobotState: RobotState{X: 0, Y: 0}, // Initialize robot at origin
		Tasks:      make(map[string]RobotTask),
	}
}
