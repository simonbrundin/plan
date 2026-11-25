// Package recurrence beräknar förekomster för återkommande mål.
//
// Förekomsterna räknas fram med RRULE (RFC 5545) via rrule-go, precis som
// kalendern gör för återkommande händelser. Skillnaden är att ett måls
// förekomster är datum utan klockslag - därför hanteras de som UTC-datum.
package recurrence

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/teambition/rrule-go"
)

// Mode avgör hur nästa förekomst räknas fram för ett återkommande mål.
type Mode string

const (
	// ModeSchedule håller serien fast relativt målets startdatum. Nästa
	// förekomst är seriens nästa datum som inte passerat - förfallna
	// tillfällen hoppas över.
	ModeSchedule Mode = "schedule"

	// ModePush skjuter fram serien. Nästa förekomst räknas från det datum
	// förekomsten faktiskt slutfördes.
	ModePush Mode = "push"
)

var (
	// ErrInvalidRule betyder att upprepningsregeln inte är giltig.
	ErrInvalidRule = errors.New("ogiltig upprepningsregel")

	// ErrNoOccurrence betyder att regeln inte ger någon förekomst.
	ErrNoOccurrence = errors.New("regeln ger inga förekomster")

	// ErrInvalidMode betyder att schemaläggningsstrategin är okänd.
	ErrInvalidMode = errors.New("ogiltig schemaläggningsstrategi")
)

const rulePrefix = "RRULE:"

// Validate kontrollerar att en upprepningsregel är giltig.
//
// rrule-go accepterar till exempel INTERVAL=0, vilket skulle ge en regel
// som antingen producerar oändligt många förekomster eller ingen alls.
// Sådana värden avvisas här i stället.
func Validate(rule string) error {
	_, err := parse(rule)
	return err
}

// ValidateMode kontrollerar att schemaläggningsstrategin är känd.
func ValidateMode(mode Mode) error {
	switch mode {
	case ModeSchedule, ModePush:
		return nil
	default:
		return fmt.Errorf("%w: %q", ErrInvalidMode, string(mode))
	}
}

// FirstOccurrence returnerar den första förekomsten på eller efter anchor.
//
// Om regeln inte ger någon förekomst returneras ErrNoOccurrence.
func FirstOccurrence(rule string, anchor time.Time) (time.Time, error) {
	parsed, err := parse(rule)
	if err != nil {
		return time.Time{}, err
	}

	parsed.DTStart(startOfDay(anchor))

	occurrence := parsed.After(startOfDay(anchor), true)
	if occurrence.IsZero() {
		return time.Time{}, fmt.Errorf("%w: serien börjar efter anchor", ErrNoOccurrence)
	}

	return startOfDay(occurrence), nil
}

// NextOccurrence returnerar nästa förekomst efter att en förekomst
// slutförts.
//
// completed är datumet för förekomsten som slutfördes och completedOn är
// det datum slutförandet skedde. anchor är målets startdatum och används i
// schedule-läge, där serien ligger fast.
//
// I schedule-läge hoppas förfallna förekomster över - nästa förekomst är det
// första seriedatumet som inte passerat. I push-läge räknas serien från
// completedOn, så att nästa förekomst följer slutförandetiden.
func NextOccurrence(rule string, mode Mode, anchor, completed, completedOn time.Time) (time.Time, error) {
	parsed, err := parse(rule)
	if err != nil {
		return time.Time{}, err
	}

	if err := ValidateMode(mode); err != nil {
		return time.Time{}, err
	}

	completedDay := startOfDay(completed)
	completedOnDay := startOfDay(completedOn)

	var seriesStart, lowerBound time.Time

	switch mode {
	case ModePush:
		seriesStart = maxTime(completedDay, completedOnDay)
		lowerBound = seriesStart.AddDate(0, 0, 1)
	case ModeSchedule:
		seriesStart = startOfDay(anchor)
		lowerBound = maxTime(completedDay.AddDate(0, 0, 1), completedOnDay)
	}

	parsed.DTStart(seriesStart)

	occurrence := parsed.After(lowerBound.Add(-time.Second), true)
	if occurrence.IsZero() {
		return time.Time{}, fmt.Errorf("%w: serien är slut", ErrNoOccurrence)
	}

	return startOfDay(occurrence), nil
}

// parse tolkar en RRULE-sträng och kontrollerar att den är giltig.
func parse(rule string) (*rrule.RRule, error) {
	normalized := strings.ToUpper(strings.TrimSpace(rule))
	if normalized == "" {
		return nil, fmt.Errorf("%w: regeln är tom", ErrInvalidRule)
	}

	normalized = strings.TrimPrefix(normalized, rulePrefix)

	if err := validateParts(normalized); err != nil {
		return nil, err
	}

	parsed, err := rrule.StrToRRule(normalized)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidRule, err.Error())
	}

	return parsed, nil
}

// validateParts kontrollerar de delar av regeln som rrule-go inte själv
// avvisar, nämligen INTERVAL och COUNT.
func validateParts(rule string) error {
	for _, part := range strings.Split(rule, ";") {
		key, value, found := strings.Cut(part, "=")
		if !found {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		switch key {
		case "INTERVAL", "COUNT":
			number, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("%w: %s har ogiltigt värde %q", ErrInvalidRule, key, value)
			}
			if number < 1 {
				return fmt.Errorf("%w: %s måste vara minst 1", ErrInvalidRule, key)
			}
		}
	}

	return nil
}

// startOfDay gör om ett klockslag till det datum det tillhör.
func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
