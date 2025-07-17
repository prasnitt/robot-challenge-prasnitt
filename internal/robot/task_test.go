package robot

import (
	"reflect"
	"testing"
)

func TestNewTask(t *testing.T) {
	type args struct {
		rawCmdSequence string
	}
	tests := []struct {
		name    string
		args    args
		want    *RobotTask
		wantErr bool
	}{
		{"Must not have empty command", args{""}, nil, true},
		{"Valid Commands reach to same position", args{"N E S W"}, &RobotTask{Commands: []RobotCommand{North, East, South, West}, State: Pending, DeltaX: 0, DeltaY: 0}, false},
		{"Valid Commands reach to position 3, 3 ", args{"N E N E N E"}, &RobotTask{Commands: []RobotCommand{North, East, North, East, North, East}, State: Pending, DeltaX: 3, DeltaY: 3}, false},
		{"Valid Commands reach to position 1, 1 ", args{"N E N E S W"}, &RobotTask{Commands: []RobotCommand{North, East, North, East, South, West}, State: Pending, DeltaX: 1, DeltaY: 1}, false},
		{"Invalid Command must fail", args{"N X S W"}, nil, true},
		{"Single Command", args{"N"}, &RobotTask{Commands: []RobotCommand{North}, State: Pending, DeltaY: 1}, false},
		{"Whitespace Only", args{"   "}, nil, true},
		{"Extra Spaces are valid", args{"  N   E   S W "}, &RobotTask{Commands: []RobotCommand{North, East, South, West}, State: Pending, DeltaX: 0, DeltaY: 0}, false},
		{"Lower case command is not allowed", args{"n e s w"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTask(tt.args.rawCmdSequence)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// For Valid cases, check the fields of the RobotTask
			if !reflect.DeepEqual(got.Commands, tt.want.Commands) {
				t.Errorf("NewTask() Commands = %v, want %v", got.Commands, tt.want.Commands)
			}
			if got.State != tt.want.State {
				t.Errorf("NewTask() State = %v, want %v", got.State, tt.want.State)
			}
			if got.DeltaX != tt.want.DeltaX {
				t.Errorf("NewTask() DeltaX = %v, want %v", got.DeltaX, tt.want.DeltaX)
			}
			if got.DeltaY != tt.want.DeltaY {
				t.Errorf("NewTask() DeltaY = %v, want %v", got.DeltaY, tt.want.DeltaY)
			}

			if got.ID == "" {
				t.Errorf("NewTask() ID should not be empty")
			}
		})
	}
}
