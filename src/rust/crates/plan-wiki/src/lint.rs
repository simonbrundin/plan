//! Lint av wikin.
//!
//! Go: `internal/wiki/lint.go`.

use std::collections::BTreeMap;
use std::time::Duration;

use crate::index::derive_summary;
use crate::wiki::Wiki;

/// En typ av lintfynd.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum LintKind {
    /// Länkar till en sida som inte finns.
    BrokenLink,
    /// Samma rubrik på flera ställen.
    Duplicate,
    /// Samma rubrik två gånger på en sida.
    DuplicateSection,
    /// Sida som ingen annan sida länkar till.
    Orphan,
    /// Sida utan sammanfattning.
    NoSummary,
    /// Sida som inte uppdaterats på länge.
    Stale,
    /// Råkälla som aldrig integrerats.
    Unintegrated,
}

impl LintKind {
    /// Den svenska etiketten.
    pub fn label(&self) -> &'static str {
        match self {
            LintKind::BrokenLink => "trasig länk",
            LintKind::Duplicate => "dubblerad sida",
            LintKind::DuplicateSection => "dubblerad sektion",
            LintKind::Orphan => "ödelapp",
            LintKind::NoSummary => "saknar sammanfattning",
            LintKind::Stale => "föråldrad",
            LintKind::Unintegrated => "ointegrerad källa",
        }
    }
}

/// Ett problem som linten hittade.
#[derive(Debug, Clone)]
pub struct LintFinding {
    /// Typen av problem.
    pub kind: LintKind,
    /// Sidan problemet gäller.
    pub page: String,
    /// Beskrivning av problemet.
    pub msg: String,
}

impl std::fmt::Display for LintFinding {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        if self.page.is_empty() {
            write!(f, "{}: {}", self.kind.label(), self.msg)
        } else {
            write!(f, "{}: {} - {}", self.kind.label(), self.page, self.msg)
        }
    }
}

/// Styr hur linten körs.
#[derive(Debug, Clone, Default)]
pub struct LintOptions {
    /// Hur länge en sida får vara ouppdaterad innan den flaggas.
    /// Noll betyder att ingen ålderskontroll körs.
    pub stale_after: Option<Duration>,
    /// Nuvarande tid - används för att göra körningen reproducerbar.
    pub now: Option<chrono::DateTime<chrono::Utc>>,
}

/// Undersöker wikin och returnerar en lista med problem.
///
/// Go: `Wiki.Lint`.
pub fn lint(wiki: &Wiki, opts: &LintOptions) -> Vec<LintFinding> {
    let mut findings = Vec::new();

    // Trasiga länkar.
    for anteckning in &wiki.notes {
        for mål in anteckning.all_related() {
            if wiki.resolve(&mål).is_none() {
                findings.push(LintFinding {
                    kind: LintKind::BrokenLink,
                    page: anteckning.id.clone(),
                    msg: format!("pekar på {mål:?} som inte finns"),
                });
            }
        }
    }

    // Dubblerade sidor: samma rubrik på flera ställen.
    let mut by_title: BTreeMap<String, Vec<String>> = BTreeMap::new();

    for anteckning in &wiki.notes {
        by_title
            .entry(anteckning.title.to_lowercase().trim().to_string())
            .or_default()
            .push(anteckning.id.clone());
    }

    for (rubrik, idn) in by_title {
        if idn.len() < 2 {
            continue;
        }

        findings.push(LintFinding {
            kind: LintKind::Duplicate,
            page: idn.join(", "),
            msg: format!("samma rubrik {rubrik:?} finns på flera ställen"),
        });
    }

    // Dubblerade sektioner: samma rubrik två gånger på en sida.
    for anteckning in &wiki.notes {
        let mut sedda: Vec<String> = Vec::new();

        for rubrik in section_headings(&anteckning.body) {
            if sedda.contains(&rubrik) {
                findings.push(LintFinding {
                    kind: LintKind::DuplicateSection,
                    page: anteckning.id.clone(),
                    msg: format!("rubriken {rubrik:?} förekommer flera gånger"),
                });
                continue;
            }

            sedda.push(rubrik);
        }
    }

    // Ödelappar: sidor som ingen annan sida länkar till.
    for anteckning in &wiki.notes {
        if wiki.backlinks(&anteckning.id).is_empty() {
            findings.push(LintFinding {
                kind: LintKind::Orphan,
                page: anteckning.id.clone(),
                msg: "inga inkommande länkar".to_string(),
            });
        }
    }

    // Sidor utan beskrivning gör indexet svårläst.
    for anteckning in &wiki.notes {
        if anteckning.summary.trim().is_empty() && derive_summary(&anteckning.body).is_empty() {
            findings.push(LintFinding {
                kind: LintKind::NoSummary,
                page: anteckning.id.clone(),
                msg: "sidan saknar beskrivning - lägg till summary i frontmatter".to_string(),
            });
        }
    }

    // Föråldrade sidor.
    if let Some(stale_after) = opts.stale_after {
        let now = opts.now.unwrap_or_else(chrono::Utc::now);
        let cutoff = now - chrono::Duration::from_std(stale_after).unwrap_or_default();

        for anteckning in &wiki.notes {
            let Ok(updated) = chrono::NaiveDate::parse_from_str(&anteckning.updated, "%Y-%m-%d")
            else {
                continue;
            };

            let updated = updated.and_hms_opt(0, 0, 0);

            let Some(updated) = updated else {
                continue;
            };

            // Go tolkar datumet som midnat i UTC och jämför mot
            // `now - stale_after`. Att jämföra mot ett lokalt klockslag
            // ger ett annat svar vid gränsen, så tiden hålls i UTC här.
            if updated.and_utc() < cutoff {
                findings.push(LintFinding {
                    kind: LintKind::Stale,
                    page: anteckning.id.clone(),
                    msg: format!("inte uppdaterad sedan {}", anteckning.updated),
                });
            }
        }
    }

    // Råkällor som aldrig integrerats.
    for källa in crate::sources::sources(wiki) {
        if !källa.indexed {
            findings.push(LintFinding {
                kind: LintKind::Unintegrated,
                page: källa.id.clone(),
                msg: "råkällan nämns inte av någon sida - skriv en sammanfattning".to_string(),
            });
        }
    }

    findings
}

/// Returnerar alla rubriker på nivå 2 i brödtexten.
fn section_headings(body: &str) -> Vec<String> {
    let mut ut = Vec::new();
    let mut i_kod = false;

    for rad in body.split('\n') {
        let trimmad = rad.trim();

        if trimmad.starts_with("```") {
            i_kod = !i_kod;
            continue;
        }

        if i_kod {
            continue;
        }

        if let Some(rubrik) = trimmad.strip_prefix("## ") {
            ut.push(rubrik.trim().to_string());
        }
    }

    ut
}
