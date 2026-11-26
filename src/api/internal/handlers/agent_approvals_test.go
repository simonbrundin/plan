package handlers

import (
	"testing"
)

// TestValidateApprovalKind kontrollerar att Plan bara godtar
// godkännandetyper som står i kontraktet med agenten.
//
// Ett förslag utan typ godtas - det är ett äldre förslag och behandlas
// som ett kodförslag.
func TestValidateApprovalKind(t *testing.T) {
	cases := []struct {
		name    string
		payload map[string]any
		wantErr bool
	}{
		{
			name:    "inget payload",
			payload: nil,
		},
		{
			name:    "utan typ",
			payload: map[string]any{"changes": []any{}},
		},
		{
			name:    "måldefinition",
			payload: map[string]any{"kind": "goal_definition"},
		},
		{
			name:    "kodändring",
			payload: map[string]any{"kind": "code_change"},
		},
		{
			name:    "okänd typ",
			payload: map[string]any{"kind": "hitta_på"},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateApprovalKind(tc.payload)
			if tc.wantErr && err == nil {
				t.Error("ett fel väntades")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("oväntat fel: %v", err)
			}
		})
	}
}
