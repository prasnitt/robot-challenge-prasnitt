package robot

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type TaskState int

// TaskState represents the state of a robot task.
const (
	Pending TaskState = iota
	InProgress
	Aborted
	Completed
)

type RobotTask struct {
	ID       string
	Commands []RobotCommand
	State    TaskState
	DeltaX   int // Change in X coordinate
	DeltaY   int // Change in Y coordinate
}

// NewTask creates a new RobotTask from a raw command sequence string.
// It parses the string into individual RobotCommand values and initializes the task state to Pending.
func NewTask(rawCmdSequence string) (*RobotTask, error) {
	commands, deltaX, deltaY, err := parseCommands(rawCmdSequence)
	if err != nil {
		return nil, err
	}

	return &RobotTask{
		ID:       uuid.New().String(),
		Commands: commands,
		State:    Pending,
		DeltaX:   deltaX,
		DeltaY:   deltaY,
	}, nil
}

// removeEmptyStrings removes empty strings from a slice of strings.
// This is useful for cleaning up command sequences that may have extra spaces.
func removeEmptyStrings(slice []string) []string {
	var result []string
	for _, s := range slice {
		if strings.TrimSpace(s) != "" {
			result = append(result, s)
		}
	}
	return result
}

// parseCommands takes a raw command sequence string and converts it into a slice of RobotCommand.
// It returns an error if any command in the sequence is invalid.
func parseCommands(raw string) ([]RobotCommand, int, int, error) {
	parts := strings.Split(raw, " ")
	parts = removeEmptyStrings(parts) // Remove any empty strings from the split
	deltaX, deltaY := 0, 0
	if len(parts) == 0 {
		return nil, deltaX, deltaY, fmt.Errorf("no commands provided")
	}

	commands := make([]RobotCommand, 0, len(parts))

	for _, p := range parts {
		switch p {
		case "N":
			deltaY++
			commands = append(commands, North)
		case "W":
			deltaX--
			commands = append(commands, West)
		case "E":
			deltaX++
			commands = append(commands, East)
		case "S":
			deltaY--
			commands = append(commands, South)
		default:
			return nil, deltaX, deltaY, fmt.Errorf("invalid command: %s", p)
		}
	}
	return commands, deltaX, deltaY, nil
}
