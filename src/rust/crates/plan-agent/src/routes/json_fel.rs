//! Go-kompatibla felmeddelanden för trasiga JSON-kroppar.
//!
//! Go använder `encoding/json`, som skriver andra meddelanden än serde.
//! Klienten jämför texterna - därför skrivs meddelandena om här.

/// Go:s felmeddelande för en kropp som inte går att tolka.
///
/// `typnamn` är Go-typens namn - det står i meddelandet när kroppen har
/// rätt form men fel typ.
pub fn json_fel(body: &[u8], typnamn: &str, err: &serde_json::Error) -> String {
    // En tom kropp ger "EOF" i Go.
    if body.iter().all(|byte| byte.is_ascii_whitespace()) {
        return "EOF".to_string();
    }

    let tecken = body
        .iter()
        .find(|byte| !byte.is_ascii_whitespace())
        .copied()
        .unwrap_or(b' ');

    // Fel typ i en befintlig struktur - Go säger vilken typ den fick.
    if err.is_data() {
        return format!(
            "json: cannot unmarshal {} into Go value of type {}",
            go_typnamn(tecken),
            typnamn
        );
    }

    // Ett ogiltigt första tecken - Go letar efter ett värde.
    if tecken.is_ascii_alphabetic() {
        let (literal, forväntat) = match tecken.to_ascii_lowercase() {
            b't' => ("true", 'u'),
            b'f' => ("false", 'a'),
            b'n' => ("null", 'u'),
            _ => ("", ' '),
        };

        if !literal.is_empty() {
            return format!(
                "invalid character '{}' in literal {} (expecting '{}')",
                tecken as char, literal, forväntat
            );
        }
    }

    if !is_json_start(tecken) {
        return format!(
            "invalid character '{}' looking for beginning of value",
            tecken as char
        );
    }

    // Resten av fallen är sällsynta - Serdes meddelande är bättre än
    // ingenting, men den är inte Go:s.
    match err.to_string().split_once(" at line ") {
        Some((del, _)) => del.to_string(),
        None => err.to_string(),
    }
}

/// Om tecknet kan inleda ett JSON-värde.
fn is_json_start(tecken: u8) -> bool {
    matches!(
        tecken,
        b'{' | b'[' | b'"' | b'-' | b'0'..=b'9' | b't' | b'f' | b'n'
    )
}

/// Go:s namn på JSON-typer.
fn go_typnamn(tecken: u8) -> &'static str {
    match tecken {
        b'"' => "string",
        b'{' => "object",
        b'[' => "array",
        b't' | b'f' => "bool",
        b'n' => "null",
        _ => "number",
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn tom_kropp_ger_eof() {
        let err = serde_json::from_slice::<serde_json::Value>(b"").unwrap_err();

        assert_eq!(json_fel(b"", "Test", &err), "EOF");
    }

    #[test]
    fn trasig_literal_ger_go_meddelandet() {
        let err = serde_json::from_slice::<serde_json::Value>(b"trasigt").unwrap_err();

        assert_eq!(
            json_fel(b"trasigt", "Test", &err),
            "invalid character 't' in literal true (expecting 'u')"
        );
    }

    #[test]
    fn ogiltigt_forsta_tecken_ser_go_meddelandet() {
        let err = serde_json::from_slice::<serde_json::Value>(b"@").unwrap_err();

        assert_eq!(
            json_fel(b"@", "Test", &err),
            "invalid character '@' looking for beginning of value"
        );
    }

    #[test]
    fn fel_typ_namnar_go_typen() {
        #[derive(serde::Deserialize, Debug)]
        struct Test {}

        let err = serde_json::from_slice::<Test>(b"\"text\"").unwrap_err();

        assert_eq!(
            json_fel(b"\"text\"", "handlers.Test", &err),
            "json: cannot unmarshal string into Go value of type handlers.Test"
        );
    }
}
