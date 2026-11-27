//! Anteckningar i wikin.
//!
//! Go: `internal/wiki/note.go`. En anteckning är en markdownfil med
//! frontmatter - metadata i nyckel-värde-form ovanför brödtexten.

use std::collections::{BTreeMap, HashSet};
use std::path::Path;

use regex::Regex;
use std::sync::OnceLock;

/// Nyckeln för anteckningens rubrik.
pub const META_TITLE: &str = "title";
/// Nyckeln för anteckningens typ.
pub const META_TYPE: &str = "type";
/// Nyckeln för anteckningens taggar.
pub const META_TAGS: &str = "tags";
/// Nyckeln för anteckningens skapandedatum.
pub const META_CREATED: &str = "created";
/// Nyckeln för anteckningens uppdateringsdatum.
pub const META_UPDATED: &str = "updated";
/// Nyckeln för relaterade sidor.
pub const META_RELATED: &str = "related";
/// Nyckeln för anteckningens status.
pub const META_STATUS: &str = "status";
/// Nyckeln för råkällor.
pub const META_SOURCES: &str = "sources";
/// Nyckeln för en enda råkälla.
pub const META_SOURCE: &str = "source";
/// Nyckeln för sammanfattningen.
pub const META_SUMMARY: &str = "summary";
/// Nyckeln för källans härkomst.
pub const META_ORIGIN: &str = "origin";
/// Nyckeln för datum källan lades in.
pub const META_ADDED: &str = "added";

/// En enskild anteckning i wikin.
#[derive(Debug, Clone, Default)]
pub struct Note {
    /// Unikt id i grafen.
    pub id: String,
    /// Mänsklig titel.
    pub title: String,
    /// Typ: entity, concept, summary, note ...
    pub note_type: String,
    /// Nyckelord för uppslagning.
    pub tags: Vec<String>,
    /// Skapad, YYYY-MM-DD.
    pub created: String,
    /// Uppdaterad, YYYY-MM-DD.
    pub updated: String,
    /// Id:n till anteckningar som hör ihop.
    pub related: Vec<String>,
    /// T.ex. active, deprecated, open.
    pub status: String,
    /// Råkällor sidan är skriven utifrån.
    pub sources: Vec<String>,
    /// Enradig beskrivning, visas i indexet.
    pub summary: String,
    /// Var källan kommer ifrån (url, fil, etc.).
    pub origin: String,
    /// Datum källan lades in (YYYY-MM-DD).
    pub added: String,
    /// Själva texten.
    pub body: String,
    /// Sökväg på disk.
    pub path: String,
}

/// Normaliserar en rubrik till ett filnamnsanpassat id.
///
/// Go: `wiki.Slug`.
pub fn slug(text: &str) -> String {
    let mut normalized = String::with_capacity(text.len());

    for tecken in text.chars() {
        match tecken {
            'å' | 'ä' | 'Å' | 'Ä' => normalized.push('a'),
            'ö' | 'Ö' => normalized.push('o'),
            'ø' => normalized.push('o'),
            'æ' => normalized.push_str("ae"),
            '_' => normalized.push('-'),
            ' ' => normalized.push('-'),
            övriga => normalized.extend(övriga.to_lowercase()),
        }
    }

    let mut slug = String::with_capacity(normalized.len());

    for tecken in normalized.chars() {
        if tecken.is_ascii_lowercase() || tecken.is_ascii_digit() {
            slug.push(tecken);
        } else {
            slug.push('-');
        }
    }

    let slug = slug.trim_matches('-').to_string();
    let mut slug = slug;

    while slug.contains("--") {
        slug = slug.replace("--", "-");
    }

    slug
}

/// Läser en anteckning från en sökväg.
///
/// Go: `wiki.ParseNote`.
pub fn parse_note(sökväg: &Path) -> Result<Note, String> {
    let data = std::fs::read_to_string(sökväg).map_err(|err| {
        format!(
            "kunde inte läsa {}: {}",
            sökväg.display(),
            crate::wiki::go_filfel("open", sökväg, &err)
        )
    })?;

    let (meta, body) = split_frontmatter(&data)
        .map_err(|err| format!("kunde inte tolka frontmatter i {}: {err}", sökväg.display()))?;

    let id = sökväg
        .file_stem()
        .map(|stem| stem.to_string_lossy().to_string())
        .unwrap_or_default();

    let mut title = meta.get(META_TITLE).cloned().unwrap_or_default();

    if title.is_empty() {
        title = heading_title(&body);
    }

    if title.is_empty() {
        title = id.clone();
    }

    Ok(Note {
        id,
        title,
        note_type: meta.get(META_TYPE).cloned().unwrap_or_default(),
        tags: split_list(meta.get(META_TAGS).map(String::as_str).unwrap_or_default()),
        created: meta.get(META_CREATED).cloned().unwrap_or_default(),
        updated: meta.get(META_UPDATED).cloned().unwrap_or_default(),
        related: split_list(
            meta.get(META_RELATED)
                .map(String::as_str)
                .unwrap_or_default(),
        ),
        status: meta.get(META_STATUS).cloned().unwrap_or_default(),
        sources: merge_sources(
            meta.get(META_SOURCES)
                .map(String::as_str)
                .unwrap_or_default(),
            meta.get(META_SOURCE)
                .map(String::as_str)
                .unwrap_or_default(),
        ),
        summary: meta.get(META_SUMMARY).cloned().unwrap_or_default(),
        origin: meta.get(META_ORIGIN).cloned().unwrap_or_default(),
        added: meta.get(META_ADDED).cloned().unwrap_or_default(),
        body: body.trim().to_string(),
        path: sökväg.to_string_lossy().to_string(),
    })
}

/// Letar upp första rubriken på nivå 1 i brödtexten.
fn heading_title(body: &str) -> String {
    for rad in body.split('\n') {
        let rad = rad.trim();

        if let Some(titel) = rad.strip_prefix("# ") {
            return titel.trim().to_string();
        }
    }

    String::new()
}

/// Delar upp en fil i metadata och brödtext.
///
/// En fil utan frontmatter tolkas som ren brödtext.
///
/// Go: `splitFrontmatter`.
fn split_frontmatter(innehåll: &str) -> Result<(BTreeMap<String, String>, String), String> {
    let normaliserat = innehåll.replace("\r\n", "\n");

    if !normaliserat.starts_with("---\n") {
        return Ok((BTreeMap::new(), normaliserat));
    }

    let rest = &normaliserat["---\n".len()..];

    let slut = rest
        .find("\n---")
        .ok_or_else(|| "frontmatter saknar avslutande ---".to_string())?;

    let header = &rest[..slut];
    let body = rest[slut + "\n---".len()..]
        .strip_prefix('\n')
        .unwrap_or(&rest[slut + "\n---".len()..]);

    let mut meta: BTreeMap<String, String> = BTreeMap::new();

    for rad in header.split('\n') {
        let rad = rad.trim();

        if rad.is_empty() || rad.starts_with('#') {
            continue;
        }

        let Some((nyckel, värde)) = rad.split_once(':') else {
            return Err(format!("ogiltig metadata-rad: {rad:?}"));
        };

        meta.insert(nyckel.trim().to_string(), värde.trim().to_string());
    }

    Ok((meta, body.to_string()))
}

/// Tolkar "a, b", "[a, b]" och "a" som en lista.
fn split_list(raw: &str) -> Vec<String> {
    let raw = raw.trim();
    let raw = raw.trim_matches(|c| c == '[' || c == ']');

    if raw.is_empty() {
        return Vec::new();
    }

    raw.split(',')
        .map(|del| del.trim())
        .filter(|del| !del.is_empty())
        .map(str::to_string)
        .collect()
}

/// Slår ihop "sources" och "source" så att båda skrivsätt fungerar.
fn merge_sources(sources: &str, source: &str) -> Vec<String> {
    let mut lista = split_list(sources);

    for källa in split_list(source) {
        if !lista.contains(&källa) {
            lista.push(källa);
        }
    }

    lista
}

/// Matchar markdown-länkar som pekar på en fil: `[text](fil.md)`.
fn markdown_länk_mönster() -> &'static Regex {
    static MÖNSTER: OnceLock<Regex> = OnceLock::new();

    MÖNSTER.get_or_init(|| Regex::new(r"\]\(([^)]+\.md)\)").expect("mönstret"))
}

/// Matchar wikilänkar: `[[Sida/Med Namn]]`.
fn wikilänk_mönster() -> &'static Regex {
    static MÖNSTER: OnceLock<Regex> = OnceLock::new();

    MÖNSTER.get_or_init(|| Regex::new(r"\[\[([^\]|]+)(?:\|[^\]]*)?\]\]").expect("mönstret"))
}

/// Matchar inline-kod: `` `...` ``.
fn kodsspann_mönster() -> &'static Regex {
    static MÖNSTER: OnceLock<Regex> = OnceLock::new();

    MÖNSTER.get_or_init(|| Regex::new("`[^`]*`").expect("mönstret"))
}

impl Note {
    /// Returnerar de länkar som finns i brödtexten.
    ///
    /// Kods spans hoppas över - prosa som beskriver länksyntax ska inte
    /// räknas som en länk.
    ///
    /// Go: `Note.BodyLinks`.
    pub fn body_links(&self) -> Vec<String> {
        let mut sett: HashSet<String> = HashSet::new();
        let mut länkar: Vec<String> = Vec::new();

        let mut lägg_till = |mål: &str| {
            let mål = mål.trim();

            if mål.is_empty() {
                return;
            }

            let nyckel = mål.to_lowercase();

            if sett.insert(nyckel) {
                länkar.push(mål.to_string());
            }
        };

        let body = kodsspann_mönster().replace_all(&self.body, " ");

        for fångst in markdown_länk_mönster().captures_iter(&body) {
            lägg_till(&fångst[1]);
        }

        for fångst in wikilänk_mönster().captures_iter(&body) {
            lägg_till(&fångst[1]);
        }

        länkar
    }

    /// Returnerar både frontmatter-länkar och länkar i brödtexten.
    ///
    /// Go: `Note.AllRelated`.
    pub fn all_related(&self) -> Vec<String> {
        let mut sett: HashSet<String> = HashSet::new();
        sett.insert(self.id.to_lowercase());

        let mut alla: Vec<String> = Vec::new();

        for id in self.related.iter().chain(self.body_links().iter()) {
            if id.is_empty() || sett.contains(&id.to_lowercase()) {
                continue;
            }

            sett.insert(id.to_lowercase());
            alla.push(id.clone());
        }

        alla
    }

    /// Id:t anteckningen får när den ingesteras.
    ///
    /// Rubriken bestämmer id:t - faller det bort används filnamnet.
    ///
    /// Go: `Note.Slug`.
    pub fn ingest_slug(&self) -> String {
        if !self.title.is_empty() {
            let slug = slug(&self.title);

            if !slug.is_empty() {
                return slug;
            }
        }

        slug(&self.id)
    }

    /// Svarar på om anteckningen har en given tagg.
    pub fn has_tag(&self, tag: &str) -> bool {
        self.tags.iter().any(|t| t.eq_ignore_ascii_case(tag))
    }

    /// Serialiserar anteckningen tillbaka till markdown med frontmatter.
    ///
    /// Go: `Note.Render`.
    pub fn render(&self) -> String {
        let mut ut = String::new();

        ut.push_str("---\n");
        ut.push_str(&format!("{META_TITLE}: {}\n", self.title));

        if !self.note_type.is_empty() {
            ut.push_str(&format!("{META_TYPE}: {}\n", self.note_type));
        }

        if !self.summary.is_empty() {
            ut.push_str(&format!("{META_SUMMARY}: {}\n", self.summary));
        }

        if !self.tags.is_empty() {
            ut.push_str(&format!("{META_TAGS}: {}\n", format_list(&self.tags)));
        }

        if !self.origin.is_empty() {
            ut.push_str(&format!("{META_ORIGIN}: {}\n", self.origin));
        }

        if !self.sources.is_empty() {
            ut.push_str(&format!("{META_SOURCES}: {}\n", format_list(&self.sources)));
        }

        if !self.created.is_empty() {
            ut.push_str(&format!("{META_CREATED}: {}\n", self.created));
        }

        if !self.updated.is_empty() {
            ut.push_str(&format!("{META_UPDATED}: {}\n", self.updated));
        }

        if !self.status.is_empty() {
            ut.push_str(&format!("{META_STATUS}: {}\n", self.status));
        }

        if !self.related.is_empty() {
            ut.push_str(&format!("{META_RELATED}: {}\n", format_list(&self.related)));
        }

        ut.push_str("---\n\n");
        ut.push_str(&self.body);
        ut.push('\n');

        ut
    }

    /// Projektnamnet anteckningen gäller, om något.
    pub fn project_of(&self) -> String {
        for tagg in &self.tags {
            if let Some(projekt) = tagg.strip_prefix(crate::project::PROJECT_TAG_PREFIX) {
                return projekt.to_string();
            }
        }

        String::new()
    }

    /// Svarar på om anteckningen gäller ett visst projekt.
    ///
    /// Go: `Note.HasProject`.
    pub fn has_project(&self, project: &str) -> bool {
        if project.is_empty() {
            return true;
        }

        self.project_of().eq_ignore_ascii_case(project)
    }
}

/// Formaterar en lista så som frontmatter gör.
fn format_list(list: &[String]) -> String {
    format!("[{}]", list.join(", "))
}

/// Slår ihop två tagg-listor utan dubbletter, sorterade.
///
/// Go: `wiki.MergeTags`.
pub fn merge_tags(a: &[String], b: &[String]) -> Vec<String> {
    let mut sett: HashSet<String> = HashSet::new();
    let mut sammanslaget: Vec<String> = Vec::new();

    for tagg in a.iter().chain(b.iter()) {
        let nyckel = tagg.to_lowercase();

        if tagg.is_empty() || sett.contains(&nyckel) {
            continue;
        }

        sett.insert(nyckel);
        sammanslaget.push(tagg.clone());
    }

    sammanslaget.sort();

    sammanslaget
}
