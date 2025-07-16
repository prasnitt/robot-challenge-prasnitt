package robot

import "fmt"

type RobotCommand int

// RobotCommand represents a command that can be executed by a robot.
// The commands are represented as integers for easy comparison and storage.
const (
	North RobotCommand = iota
	West
	East
	South
)

func (c RobotCommand) String() string {
	switch c {
	case North:
		return "N"
	case West:
		return "W"
	case East:
		return "E"
	case South:
		return "S"
	default:
		return fmt.Sprintf("Unknown Command %d", c)
	}
}
