package tools

import (
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		schema  map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil schema",
			args:    map[string]any{"foo": "bar"},
			schema:  nil,
			wantErr: false,
		},
		{
			name:    "empty schema",
			args:    map[string]any{"foo": "bar"},
			schema:  map[string]any{},
			wantErr: false,
		},
		{
			name:    "required field missing",
			args:    map[string]any{},
			schema:  map[string]any{"required": []any{"name"}},
			wantErr: true,
			errMsg:  "name",
		},
		{
			name:    "required field present",
			args:    map[string]any{"name": "test"},
			schema:  map[string]any{"required": []any{"name"}},
			wantErr: false,
		},
		{
			name:    "string type valid",
			args:    map[string]any{"name": "test"},
			schema:  map[string]any{"properties": map[string]any{"name": map[string]any{"type": "string"}}},
			wantErr: false,
		},
		{
			name:    "string type invalid",
			args:    map[string]any{"name": 123},
			schema:  map[string]any{"properties": map[string]any{"name": map[string]any{"type": "string"}}},
			wantErr: true,
			errMsg:  "expected string",
		},
		{
			name:    "number type valid",
			args:    map[string]any{"count": 42},
			schema:  map[string]any{"properties": map[string]any{"count": map[string]any{"type": "number"}}},
			wantErr: false,
		},
		{
			name:    "number type valid float",
			args:    map[string]any{"count": 3.14},
			schema:  map[string]any{"properties": map[string]any{"count": map[string]any{"type": "number"}}},
			wantErr: false,
		},
		{
			name:    "number type invalid",
			args:    map[string]any{"count": "42"},
			schema:  map[string]any{"properties": map[string]any{"count": map[string]any{"type": "number"}}},
			wantErr: true,
			errMsg:  "expected number",
		},
		{
			name:    "integer type valid",
			args:    map[string]any{"count": 42},
			schema:  map[string]any{"properties": map[string]any{"count": map[string]any{"type": "integer"}}},
			wantErr: false,
		},
		{
			name:    "integer type invalid",
			args:    map[string]any{"count": 3.14},
			schema:  map[string]any{"properties": map[string]any{"count": map[string]any{"type": "integer"}}},
			wantErr: true,
			errMsg:  "expected integer",
		},
		{
			name:    "boolean type valid",
			args:    map[string]any{"active": true},
			schema:  map[string]any{"properties": map[string]any{"active": map[string]any{"type": "boolean"}}},
			wantErr: false,
		},
		{
			name:    "boolean type invalid",
			args:    map[string]any{"active": "true"},
			schema:  map[string]any{"properties": map[string]any{"active": map[string]any{"type": "boolean"}}},
			wantErr: true,
			errMsg:  "expected boolean",
		},
		{
			name:    "array type valid",
			args:    map[string]any{"items": []any{1, 2, 3}},
			schema:  map[string]any{"properties": map[string]any{"items": map[string]any{"type": "array"}}},
			wantErr: false,
		},
		{
			name:    "array type invalid",
			args:    map[string]any{"items": "not an array"},
			schema:  map[string]any{"properties": map[string]any{"items": map[string]any{"type": "array"}}},
			wantErr: true,
			errMsg:  "expected array",
		},
		{
			name:    "object type valid",
			args:    map[string]any{"data": map[string]any{"key": "value"}},
			schema:  map[string]any{"properties": map[string]any{"data": map[string]any{"type": "object"}}},
			wantErr: false,
		},
		{
			name:    "minLength valid",
			args:    map[string]any{"name": "hello"},
			schema:  map[string]any{"properties": map[string]any{"name": map[string]any{"type": "string", "minLength": float64(3)}}},
			wantErr: false,
		},
		{
			name:    "minLength violated",
			args:    map[string]any{"name": "hi"},
			schema:  map[string]any{"properties": map[string]any{"name": map[string]any{"type": "string", "minLength": float64(3)}}},
			wantErr: true,
			errMsg:  "at least 3 characters",
		},
		{
			name:    "maxLength valid",
			args:    map[string]any{"name": "hello"},
			schema:  map[string]any{"properties": map[string]any{"name": map[string]any{"type": "string", "maxLength": float64(10)}}},
			wantErr: false,
		},
		{
			name:    "maxLength violated",
			args:    map[string]any{"name": "hello world"},
			schema:  map[string]any{"properties": map[string]any{"name": map[string]any{"type": "string", "maxLength": float64(5)}}},
			wantErr: true,
			errMsg:  "at most 5 characters",
		},
		{
			name:    "minimum valid",
			args:    map[string]any{"count": 10},
			schema:  map[string]any{"properties": map[string]any{"count": map[string]any{"type": "number", "minimum": float64(5)}}},
			wantErr: false,
		},
		{
			name:    "minimum violated",
			args:    map[string]any{"count": 3},
			schema:  map[string]any{"properties": map[string]any{"count": map[string]any{"type": "number", "minimum": float64(5)}}},
			wantErr: true,
			errMsg:  "must be >=",
		},
		{
			name:    "maximum valid",
			args:    map[string]any{"count": 10},
			schema:  map[string]any{"properties": map[string]any{"count": map[string]any{"type": "number", "maximum": float64(100)}}},
			wantErr: false,
		},
		{
			name:    "maximum violated",
			args:    map[string]any{"count": 150},
			schema:  map[string]any{"properties": map[string]any{"count": map[string]any{"type": "number", "maximum": float64(100)}}},
			wantErr: true,
			errMsg:  "must be <=",
		},
		{
			name:    "pattern match",
			args:    map[string]any{"email": "test@example.com"},
			schema:  map[string]any{"properties": map[string]any{"email": map[string]any{"type": "string", "pattern": "^[\\w.-]+@[\\w.-]+\\.\\w+$"}}},
			wantErr: false,
		},
		{
			name:    "pattern no match",
			args:    map[string]any{"email": "invalid"},
			schema:  map[string]any{"properties": map[string]any{"email": map[string]any{"type": "string", "pattern": "^[\\w.-]+@[\\w.-]+\\.\\w+$"}}},
			wantErr: true,
			errMsg:  "does not match pattern",
		},
		{
			name:    "enum match",
			args:    map[string]any{"status": "active"},
			schema:  map[string]any{"properties": map[string]any{"status": map[string]any{"type": "string", "enum": []any{"active", "inactive", "pending"}}}},
			wantErr: false,
		},
		{
			name:    "enum no match",
			args:    map[string]any{"status": "unknown"},
			schema:  map[string]any{"properties": map[string]any{"status": map[string]any{"type": "string", "enum": []any{"active", "inactive", "pending"}}}},
			wantErr: true,
			errMsg:  "must be one of",
		},
		{
			name:    "array items valid",
			args:    map[string]any{"ids": []any{1, 2, 3}},
			schema:  map[string]any{"properties": map[string]any{"ids": map[string]any{"type": "array", "items": map[string]any{"type": "number"}}}},
			wantErr: false,
		},
		{
			name:    "array items invalid",
			args:    map[string]any{"ids": []any{1, "two", 3}},
			schema:  map[string]any{"properties": map[string]any{"ids": map[string]any{"type": "array", "items": map[string]any{"type": "number"}}}},
			wantErr: true,
			errMsg:  "expected number",
		},
		{
			name: "combined schema",
			args: map[string]any{"name": "test", "count": 5, "active": true},
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name":   map[string]any{"type": "string", "minLength": float64(1), "maxLength": float64(50)},
					"count":  map[string]any{"type": "number", "minimum": float64(1), "maximum": float64(100)},
					"active": map[string]any{"type": "boolean"},
				},
				"required": []any{"name"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.args, tt.schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && err != nil {
				if err.Error() == "" {
					t.Error("Validate() returned error with empty message")
				}
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{Field: "test", Message: "test error"}
	expected := "validation failed for 'test': test error"
	if err.Error() != expected {
		t.Errorf("ValidationError.Error() = %q, want %q", err.Error(), expected)
	}
}

func TestIsInteger(t *testing.T) {
	tests := []struct {
		val   any
		isInt bool
	}{
		{int(42), true},
		{int64(42), true},
		{int32(42), true},
		{float64(42), true},
		{float64(42.5), false},
		{"42", false},
		{true, false},
		{nil, false},
	}

	for _, tt := range tests {
		if got := isInteger(tt.val); got != tt.isInt {
			t.Errorf("isInteger(%v) = %v, want %v", tt.val, got, tt.isInt)
		}
	}
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		val    any
		result float64
		ok     bool
	}{
		{int(42), 42, true},
		{int64(42), 42, true},
		{int32(42), 42, true},
		{float64(42.5), 42.5, true},
		{float32(42.5), 42.5, true},
		{"42", 0, false},
		{true, 0, false},
	}

	for _, tt := range tests {
		if got, ok := toFloat64(tt.val); ok != tt.ok || (ok && got != tt.result) {
			t.Errorf("toFloat64(%v) = (%v, %v), want (%v, %v)", tt.val, got, ok, tt.result, tt.ok)
		}
	}
}

func TestValidatePath(t *testing.T) {
	validator := NewPathValidator(nil)

	tests := []struct {
		path    string
		wantErr bool
	}{
		{"src/main.go", false},
		{"", true},
		{"../etc/passwd", true}, // Block path traversal
		{"./src/main.go", false},
		{"path/with spaces", false},
	}

	for _, tt := range tests {
		err := validator.Validate(tt.path)
		if (err != nil) != tt.wantErr {
			t.Errorf("PathValidator.Validate(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
		}
	}
}

func TestValidateCommand(t *testing.T) {
	tests := []struct {
		cmd     string
		wantErr bool
	}{
		{"ls -la", false},
		{"git status", false},
		{"rm -rf /tmp/test", true}, // Blocks dangerous patterns
		{"", true},
		{":(){:|:&};:", true}, // Fork bomb
	}

	for _, tt := range tests {
		err := ValidateCommand(tt.cmd)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateCommand(%q) error = %v, wantErr %v", tt.cmd, err, tt.wantErr)
		}
	}
}

// BenchmarkValidate benchmarks the validation function
func BenchmarkValidate(b *testing.B) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name":   map[string]any{"type": "string", "minLength": float64(1), "maxLength": float64(50)},
			"count":  map[string]any{"type": "number", "minimum": float64(1), "maximum": float64(100)},
			"active": map[string]any{"type": "boolean"},
			"tags":   map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		},
		"required": []any{"name"},
	}

	args := map[string]any{
		"name":   "test-project",
		"count":  42,
		"active": true,
		"tags":   []any{"go", "backend", "api"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Validate(args, schema)
	}
}
