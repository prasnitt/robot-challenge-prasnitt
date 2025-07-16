package robot

import "testing"

func TestRobotCommand_String(t *testing.T) {
	tests := []struct {
		name string
		c    RobotCommand
		want string
	}{
		{"North", North, "N"},
		{"West", West, "W"},
		{"East", East, "E"},
		{"South", South, "S"},
		{"Unknown", RobotCommand(999), "Unknown Command 999"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.String(); got != tt.want {
				t.Errorf("RobotCommand.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
