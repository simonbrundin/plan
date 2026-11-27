//! Inmatningsscheman för verktygen.
//!
//! Go: `internal/core/tools/schemas.go`.

use serde_json::json;

/// Schemat för `bash`.
pub fn bash_schema() -> serde_json::Value {
    json!({
        "type": "object",
        "properties": {
            "command": {
                "type": "string",
                "description": "The bash command to execute",
                "minLength": 1,
                "maxLength": 10000,
            },
            "timeout": {
                "type": "number",
                "description": "Timeout in seconds (default 60)",
                "minimum": 1,
                "maximum": 3600,
            },
            "cwd": {
                "type": "string",
                "description": "Working directory (default current)",
            },
        },
        "required": ["command"],
    })
}

/// Schemat för `read`.
pub fn fil_schema() -> serde_json::Value {
    json!({
        "type": "object",
        "properties": {
            "path": {
                "type": "string",
                "description": "Path to the file to read",
                "minLength": 1,
            },
            "offset": {
                "type": "number",
                "description": "Starting line number (1-indexed)",
                "minimum": 1,
            },
            "limit": {
                "type": "number",
                "description": "Maximum number of lines to read",
                "minimum": 1,
                "maximum": 10000,
            },
        },
        "required": ["path"],
    })
}

/// Schemat för `write`.
pub fn skriv_schema() -> serde_json::Value {
    json!({
        "type": "object",
        "properties": {
            "path": {
                "type": "string",
                "description": "Path to the file to write",
                "minLength": 1,
            },
            "content": {
                "type": "string",
                "description": "Content to write to the file",
            },
        },
        "required": ["path", "content"],
    })
}

/// Schemat för `edit`.
pub fn redigera_schema() -> serde_json::Value {
    json!({
        "type": "object",
        "properties": {
            "path": {
                "type": "string",
                "description": "Path to the file to edit",
                "minLength": 1,
            },
            "oldText": {
                "type": "string",
                "description": "The exact text to replace",
            },
            "newText": {
                "type": "string",
                "description": "The replacement text",
            },
        },
        "required": ["path", "oldText", "newText"],
    })
}

/// Schemat för `glob`.
pub fn glob_schema() -> serde_json::Value {
    json!({
        "type": "object",
        "properties": {
            "pattern": {
                "type": "string",
                "description": "Glob pattern to match (e.g., **/*.go)",
                "minLength": 1,
                "pattern": r"^[\w.*?/\[\]]+$",
            },
            "base": {
                "type": "string",
                "description": "Base directory to search from (default current)",
            },
        },
        "required": ["pattern"],
    })
}

/// Schemat för `grep`.
pub fn grep_schema() -> serde_json::Value {
    json!({
        "type": "object",
        "properties": {
            "pattern": {
                "type": "string",
                "description": "Regular expression pattern to search for",
                "minLength": 1,
            },
            "path": {
                "type": "string",
                "description": "Path to the file to search",
                "minLength": 1,
            },
            "ignoreCase": {
                "type": "boolean",
                "description": "Case insensitive search (default false)",
            },
        },
        "required": ["pattern", "path"],
    })
}

/// Schemat för `read_multiple`.
pub fn läs_flera_schema() -> serde_json::Value {
    json!({
        "type": "object",
        "properties": {
            "paths": {
                "type": "array",
                "items": {
                    "type": "string",
                    "minLength": 1,
                },
                "description": "Array of file paths to read",
                "minLength": 1,
                "maxLength": 100,
            },
        },
        "required": ["paths"],
    })
}
