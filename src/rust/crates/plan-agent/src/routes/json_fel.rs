//! Go-kompatibla felmeddelanden för trasiga JSON-kroppar.
//!
//! Go använder `encoding/json`, som skriver andra meddelanden än serde.
//! Klienten jämför texterna - därför skrivs meddelandena om här.
//!
//! Meddelandena byggs upp så här:
//!
//! - `json: cannot unmarshal <typ> into Go value of type <paket>.<typ>`
//! - `json: cannot unmarshal <typ> into Go struct field <Typ>.<fält> of type <typ>`
//!
//! Serdes felmeddelande innehåller både den faktiska typen och den
//! förväntade - dem översätts till Go:s format.

use serde_path_to_error::Error as Sökvägsfel;

/// Go:s felmeddelande för en kropp som inte går att tolka.
///
/// `typnamn` är Go-typens namn med paket - som "handlers.CreateMessageRequest".
pub fn json_fel(body: &[u8], typnamn: &str, fel: &Sökvägsfel<serde_json::Error>) -> String {
    // En tom kropp ger "EOF" i Go.
    if body.iter().all(|byte| byte.is_ascii_whitespace()) {
        return "EOF".to_string();
    }

    let tecken = body
        .iter()
        .find(|byte| !byte.is_ascii_whitespace())
        .copied()
        .unwrap_or(b' ');

    // Ett syntaxfel - Go har egna meddelanden för dem.
    if fel.inner().is_syntax() || fel.inner().is_eof() {
        if let Some(meddelande) = go_syntaxfel(body) {
            return meddelande;
        }
    }

    // Fel i ett fält - Go skriver vilket fält det var.
    let sökväg = fel.path().to_string();
    let har_fält = fel.path().iter().count() > 0 && !sökväg.contains('?');

    if har_fält {
        let (faktisk, förväntad) = beskriv_typer(fel.inner());

        let (paket, typ) = dela_typnamn(typnamn);

        let förväntad = if förväntad.starts_with("struct ") {
            let namn = förväntad.trim_start_matches("struct ");
            format!("{paket}.{namn}")
        } else {
            förväntad
        };

        return format!(
            "json: cannot unmarshal {faktisk} into Go struct field {typ}.{sökväg} of type {förväntad}"
        );
    }

    // Fel typ i hela kroppen - Go skriver vilken typ den fick.
    if fel.inner().is_data() {
        let (faktisk, _) = beskriv_typer(fel.inner());

        return format!("json: cannot unmarshal {faktisk} into Go value of type {typnamn}");
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
    match fel.inner().to_string().split_once(" at line ") {
        Some((del, _)) => del.to_string(),
        None => fel.inner().to_string(),
    }
}

/// Delar upp ett typnamn i paket och typ.
///
/// Go skriver paketnamnet i vissa av sina meddelanden och inte i andra.
fn dela_typnamn(typnamn: &str) -> (&str, &str) {
    match typnamn.split_once('.') {
        Some((paket, typ)) => (paket, typ),
        None => ("", typnamn),
    }
}

/// Beskriver den faktiska och den förväntade typen.
///
/// Serde skriver "invalid type: integer `123`, expected a string".
/// Go skriver "number" och "string".
fn beskriv_typer(fel: &serde_json::Error) -> (String, String) {
    let text = fel.to_string();

    let Some(rest) = text.strip_prefix("invalid type: ") else {
        return ("object".to_string(), "interface {}".to_string());
    };

    let (typ, rest) = match rest.split_once(", expected ") {
        Some((typ, rest)) => (typ.trim(), rest.trim()),
        None => (rest.trim(), ""),
    };

    let förväntad = rest.split(" at line ").next().unwrap_or("").trim();

    (go_typnamn_faktisk(typ), go_typnamn_förväntad(förväntad))
}

/// Go:s namn på den faktiska typen.
fn go_typnamn_faktisk(typ: &str) -> String {
    if let Some(värde) = typ.strip_prefix("floating point ") {
        // Go skriver värdet med - "number 1.5".
        return format!("number {}", trimma_backticks(värde));
    }

    if typ.starts_with("integer") {
        return "number".to_string();
    }

    if typ.starts_with("string") {
        return "string".to_string();
    }

    if typ.starts_with("boolean") {
        return "bool".to_string();
    }

    match typ {
        "sequence" => "array".to_string(),
        "map" => "object".to_string(),
        "null" => "null".to_string(),
        _ => "object".to_string(),
    }
}

/// Go:s namn på den förväntade typen.
fn go_typnamn_förväntad(förväntad: &str) -> String {
    match förväntad {
        "a string" => "string".to_string(),
        "a boolean" => "bool".to_string(),
        "a sequence" => "[]interface {}".to_string(),
        "a map" => "map[string]interface {}".to_string(),
        "any" | "a value" => "interface {}".to_string(),
        "i64" => "int64".to_string(),
        "u64" | "usize" => "uint64".to_string(),
        "i32" => "int32".to_string(),
        "u32" => "uint32".to_string(),
        "f64" => "float64".to_string(),
        "f32" => "float32".to_string(),
        "" => "interface {}".to_string(),
        annan => annan.to_string(),
    }
}

/// Tar bort backticks runt ett värde.
fn trimma_backticks(värde: &str) -> String {
    värde.trim_matches('`').to_string()
}

/// Om tecknet kan inleda ett JSON-värde.
fn is_json_start(tecken: u8) -> bool {
    matches!(
        tecken,
        b'{' | b'[' | b'"' | b'-' | b'0'..=b'9' | b't' | b'f' | b'n'
    )
}

/// Tolkar en kropp på Go:s sätt - med Go:s felmeddelanden.
///
/// `typnamn` är Go-typens namn med paket - som "handlers.CreateMessageRequest".
pub fn tolka<T: serde::de::DeserializeOwned + Default>(
    body: &[u8],
    typnamn: &str,
) -> Result<T, String> {
    // Go läser kroppen som ett objekt - en lista eller ett enstaka
    // värde ger ett annat felmeddelande än serde.
    // Ett syntaxfel fångas först - då är kroppen inte JSON alls.
    if let Some(meddelande) = go_syntaxfel(body) {
        return Err(meddelande);
    }

    if let Ok(värde) = serde_json::from_slice::<serde_json::Value>(body) {
        // Go läser "null" som ett tomt objekt - fälten lämnas som de var.
        if värde.is_null() {
            return Ok(T::default());
        }

        if !värde.is_object() {
            let typ = match &värde {
                serde_json::Value::Null => "null",
                serde_json::Value::Bool(_) => "bool",
                serde_json::Value::Number(_) => "number",
                serde_json::Value::String(_) => "string",
                serde_json::Value::Array(_) => "array",
                serde_json::Value::Object(_) => "object",
            };

            return Err(format!(
                "json: cannot unmarshal {typ} into Go value of type {typnamn}"
            ));
        }
    }

    let mut tolkare = serde_json::Deserializer::from_slice(body);

    match serde_path_to_error::deserialize::<_, T>(&mut tolkare) {
        Ok(värde) => Ok(värde),
        Err(fel) => Err(json_fel(body, typnamn, &fel)),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[derive(serde::Deserialize, Debug)]
    #[allow(dead_code)]
    struct Mål {
        #[serde(rename = "id")]
        id: i64,
    }

    #[derive(serde::Deserialize, Debug)]
    #[allow(dead_code)]
    struct Test {
        content: String,
        #[serde(rename = "goal_context")]
        goal_context: Option<Mål>,
    }

    /// Tolkar en kropp och returnerar felet.
    fn felsök<T: serde::de::DeserializeOwned + std::fmt::Debug>(
        body: &[u8],
    ) -> Sökvägsfel<serde_json::Error> {
        let mut tolkare = serde_json::Deserializer::from_slice(body);

        match serde_path_to_error::deserialize::<_, T>(&mut tolkare) {
            Ok(_) => panic!("kroppen förväntades att vara trasig"),
            Err(fel) => fel,
        }
    }

    #[test]
    fn tom_kropp_ger_eof() {
        let fel = felsök::<serde_json::Value>(b"");

        assert_eq!(json_fel(b"", "handlers.Test", &fel), "EOF");
    }

    #[test]
    fn avbruten_json_ger_go_meddelandet() {
        let fel = felsök::<serde_json::Value>(b"{");

        assert_eq!(
            json_fel(b"{", "handlers.Test", &fel),
            "unexpected end of JSON input"
        );
    }

    #[test]
    fn trasig_literal_ger_go_meddelandet() {
        let fel = felsök::<serde_json::Value>(b"trasigt");

        assert_eq!(
            json_fel(b"trasigt", "handlers.Test", &fel),
            "invalid character 'a' in literal true (expecting 'u')"
        );
    }

    #[test]
    fn ogiltigt_forsta_tecken_ser_go_meddelandet() {
        let fel = felsök::<serde_json::Value>(b"@");

        assert_eq!(
            json_fel(b"@", "handlers.Test", &fel),
            "invalid character '@' looking for beginning of value"
        );
    }

    #[test]
    fn fel_typ_namnar_go_typen() {
        let fel = felsök::<Test>(b"\"text\"");

        assert_eq!(
            json_fel(b"\"text\"", "handlers.Test", &fel),
            "json: cannot unmarshal string into Go value of type handlers.Test"
        );
    }

    #[test]
    fn fel_i_fältet_namnar_fältet() {
        let fel = felsök::<Test>(b"{\"content\":123}");

        assert_eq!(
            json_fel(b"{\"content\":123}", "handlers.Test", &fel),
            "json: cannot unmarshal number into Go struct field Test.content of type string"
        );
    }

    #[test]
    fn fel_i_nestlat_fält_namnar_stigen() {
        let kropp = b"{\"content\":\"x\",\"goal_context\":{\"id\":\"a\"}}";
        let fel = felsök::<Test>(kropp);

        assert_eq!(
            json_fel(kropp, "handlers.Test", &fel),
            "json: cannot unmarshal string into Go struct field Test.goal_context.id of type int64"
        );
    }
}

/// Go:s felmeddelande för en trasig JSON.
///
/// Go använder `encoding/json`, som ger avvikande meddelanden jämfört
/// med serde - "invalid character 'x' looking for beginning of value"
/// i stället för serdes beskrivning.
///
/// Returnerar `None` om kroppen är giltig JSON.
pub fn go_syntaxfel(body: &[u8]) -> Option<String> {
    let mut index = 0;

    hoppa_ut(body, &mut index);

    if index >= body.len() {
        return Some("unexpected end of JSON input".to_string());
    }

    if let Some(fel) = go_värde(body, &mut index) {
        return Some(fel);
    }

    hoppa_ut(body, &mut index);

    if index < body.len() {
        return Some(format!(
            "invalid character '{}' after top-level value",
            body[index] as char
        ));
    }

    None
}

/// Hoppar över mellanrum.
fn hoppa_ut(body: &[u8], index: &mut usize) {
    while *index < body.len() && body[*index].is_ascii_whitespace() {
        *index += 1;
    }
}

/// Läser ett JSON-värde.
fn go_värde(body: &[u8], index: &mut usize) -> Option<String> {
    if *index >= body.len() {
        return Some("unexpected end of JSON input".to_string());
    }

    match body[*index] {
        b'{' => go_objekt(body, index),
        b'[' => go_lista(body, index),
        b'"' => go_sträng(body, index),
        b't' => go_literal(body, index, "true"),
        b'f' => go_literal(body, index, "false"),
        b'n' => go_literal(body, index, "null"),
        b'-' | b'0'..=b'9' => go_nummer(body, index),
        tecken => Some(format!(
            "invalid character '{}' looking for beginning of value",
            tecken as char
        )),
    }
}

/// Läser ett objekt.
fn go_objekt(body: &[u8], index: &mut usize) -> Option<String> {
    *index += 1;
    hoppa_ut(body, index);

    if *index >= body.len() {
        return Some("unexpected end of JSON input".to_string());
    }

    if body[*index] == b'}' {
        *index += 1;
        return None;
    }

    loop {
        hoppa_ut(body, index);

        if *index >= body.len() {
            return Some("unexpected end of JSON input".to_string());
        }

        if body[*index] != b'"' {
            return Some(format!(
                "invalid character '{}' looking for beginning of object key string",
                body[*index] as char
            ));
        }

        if let Some(fel) = go_sträng(body, index) {
            return Some(fel);
        }

        hoppa_ut(body, index);

        if *index >= body.len() {
            return Some("unexpected end of JSON input".to_string());
        }

        if body[*index] != b':' {
            return Some(format!(
                "invalid character '{}' after object key",
                body[*index] as char
            ));
        }

        *index += 1;
        hoppa_ut(body, index);

        if let Some(fel) = go_värde(body, index) {
            return Some(fel);
        }

        hoppa_ut(body, index);

        if *index >= body.len() {
            return Some("unexpected end of JSON input".to_string());
        }

        match body[*index] {
            b',' => *index += 1,
            b'}' => {
                *index += 1;
                return None;
            }
            tecken => {
                return Some(format!(
                    "invalid character '{}' after object key:value pair",
                    tecken as char
                ));
            }
        }
    }
}

/// Läser en lista.
fn go_lista(body: &[u8], index: &mut usize) -> Option<String> {
    *index += 1;
    hoppa_ut(body, index);

    if *index >= body.len() {
        return Some("unexpected end of JSON input".to_string());
    }

    if body[*index] == b']' {
        *index += 1;
        return None;
    }

    loop {
        hoppa_ut(body, index);

        if let Some(fel) = go_värde(body, index) {
            return Some(fel);
        }

        hoppa_ut(body, index);

        if *index >= body.len() {
            return Some("unexpected end of JSON input".to_string());
        }

        match body[*index] {
            b',' => *index += 1,
            b']' => {
                *index += 1;
                return None;
            }
            tecken => {
                return Some(format!(
                    "invalid character '{}' after array element",
                    tecken as char
                ));
            }
        }
    }
}

/// Läser en sträng.
fn go_sträng(body: &[u8], index: &mut usize) -> Option<String> {
    *index += 1;

    while *index < body.len() {
        match body[*index] {
            b'\\' => {
                *index += 2;
            }
            b'"' => {
                *index += 1;
                return None;
            }
            _ => *index += 1,
        }
    }

    Some("unexpected end of JSON input".to_string())
}

/// Läser ett literal - som "true", "false" eller "null".
fn go_literal(body: &[u8], index: &mut usize, ord: &str) -> Option<String> {
    let bytes = ord.as_bytes();

    for (position, förväntad) in bytes.iter().enumerate() {
        let index = *index + position;

        if index >= body.len() {
            return Some("unexpected end of JSON input".to_string());
        }

        if body[index] != *förväntad {
            return Some(format!(
                "invalid character '{}' in literal {ord} (expecting '{}')",
                body[index] as char, *förväntad as char
            ));
        }
    }

    *index += bytes.len();

    None
}

/// Läser ett tal.
///
/// Go följer JSON:s talregler - ett nollan一开始 nollan, och ett
/// tal måste ha siffror efter punkt och exponent.
fn go_nummer(body: &[u8], index: &mut usize) -> Option<String> {
    let mut i = *index;

    if body[i] == b'-' {
        i += 1;

        if i >= body.len() {
            return Some("unexpected end of JSON input".to_string());
        }
    }

    match body[i] {
        b'0' => {
            i += 1;
        }
        b'1'..=b'9' => {
            while i < body.len() && body[i].is_ascii_digit() {
                i += 1;
            }
        }
        tecken => {
            return Some(format!("invalid character '{tecken}' in numeric literal"));
        }
    }

    // Bråktal.
    if i < body.len() && body[i] == b'.' {
        i += 1;

        let start = i;

        while i < body.len() && body[i].is_ascii_digit() {
            i += 1;
        }

        if i == start {
            return Some("unexpected end of JSON input".to_string());
        }
    }

    // Exponent.
    if i < body.len() && (body[i] == b'e' || body[i] == b'E') {
        i += 1;

        if i < body.len() && (body[i] == b'+' || body[i] == b'-') {
            i += 1;
        }

        let start = i;

        while i < body.len() && body[i].is_ascii_digit() {
            i += 1;
        }

        if i == start {
            return Some("unexpected end of JSON input".to_string());
        }
    }

    *index = i;

    None
}

#[cfg(test)]
mod syntax_test {
    use super::go_syntaxfel;

    /// Go:s meddelanden för trasig JSON - insamlade från Go.
    #[test]
    fn go_meddelanden_matchar() {
        let fall = [
            (
                "{trasigt",
                "invalid character 't' looking for beginning of object key string",
            ),
            ("{\"a\" 1}", "invalid character '1' after object key"),
            (
                "{\"a\":}",
                "invalid character '}' looking for beginning of value",
            ),
            (
                "{\"a\":1,}",
                "invalid character '}' looking for beginning of object key string",
            ),
            ("[1,2", "unexpected end of JSON input"),
            (
                "{\"a\":1} extra",
                "invalid character 'e' after top-level value",
            ),
            ("nul", "unexpected end of JSON input"),
            ("tru", "unexpected end of JSON input"),
            ("{\"a\":\"b\"", "unexpected end of JSON input"),
            ("{", "unexpected end of JSON input"),
            ("}", "invalid character '}' looking for beginning of value"),
            ("@", "invalid character '@' looking for beginning of value"),
            ("", "unexpected end of JSON input"),
            (
                "nulx",
                "invalid character 'x' in literal null (expecting 'l')",
            ),
            ("truex", "invalid character 'x' after top-level value"),
            (
                "falsx",
                "invalid character 'x' in literal false (expecting 'e')",
            ),
            ("1x", "invalid character 'x' after top-level value"),
            ("-", "unexpected end of JSON input"),
            ("01", "invalid character '1' after top-level value"),
            ("[1,2}", "invalid character '}' after array element"),
            (
                "{,}",
                "invalid character ',' looking for beginning of object key string",
            ),
            (
                "[,]",
                "invalid character ',' looking for beginning of value",
            ),
            ("{\"a\"}", "invalid character '}' after object key"),
        ];

        for (kropp, förväntat) in fall {
            assert_eq!(
                go_syntaxfel(kropp.as_bytes()).as_deref(),
                Some(förväntat),
                "kroppen {kropp:?}"
            );
        }
    }

    #[test]
    fn giltig_json_ger_inget_fel() {
        assert_eq!(go_syntaxfel(b"{\"a\":[1,2],\"b\":null}"), None);
        assert_eq!(go_syntaxfel(b"[]"), None);
        assert_eq!(go_syntaxfel(b"\"text\""), None);
        assert_eq!(go_syntaxfel(b"1.5"), None);
    }
}
