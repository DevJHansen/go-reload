package builder

import (
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		runCmd  string
		wantErr bool
	}{
		{
			name:    "valid command",
			runCmd:  "./app",
			wantErr: false,
		},
		{
			name:    "valid command with args",
			runCmd:  "./app --flag value",
			wantErr: false,
		},
		{
			name:    "empty string",
			runCmd:  "",
			wantErr: true,
		},
		{
			name:    "white space only",
			runCmd:  "    ",
			wantErr: true,
		},
		{
			name:    "tabs only",
			runCmd:  "\t\t",
			wantErr: true,
		},
		{
			name:    "new lines only",
			runCmd:  "\n\n",
			wantErr: true,
		},
		{
			name:    "only white space, new lines and tabs combo",
			runCmd:  "\n\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder, err := New(tt.runCmd)

			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if builder == nil {
					t.Errorf("New() returned nil builder")
					return
				}
				if builder.cmd != nil {
					t.Errorf("New() builder.cmd should be nil")
					return
				}
				if builder.runCmd != tt.runCmd {
					t.Errorf("New() builder.runCmd = %v, want %v", builder.runCmd, tt.runCmd)
					return
				}
			}
		})
	}
}
