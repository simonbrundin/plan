//! Ingest av inboxen in i kunskapsgrafen.
//!
//! Go: `internal/wiki/ingest.go`. Inlägg i `inbox/` blir sidor i
//! `wiki/notes/` och arkiveras i `archive/`.

use std::collections::BTreeSet;
use std::path::Path;

use chrono::{DateTime, Local};

use crate::note::{merge_tags, slug, Note};
use crate::wiki::{open, write_file, ARCHIVE_DIR, INBOX_DIR, NOTES_DIR, PAGES_DIR};

/// Styr hur ingesten körs.
#[derive(Debug, Clone, Default)]
pub struct IngestOptions {
    /// Räknar bara fram, skriver inget.
    pub dry_run: bool,
    /// Gör körningen reproducerbar i tester.
    pub now: Option<DateTime<Local>>,
}

/// Ett förslag på att slå ihop två sidor.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Merge {
    /// Den första sidan.
    pub a: String,
    /// Den andra sidan.
    pub b: String,
    /// Taggar som sidorna delar.
    pub tags: Vec<String>,
}

impl std::fmt::Display for Merge {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(
            f,
            "{} + {} (delade taggar: {})",
            self.a,
            self.b,
            self.tags.join(", ")
        )
    }
}

/// Resultatet av en ingest.
#[derive(Debug, Clone, Default)]
pub struct IngestReport {
    /// Id:n för nya sidor.
    pub created: Vec<String>,
    /// Id:n för uppdaterade sidor.
    pub updated: Vec<String>,
    /// Arkiverade inbox-filer.
    pub archived: Vec<String>,
    /// Förslag på sammanslagning.
    pub suggest: Vec<Merge>,
}

/// Läser in alla inlägg i inboxen, sorterade på sökväg.
///
/// Go: `readInbox`.
pub fn read_inbox(dir: &Path) -> Result<Vec<Note>, String> {
    let inbox = dir.join(INBOX_DIR);
    let mut anteckningar = Vec::new();

    let poster = match std::fs::read_dir(&inbox) {
        Ok(poster) => poster,
        Err(err) if err.kind() == std::io::ErrorKind::NotFound => return Ok(anteckningar),
        Err(err) => return Err(format!("kunde inte läsa {}: {err}", inbox.display())),
    };

    let mut filer: Vec<std::path::PathBuf> = poster
        .filter_map(|poster| poster.ok())
        .map(|poster| poster.path())
        .filter(|sökväg| sökväg.is_file())
        .filter(|sökväg| {
            sökväg
                .extension()
                .map(|ändelse| ändelse == "md")
                .unwrap_or(false)
        })
        .filter(|sökväg| {
            let namn = sökväg
                .file_name()
                .map(|namn| namn.to_string_lossy().to_string())
                .unwrap_or_default();

            namn != "README.md" && !namn.starts_with('_')
        })
        .collect();

    filer.sort();

    for fil in filer {
        anteckningar.push(crate::note::parse_note(&fil)?);
    }

    Ok(anteckningar)
}

/// Lägger till en daterad sektion i slutet av en sida.
///
/// Go: `appendUpdate`.
pub fn append_update(body: &str, addition: &str, date: &str) -> String {
    if addition.trim().is_empty() {
        return body.to_string();
    }

    if body.trim().is_empty() {
        return addition.trim().to_string();
    }

    format!(
        "{}\n\n## Uppdatering {date}\n\n{}\n",
        body.trim_end_matches('\n'),
        addition.trim()
    )
}

/// Lägger in inboxen i grafen.
///
/// Varje inlägg blir en not i `wiki/notes/`. Finns det redan en not
/// med samma id blir den uppdaterad - det är så upprepade lärdomar
/// inte skapar dubbletter.
///
/// Go: `wiki.Ingest`.
pub fn ingest(dir: &Path, opts: &IngestOptions) -> Result<IngestReport, String> {
    let nu = opts.now.unwrap_or_else(Local::now);
    let date = nu.format("%Y-%m-%d").to_string();

    let mut report = IngestReport::default();

    let mut poster = read_inbox(dir)?;

    let mut w = open(dir)?;

    let archive_dir = dir.join(ARCHIVE_DIR);

    if !opts.dry_run {
        std::fs::create_dir_all(&archive_dir)
            .map_err(|err| format!("kunde inte skapa {}: {err}", archive_dir.display()))?;
    }

    for item in poster.drain(..) {
        let id = item.ingest_slug();

        if id.is_empty() {
            continue;
        }

        let page_id = format!("{NOTES_DIR}/{id}");

        if let Some(existing) = w.notes.iter().find(|anteckning| anteckning.id == page_id) {
            let mut existing = existing.clone();

            existing.body = append_update(&existing.body, &item.body, &date);
            existing.tags = merge_tags(&existing.tags, &item.tags);
            existing.updated = date.clone();

            if existing.status.is_empty() {
                existing.status = "active".to_string();
            }

            existing.related = merge_tags(&existing.related, &item.all_related());

            report.updated.push(page_id.clone());

            if !opts.dry_run {
                write_file(Path::new(&existing.path), &existing.render())?;
            }

            // Anteckningen i grafen måste spegla den uppdaterade.
            if let Some(plats) = w.notes.iter_mut().find(|a| a.id == page_id) {
                *plats = existing;
            }
        } else {
            let mut anteckning = Note {
                id: page_id.clone(),
                title: item.title.clone(),
                note_type: item.note_type.clone(),
                tags: item.tags.clone(),
                created: item.created.clone(),
                updated: item.updated.clone(),
                related: item.all_related(),
                status: item.status.clone(),
                sources: item.sources.clone(),
                body: item.body.clone(),
                path: dir
                    .join(PAGES_DIR)
                    .join(NOTES_DIR)
                    .join(format!("{id}.md"))
                    .to_string_lossy()
                    .to_string(),
                summary: item.summary.clone(),
                origin: item.origin.clone(),
                added: item.added.clone(),
            };

            if anteckning.created.is_empty() {
                anteckning.created = date.clone();
            }

            if anteckning.updated.is_empty() {
                anteckning.updated = date.clone();
            }

            if anteckning.status.is_empty() {
                anteckning.status = "active".to_string();
            }

            if anteckning.note_type.is_empty() {
                anteckning.note_type = "note".to_string();
            }

            report.created.push(page_id.clone());

            if !opts.dry_run {
                write_file(Path::new(&anteckning.path), &anteckning.render())?;
            }

            w.notes.push(anteckning);
        }

        let arkiverad = archive_dir.join(format!("{date}-{id}.md"));

        report.archived.push(
            arkiverad
                .file_name()
                .map(|namn| namn.to_string_lossy().to_string())
                .unwrap_or_default(),
        );

        if !opts.dry_run {
            let källa = dir.join(INBOX_DIR).join(format!("{id}.md"));

            if källa.exists() {
                std::fs::rename(&källa, &arkiverad).map_err(|err| {
                    format!(
                        "kunde inte arkivera {}: rename {} {}: {}",
                        källa.display(),
                        källa.display(),
                        arkiverad.display(),
                        crate::wiki::go_errno(&err)
                    )
                })?;
            }
        }
    }

    w.notes.sort_by(|a, b| a.id.cmp(&b.id));

    report.suggest = suggest_merges(&w);

    if !opts.dry_run {
        w.write_index()?;
        w.write_catalog()?;
        w.write_graph()?;

        append_log(dir, &date, &report)?;
    }

    Ok(report)
}

/// Skriver ingesten till loggen.
///
/// Go: `appendLog`.
fn append_log(dir: &Path, date: &str, report: &IngestReport) -> Result<(), String> {
    let mut poster: Vec<crate::log::LogEntry> = Vec::new();
    let mut detaljer: Vec<String> = Vec::new();

    for id in &report.created {
        poster.push(crate::log::LogEntry {
            date: date.to_string(),
            op: "ingest".to_string(),
            title: format!("create {id}"),
        });
        detaljer.push(format!("[create] [[{id}]] — ny sida från inboxen"));
    }

    for id in &report.updated {
        poster.push(crate::log::LogEntry {
            date: date.to_string(),
            op: "ingest".to_string(),
            title: format!("update {id}"),
        });
        detaljer.push(format!("[update] [[{id}]] — tillägg från inboxen"));
    }

    for namn in &report.archived {
        detaljer.push(format!("[archive] {namn}"));
    }

    for sammanslagning in &report.suggest {
        detaljer.push(format!("[merge?] {sammanslagning}"));
    }

    if poster.is_empty() {
        return Ok(());
    }

    crate::log::append_log(dir, &poster, &detaljer)
}

/// Hittar par av sidor som sannolikt beskriver samma sak.
///
/// Tre villkor måste vara uppfyllda:
/// - samma sidtyp (en concept och en summary är inte samma sak),
/// - nästan lika titel (slug lika, eller den ena titeln är ett prefix),
/// - minst en gemensam tagg.
///
/// Go: `SuggestMerges`.
pub fn suggest_merges(w: &crate::wiki::Wiki) -> Vec<Merge> {
    let mut ut = Vec::new();

    for i in 0..w.notes.len() {
        for j in (i + 1)..w.notes.len() {
            let a = &w.notes[i];
            let b = &w.notes[j];

            if a.note_type != b.note_type {
                continue;
            }

            if !similar_titles(&a.title, &b.title) {
                continue;
            }

            let delade = shared_tags(&a.tags, &b.tags);

            if delade.is_empty() {
                continue;
            }

            ut.push(Merge {
                a: a.id.clone(),
                b: b.id.clone(),
                tags: delade,
            });
        }
    }

    ut.sort_by(|a, b| a.a.cmp(&b.a).then_with(|| a.b.cmp(&b.b)));

    ut
}

/// Avgör om två titlar troligen beskriver samma sak.
///
/// Jämförelsen sker på slug utan bindestreck, så "FluxOperator" och
/// "flux-operator" räknas som samma. Prefix räknas bara när det följs av
/// en ny ord - "GitOps" är inte samma som "Flux GitOps".
fn similar_titles(a: &str, b: &str) -> bool {
    let sa = slug(a);
    let sb = slug(b);

    if strip_separators(&sa) == strip_separators(&sb) {
        return true;
    }

    is_word_prefix(&sa, &sb) || is_word_prefix(&sb, &sa)
}

/// Avgör om a är ett prefix av b på ordgräns.
fn is_word_prefix(a: &str, b: &str) -> bool {
    if a.len() >= b.len() {
        return false;
    }

    if !b.starts_with(a) {
        return false;
    }

    b.as_bytes()[a.len()] == b'-'
}

/// Tar bort bindestreck och understreck ur en slug.
fn strip_separators(s: &str) -> String {
    s.replace(['-', '_', ' '], "")
}

/// Taggar som två listor delar.
fn shared_tags(a: &[String], b: &[String]) -> Vec<String> {
    let i_b: BTreeSet<String> = b.iter().map(|tagg| tagg.to_lowercase()).collect();

    let mut ut: Vec<String> = Vec::new();
    let mut sedda: BTreeSet<String> = BTreeSet::new();

    for tagg in a {
        let nyckel = tagg.to_lowercase();

        if i_b.contains(&nyckel) && sedda.insert(nyckel) {
            ut.push(tagg.clone());
        }
    }

    ut
}
