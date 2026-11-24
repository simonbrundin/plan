package tools

// Input schemas for all tools

func bashInputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"command": map[string]any{
				"type":        "string",
				"description": "The bash command to execute",
				"minLength":   float64(1),
				"maxLength":   float64(10000),
			},
			"timeout": map[string]any{
				"type":        "number",
				"description": "Timeout in seconds (default 60)",
				"minimum":     float64(1),
				"maximum":     float64(3600),
			},
			"cwd": map[string]any{
				"type":        "string",
				"description": "Working directory (default current)",
			},
		},
		"required": []any{"command"},
	}
}

func fileInputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "Path to the file to read",
				"minLength":   float64(1),
			},
			"offset": map[string]any{
				"type":        "number",
				"description": "Starting line number (1-indexed)",
				"minimum":     float64(1),
			},
			"limit": map[string]any{
				"type":        "number",
				"description": "Maximum number of lines to read",
				"minimum":     float64(1),
				"maximum":     float64(10000),
			},
		},
		"required": []any{"path"},
	}
}

func writeInputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "Path to the file to write",
				"minLength":   float64(1),
			},
			"content": map[string]any{
				"type":        "string",
				"description": "Content to write to the file",
			},
		},
		"required": []any{"path", "content"},
	}
}

func editInputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "Path to the file to edit",
				"minLength":   float64(1),
			},
			"oldText": map[string]any{
				"type":        "string",
				"description": "The exact text to replace",
			},
			"newText": map[string]any{
				"type":        "string",
				"description": "The replacement text",
			},
		},
		"required": []any{"path", "oldText", "newText"},
	}
}

func globInputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"pattern": map[string]any{
				"type":        "string",
				"description": "Glob pattern to match (e.g., **/*.go)",
				"minLength":   float64(1),
				"pattern":     `^[\w.*?/\[\]]+$`,
			},
			"base": map[string]any{
				"type":        "string",
				"description": "Base directory to search from (default current)",
			},
		},
		"required": []any{"pattern"},
	}
}

func grepInputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"pattern": map[string]any{
				"type":        "string",
				"description": "Regular expression pattern to search for",
				"minLength":   float64(1),
			},
			"path": map[string]any{
				"type":        "string",
				"description": "Path to the file to search",
				"minLength":   float64(1),
			},
			"ignoreCase": map[string]any{
				"type":        "boolean",
				"description": "Case insensitive search (default false)",
			},
		},
		"required": []any{"pattern", "path"},
	}
}

func readMultipleInputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"paths": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type":      "string",
					"minLength": float64(1),
				},
				"description": "Array of file paths to read",
				"minLength":   float64(1),
				"maxLength":   float64(100),
			},
		},
		"required": []any{"paths"},
	}
}
