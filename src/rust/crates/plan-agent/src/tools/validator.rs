//! Validering av verktygsargument.
//!
//! Go: `internal/core/tools/validator_core.go`, `validator_types.go`.
//! Schemat är ett JSON-schema-liknande subset: typ, required,
//! properties, minLength, maxLength, minimum, maximum, pattern, enum
//! och items.

use regex::Regex;
use serde_json::Value;

/// Ett valideringsfel.
#[derive(Debug, Clone)]
pub struct ValidationError {
    /// Fältet som var fel.
    pub field: String,
    /// Beskrivningen av felet.
    pub message: String,
}

impl std::fmt::Display for ValidationError {
    fn fmt(&self, form: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(
            form,
            "validation failed for '{}': {}",
            self.field, self.message
        )
    }
}

impl std::error::Error for ValidationError {}

/// Typen på ett värde - så som Go skriver den.
fn typnamn(värde: &Value) -> &'static str {
    match värde {
        Value::Null => "<nil>",
        Value::Bool(_) => "bool",
        Value::Number(_) => "float64",
        Value::String(_) => "string",
        Value::Array(_) => "[]interface {}",
        Value::Object(_) => "map[string]interface {}",
    }
}

/// Gör om ett värde till ett tal.
fn till_f64(värde: &Value) -> Option<f64> {
    värde.as_f64()
}

/// Om värdet är ett heltal.
fn är_heltal(värde: &Value) -> bool {
    match värde {
        Value::Number(num) => num.as_f64().map(|f| f == f.trunc()).unwrap_or(false),
        _ => false,
    }
}

/// Validerar argumenten mot ett schema.
///
/// Go: `tools.ValidateArguments`.
pub fn validera_argument(args: &Value, schema: &Value) -> Result<(), ValidationError> {
    if schema.is_null() {
        return Ok(());
    }

    // Obligatoriska fält.
    if let Some(obligatoriska) = schema.get("required").and_then(|v| v.as_array()) {
        for obligatoriskt in obligatoriska {
            let Some(namn) = obligatoriskt.as_str() else {
                continue;
            };

            if args.get(namn).is_none() {
                return Err(ValidationError {
                    field: namn.to_string(),
                    message: "required field missing".to_string(),
                });
            }
        }
    }

    // Egenskaper.
    if let Some(egenskaper) = schema.get("properties").and_then(|v| v.as_object()) {
        for (namn, egenskap) in egenskaper {
            if !egenskap.is_object() {
                continue;
            }

            if let Some(värde) = args.get(namn) {
                validera_värde(namn, värde, egenskap)?;
            }
        }
    }

    Ok(())
}

/// Validerar ett värde mot ett schema.
fn validera_värde(namn: &str, värde: &Value, schema: &Value) -> Result<(), ValidationError> {
    // Typen.
    if let Some(typ) = schema.get("type").and_then(|v| v.as_str()) {
        validera_typ(namn, värde, typ)?;
    }

    // Uppslagningsvärden.
    if let Some(ramper) = schema.get("enum").and_then(|v| v.as_array()) {
        validera_enum(namn, värde, ramper)?;
    }

    // Strängspecifikationer.
    if let Some(text) = värde.as_str() {
        validera_text(namn, text, schema)?;
    }

    // Talspecifikationer.
    if let Some(tal) = till_f64(värde) {
        validera_tal(namn, tal, schema)?;
    }

    // Arrayens poster.
    if let Some(lista) = värde.as_array() {
        if let Some(poster) = schema.get("items").filter(|v| v.is_object()) {
            for (index, post) in lista.iter().enumerate() {
                validera_värde(&format!("{namn}[{index}]"), post, poster)?;
            }
        }
    }

    Ok(())
}

/// Validerar typen på ett värde.
fn validera_typ(namn: &str, värde: &Value, förväntad: &str) -> Result<(), ValidationError> {
    let stämmer = match förväntad {
        "string" => värde.is_string(),
        "number" => till_f64(värde).is_some(),
        "integer" => är_heltal(värde),
        "boolean" => värde.is_boolean(),
        "array" => värde.is_array(),
        "object" => värde.is_object(),
        _ => true,
    };

    if stämmer {
        return Ok(());
    }

    Err(ValidationError {
        field: namn.to_string(),
        message: format!("expected {förväntad}, got {}", typnamn(värde)),
    })
}

/// Validerar att värdet finns i uppslagningslistan.
fn validera_enum(namn: &str, värde: &Value, ramper: &[Value]) -> Result<(), ValidationError> {
    for rampe in ramper {
        if rampe == värde {
            return Ok(());
        }
    }

    Err(ValidationError {
        field: namn.to_string(),
        message: format!(
            "value must be one of {}, got {}",
            go_values(ramper),
            go_value(värde)
        ),
    })
}

/// Formaterar en lista av värden så som Go gör.
fn go_values(värden: &[Value]) -> String {
    let delar: Vec<String> = värden.iter().map(go_value).collect();

    format!("[{}]", delar.join(" "))
}

/// Formaterar ett värde så som Go gör.
fn go_value(värde: &Value) -> String {
    match värde {
        Value::String(text) => text.clone(),
        Value::Null => "<nil>".to_string(),
        Value::Bool(värde) => värde.to_string(),
        Value::Number(num) => num.to_string(),
        Value::Array(_) => "[...]".to_string(),
        Value::Object(_) => "map[...]".to_string(),
    }
}

/// Validerar strängens längd och mönster.
fn validera_text(namn: &str, text: &str, schema: &Value) -> Result<(), ValidationError> {
    if let Some(minsta) = schema.get("minLength").and_then(|v| v.as_f64()) {
        if text.len() < minsta as usize {
            return Err(ValidationError {
                field: namn.to_string(),
                message: format!(
                    "string must be at least {} characters, got {}",
                    minsta as usize,
                    text.len()
                ),
            });
        }
    }

    if let Some(maxta) = schema.get("maxLength").and_then(|v| v.as_f64()) {
        if text.len() > maxta as usize {
            return Err(ValidationError {
                field: namn.to_string(),
                message: format!(
                    "string must be at most {} characters, got {}",
                    maxta as usize,
                    text.len()
                ),
            });
        }
    }

    if let Some(mönster) = schema.get("pattern").and_then(|v| v.as_str()) {
        match Regex::new(mönster) {
            Ok(regex) => {
                if !regex.is_match(text) {
                    return Err(ValidationError {
                        field: namn.to_string(),
                        message: format!("string does not match pattern '{mönster}'"),
                    });
                }
            }
            Err(fel) => {
                return Err(ValidationError {
                    field: namn.to_string(),
                    message: format!("invalid pattern: {fel}"),
                });
            }
        }
    }

    Ok(())
}

/// Validerar talets gränser.
fn validera_tal(namn: &str, tal: f64, schema: &Value) -> Result<(), ValidationError> {
    if let Some(minsta) = schema.get("minimum").and_then(|v| v.as_f64()) {
        if tal < minsta {
            return Err(ValidationError {
                field: namn.to_string(),
                message: format!("number must be >= {minsta}, got {tal}"),
            });
        }
    }

    if let Some(maxta) = schema.get("maximum").and_then(|v| v.as_f64()) {
        if tal > maxta {
            return Err(ValidationError {
                field: namn.to_string(),
                message: format!("number must be <= {maxta}, got {tal}"),
            });
        }
    }

    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::json;

    #[test]
    fn obligatoriskt_falt_saknas() {
        let schema = json!({
            "type": "object",
            "properties": { "path": { "type": "string" } },
            "required": ["path"],
        });

        let fel = validera_argument(&json!({}), &schema).unwrap_err();

        assert_eq!(fel.field, "path");
        assert_eq!(fel.message, "required field missing");
    }

    #[test]
    fn fel_typ_rapporteras_som_i_go() {
        let schema = json!({
            "type": "object",
            "properties": { "count": { "type": "number" } },
        });

        let fel = validera_argument(&json!({ "count": "text" }), &schema).unwrap_err();

        assert_eq!(fel.message, "expected number, got string");
    }

    #[test]
    fn langd_granser_kontrolleras() {
        let schema = json!({
            "type": "object",
            "properties": { "command": { "type": "string", "minLength": 1 } },
        });

        let fel = validera_argument(&json!({ "command": "" }), &schema).unwrap_err();

        assert_eq!(fel.message, "string must be at least 1 characters, got 0");
    }

    #[test]
    fn enum_godkanns() {
        let schema = json!({
            "type": "object",
            "properties": { "mode": { "enum": ["a", "b"] } },
        });

        assert!(validera_argument(&json!({ "mode": "a" }), &schema).is_ok());

        let fel = validera_argument(&json!({ "mode": "c" }), &schema).unwrap_err();

        assert_eq!(fel.message, "value must be one of [a b], got c");
    }
}
