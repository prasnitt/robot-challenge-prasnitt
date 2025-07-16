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
		{"Valid Commands", args{"N E S W"}, &RobotTask{RawCmdSequence: "N E S W", Commands: []RobotCommand{North, East, South, West}, State: Pending}, false},
		{"Invalid Command must fail", args{"N X S W"}, nil, true},
		{"Single Command", args{"N"}, &RobotTask{RawCmdSequence: "N", Commands: []RobotCommand{North}, State: Pending}, false},
		{"Whitespace Only", args{"   "}, nil, true},
		{"Extra Spaces are valid", args{"  N   E   S W "}, &RobotTask{RawCmdSequence: "  N   E   S W ", Commands: []RobotCommand{North, East, South, West}, State: Pending}, false},
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
			if !reflect.DeepEqual(got.RawCmdSequence, tt.want.RawCmdSequence) {
				t.Errorf("NewTask() RawCmdSequence  got = %v, want %v", got.RawCmdSequence, tt.want.RawCmdSequence)
			}
			if !reflect.DeepEqual(got.Commands, tt.want.Commands) {
				t.Errorf("NewTask() Commands = %v, want %v", got.Commands, tt.want.Commands)
			}
			if got.State != tt.want.State {
				t.Errorf("NewTask() State = %v, want %v", got.State, tt.want.State)
			}
			if got.ID == "" {
				t.Errorf("NewTask() ID should not be empty")
			}
		})
	}
}
