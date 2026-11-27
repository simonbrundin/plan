//! Sidor i wiki-lagret.
//!
//! Go: `internal/wiki/page.go`. Varje sidtyp har en egen katalog, och
//! en sida skapas eller uppdateras med `write_page`.

use std::path::{Path, PathBuf};

use chrono::{DateTime, Local};

use crate::note::{merge_tags, parse_note, slug, Note};
use crate::wiki::{open, write_file, PAGES_DIR};

/// En sidtyp i wiki-lagret.
#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord)]
pub enum PageKind {
    /// En anteckning.
    Note,
    /// En entitet - något konkret.
    Entity,
    /// Ett koncept - en idé.
    Concept,
    /// En sammanfattning av en råkälla.
    Summary,
    /// En syntes - flera sidor sammanvävda.
    Synthesis,
}

/// Sidtyperna i den ordning de presenteras i indexet.
pub const KINDS: [PageKind; 5] = [
    PageKind::Entity,
    PageKind::Concept,
    PageKind::Summary,
    PageKind::Synthesis,
    PageKind::Note,
];

impl PageKind {
    /// Katalogen sidtypen ligger i.
    ///
    /// Go: `PageKind.KindDir`.
    pub fn kind_dir(&self) -> &'static str {
        match self {
            PageKind::Entity => "entities",
            PageKind::Concept => "concepts",
            PageKind::Summary => "summaries",
            PageKind::Synthesis => "synthesis",
            PageKind::Note => crate::wiki::NOTES_DIR,
        }
    }

    /// Svenska rubriker för indexet.
    ///
    /// Go: `PageKind.KindTitle`.
    pub fn kind_title(&self) -> &'static str {
        match self {
            PageKind::Entity => "Entiteter",
            PageKind::Concept => "Koncept",
            PageKind::Summary => "Sammanfattningar",
            PageKind::Synthesis => "Synteser",
            PageKind::Note => "Noter",
        }
    }

    /// Namnet på sidtypen.
    pub fn name(&self) -> &'static str {
        match self {
            PageKind::Note => "note",
            PageKind::Entity => "entity",
            PageKind::Concept => "concept",
            PageKind::Summary => "summary",
            PageKind::Synthesis => "synthesis",
        }
    }
}

/// Tolkar ett sidtypnamn.
///
/// Go: `wiki.ParseKind`.
pub fn parse_kind(s: &str) -> Result<PageKind, String> {
    match s.trim().to_lowercase().as_str() {
        "note" | "" => Ok(PageKind::Note),
        "entity" => Ok(PageKind::Entity),
        "concept" => Ok(PageKind::Concept),
        "summary" => Ok(PageKind::Summary),
        "synthesis" => Ok(PageKind::Synthesis),
        _ => Err(format!(
            "okänd sidtyp {s:?} (giltiga: note, entity, concept, summary, synthesis)"
        )),
    }
}

/// Styr hur en sida skrivs.
#[derive(Debug, Clone, Default)]
pub struct WriteOptions {
    /// Sidtypen.
    pub kind: Option<PageKind>,
    /// Rubriken.
    pub title: String,
    /// Brödtexten.
    pub body: String,
    /// Sammanfattningen som syns i indexet.
    pub summary: String,
    /// Taggar.
    pub tags: Vec<String>,
    /// Relaterade sidor.
    pub related: Vec<String>,
    /// Råkällor.
    pub sources: Vec<String>,
    /// Status, t.ex. active.
    pub status: String,
    /// Ersätter brödtexten i stället för att lägga till en daterad
    /// uppdateringssektion.
    pub replace: bool,
    /// Gör körningen reproducerbar i tester.
    pub now: Option<DateTime<Local>>,
}

/// Skriver en sida i wiki-lagret och returnerar dess id.
///
/// Finns sidan redan uppdateras den: metadata slås ihop, taggar och
/// relaterade sidor adderas, och brödtexten antingen ersätts eller läggs
/// till under en daterad uppdateringsrubrik.
///
/// Go: `wiki.WritePage`.
pub fn write_page(dir: &Path, opts: &WriteOptions) -> Result<String, String> {
    if opts.title.trim().is_empty() {
        return Err("en sida måste ha en rubrik".to_string());
    }

    let now = opts.now.unwrap_or_else(Local::now);
    let date = now.format("%Y-%m-%d").to_string();

    let kind = opts.kind.unwrap_or(PageKind::Note);

    let id = slug(&opts.title);

    if id.is_empty() {
        return Err(format!(
            "rubriken {:?} går inte att göra till ett id",
            opts.title
        ));
    }

    let page_id = format!("{}/{id}", kind.kind_dir());
    let path = dir
        .join(PAGES_DIR)
        .join(kind.kind_dir())
        .join(format!("{id}.md"));

    let existing = match parse_note(&path) {
        Ok(anteckning) => Some(anteckning),
        Err(err) if err.contains("kunde inte läsa") => None,
        Err(err) => return Err(err),
    };

    let Some(mut existing) = existing else {
        let mut anteckning = Note {
            id: page_id.clone(),
            title: opts.title.clone(),
            note_type: kind.name().to_string(),
            tags: opts.tags.clone(),
            related: opts.related.clone(),
            sources: opts.sources.clone(),
            status: opts.status.clone(),
            summary: opts.summary.clone(),
            created: date.clone(),
            updated: date.clone(),
            body: opts.body.trim().to_string(),
            path: path.to_string_lossy().to_string(),
            ..Default::default()
        };

        if anteckning.status.is_empty() {
            anteckning.status = "active".to_string();
        }

        write_file(&path, &anteckning.render())?;

        return Ok(page_id);
    };

    existing.title = opts.title.clone();
    existing.note_type = kind.name().to_string();
    existing.tags = merge_tags(&existing.tags, &opts.tags);
    existing.related = merge_tags(&existing.related, &opts.related);
    existing.sources = merge_tags(&existing.sources, &opts.sources);

    if !opts.summary.is_empty() {
        existing.summary = opts.summary.clone();
    }

    if !opts.status.is_empty() {
        existing.status = opts.status.clone();
    }

    existing.updated = date.clone();

    // En tom body lämnar den befintliga texten orörd.
    let body = opts.body.trim();

    if !body.is_empty() {
        if opts.replace {
            existing.body = body.to_string();
        } else {
            existing.body = crate::ingest::append_update(&existing.body, body, &date);
        }
    }

    write_file(&path, &existing.render())?;

    Ok(page_id)
}

/// Lägger till en länk från en sida till en annan.
///
/// Länken hamnar både i frontmatter (related) och i brödtextens
/// "## Relaterade"-sektion, så att både maskin och läsare ser den.
///
/// Go: `wiki.Link`.
pub fn link(dir: &Path, from_id: &str, to_id: &str) -> Result<(), String> {
    let mut from = find_page_by_id(dir, from_id)?;

    if find_page_by_id(dir, to_id).is_err() {
        return Err(format!("målsidan {to_id:?} finns inte"));
    }

    from.related = merge_tags(&from.related, &[to_id.to_string()]);
    from.updated = Local::now().format("%Y-%m-%d").to_string();

    let mut body = from.body.clone();

    if !has_related_entry(&body, to_id) {
        body = append_related(&body, to_id);
    }

    from.body = body;

    write_file(Path::new(&from.path), &from.render())
}

/// Tar bort en länk mellan två sidor.
///
/// Go: `wiki.Unlink`.
pub fn unlink(dir: &Path, from_id: &str, to_id: &str) -> Result<(), String> {
    let mut from = find_page_by_id(dir, from_id)?;

    from.related = from
        .related
        .iter()
        .filter(|tagg| !tagg.eq_ignore_ascii_case(to_id))
        .cloned()
        .collect();
    from.updated = Local::now().format("%Y-%m-%d").to_string();
    from.body = remove_related(&from.body, to_id);

    write_file(Path::new(&from.path), &from.render())
}

/// Hittar en sida på dess id (case-insensitive) eller, om inget hittas,
/// på sista segmentet av id:t.
///
/// Go: `findPageByID`.
pub fn find_page_by_id(dir: &Path, id: &str) -> Result<Note, String> {
    let wiki = open(dir)?;

    if let Some(anteckning) = wiki.note(id) {
        return Ok(anteckning.clone());
    }

    let bas = Path::new(&id.replace('\\', "/"))
        .file_name()
        .map(|namn| namn.to_string_lossy().to_string())
        .unwrap_or_else(|| id.to_string());

    if let Some(anteckning) = wiki.notes.iter().find(|anteckning| {
        Path::new(&anteckning.id)
            .file_name()
            .map(|namn| namn.to_string_lossy().eq_ignore_ascii_case(&bas))
            .unwrap_or(false)
    }) {
        return Ok(anteckning.clone());
    }

    Err(format!("sidan {id:?} finns inte"))
}

/// Svarar om brödtexten redan länkar till målet.
fn has_related_entry(body: &str, to_id: &str) -> bool {
    for rad in body.split('\n') {
        if !rad.trim().starts_with("- ") {
            continue;
        }

        if rad.contains(&format!("[[{to_id}]]")) || rad.contains(&format!("[[{to_id}|")) {
            return true;
        }
    }

    false
}

/// Lägger till en post i "## Relaterade"-sektionen.
///
/// Listan hålls sammanhållen: inga tomma rader mellan posterna.
fn append_related(body: &str, to_id: &str) -> String {
    const RUBRIK: &str = "## Relaterade";
    let post = format!("- [[{to_id}]]");

    let rader: Vec<&str> = body.split('\n').collect();

    for (index, rad) in rader.iter().enumerate() {
        if rad.trim() != RUBRIK {
            continue;
        }

        // Samla in posterna som hör till sektionen.
        let mut poster: Vec<String> = Vec::new();
        let mut j = index + 1;

        while j < rader.len() {
            let trimmad = rader[j].trim();

            if trimmad.is_empty() {
                j += 1;
                continue;
            }

            if !trimmad.starts_with("- ") {
                break;
            }

            poster.push(trimmad.to_string());
            j += 1;
        }

        // Den nya posten hamnar först i listan.
        poster.insert(0, post);

        let mut ut: Vec<String> = rader[..index + 1]
            .iter()
            .map(|rad| rad.to_string())
            .collect();

        ut.push(String::new());
        ut.extend(poster);
        ut.extend(rader[j..].iter().map(|rad| rad.to_string()));

        return ut.join("\n");
    }

    if body.trim().is_empty() {
        return format!("{RUBRIK}\n\n{post}");
    }

    format!("{}\n\n{RUBRIK}\n\n{post}\n", body.trim_end_matches('\n'))
}

/// Tar bort en post ur "## Relaterade"-sektionen.
fn remove_related(body: &str, to_id: &str) -> String {
    let mut ut: Vec<&str> = Vec::new();

    for rad in body.split('\n') {
        let trimmad = rad.trim();

        if trimmad.starts_with("- ")
            && (rad.contains(&format!("[[{to_id}]]")) || rad.contains(&format!("[[{to_id}|")))
        {
            continue;
        }

        ut.push(rad);
    }

    ut.join("\n")
}

/// Sökvägen till en sida i en katalog.
pub fn page_path(dir: &Path, kind: PageKind, id: &str) -> PathBuf {
    dir.join(PAGES_DIR)
        .join(kind.kind_dir())
        .join(format!("{id}.md"))
}
