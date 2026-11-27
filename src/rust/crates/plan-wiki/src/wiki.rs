//! Kunskapsgrafen - en laddad wiki med sidor, länkar och sökning.
//!
//! Go: `internal/wiki/wiki.go`.

use std::collections::BTreeMap;
use std::path::{Path, PathBuf};

use crate::note::{parse_note, Note};

/// Katalogen med genererade sidor.
pub const PAGES_DIR: &str = "wiki";
/// Katalogen där inlägg landar.
pub const INBOX_DIR: &str = "inbox";
/// Katalogen där inlägg arkiveras.
pub const ARCHIVE_DIR: &str = "archive";
/// Katalogen för sandlådan.
pub const SANDBOX_DIR: &str = "sandbox";
/// Katalogen med immutabla råkällor.
pub const RAW_DIR: &str = "raw";
/// Den genererade katalogen.
pub const CATALOG_FILE: &str = "catalog.md";
/// Den genererade länkgraphen.
pub const GRAPH_FILE: &str = "graph.md";
/// Loggen över händelser.
pub const LOG_FILE: &str = "log.md";
/// Katalogen där arbetsnoter landar.
pub const NOTES_DIR: &str = "notes";

/// En laddad kunskapsgraf.
#[derive(Debug, Clone, Default)]
pub struct Wiki {
    /// Katalogen wikin ligger i.
    pub dir: PathBuf,
    /// Alla sidor i grafen.
    pub notes: Vec<Note>,
}

/// Laddar alla sidor från `dir/wiki`.
///
/// Go: `wiki.Open`.
pub fn open(dir: &Path) -> Result<Wiki, String> {
    let pages_root = dir.join(PAGES_DIR);
    let mut wiki = Wiki {
        dir: dir.to_path_buf(),
        notes: Vec::new(),
    };

    samla_sidor(&pages_root, &pages_root, &mut wiki.notes)?;

    wiki.sort_notes();

    Ok(wiki)
}

/// Går igenom en katalog rekursivt och läser alla markdownfiler.
fn samla_sidor(rot: &Path, aktuell: &Path, anteckningar: &mut Vec<Note>) -> Result<(), String> {
    let poster = match std::fs::read_dir(aktuell) {
        Ok(poster) => poster,
        Err(err) if err.kind() == std::io::ErrorKind::NotFound => return Ok(()),
        Err(err) => {
            return Err(format!(
                "kunde inte läsa {}: {}",
                aktuell.display(),
                go_filfel("open", aktuell, &err)
            ))
        }
    };

    let mut namn: Vec<PathBuf> = poster
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

    namn.sort();

    for sökväg in namn {
        let mut anteckning = parse_note(&sökväg)
            .map_err(|err| format!("kunde inte läsa {}: {err}", sökväg.display()))?;

        let relativ = sökväg
            .strip_prefix(rot)
            .map_err(|err| format!("kunde inte hitta relativ sökväg: {err}"))?;

        anteckning.id = relativ
            .to_string_lossy()
            .replace('\\', "/")
            .trim_end_matches(".md")
            .to_string();

        anteckningar.push(anteckning);
    }

    // Underkataloger kommer efter sidorna i samma katalog - ordningen
    // påverkar inte innehållet, bara läsningen.
    let under: Vec<PathBuf> = std::fs::read_dir(aktuell)
        .map_err(|err| {
            format!(
                "kunde inte läsa {}: {}",
                aktuell.display(),
                go_filfel("open", aktuell, &err)
            )
        })?
        .filter_map(|poster| poster.ok())
        .map(|poster| poster.path())
        .filter(|sökväg| sökväg.is_dir())
        .collect();

    for katalog in under {
        samla_sidor(rot, &katalog, anteckningar)?;
    }

    Ok(())
}

impl Wiki {
    /// Sorterar sidorna på id.
    fn sort_notes(&mut self) {
        self.notes.sort_by(|a, b| a.id.cmp(&b.id));
    }

    /// Returnerar anteckningen med givet id, eller ingenting.
    pub fn note(&self, id: &str) -> Option<&Note> {
        self.notes
            .iter()
            .find(|anteckning| anteckning.id.eq_ignore_ascii_case(id))
    }

    /// Hittar sidan en länk pekar på.
    ///
    /// Länkar kan skrivas som `[[concepts/Flux]]`, `[[Concepts/Flux]]` eller
    /// som markdown-länkar, och matchningen är tolerant mot skiftläge,
    /// bindestreck och slug-form.
    ///
    /// Go: `Wiki.Resolve`.
    pub fn resolve(&self, target: &str) -> Option<&Note> {
        let target = target.trim().trim_end_matches(".md");

        if target.is_empty() {
            return None;
        }

        // 1. Exakt id, case-insensitive.
        if let Some(anteckning) = self.note(target) {
            return Some(anteckning);
        }

        // 2. Samma sista segment - "Summaries/Longhorn disk guide" ->
        //    "summaries/longhorn-disk-guide".
        let target_base = Path::new(&target.replace('\\', "/"))
            .file_name()
            .map(|namn| namn.to_string_lossy().to_string())
            .unwrap_or_else(|| target.to_string());

        if let Some(anteckning) = self.notes.iter().find(|anteckning| {
            Path::new(&anteckning.id)
                .file_name()
                .map(|namn| namn.to_string_lossy().eq_ignore_ascii_case(&target_base))
                .unwrap_or(false)
        }) {
            return Some(anteckning);
        }

        // 3. Samma slug - "Longhorn disk guide" -> "longhorn-disk-guide".
        let target_slug = crate::note::slug(&target_base);

        self.notes.iter().find(|anteckning| {
            Path::new(&anteckning.id)
                .file_name()
                .map(|namn| crate::note::slug(&namn.to_string_lossy()) == target_slug)
                .unwrap_or(false)
        })
    }

    /// Hittar sidor som matchar frågan, rangordnade efter relevans.
    ///
    /// Rubriken och taggarna väger tyngre än brödtexten.
    ///
    /// Go: `Wiki.Search`.
    pub fn search(&self, query: &str) -> Vec<Match> {
        let terms: Vec<String> = query
            .to_lowercase()
            .split_whitespace()
            .map(str::to_string)
            .collect();

        if terms.is_empty() {
            return Vec::new();
        }

        let mut träffar: Vec<Match> = Vec::new();

        for anteckning in &self.notes {
            let mut poäng: i32 = 0;
            let rubrik = anteckning.title.to_lowercase();
            let taggar = anteckning.tags.join(" ").to_lowercase();
            let body = anteckning.body.to_lowercase();

            for term in &terms {
                if rubrik.contains(term) {
                    poäng += 10;
                }

                if taggar.contains(term) {
                    poäng += 6;
                }

                poäng += body.matches(term.as_str()).count() as i32;
            }

            if poäng == 0 {
                continue;
            }

            träffar.push(Match {
                note: anteckning.clone(),
                score: poäng,
                lines: context_lines(anteckning, &terms),
            });
        }

        träffar.sort_by(|a, b| {
            b.score
                .cmp(&a.score)
                .then_with(|| a.note.id.cmp(&b.note.id))
        });

        träffar
    }

    /// Returnerar sidor som länkar till det givna id:t.
    ///
    /// Go: `Wiki.Backlinks`.
    pub fn backlinks(&self, id: &str) -> Vec<&Note> {
        let mut ut = Vec::new();

        for anteckning in &self.notes {
            for länk in self.links_of(anteckning) {
                if länk.id.eq_ignore_ascii_case(id) {
                    ut.push(anteckning);
                    break;
                }
            }
        }

        ut
    }

    /// Returnerar de sidor anteckningen pekar på.
    pub fn links_of(&self, anteckning: &Note) -> Vec<&Note> {
        anteckning
            .all_related()
            .iter()
            .filter_map(|mål| self.resolve(mål))
            .collect()
    }

    /// Returnerar alla taggar i wikin med sin frekvens.
    pub fn tags(&self) -> BTreeMap<String, usize> {
        let mut counts: BTreeMap<String, usize> = BTreeMap::new();

        for anteckning in &self.notes {
            for tagg in &anteckning.tags {
                *counts.entry(tagg.clone()).or_insert(0) += 1;
            }
        }

        counts
    }

    /// Renderar den genererade katalogen.
    ///
    /// Go: `Wiki.Catalog`.
    pub fn catalog(&self) -> String {
        let mut ut = String::new();

        ut.push_str("<!-- GENERERAD av `wikictl index`. Ändra inte manuellt. -->\n");
        ut.push_str("# Katalog\n\n");
        ut.push_str("Alla sidor i wikin, sorterade på id. Den kuraterade katalogen\n");
        ut.push_str("finns i `index.md`.\n\n");

        if self.notes.is_empty() {
            ut.push_str("_Inga sidor ännu._\n");
            return ut;
        }

        ut.push_str("| Sida | Titel | Typ | Taggar | Uppdaterad | Länkar |\n");
        ut.push_str("|---|---|---|---|---|---|\n");

        for anteckning in &self.notes {
            let rubrik = anteckning.title.replace('|', "\\|");
            let taggar = if anteckning.tags.is_empty() {
                "-".to_string()
            } else {
                anteckning.tags.join(", ")
            };
            let typ = if anteckning.note_type.is_empty() {
                "-".to_string()
            } else {
                anteckning.note_type.clone()
            };
            let uppdaterad = if anteckning.updated.is_empty() {
                "-".to_string()
            } else {
                anteckning.updated.clone()
            };

            ut.push_str(&format!(
                "| [[{}]] | {} | {} | {} | {} | {} |\n",
                anteckning.id,
                rubrik,
                typ,
                taggar,
                uppdaterad,
                self.links_of(anteckning).len()
            ));
        }

        ut.push_str("\n## Taggar\n\n");

        for (tagg, antal) in self.tags() {
            ut.push_str(&format!("- `{tagg}` ({antal})\n"));
        }

        ut
    }

    /// Renderar den genererade länkgraphen.
    ///
    /// Go: `Wiki.Graph`.
    pub fn graph(&self) -> String {
        let mut ut = String::new();

        ut.push_str("<!-- GENERERAD av `wikictl graph`. Ändra inte manuellt. -->\n");
        ut.push_str("# Länkgraph\n\n");
        ut.push_str("Kunskapsgrafen: varje nod är en sida, varje kant är en länk mellan dem.\n\n");
        ut.push_str("```mermaid\n");
        ut.push_str("graph LR\n");

        for anteckning in &self.notes {
            ut.push_str(&format!(
                "  {}[\"{}\"]\n",
                node_id(&anteckning.id),
                escape_label(&anteckning.title)
            ));
        }

        for anteckning in &self.notes {
            for länk in self.links_of(anteckning) {
                ut.push_str(&format!(
                    "  {} --> {}\n",
                    node_id(&anteckning.id),
                    node_id(&länk.id)
                ));
            }
        }

        ut.push_str("```\n\n");
        ut.push_str("## Backlinks\n\n");

        for anteckning in &self.notes {
            let tillbaka = self.backlinks(&anteckning.id);

            if tillbaka.is_empty() {
                continue;
            }

            let namn: Vec<String> = tillbaka
                .iter()
                .map(|anteckning| format!("[[{}]]", anteckning.id))
                .collect();

            ut.push_str(&format!("- **{}** <- {}\n", anteckning.id, namn.join(", ")));
        }

        ut
    }

    /// Skriver index.md.
    pub fn write_index(&self) -> Result<(), String> {
        write_file(
            &self.dir.join(crate::index::INDEX_FILE),
            &crate::index::index_by_kind(self),
        )
    }

    /// Skriver catalog.md.
    pub fn write_catalog(&self) -> Result<(), String> {
        write_file(&self.dir.join(CATALOG_FILE), &self.catalog())
    }

    /// Skriver graph.md.
    pub fn write_graph(&self) -> Result<(), String> {
        write_file(&self.dir.join(GRAPH_FILE), &self.graph())
    }
}

/// Ett sökträff i en sida.
#[derive(Debug, Clone)]
pub struct Match {
    /// Sidan som träffade.
    pub note: Note,
    /// Hur viktigt träffen är.
    pub score: i32,
    /// Rader ur brödtexten som visar träffen.
    pub lines: Vec<String>,
}

/// Returnerar upp till tre brödtextrader som innehåller söktermerna.
fn context_lines(anteckning: &Note, terms: &[String]) -> Vec<String> {
    const MAX_LINES: usize = 3;

    let mut rader: Vec<String> = Vec::new();

    for (index, rad) in anteckning.body.split('\n').enumerate() {
        let lägre = rad.to_lowercase();

        for term in terms {
            if lägre.contains(term) {
                rader.push(format!("{:4}| {}", index + 1, rad.trim()));
                break;
            }
        }

        if rader.len() == MAX_LINES {
            break;
        }
    }

    rader
}

/// Gör ett id säkert att använda som mermaid-nod.
fn node_id(id: &str) -> String {
    format!("n_{}", id.replace(['/', ' ', '-', '.'], "_"))
}

/// Escapes citattecken i mermaid-etiketter.
fn escape_label(label: &str) -> String {
    label.replace('"', "'")
}

/// Skriver en fil och skapar föräldrakatalogerna vid behov.
pub fn write_file(sökväg: &Path, innehåll: &str) -> Result<(), String> {
    if let Some(förälder) = sökväg.parent() {
        std::fs::create_dir_all(förälder).map_err(|err| {
            format!(
                "kunde inte skapa {}: {go_filfel}",
                förälder.display(),
                go_filfel = go_filfel("mkdir", förälder, &err)
            )
        })?;
    }

    std::fs::write(sökväg, innehåll).map_err(|err| {
        format!(
            "kunde inte skriva {}: {go_filfel}",
            sökväg.display(),
            go_filfel = go_filfel("open", sökväg, &err)
        )
    })
}

/// Formaterar ett filfel på samma sätt som Go gör.
///
/// Go använder `open filnamn: no such file or directory` - Rust ger
/// `No such file or directory (os error 2)`. Skillnaden syns i
/// felmeddelandena, så den måste matchas.
///
/// Go: `open <sökväg>: <errno>`.
pub fn go_filfel(verb: &str, sökväg: &Path, err: &std::io::Error) -> String {
    format!("{verb} {}: {}", sökväg.display(), go_errno(err))
}

/// Go:s errno-texter.
pub fn go_errno(err: &std::io::Error) -> String {
    match err.raw_os_error() {
        Some(1) => "operation not permitted".to_string(),
        Some(2) => "no such file or directory".to_string(),
        Some(13) => "permission denied".to_string(),
        Some(17) => "file exists".to_string(),
        Some(20) => "not a directory".to_string(),
        Some(21) => "is a directory".to_string(),
        Some(22) => "invalid argument".to_string(),
        Some(39) => "directory not empty".to_string(),
        _ => err.to_string(),
    }
}

/// Sätter ihop sökvägar som Go:s `filepath.Join` gör.
pub fn go_join(base: &str, namn: &str) -> String {
    let samman = if base.is_empty() {
        namn.to_string()
    } else {
        format!("{base}/{namn}")
    };

    let mut städad = samman.replace("//", "/");

    while städad.starts_with("./") {
        städad = städad[2..].to_string();
    }

    städad
}

/// Hittar wikin.
///
/// Prioritet: explicit sökväg, sedan `$WIKI_DIR`, sedan en `wiki/`
/// som hittas genom att gå uppåt från arbetskatalogen, och till sist
/// den globala `~/wiki`.
///
/// Go: `wiki.ResolveDir`.
pub fn resolve_dir(explicit: &str) -> Result<PathBuf, String> {
    if !explicit.is_empty() {
        return Ok(PathBuf::from(explicit));
    }

    if let Some(env) = std::env::var("WIKI_DIR").ok().filter(|v| !v.is_empty()) {
        return Ok(PathBuf::from(env));
    }

    let mut dir =
        std::env::current_dir().map_err(|err| format!("kunde inte läsa arbetskatalogen: {err}"))?;

    loop {
        let kandidat = dir.join(PAGES_DIR);

        if kandidat.join(PAGES_DIR).is_dir() {
            return Ok(kandidat);
        }

        match dir.parent() {
            Some(förälder) if förälder != dir => dir = förälder.to_path_buf(),
            _ => break,
        }
    }

    if let Some(hem) = hemkatalog() {
        let global = hem.join(PAGES_DIR);

        if global.join(PAGES_DIR).is_dir() {
            return Ok(global);
        }
    }

    Err("hittade inte wikin - kör från repot eller sätt WIKI_DIR".to_string())
}

/// Hemkatalogen.
pub fn hemkatalog() -> Option<PathBuf> {
    std::env::var("HOME")
        .ok()
        .filter(|hem| !hem.is_empty())
        .map(PathBuf::from)
}
