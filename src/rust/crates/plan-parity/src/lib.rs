//! Paritetsverktyg för migreringen från Go till Rust.
//!
//! Under migreringen ska den Rust-tjänsten ha exakt samma vägar och
//! metoder som den Go-tjänst den ersätter. Det här verktyget läser
//! route-registreringarna ur Go-källan, så att skillnader upptäcks i CI
//! i stället för i produktion.
//!
//! Två sorters fel hittas:
//!
//! * en route som finns i Go men inte i Rust (en glömd endpoint),
//! * en route som finns i Rust men inte i Go (en endpoint som driftar).
//!
//! Go registrerar de flesta vägarna på en grupp (`api := r.Group("/api/v1")`),
//! vilket gör att sökvägarna i källan är relativa. Gruppenamnet och
//! prefixet plockas ut här, så att jämförelsen sker på fullständiga
//! sökvägar.

use std::collections::BTreeSet;

/// Läser route-registreringarna ur Go-källan.
///
/// Returnerar `(metod, sökväg)` med axum-stil: `:id` blir `{id}` och
/// `*any` blir `{*any}`.
///
/// Parsningen sker över hela källan, inte rad för rad - en route kan
/// registreras över flera rader (`r.GET("/docs/*any", openapiui.WrapHandler(`).
pub fn go_routes(kalla: &str) -> BTreeSet<(String, String)> {
    let grupper = hitta_grupper(kalla);
    let mut routes = BTreeSet::new();

    for (namn, metod, sokvag) in hitta_anrop(kalla) {
        // Endast grupper registrerar routes - handler-anrop ignoreras.
        let prefix = match grupper.iter().rev().find(|(grupp, _)| grupp == &namn) {
            Some((_, prefix)) => prefix.clone(),
            None => continue,
        };

        routes.insert((metod, normalisera(&format!("{prefix}{sokvag}"))));
    }

    routes
}

/// Hittar grupperna: `api := r.Group("/api/v1")`.
///
/// Returnerar gruppnamn och prefix i den ordning de deklareras.
fn hitta_grupper(kalla: &str) -> Vec<(String, String)> {
    let mut grupper = vec![("r".to_string(), String::new())];
    let mut position = 0;

    while let Some(funnet) = kalla[position..].find(".Group(") {
        let start = position + funnet;
        let efter = start + ".Group(".len();

        // Prefixet är det första citattecknet efter ".Group(".
        let Some(citat_start) = kalla[efter..].find('"').map(|index| efter + index) else {
            break;
        };
        let Some(slut) = kalla[citat_start + 1..]
            .find('"')
            .map(|index| citat_start + 1 + index)
        else {
            break;
        };
        let prefix = kalla[citat_start + 1..slut].to_string();

        // Variabelnamnet står före `:=` - gå bakåt från ".Group(".
        let mut index = start;
        index = gå_bakåt_över_identifikator(kalla, index);
        index = gå_bakåt_över_blank(kalla, index);

        let mut namn = String::new();
        if kalla[..index].ends_with(":=") {
            index -= 2;
            index = gå_bakåt_över_blank(kalla, index);
            let början = gå_bakåt_över_identifikator(kalla, index);
            namn = kalla[början..index].to_string();
        } else if kalla[..index].ends_with('=') {
            index -= 1;
            index = gå_bakåt_över_blank(kalla, index);
            let början = gå_bakåt_över_identifikator(kalla, index);
            namn = kalla[början..index].to_string();
        }

        if !namn.is_empty() {
            grupper.push((namn, prefix));
        }

        position = slut + 1;
    }

    grupper
}

/// Går bakåt över identifierare (bokstäver, siffror, understreck).
fn gå_bakåt_över_identifikator(kalla: &str, index: usize) -> usize {
    let mut index = index;
    while index > 0 {
        let tecken = kalla[..index].chars().next_back().unwrap();
        if tecken.is_alphanumeric() || tecken == '_' {
            index -= tecken.len_utf8();
        } else {
            break;
        }
    }
    index
}

/// Går bakåt över blanksteg.
fn gå_bakåt_över_blank(kalla: &str, index: usize) -> usize {
    let mut index = index;
    while index > 0 {
        let tecken = kalla[..index].chars().next_back().unwrap();
        if tecken.is_whitespace() {
            index -= tecken.len_utf8();
        } else {
            break;
        }
    }
    index
}

/// Hittar alla route-anrop: `<grupp>.<METOD>("<sökväg>")`.
fn hitta_anrop(kalla: &str) -> Vec<(String, String, String)> {
    let mut anrop = Vec::new();
    let metoder = ["GET", "POST", "PUT", "PATCH", "DELETE"];

    for metod in metoder {
        let mönster = format!(".{metod}(");
        let mut position = 0;

        while let Some(funnet) = kalla[position..].find(&mönster) {
            let start = position + funnet;
            let efter = start + mönster.len();

            // Sökvägen är det första citattecknet efter "(".
            let rest = &kalla[efter..];
            let Some(citat_start) = rest.find('"') else {
                break;
            };
            let Some(slut) = rest[citat_start + 1..].find('"') else {
                break;
            };
            let sokvag = rest[citat_start + 1..citat_start + 1 + slut].to_string();

            // Gruppnamnet står direkt före punkten.
            let början = gå_bakåt_över_identifikator(kalla, start);
            let namn = kalla[början..start].to_string();

            anrop.push((namn, metod.to_string(), sokvag));
            position = efter + citat_start + slut + 2;
        }
    }

    anrop
}

/// Gör en Go-sökväg jämförbar med en axum-sökväg.
pub fn normalisera(sokvag: &str) -> String {
    sokvag
        .split('/')
        .map(|del| {
            if let Some(namn) = del.strip_prefix(':') {
                format!("{{{namn}}}")
            } else if let Some(namn) = del.strip_prefix('*') {
                format!("{{*{namn}}}")
            } else {
                del.to_string()
            }
        })
        .collect::<Vec<_>>()
        .join("/")
}

/// Jämför Rust-tabellen mot Go-vägarna.
///
/// `rust_vägar` är `(metod, sökväg)`-par. Returnerar vad som saknas i
/// Rust respektive vad som finns i Rust men inte i Go.
pub fn jämför(
    go_vägar: &BTreeSet<(String, String)>,
    rust_vägar: &[(String, String)],
) -> (Vec<String>, Vec<String>) {
    let rust: BTreeSet<(String, String)> = rust_vägar.iter().cloned().collect();

    let saknas: Vec<String> = go_vägar
        .iter()
        .filter(|(_, sokvag)| !sokvag.starts_with("/docs"))
        .filter(|(metod, sokvag)| !rust.contains(&(metod.clone(), sokvag.clone())))
        .map(|(metod, sokvag)| format!("{metod} {sokvag}"))
        .collect();

    let extra: Vec<String> = rust_vägar
        .iter()
        .filter(|(metod, sokvag)| !go_vägar.contains(&(metod.clone(), sokvag.clone())))
        .map(|(metod, sokvag)| format!("{metod} {sokvag}"))
        .collect();

    (saknas, extra)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn gruppens_prefix_laggs_på() {
        let kalla = r#"
        r.GET("/health", h)
        api := r.Group("/api/v1")
        api.GET("/goals", h)
        api.PATCH("/goals/:id", h)
        "#;

        let routes = go_routes(kalla);

        assert!(routes.contains(&("GET".to_string(), "/health".to_string())));
        assert!(routes.contains(&("GET".to_string(), "/api/v1/goals".to_string())));
        assert!(routes.contains(&("PATCH".to_string(), "/api/v1/goals/{id}".to_string())));
    }

    #[test]
    fn handlers_raknas_inte_som_routes() {
        let kalla = r#"
        api := r.Group("/api/v1")
        api.GET("/goals", goalHandler.GetGoals)
        "#;

        let routes = go_routes(kalla);

        assert_eq!(routes.len(), 1);
        assert!(routes.contains(&("GET".to_string(), "/api/v1/goals".to_string())));
    }

    #[test]
    fn skiljer_saknas_fran_extra() {
        let go: BTreeSet<(String, String)> = [
            ("GET".to_string(), "/api/v1/goals".to_string()),
            ("POST".to_string(), "/api/v1/goals".to_string()),
        ]
        .into_iter()
        .collect();

        let rust = vec![
            ("GET".to_string(), "/api/v1/goals".to_string()),
            ("DELETE".to_string(), "/api/v1/goals/{id}".to_string()),
        ];

        let (saknas, extra) = jämför(&go, &rust);

        assert_eq!(saknas, vec!["POST /api/v1/goals".to_string()]);
        assert_eq!(extra, vec!["DELETE /api/v1/goals/{id}".to_string()]);
    }
}
