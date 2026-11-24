package tools

import (
	"fmt"
	"regexp"
)

// ValidationError represents a detailed validation error
type ValidationError struct {
	Field   string
	Message string
}

// Error returns a formatted error message
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for '%s': %s", e.Field, e.Message)
}

// Validate validates tool arguments against a JSON Schema.
// Supports: type, required, properties, minLength, maxLength, minimum, maximum,
//
//	pattern, enum, items
func Validate(args map[string]any, schema map[string]any) error {
	if schema == nil {
		return nil
	}

	// Check required fields
	if required, ok := schema["required"].([]any); ok {
		for _, r := range required {
			name, ok := r.(string)
			if !ok {
				continue
			}
			if _, exists := args[name]; !exists {
				return &ValidationError{Field: name, Message: "required field missing"}
			}
		}
	}

	// Validate properties
	if props, ok := schema["properties"].(map[string]any); ok {
		for name, prop := range props {
			propMap, ok := prop.(map[string]any)
			if !ok {
				continue
			}

			if val, exists := args[name]; exists {
				if err := validateValue(name, val, propMap); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// validateValue validates a single value against its schema
func validateValue(name string, value any, schema map[string]any) error {
	// Type validation
	if typ, ok := schema["type"].(string); ok {
		if err := validateType(name, value, typ); err != nil {
			return err
		}
	}

	// Enum validation
	if enum, ok := schema["enum"].([]any); ok {
		if err := validateEnum(name, value, enum); err != nil {
			return err
		}
	}

	// String-specific validations
	if str, ok := value.(string); ok {
		if err := validateString(name, str, schema); err != nil {
			return err
		}
	}

	// Number-specific validations
	if num, ok := toFloat64(value); ok {
		if err := validateNumber(name, num, schema); err != nil {
			return err
		}
	}

	// Array validation
	if arr, ok := value.([]any); ok {
		if items, ok := schema["items"].(map[string]any); ok {
			for i, item := range arr {
				if err := validateValue(fmt.Sprintf("%s[%d]", name, i), item, items); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// validateString validates string constraints
func validateString(name, value string, schema map[string]any) error {
	if minLen, ok := schema["minLength"].(float64); ok {
		if len(value) < int(minLen) {
			return &ValidationError{
				Field:   name,
				Message: fmt.Sprintf("string must be at least %d characters, got %d", int(minLen), len(value)),
			}
		}
	}

	if maxLen, ok := schema["maxLength"].(float64); ok {
		if len(value) > int(maxLen) {
			return &ValidationError{
				Field:   name,
				Message: fmt.Sprintf("string must be at most %d characters, got %d", int(maxLen), len(value)),
			}
		}
	}

	if pattern, ok := schema["pattern"].(string); ok {
		if matched, err := regexp.MatchString(pattern, value); err != nil {
			return &ValidationError{
				Field:   name,
				Message: fmt.Sprintf("invalid pattern: %v", err),
			}
		} else if !matched {
			return &ValidationError{
				Field:   name,
				Message: fmt.Sprintf("string does not match pattern '%s'", pattern),
			}
		}
	}

	return nil
}

// validateNumber validates number constraints
func validateNumber(name string, value float64, schema map[string]any) error {
	if min, ok := schema["minimum"].(float64); ok {
		if value < min {
			return &ValidationError{
				Field:   name,
				Message: fmt.Sprintf("number must be >= %v, got %v", min, value),
			}
		}
	}

	if max, ok := schema["maximum"].(float64); ok {
		if value > max {
			return &ValidationError{
				Field:   name,
				Message: fmt.Sprintf("number must be <= %v, got %v", max, value),
			}
		}
	}

	return nil
}
