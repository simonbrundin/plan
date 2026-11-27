//! Råkällor - de immutabla källorna wikin syntetiserar.
//!
//! Go: `internal/wiki/index.go`.

use std::path::{Path, PathBuf};

use chrono::{DateTime, Local};

use crate::note::{parse_note, slug};
use crate::wiki::{write_file, Wiki};

/// En råkälla som lagts in i `raw/`.
#[derive(Debug, Clone)]
pub struct Source {
    /// Filnamnet utan `.md`.
    pub id: String,
    /// Rubriken.
    pub title: String,
    /// Varifrån källan kommer.
    pub origin: String,
    /// Datum källan lades in.
    pub added: String,
    /// Sökvägen på disk.
    pub path: PathBuf,
    /// Sant om någon sida i wikin hänvisar till källan.
    pub indexed: bool,
}

/// Kopierar en fil in i `raw/` som en oförändlig råkälla.
///
/// Källan får ett header med metadata så att den går att hitta igen.
/// Ligger originalet redan i `raw/` flyttas det - annars finns samma
/// text två gånger i källsamlingen.
///
/// Go: `wiki.SourceFile`.
pub fn source_file(
    dir: &Path,
    src_path: &str,
    title: &str,
    origin: &str,
    now: DateTime<Local>,
) -> Result<PathBuf, String> {
    let data = std::fs::read_to_string(src_path)
        .map_err(|err| format!("kunde inte läsa {src_path}: {err}"))?;

    let mut rubrik = title.trim().to_string();

    if rubrik.is_empty() {
        rubrik = Path::new(src_path)
            .file_stem()
            .map(|namn| namn.to_string_lossy().to_string())
            .unwrap_or_default();
    }

    let date = now.format("%Y-%m-%d").to_string();
    let namn = format!("{date}-{}.md", slug(&rubrik));

    let mål = dir.join(crate::wiki::RAW_DIR).join(namn);

    if mål.exists() {
        return Err(format!("råkällan finns redan: {}", mål.display()));
    }

    let mut ut = String::new();

    ut.push_str("---\n");
    ut.push_str(&format!("title: {rubrik}\n"));
    ut.push_str("type: source\n");
    ut.push_str(&format!("added: {date}\n"));

    if !origin.is_empty() {
        ut.push_str(&format!("origin: {origin}\n"));
    }

    ut.push_str("immutable: true\n");
    ut.push_str("---\n\n");
    ut.push_str(&data);

    write_file(&mål, &ut)?;

    // Om källan redan låg i raw/ tar vi bort originalet.
    if is_in_raw_dir(dir, Path::new(src_path))? {
        std::fs::remove_file(src_path)
            .map_err(|err| format!("kunde inte ta bort {src_path}: {err}"))?;
    }

    Ok(mål)
}

/// Svarar om sökvägen ligger i katalogens `raw/`-katalog.
fn is_in_raw_dir(dir: &Path, path: &Path) -> Result<bool, String> {
    let raw_dir = dir.join(crate::wiki::RAW_DIR);

    let abs_raw = std::fs::canonicalize(&raw_dir)
        .map_err(|err| format!("kunde inte hitta {}: {err}", raw_dir.display()))?;
    let abs_path = std::fs::canonicalize(path)
        .map_err(|err| format!("kunde inte hitta {}: {err}", path.display()))?;

    Ok(abs_path.starts_with(&abs_raw))
}

/// Listar råkällorna i `raw/` och markerar de som redan integrerats.
///
/// Go: `Wiki.Sources`.
pub fn sources(wiki: &Wiki) -> Vec<Source> {
    let raw_dir = wiki.dir.join(crate::wiki::RAW_DIR);

    let Ok(poster) = std::fs::read_dir(&raw_dir) else {
        return Vec::new();
    };

    // Samla ihop alla källor som nämns i sidornas `sources`.
    let mut nämnda: Vec<String> = Vec::new();

    for anteckning in &wiki.notes {
        for src in &anteckning.sources {
            nämnda.push(
                Path::new(src)
                    .file_stem()
                    .map(|namn| namn.to_string_lossy().to_lowercase())
                    .unwrap_or_else(|| src.to_lowercase()),
            );
            nämnda.push(src.trim_end_matches(".md").to_lowercase());
        }
    }

    let mut filer: Vec<PathBuf> = poster
        .filter_map(|poster| poster.ok())
        .map(|poster| poster.path())
        .filter(|sökväg| sökväg.is_file())
        .filter(|sökväg| {
            sökväg
                .extension()
                .map(|ändelse| ändelse == "md")
                .unwrap_or(false)
        })
        .collect();

    filer.sort();

    let mut ut = Vec::new();

    for fil in filer {
        let Ok(anteckning) = parse_note(&fil) else {
            continue;
        };

        let id = fil
            .file_stem()
            .map(|namn| namn.to_string_lossy().to_string())
            .unwrap_or_default();

        let indexed = nämnda.contains(&id.to_lowercase());

        ut.push(Source {
            id,
            title: anteckning.title,
            origin: anteckning.origin,
            added: anteckning.added,
            path: fil,
            indexed,
        });
    }

    ut
}
