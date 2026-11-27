//! Projektkoppling - vilket repo en anteckning gäller.
//!
//! Go: `internal/wiki/project.go`.

use std::path::Path;

/// Prefixet som markerar projektet en anteckning gäller.
pub const PROJECT_TAG_PREFIX: &str = "projekt:";

/// Returnerar namnet på projektet (repot) som katalogen tillhör.
///
/// Det görs genom att fråga git om repo-roten; faller det tillbaka på
/// katalognamnet.
///
/// Go: `wiki.DetectProject`.
pub fn detect_project(dir: &str) -> String {
    let katalog = if dir.is_empty() {
        match std::env::current_dir() {
            Ok(katalog) => katalog,
            Err(_) => return String::new(),
        }
    } else {
        Path::new(dir).to_path_buf()
    };

    let namn = git_repo_name(&katalog);

    if !namn.is_empty() {
        return namn;
    }

    katalog
        .file_name()
        .map(|namn| namn.to_string_lossy().to_string())
        .unwrap_or_default()
}

/// Frågar git vilket repo katalogen tillhör.
fn git_repo_name(dir: &Path) -> String {
    let utdata = std::process::Command::new("git")
        .arg("-C")
        .arg(dir)
        .args(["rev-parse", "--show-toplevel"])
        .output();

    let Ok(utdata) = utdata else {
        return String::new();
    };

    if !utdata.status.success() {
        return String::new();
    }

    Path::new(String::from_utf8_lossy(&utdata.stdout).trim())
        .file_name()
        .map(|namn| namn.to_string_lossy().to_string())
        .unwrap_or_default()
}

/// Bygger taggen som markerar ett projekt.
///
/// Go: `wiki.ProjectTag`.
pub fn project_tag(project: &str) -> String {
    if project.is_empty() {
        return String::new();
    }

    format!("{PROJECT_TAG_PREFIX}{project}")
}
