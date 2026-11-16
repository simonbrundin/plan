package models

import (
	"encoding/json"
	"testing"
)

func TestUpdateGoalRequestStartedDecoding(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantSet   bool
		wantValue *string
	}{
		{
			name:      "null clears the timestamp",
			body:      `{"started": null}`,
			wantSet:   true,
			wantValue: nil,
		},
		{
			name:      "a value sets the timestamp",
			body:      `{"started": "2026-01-02T03:04:05Z"}`,
			wantSet:   true,
			wantValue: strPtr("2026-01-02T03:04:05Z"),
		},
		{
			name:      "an empty string clears the timestamp",
			body:      `{"started": ""}`,
			wantSet:   true,
			wantValue: strPtr(""),
		},
		{
			name:      "a missing field leaves the timestamp untouched",
			body:      `{"title": "Ny titel"}`,
			wantSet:   false,
			wantValue: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req UpdateGoalRequest
			if err := json.Unmarshal([]byte(tt.body), &req); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if req.Started.Set != tt.wantSet {
				t.Errorf("Started.Set = %v, want %v", req.Started.Set, tt.wantSet)
			}

			switch {
			case tt.wantValue == nil && req.Started.Value != nil:
				t.Errorf("Started.Value = %q, want nil", *req.Started.Value)
			case tt.wantValue != nil && req.Started.Value == nil:
				t.Errorf("Started.Value = nil, want %q", *tt.wantValue)
			case tt.wantValue != nil && *req.Started.Value != *tt.wantValue:
				t.Errorf("Started.Value = %q, want %q", *req.Started.Value, *tt.wantValue)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}
