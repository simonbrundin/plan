//! Loggen över händelser i wikin.
//!
//! Go: `internal/wiki/log.go`. Loggen är append-only och medvetet
//! grep-bart: `grep "^## \[" log.md | tail -5`.

use std::path::Path;

/// En rad i loggen.
#[derive(Debug, Clone)]
pub struct LogEntry {
    /// Datum, YYYY-MM-DD.
    pub date: String,
    /// Vad som hände, t.ex. `ingest`.
    pub op: String,
    /// Vad det gällde.
    pub title: String,
}

impl std::fmt::Display for LogEntry {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "## [{}] {} | {}", self.date, self.op, self.title)
    }
}

/// Lägger till poster i loggen.
///
/// Go: `wiki.AppendLog`.
pub fn append_log(dir: &Path, entries: &[LogEntry], details: &[String]) -> Result<(), String> {
    let path = dir.join(crate::wiki::LOG_FILE);

    let existing = match std::fs::read_to_string(&path) {
        Ok(innehåll) => innehåll,
        Err(err) if err.kind() == std::io::ErrorKind::NotFound => String::new(),
        Err(err) => return Err(format!("kunde inte läsa {}: {err}", path.display())),
    };

    let mut ut = existing.clone();

    if !existing.is_empty() && !existing.ends_with('\n') {
        ut.push('\n');
    }

    if !existing.is_empty() {
        ut.push('\n');
    }

    for (index, entry) in entries.iter().enumerate() {
        ut.push_str(&entry.to_string());
        ut.push('\n');

        if index == 0 {
            for detalj in details {
                ut.push_str(&format!("- {detalj}\n"));
            }
        }
    }

    ut.push('\n');

    crate::wiki::write_file(&path, &ut)
}
