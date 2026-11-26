//! Taggar.
//!
//! Port av `src/api/internal/models/tag.go`. Namnen normaliseras: gemener,
//! inga dubbla blanksteg och inget inledande "#". Det gör att "Ute" och
//! "ute " blir samma tagg.

use serde::{Deserialize, Serialize};
use utoipa::ToSchema;

/// Längsta tillåtna taggnamn (antal tecken).
pub const MAX_TAG_NAME_LENGTH: usize = 32;

/// En etikett som kan sättas på ett mål.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, ToSchema)]
pub struct Tag {
    /// Taggnamnet.
    pub name: String,
    /// Antalet mål som har taggen.
    pub goal_count: i32,
}

/// Normaliserar ett taggnamn.
///
/// Returnerar ett felmeddelande som i Go: "invalid tag name: ...".
pub fn normalize_tag_name(raw: &str) -> Result<String, String> {
    let mut namn = raw.trim().to_string();

    // "#ute" och "ute" ska bli samma tagg - #-prefixet är bara en
    // bekvämlighet i gränssnittet.
    if let Some(restad) = namn.strip_prefix('#') {
        namn = restad.to_string();
    }

    namn = namn.to_lowercase();

    // Flera blanksteg blir ett.
    namn = namn.split_whitespace().collect::<Vec<&str>>().join(" ");

    if namn.is_empty() {
        return Err("invalid tag name: namnet får inte vara tomt".to_string());
    }

    if namn.chars().count() > MAX_TAG_NAME_LENGTH {
        return Err(format!(
            "invalid tag name: namnet får vara max {MAX_TAG_NAME_LENGTH} tecken"
        ));
    }

    Ok(namn)
}

/// Normaliserar en lista med taggnamn och tar bort dubbletter.
///
/// Ordningen behålls - den första förekomsten vinner.
pub fn normalize_tag_names(raw: &[String]) -> Result<Vec<String>, String> {
    let mut sedda: Vec<String> = Vec::new();
    let mut ut: Vec<String> = Vec::with_capacity(raw.len());

    for rått in raw {
        let namn = normalize_tag_name(rått)?;

        if sedda.contains(&namn) {
            continue;
        }

        sedda.push(namn.clone());
        ut.push(namn);
    }

    Ok(ut)
}

/// Request-body för att ersätta ett måls taggar.
///
/// En tom lista tar bort alla taggar. Namn som inte finns ännu skapas.
#[derive(Debug, Clone, Deserialize, ToSchema)]
pub struct SetGoalTagsRequest {
    /// Taggnamnen.
    #[serde(default)]
    pub tags: Vec<String>,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn namn_normaliseras() {
        assert_eq!(normalize_tag_name("Ute").unwrap(), "ute");
        assert_eq!(normalize_tag_name("#ute").unwrap(), "ute");
        assert_eq!(normalize_tag_name("  ute  ").unwrap(), "ute");
        assert_eq!(normalize_tag_name("Ute   hem").unwrap(), "ute hem");
    }

    #[test]
    fn tomma_och_langa_namn_avvisas() {
        assert_eq!(
            normalize_tag_name("   ").unwrap_err(),
            "invalid tag name: namnet får inte vara tomt"
        );
        assert_eq!(
            normalize_tag_name(&"a".repeat(33)).unwrap_err(),
            "invalid tag name: namnet får vara max 32 tecken"
        );
        assert!(normalize_tag_name(&"a".repeat(32)).is_ok());
    }

    #[test]
    fn dubbletter_tas_bort_ur_listan() {
        let namn = vec![
            "Ute".to_string(),
            "ute".to_string(),
            "hem".to_string(),
            "#hem".to_string(),
        ];

        assert_eq!(
            normalize_tag_names(&namn).unwrap(),
            vec!["ute".to_string(), "hem".to_string()]
        );
    }
}
