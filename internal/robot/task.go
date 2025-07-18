package robot

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// TaskState represents the state of a robot task.
// @Description Current state of the robot task
// @Enum Pending InProgress Aborted RequestCancellation Canceled Completed Invalid
type TaskState int

// RobotCommands represents a slice of RobotCommand values.
// It provides methods for string representation and JSON marshaling.
// @Description A string containing space-separated robot commands
// @Example "N E S W"
type RobotCommands []RobotCommand

// CommandDuration represents the duration between commands.
// @Description Duration between executing commands, can be used to control the speed of command execution.
// @Example "1s"
type CommandDuration time.Duration

const defaultDelayBetweenCommands = CommandDuration(time.Second) // Default delay between commands is 1 second

func (rc RobotCommands) String() string {
	var builder strings.Builder
	for _, cmd := range rc {
		builder.WriteString(cmd.String() + " ")
	}
	commandsStr := builder.String()

	return strings.TrimRight(commandsStr, " ")
}

func (rc RobotCommands) MarshalJSON() ([]byte, error) {
	return json.Marshal(rc.String())
}

func (cd CommandDuration) String() string {
	return time.Duration(cd).String()
}

func (cd CommandDuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(cd.String())
}

// TaskState represents the state of a robot task.
// @Description Current state of the robot task
// @Enum Pending InProgress Aborted RequestCancellation Canceled Completed Invalid
const (
	Pending TaskState = iota
	InProgress
	Aborted
	RequestCancellation // Represents a task that has been requested for cancellation via API
	Canceled            // Represents a task that has been cancelled after cancellation request
	Completed

	Invalid // Represents an invalid state, can be used for error handling
)

// Convert TaskState to string for easy representation.
func (s TaskState) String() string {
	switch s {
	case Pending:
		return "Pending"
	case InProgress:
		return "InProgress"
	case Aborted:
		return "Aborted"
	case RequestCancellation:
		return "RequestCancellation"
	case Canceled:
		return "Canceled"
	case Completed:
		return "Completed"
	case Invalid:
		return "Invalid"
	default:
		return fmt.Sprintf("Unknown State %d", s)
	}
}

func (s TaskState) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

type RobotTask struct {
	ID                   string          `json:"id"`                                                       // Unique identifier for the task
	Commands             RobotCommands   `json:"commands" swaggertype:"string" example:"N E S W"`          // List of commands to be executed by the robot
	State                TaskState       `json:"state" swaggertype:"string" example:"Pending"`             // Current state of the task
	DelayBetweenCommands CommandDuration `json:"delay_between_commands" swaggertype:"string" example:"1s"` // Delay between executing commands

	SequenceNum int    `json:"sequence_num"` // Sequence number for the task, used for ordering tasks in the queue
	Error       string `json:"error"`        // Error message if the task fails

	// DeltaX and DeltaY represent the change in robot's position after executing the commands
	DeltaX int `json:"-"` // Change in X coordinate
	DeltaY int `json:"-"` // Change in Y coordinate
}

// NewTask creates a new RobotTask from a raw command sequence string.
// It parses the string into individual RobotCommand values and initializes the task state to Pending.
func NewTask(rawCmdSequence string, delayBetweenCommandsStr string) (*RobotTask, error) {

	delayBetweenCommands := defaultDelayBetweenCommands // Default delay is set to 1 second

	if delayBetweenCommandsStr != "" {
		// Parse the delay and set it in the task
		duration, err := time.ParseDuration(delayBetweenCommandsStr)
		if err != nil {
			return nil, fmt.Errorf("invalid delay format: %v", err)
		}
		delayBetweenCommands = CommandDuration(duration) // Set the delay in the task
	}

	commands, deltaX, deltaY, err := parseCommands(rawCmdSequence)
	if err != nil {
		return nil, err
	}

	return &RobotTask{
		ID:                   uuid.New().String(),
		Commands:             commands,
		DelayBetweenCommands: delayBetweenCommands,
		State:                Pending,
		DeltaX:               deltaX,
		DeltaY:               deltaY,
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
