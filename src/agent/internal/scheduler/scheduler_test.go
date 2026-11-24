package scheduler

import (
	"testing"

	"plan-agent/internal/planclient"
)

func TestIsAgentDefineGoal(t *testing.T) {
	tests := []struct {
		name string
		goal planclient.Goal
		want bool
	}{
		{
			name: "agent-marked goal in define",
			goal: planclient.Goal{Agent: true, StatusID: 8},
			want: true,
		},
		{
			name: "unmarked goal in define",
			goal: planclient.Goal{Agent: false, StatusID: 8},
			want: false,
		},
		{
			name: "agent-marked goal in another phase",
			goal: planclient.Goal{Agent: true, StatusID: 13},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAgentDefineGoal(tt.goal, 8); got != tt.want {
				t.Fatalf("isAgentDefineGoal(%+v, 8) = %v, want %v", tt.goal, got, tt.want)
			}
		})
	}
}
