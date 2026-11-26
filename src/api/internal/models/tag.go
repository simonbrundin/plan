package models

import (
	"errors"
	"fmt"
	"strings"
)

// MaxTagNameLength är längsta tillåtna taggnamn (antal tecken).
const MaxTagNameLength = 32

// ErrInvalidTagName betyder att ett taggnamn inte kunde användas.
var ErrInvalidTagName = errors.New("invalid tag name")

// Tag är en etikett som kan sättas på ett mål ("ute", "hem", "vinter").
//
// Namnen är normaliserade: gemener, inga dubbla blanksteg och inget
// inledande "#". Det gör att "Ute" och "ute " blir samma tagg.
type Tag struct {
	Name string `json:"name" example:"ute"`
	// GoalCount är antalet mål som har taggen. Används för att se
	// vilka taggar som faktiskt används.
	GoalCount int `json:"goal_count" example:"3"`
}

// NormalizeTagName gör om ett taggnamn till den form som sparas i databasen.
//
// Returnerar ErrInvalidTagName om namnet är tomt eller för långt.
func NormalizeTagName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	// "#ute" och "ute" ska bli samma tagg - #-prefixet är bara en
	// bekvämlighet i gränssnittet.
	name = strings.TrimPrefix(name, "#")
	name = strings.ToLower(name)
	// Flera blanksteg blir ett.
	name = strings.Join(strings.Fields(name), " ")

	if name == "" {
		return "", fmt.Errorf("%w: namnet får inte vara tomt", ErrInvalidTagName)
	}
	if len([]rune(name)) > MaxTagNameLength {
		return "", fmt.Errorf("%w: namnet får vara max %d tecken", ErrInvalidTagName, MaxTagNameLength)
	}

	return name, nil
}

// NormalizeTagNames normaliserar och tar bort dubbletter ur en lista
// med taggnamn. Ordningen behålls - den första förekomsten vinner.
func NormalizeTagNames(raw []string) ([]string, error) {
	seen := map[string]bool{}
	out := make([]string, 0, len(raw))

	for _, r := range raw {
		name, err := NormalizeTagName(r)
		if err != nil {
			return nil, err
		}
		if seen[name] {
			continue
		}
		seen[name] = true
		out = append(out, name)
	}

	return out, nil
}

// SetGoalTagsRequest är request-bodyn för att ersätta ett måls taggar.
//
// En tom lista tar bort alla taggar. Namn som inte finns ännu skapas.
type SetGoalTagsRequest struct {
	Tags []string `json:"tags"`
}
