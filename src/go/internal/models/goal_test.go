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

func TestUpdateGoalRequestDescriptionDecoding(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantSet bool
		want    string
	}{
		{
			name:    "a description sets the description",
			body:    `{"description": "En längre beskrivning"}`,
			wantSet: true,
			want:    "En längre beskrivning",
		},
		{
			name:    "an empty description clears the description",
			body:    `{"description": ""}`,
			wantSet: true,
			want:    "",
		},
		{
			name:    "a missing field leaves the description untouched",
			body:    `{"title": "Ny titel"}`,
			wantSet: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req UpdateGoalRequest
			if err := json.Unmarshal([]byte(tt.body), &req); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if (req.Description != nil) != tt.wantSet {
				t.Fatalf("Description set = %v, want %v", req.Description != nil, tt.wantSet)
			}
			if tt.wantSet && *req.Description != tt.want {
				t.Errorf("Description = %q, want %q", *req.Description, tt.want)
			}
		})
	}
}

func TestUpdateGoalRequestEternalDecoding(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		wantSet  bool
		wantEtal bool
	}{
		{
			name:     "true marks the goal as eternal",
			body:     `{"eternal": true}`,
			wantSet:  true,
			wantEtal: true,
		},
		{
			name:     "false removes the eternal mark",
			body:     `{"eternal": false}`,
			wantSet:  true,
			wantEtal: false,
		},
		{
			name:     "a missing field leaves the goal untouched",
			body:     `{"title": "Ny titel"}`,
			wantSet:  false,
			wantEtal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req UpdateGoalRequest
			if err := json.Unmarshal([]byte(tt.body), &req); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if (req.Eternal != nil) != tt.wantSet {
				t.Fatalf("Eternal set = %v, want %v", req.Eternal != nil, tt.wantSet)
			}
			if tt.wantSet && *req.Eternal != tt.wantEtal {
				t.Errorf("Eternal = %v, want %v", *req.Eternal, tt.wantEtal)
			}
		})
	}
}

func TestUpdateGoalRequestFinishesGoal(t *testing.T) {
	tests := []struct {
		name string
		body string
		want bool
	}{
		{
			name: "a timestamp finishes the goal",
			body: `{"finished": "2026-01-02T03:04:05Z"}`,
			want: true,
		},
		{
			name: "clearing the timestamp does not finish the goal",
			body: `{"finished": null}`,
			want: false,
		},
		{
			name: "an empty string does not finish the goal",
			body: `{"finished": ""}`,
			want: false,
		},
		{
			name: "a missing field does not finish the goal",
			body: `{"title": "Ny titel"}`,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req UpdateGoalRequest
			if err := json.Unmarshal([]byte(tt.body), &req); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got := req.FinishesGoal(); got != tt.want {
				t.Errorf("FinishesGoal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGoalMarshalsNullTimestamps(t *testing.T) {
	data, err := json.Marshal(Goal{ID: 1, Title: "Mål"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, field := range []string{"started", "finished"} {
		value, ok := raw[field]
		if !ok {
			t.Errorf("fältet %q saknas i JSON-svaret", field)
			continue
		}
		if value != nil {
			t.Errorf("%s = %v, want null", field, value)
		}
	}
}
