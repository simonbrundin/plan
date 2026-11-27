//! Säkerhetskontroller för sökvägar, kommandon och innehåll.
//!
//! Go: `internal/core/tools/sanitize.go` och `validator_security.go`.

use regex::Regex;
use std::path::Path;

/// Resultatet av en kontroll.
#[derive(Debug, Clone)]
pub struct ValidationResult {
    /// Om värdet är tillåtet.
    pub valid: bool,
    /// Förklaringen när värdet inte är tillåtet.
    pub message: String,
}

impl ValidationResult {
    /// Ett godkänt resultat.
    pub fn ok() -> Self {
        Self {
            valid: true,
            message: String::new(),
        }
    }

    /// Ett underkänt resultat.
    pub fn fel(text: impl Into<String>) -> Self {
        Self {
            valid: false,
            message: text.into(),
        }
    }
}

/// Sökvägar som alltid tillåts.
pub fn path_allowlist() -> Vec<String> {
    vec![
        "/home/simon".to_string(),
        "/tmp".to_string(),
        "/var".to_string(),
    ]
}

/// Sökvägar som alltid förbjuds.
pub fn path_denylist() -> Vec<String> {
    vec![
        "/etc/shadow".to_string(),
        "/etc/passwd".to_string(),
        "/etc/sudoers".to_string(),
        "/root/.ssh".to_string(),
        "/home/*/.ssh".to_string(),
        "/.git/objects".to_string(),
        "/.git/config".to_string(),
    ]
}

/// Kommandon som alltid förbjuds.
pub fn command_blocklist() -> Vec<&'static str> {
    vec![
        r"rm\s+-rf\s+/",
        r"dd\s+if=",
        r":\(\): \{ :\|:& \};:",
        r"fork\(\)",
        r"/proc/self",
        r"curl.*-X.*--data",
        r"wget.*--post-data",
        r"nc\s+-e",
        r"bash\s+-i",
        r"python.*-c\s+import\s+os",
        r"perl.*-e.*system",
    ]
}

/// Kontrollerar att en sökväg får användas.
///
/// Go: `ValidatePathStrict`.
pub fn validera_sökväg(sökväg: &str) -> ValidationResult {
    // Null byte-injektion.
    if sökväg.contains('\0') {
        return ValidationResult::fel("path contains null byte");
    }

    let rensad = rensa_sökväg(sökväg);

    // Traversering.
    if sökväg.contains("..") {
        return ValidationResult::fel("path traversal not allowed");
    }

    // Misstänkta mönster - alltid förbjudna.
    let misstänkta = [
        "/.git/hooks",
        "/.github/workflows",
        "/__pycache__",
        "/node_modules/.bin",
    ];

    for mönster in misstänkta {
        if rensad.contains(mönster) {
            return ValidationResult::fel(format!("suspicious path pattern: {mönster}"));
        }
    }

    // Nekad lista.
    for nekad in path_denylist() {
        if rensad.starts_with(&nekad) || rensad == nekad {
            return ValidationResult::fel("path is in denylist");
        }
    }

    // Tillåten lista.
    let tillåten = path_allowlist();

    if !tillåten.is_empty() && !tillåten.iter().any(|väg| rensad.starts_with(väg)) {
        return ValidationResult::fel("path not in allowlist");
    }

    ValidationResult::ok()
}

/// Kontrollerar att ett kommando är säkert att köra.
///
/// Go: `ValidateCommandStrict`.
pub fn validera_kommando(kommando: &str) -> ValidationResult {
    if kommando.contains('\0') {
        return ValidationResult::fel("command contains null byte");
    }

    for blockerat in command_blocklist() {
        let Ok(regex) = Regex::new(blockerat) else {
            continue;
        };

        if regex.is_match(kommando) {
            return ValidationResult::fel("command matches blocklist pattern");
        }
    }

    // Farliga omdirigeringar.
    let omdirigeringar = [
        ">/etc/passwd",
        ">/etc/shadow",
        ">/~/.ssh",
        ">/dev/sda",
        ">/dev/null",
        "2>&1",
    ];

    for omdirigering in omdirigeringar {
        if kommando.contains(omdirigering) {
            return ValidationResult::fel(format!("dangerous redirection: {omdirigering}"));
        }
    }

    // Rör till ett skal.
    if let Ok(regex) = Regex::new(r"\|\s*sh\b") {
        if regex.is_match(kommando) {
            return ValidationResult::fel("pipe to shell not allowed");
        }
    }

    // Kommandotsubstitution.
    if kommando.contains("$(") || kommando.contains('`') {
        return ValidationResult::fel("command substitution not allowed");
    }

    // Miljövariabelutvidgning.
    if let Ok(regex) = Regex::new(r"\$\{?[A-Za-z_][A-Za-z0-9_]*\}?") {
        if regex.is_match(kommando) {
            return ValidationResult::fel("environment variable expansion not allowed");
        }
    }

    ValidationResult::ok()
}

/// Kontrollerar en sökväg utan den strikta listan.
///
/// Go: `PathValidator.Validate`.
pub fn validera_sökväg_lätt(sökväg: &str) -> Result<(), ValidationError> {
    if sökväg.is_empty() {
        return Err(ValidationError {
            field: "path".to_string(),
            message: "path cannot be empty".to_string(),
        });
    }

    let farliga = ["../", "~", "$", "`", ";", "|", "<", ">"];

    for sekvens in farliga {
        if sökväg.contains(sekvens) && !sökväg.starts_with("./") {
            return Err(ValidationError {
                field: "path".to_string(),
                message: format!("path contains potentially dangerous sequence: {sekvens}"),
            });
        }
    }

    Ok(())
}

/// Kontrollerar ett kommando utan den strikta listan.
///
/// Go: `ValidateCommand`.
pub fn validera_kommando_lätt(kommando: &str) -> Result<(), ValidationError> {
    if kommando.is_empty() {
        return Err(ValidationError {
            field: "command".to_string(),
            message: "command cannot be empty".to_string(),
        });
    }

    let farliga = [
        "rm -rf /",
        ":(){:|:&};:",
        "dd if=/dev/zero",
        "> /dev/sda",
        "mkfs",
        "chmod -R 777 /",
    ];

    let lägre = kommando.to_lowercase();

    for mönster in farliga {
        if lägre.contains(&mönster.to_lowercase()) {
            return Err(ValidationError {
                field: "command".to_string(),
                message: "command contains potentially dangerous pattern".to_string(),
            });
        }
    }

    Ok(())
}

/// Rensar en sökväg.
///
/// Go: `SanitizePathStrict`.
pub fn rensa_sökväg(sökväg: &str) -> String {
    let utan_null = sökväg.replace('\0', "");

    rensad(&utan_null).trim().to_string()
}

/// Rensar ett kommando.
///
/// Go: `SanitizeCommandStrict`.
pub fn rensa_kommando(kommando: &str) -> String {
    let utan_null = kommando.replace('\0', "");

    let utan_backtick = utan_null.replace('`', "");

    match Regex::new(r"\$\([^)]*\)") {
        Ok(regex) => regex.replace_all(&utan_backtick, "").to_string(),
        Err(_) => utan_backtick,
    }
}

/// Rensar en sökväg som Go:s `filepath.Clean` gör.
pub fn rensad(sökväg: &str) -> String {
    if sökväg.is_empty() {
        return ".".to_string();
    }

    let absolut = sökväg.starts_with('/');

    let mut delar: Vec<&str> = Vec::new();

    for del in sökväg.split('/') {
        match del {
            "" | "." => {}
            ".." => {
                if let Some(sista) = delar.last() {
                    if *sista != ".." {
                        delar.pop();
                        continue;
                    }
                }

                if !absolut {
                    delar.push("..");
                }
            }
            annan => delar.push(annan),
        }
    }

    let sammansatt = delar.join("/");

    if absolut {
        format!("/{sammansatt}")
    } else if sammansatt.is_empty() {
        ".".to_string()
    } else {
        sammansatt
    }
}

/// En validator av innehåll som ska skrivas till disk.
pub struct ContentValidator {
    /// Högsta tillåtna storlek.
    pub max_size: usize,
    /// Förbjudna filändelser.
    pub blocked_exts: Vec<String>,
}

/// En validator av innehåll med standardinställningar.
impl Default for ContentValidator {
    fn default() -> Self {
        Self {
            max_size: 10 * 1024 * 1024,
            blocked_exts: vec![
                ".exe".to_string(),
                ".dll".to_string(),
                ".so".to_string(),
                ".dylib".to_string(),
            ],
        }
    }
}

impl ContentValidator {
    /// Kontrollerar att innehållet får skrivas.
    ///
    /// Go: `ContentValidator.ValidateContent`.
    pub fn validera_innehåll(&self, innehåll: &[u8], filnamn: &str) -> ValidationResult {
        if innehåll.len() > self.max_size {
            return ValidationResult::fel("content exceeds maximum size");
        }

        let ändelse = Path::new(filnamn)
            .extension()
            .map(|e| e.to_string_lossy().to_lowercase())
            .unwrap_or_default();

        for blockerad in &self.blocked_exts {
            if ändelse == *blockerad {
                return ValidationResult::fel(format!("file extension not allowed: {ändelse}"));
            }
        }

        // Shebang-manipulation.
        if innehåll.len() > 2 && innehåll[0] == b'#' && innehåll[1] == b'!' {
            return ValidationResult::fel("shebang not allowed");
        }

        if innehåll.contains(&0) {
            return ValidationResult::fel("content contains null bytes");
        }

        ValidationResult::ok()
    }
}

/// Ett valideringsfel.
#[derive(Debug, Clone)]
pub struct ValidationError {
    /// Fältet som var fel.
    pub field: String,
    /// Beskrivningen av felet.
    pub message: String,
}

impl std::fmt::Display for ValidationError {
    fn fmt(&self, form: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(
            form,
            "validation failed for '{}': {}",
            self.field, self.message
        )
    }
}

impl std::error::Error for ValidationError {}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn null_byte_i_sokvag_avvisas() {
        let resultat = validera_sökväg("/tmp/a\0b");

        assert!(!resultat.valid);
        assert_eq!(resultat.message, "path contains null byte");
    }

    #[test]
    fn traversal_avvisas() {
        let resultat = validera_sökväg("/tmp/../etc/passwd");

        assert!(!resultat.valid);
        assert_eq!(resultat.message, "path traversal not allowed");
    }

    #[test]
    fn kommando_substitution_avvisas() {
        let resultat = validera_kommando("echo $(whoami)");

        assert!(!resultat.valid);
        assert_eq!(resultat.message, "command substitution not allowed");
    }

    #[test]
    fn miljovariabel_avvisas() {
        let resultat = validera_kommando("echo $HOME");

        assert!(!resultat.valid);
        assert_eq!(
            resultat.message,
            "environment variable expansion not allowed"
        );
    }

    #[test]
    fn vanligt_kommando_slapper_igenom() {
        assert!(validera_kommando("ls -la /tmp").valid);
    }

    #[test]
    fn shebang_i_innehall_avvisas() {
        let validator = ContentValidator::default();

        let resultat = validator.validera_innehåll(b"#!/bin/sh\necho hej", "/tmp/a.sh");

        assert!(!resultat.valid);
        assert_eq!(resultat.message, "shebang not allowed");
    }

    #[test]
    fn soker_igenom_och_tar_bort_dubblar() {
        assert_eq!(rensad("/tmp//a/./b"), "/tmp/a/b");
        assert_eq!(rensad("/tmp/a/../b"), "/tmp/b");
        assert_eq!(rensad("a/../.."), "..");
    }
}
