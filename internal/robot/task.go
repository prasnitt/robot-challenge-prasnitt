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
	ID             string
	RawCmdSequence string
	Commands       []RobotCommand
	State          TaskState
}

// NewTask creates a new RobotTask from a raw command sequence string.
// It parses the string into individual RobotCommand values and initializes the task state to Pending.
func NewTask(rawCmdSequence string) (*RobotTask, error) {
	commands, err := parseCommands(rawCmdSequence)
	if err != nil {
		return nil, err
	}

	return &RobotTask{
		ID:             uuid.New().String(),
		RawCmdSequence: rawCmdSequence,
		Commands:       commands,
		State:          Pending,
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
func parseCommands(raw string) ([]RobotCommand, error) {
	parts := strings.Split(raw, " ")
	parts = removeEmptyStrings(parts) // Remove any empty strings from the split
	if len(parts) == 0 {
		return nil, fmt.Errorf("no commands provided")
	}

	commands := make([]RobotCommand, 0, len(parts))

	for _, p := range parts {
		switch p {
		case "N":
			commands = append(commands, North)
		case "W":
			commands = append(commands, West)
		case "E":
			commands = append(commands, East)
		case "S":
			commands = append(commands, South)
		default:
			return nil, fmt.Errorf("invalid command: %s", p)
		}
	}
	return commands, nil
}
