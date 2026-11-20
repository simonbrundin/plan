package models

import (
	"encoding/json"
	"testing"
)

// Uppskattad tidsåtgång kan sättas, rensas och lämnas orörd - samma
// tre tillstånd som due-datumet har.
func TestUpdateGoalRequestEstimatedMinutesDecoding(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantSet   bool
		wantValue *int
	}{
		{
			name:      "null clears the estimate",
			body:      `{"estimated_minutes": null}`,
			wantSet:   true,
			wantValue: nil,
		},
		{
			name:      "a number sets the estimate",
			body:      `{"estimated_minutes": 90}`,
			wantSet:   true,
			wantValue: intPtr(90),
		},
		{
			name:      "zero is a valid estimate",
			body:      `{"estimated_minutes": 0}`,
			wantSet:   true,
			wantValue: intPtr(0),
		},
		{
			name:      "a missing field leaves the estimate untouched",
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
			if req.EstimatedMinutes.Set != tt.wantSet {
				t.Errorf("EstimatedMinutes.Set = %v, want %v", req.EstimatedMinutes.Set, tt.wantSet)
			}
			switch {
			case tt.wantValue == nil && req.EstimatedMinutes.Value != nil:
				t.Errorf("EstimatedMinutes.Value = %d, want nil", *req.EstimatedMinutes.Value)
			case tt.wantValue != nil && req.EstimatedMinutes.Value == nil:
				t.Errorf("EstimatedMinutes.Value = nil, want %d", *tt.wantValue)
			case tt.wantValue != nil && *req.EstimatedMinutes.Value != *tt.wantValue:
				t.Errorf("EstimatedMinutes.Value = %d, want %d", *req.EstimatedMinutes.Value, *tt.wantValue)
			}
		})
	}
}

func intPtr(i int) *int { return &i }
