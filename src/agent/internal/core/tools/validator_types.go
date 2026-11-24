package tools

import "fmt"

// validateType validates the type of a value
func validateType(name string, value any, expectedType string) error {
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return &ValidationError{
				Field:   name,
				Message: fmt.Sprintf("expected string, got %T", value),
			}
		}
	case "number":
		if _, ok := toFloat64(value); !ok {
			return &ValidationError{
				Field:   name,
				Message: fmt.Sprintf("expected number, got %T", value),
			}
		}
	case "integer":
		if !isInteger(value) {
			return &ValidationError{
				Field:   name,
				Message: fmt.Sprintf("expected integer, got %T", value),
			}
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return &ValidationError{
				Field:   name,
				Message: fmt.Sprintf("expected boolean, got %T", value),
			}
		}
	case "array":
		if _, ok := value.([]any); !ok {
			return &ValidationError{
				Field:   name,
				Message: fmt.Sprintf("expected array, got %T", value),
			}
		}
	case "object":
		if _, ok := value.(map[string]any); !ok {
			return &ValidationError{
				Field:   name,
				Message: fmt.Sprintf("expected object, got %T", value),
			}
		}
	}
	return nil
}

// validateEnum validates that a value is in the allowed set
func validateEnum(name string, value any, enum []any) error {
	for _, e := range enum {
		if e == value {
			return nil
		}
	}
	// Special case for string enums
	if str, ok := value.(string); ok {
		for _, e := range enum {
			if eStr, ok := e.(string); ok && eStr == str {
				return nil
			}
		}
	}
	return &ValidationError{
		Field:   name,
		Message: fmt.Sprintf("value must be one of %v, got %v", enum, value),
	}
}

// toFloat64 converts various numeric types to float64
func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case int32:
		return float64(n), true
	}
	return 0, false
}

// isInteger checks if a value is an integer type
func isInteger(v any) bool {
	switch v.(type) {
	case int, int64, int32:
		return true
	}
	// Also check if it's a float that represents an integer
	if f, ok := toFloat64(v); ok {
		return f == float64(int64(f))
	}
	return false
}
