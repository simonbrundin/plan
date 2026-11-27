//! Genererat index - den innehållsorienterade katalogen över wikin.
//!
//! Go: `internal/wiki/index.go`.

use std::collections::BTreeMap;
use std::path::Path;

use crate::note::Note;
use crate::page::{PageKind, KINDS};
use crate::wiki::Wiki;

/// Filen med det genererade indexet.
pub const INDEX_FILE: &str = "index.md";

/// Returnerar de sidor som ligger i en given typ, sorterade på titel.
///
/// Go: `Wiki.pagesOfKind`.
pub fn pages_of_kind(wiki: &Wiki, kind: PageKind) -> Vec<&Note> {
    let mut sidor: Vec<&Note> = wiki
        .notes
        .iter()
        .filter(|anteckning| kind_of(anteckning) == kind)
        .collect();

    sidor.sort_by_key(|a| a.title.to_lowercase());

    sidor
}

/// Gissar sidtypen utifrån frontmatter, med katalogen som fallback.
///
/// Go: `index.kindOf`.
pub fn kind_of(anteckning: &Note) -> PageKind {
    match anteckning.note_type.to_lowercase().as_str() {
        "entity" => return PageKind::Entity,
        "concept" => return PageKind::Concept,
        "summary" => return PageKind::Summary,
        "synthesis" => return PageKind::Synthesis,
        "note" => return PageKind::Note,
        _ => {}
    }

    match anteckning.id.split('/').next().unwrap_or_default() {
        "entities" => PageKind::Entity,
        "concepts" => PageKind::Concept,
        "summaries" => PageKind::Summary,
        "synthesis" => PageKind::Synthesis,
        _ => PageKind::Note,
    }
}

/// Returnerar sidtypen ett id hör till.
///
/// Go: `index.kindOf` för ett id.
pub fn kind_av_id(id: &str) -> PageKind {
    match id.split('/').next().unwrap_or_default() {
        "entities" => PageKind::Entity,
        "concepts" => PageKind::Concept,
        "summaries" => PageKind::Summary,
        "synthesis" => PageKind::Synthesis,
        _ => PageKind::Note,
    }
}

/// Returnerar en rad som beskriver en sida.
///
/// Frontmatter-fältet `summary` används om det finns, annars härleds
/// en kort beskrivning ur brödtextens första mening.
///
/// Go: `Note.SummaryLine`.
pub fn summary_line(anteckning: &Note) -> String {
    if !anteckning.summary.trim().is_empty() {
        return anteckning.summary.clone();
    }

    derive_summary(&anteckning.body)
}

/// Plockar ut första meningen i en text och kortar av den.
///
/// Go: `index.DeriveSummary`.
pub fn derive_summary(body: &str) -> String {
    for rad in body.split('\n') {
        let rad = rad.trim();

        if rad.is_empty() || rad.starts_with('#') || rad.starts_with('-') {
            continue;
        }

        if rad.starts_with('|') || rad.starts_with("```") {
            continue;
        }

        let mut summary = rad.to_string();

        if let Some(index) = summary.find(". ") {
            if index > 0 {
                summary = summary[..index + 1].to_string();
            }
        }

        let summary = summary.trim().to_string();

        if summary.chars().count() > 140 {
            let kort: String = summary.chars().take(137).collect();
            return format!("{kort}...");
        }

        return summary;
    }

    String::new()
}

/// Renderar index.md - en katalog över alla sidor, uppdelad efter
/// sidtyp, med en rad per sida.
///
/// Go: `Wiki.IndexByKind`.
pub fn index_by_kind(wiki: &Wiki) -> String {
    let mut ut = String::new();

    ut.push_str("<!-- GENERERAD av `wikictl index`. Ändra inte manuellt. -->\n");
    ut.push_str("# Wiki Index\n\n");
    ut.push_str("Katalog över allt i wikin. **Läs den här först** när du söker,\n");
    ut.push_str("och gå sedan ner på sidorna som är relevanta.\n\n");

    if wiki.notes.is_empty() {
        ut.push_str("_Inga sidor ännu._\n");
        return ut;
    }

    // Räkna per typ.
    let mut counts: BTreeMap<PageKind, usize> = BTreeMap::new();

    for anteckning in &wiki.notes {
        *counts.entry(kind_of(anteckning)).or_insert(0) += 1;
    }

    ut.push_str("| Typ | Antal |\n|---|---|\n");

    for kind in KINDS {
        let antal = counts.get(&kind).copied().unwrap_or(0);

        if antal == 0 {
            continue;
        }

        ut.push_str(&format!("| {} | {antal} |\n", kind.kind_title()));
    }

    ut.push('\n');

    for kind in KINDS {
        let sidor = pages_of_kind(wiki, kind);

        if sidor.is_empty() {
            continue;
        }

        ut.push_str(&format!("## {}\n\n", kind.kind_title()));

        for anteckning in sidor {
            let updated = if anteckning.updated.is_empty() {
                "-".to_string()
            } else {
                anteckning.updated.clone()
            };

            ut.push_str(&format!(
                "- [[{}]] - {} _(uppdaterad {})_\n",
                anteckning.id,
                summary_line(anteckning),
                updated
            ));
        }

        ut.push('\n');
    }

    ut
}

/// Bygger indexet för en katalog på disk.
pub fn build_index(dir: &Path) -> Result<String, String> {
    let wiki = crate::wiki::open(dir)?;

    Ok(index_by_kind(&wiki))
}
