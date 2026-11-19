package recurrence

import (
	"errors"
	"testing"
	"time"
)

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		rule    string
		wantErr bool
	}{
		{name: "daglig", rule: "FREQ=DAILY"},
		{name: "varannan dag", rule: "FREQ=DAILY;INTERVAL=2"},
		{name: "veckodag", rule: "FREQ=WEEKLY;BYDAY=MO"},
		{name: "med prefix", rule: "RRULE:FREQ=DAILY"},
		{name: "tom", rule: "", wantErr: true},
		{name: "mellanrum", rule: "   ", wantErr: true},
		{name: "intervall noll", rule: "FREQ=DAILY;INTERVAL=0", wantErr: true},
		{name: "negativt intervall", rule: "FREQ=DAILY;INTERVAL=-2", wantErr: true},
		{name: "count noll", rule: "FREQ=DAILY;COUNT=0", wantErr: true},
		{name: "okänd frekvens", rule: "FREQ=BOGUS", wantErr: true},
		{name: "trasig", rule: "FREQ=DAILY;INTERVAL=abc", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.rule)
			if tt.wantErr && err == nil {
				t.Fatalf("Validate(%q) = nil, want error", tt.rule)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("Validate(%q) = %v, want nil", tt.rule, err)
			}
		})
	}
}

func TestFirstOccurrence(t *testing.T) {
	tests := []struct {
		name   string
		rule   string
		anchor time.Time
		want   time.Time
	}{
		{
			name:   "daglig börjar på anchor",
			rule:   "FREQ=DAILY",
			anchor: date(2026, time.January, 5),
			want:   date(2026, time.January, 5),
		},
		{
			name:   "varannan dag börjar på anchor",
			rule:   "FREQ=DAILY;INTERVAL=2",
			anchor: date(2026, time.January, 5),
			want:   date(2026, time.January, 5),
		},
		{
			name:   "veckodag hoppar till nästa matchning",
			rule:   "FREQ=WEEKLY;BYDAY=MO",
			anchor: date(2026, time.January, 7),
			want:   date(2026, time.January, 12),
		},
		{
			name:   "månadsvis",
			rule:   "FREQ=MONTHLY",
			anchor: date(2026, time.January, 5),
			want:   date(2026, time.January, 5),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FirstOccurrence(tt.rule, tt.anchor)
			if err != nil {
				t.Fatalf("FirstOccurrence(%q) = error %v", tt.rule, err)
			}
			if !got.Equal(tt.want) {
				t.Fatalf("FirstOccurrence(%q) = %s, want %s", tt.rule, got.Format(time.RFC3339), tt.want.Format(time.RFC3339))
			}
		})
	}
}

func TestFirstOccurrenceWithoutOccurrences(t *testing.T) {
	_, err := FirstOccurrence("FREQ=DAILY;UNTIL=20200101T000000Z", date(2026, time.January, 5))
	if !errors.Is(err, ErrNoOccurrence) {
		t.Fatalf("err = %v, want ErrNoOccurrence", err)
	}
}

func TestNextOccurrenceSchedule(t *testing.T) {
	anchor := date(2026, time.January, 1)

	tests := []struct {
		name        string
		rule        string
		completed   time.Time
		completedOn time.Time
		want        time.Time
	}{
		{
			name:        "i tid",
			rule:        "FREQ=DAILY;INTERVAL=2",
			completed:   date(2026, time.January, 5),
			completedOn: date(2026, time.January, 5),
			want:        date(2026, time.January, 7),
		},
		{
			name:        "sen hoppar över förfallna",
			rule:        "FREQ=DAILY;INTERVAL=2",
			completed:   date(2026, time.January, 5),
			completedOn: date(2026, time.January, 10),
			want:        date(2026, time.January, 11),
		},
		{
			name:        "veckan hålls",
			rule:        "FREQ=WEEKLY;BYDAY=MO",
			completed:   date(2026, time.January, 5),
			completedOn: date(2026, time.January, 5),
			want:        date(2026, time.January, 12),
		},
		{
			name:        "tidig slutförande",
			rule:        "FREQ=DAILY;INTERVAL=2",
			completed:   date(2026, time.January, 7),
			completedOn: date(2026, time.January, 5),
			want:        date(2026, time.January, 9),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NextOccurrence(tt.rule, ModeSchedule, anchor, tt.completed, tt.completedOn)
			if err != nil {
				t.Fatalf("NextOccurrence = error %v", err)
			}
			if !got.Equal(tt.want) {
				t.Fatalf("NextOccurrence = %s, want %s", got.Format(time.RFC3339), tt.want.Format(time.RFC3339))
			}
		})
	}
}

func TestNextOccurrencePush(t *testing.T) {
	anchor := date(2026, time.January, 1)

	tests := []struct {
		name        string
		rule        string
		completed   time.Time
		completedOn time.Time
		want        time.Time
	}{
		{
			name:        "i tid",
			rule:        "FREQ=DAILY;INTERVAL=2",
			completed:   date(2026, time.January, 5),
			completedOn: date(2026, time.January, 5),
			want:        date(2026, time.January, 7),
		},
		{
			name:        "sen glider serien",
			rule:        "FREQ=DAILY;INTERVAL=2",
			completed:   date(2026, time.January, 5),
			completedOn: date(2026, time.January, 10),
			want:        date(2026, time.January, 12),
		},
		{
			name:        "veckodag följer veckodagen",
			rule:        "FREQ=WEEKLY;BYDAY=MO",
			completed:   date(2026, time.January, 5),
			completedOn: date(2026, time.January, 6),
			want:        date(2026, time.January, 12),
		},
		{
			name:        "tidig slutförande",
			rule:        "FREQ=DAILY;INTERVAL=2",
			completed:   date(2026, time.January, 7),
			completedOn: date(2026, time.January, 5),
			want:        date(2026, time.January, 9),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NextOccurrence(tt.rule, ModePush, anchor, tt.completed, tt.completedOn)
			if err != nil {
				t.Fatalf("NextOccurrence = error %v", err)
			}
			if !got.Equal(tt.want) {
				t.Fatalf("NextOccurrence = %s, want %s", got.Format(time.RFC3339), tt.want.Format(time.RFC3339))
			}
		})
	}
}

func TestNextOccurrenceUnknownMode(t *testing.T) {
	_, err := NextOccurrence("FREQ=DAILY", Mode("nonsense"), date(2026, time.January, 1), date(2026, time.January, 1), date(2026, time.January, 1))
	if !errors.Is(err, ErrInvalidMode) {
		t.Fatalf("err = %v, want ErrInvalidMode", err)
	}
}

func TestNextOccurrenceWhenSeriesEnds(t *testing.T) {
	anchor := date(2026, time.January, 1)

	_, err := NextOccurrence("FREQ=DAILY;COUNT=2", ModeSchedule, anchor, date(2026, time.January, 2), date(2026, time.January, 2))
	if !errors.Is(err, ErrNoOccurrence) {
		t.Fatalf("err = %v, want ErrNoOccurrence", err)
	}
}
