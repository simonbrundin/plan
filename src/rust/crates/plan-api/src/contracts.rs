//! Kontraktet mellan Plan och agenten.
//!
//! Kontraktet ligger i `src/contracts/plan-agent.yaml` och beskriver de
//! delar agenten använder: faser, tillstånd, triggrar, profiler och
//! godkännandetyper. Go-tjänsten läser samma fil
//! (`src/api/internal/contracts`) - därför söker även Rust-tjänsten uppåt
//! i katalogträdet efter filen.
//!
//! Finns filen inte - till exempel i produktionsbehållaren, där bara
//! binären finns med - tillåts alla godkännandetyper. Ett förslag ska inte
//! stoppas av att en fil saknas.

use std::path::{Path, PathBuf};
use std::sync::OnceLock;

use serde::Deserialize;

/// Filnamnet på kontraktet.
pub const FILENAME: &str = "plan-agent.yaml";

/// Kontraktet mellan Plan och agenten.
#[derive(Debug, Clone, Deserialize)]
pub struct Contract {
    /// Versionen av kontraktet.
    #[serde(default)]
    pub version: i64,
    /// Agentkörningens faser.
    #[serde(default)]
    pub phases: Vec<String>,
    /// Agentkörningens tillstånd.
    #[serde(default)]
    pub statuses: Vec<String>,
    /// Tillstånd som en körning inte kan lämna.
    #[serde(default)]
    pub terminal_statuses: Vec<String>,
    /// Tillstånd där en människa måste agera.
    #[serde(default)]
    pub status_needing_human: Vec<String>,
    /// Vad som startar en körning.
    #[serde(default)]
    pub triggers: Vec<String>,
    /// Vilken sorts agent som utför körningen.
    #[serde(default)]
    pub profiles: Vec<String>,
    /// Typer av godkännanden agenten kan skapa.
    #[serde(default)]
    pub approval_kinds: Vec<String>,
}

impl Contract {
    /// Om en godkännandetyp finns i kontraktet.
    pub fn har_godkännandetyp(&self, kind: &str) -> bool {
        self.approval_kinds.iter().any(|typ| typ == kind)
    }
}

/// Cachar kontraktet - filen läses en gång per process.
static KONTRAKT: OnceLock<Option<Contract>> = OnceLock::new();

/// Läser kontraktet, om det finns.
///
/// `None` betyder att filen inte kunde läsas - då tillåts alla typer.
pub fn shared() -> Option<&'static Contract> {
    KONTRAKT.get_or_init(läs_kontrakt).as_ref()
}

/// Sökvägen till kontraktet.
///
/// Sökningen börjar i den katalog processen startade i och går uppåt tills
/// filen hittas - så fungerar den oavsett vilken katalog tjänsten startas
/// från.
pub fn sökväg() -> Option<PathBuf> {
    if let Ok(sätt) = std::env::var("PLAN_CONTRACT_PATH") {
        let sökväg = PathBuf::from(sätt);
        if sökväg.is_file() {
            return Some(sökväg);
        }
    }

    if let Ok(katalog) = std::env::current_dir() {
        if let Some(träff) = sök_uppåt(&katalog) {
            return Some(träff);
        }
    }

    // I tester och i utvecklingsläge når källkoden filen via sin egen
    // katalog - samma trick som Go-tjänsten använder.
    let manifest = PathBuf::from(env!("CARGO_MANIFEST_DIR"));
    sök_uppåt(&manifest)
}

/// Söker uppåt efter `contracts/plan-agent.yaml`.
fn sök_uppåt(start: &Path) -> Option<PathBuf> {
    let mut katalog = start.to_path_buf();

    loop {
        let kandidat = katalog.join("contracts").join(FILENAME);
        if kandidat.is_file() {
            return Some(kandidat);
        }

        match katalog.parent() {
            Some(förälder) if förälder != katalog => katalog = förälder.to_path_buf(),
            _ => return None,
        }
    }
}

/// Läser och tolkar kontraktet.
fn läs_kontrakt() -> Option<Contract> {
    let sökväg = sökväg()?;

    let rå = std::fs::read_to_string(&sökväg).ok()?;
    serde_yaml::from_str(&rå).ok()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn kontraktet_kan_lasas_och_har_godkannandetyper() {
        let kontrakt = shared().expect("kontraktet ska finnas i repot");

        assert!(kontrakt.har_godkännandetyp("code_change"));
        assert!(kontrakt.har_godkännandetyp("goal_definition"));
        assert!(!kontrakt.har_godkännandetyp("nonsense"));
    }

    #[test]
    fn faserna_och_tillstanden_overensstammer_med_modellerna() {
        let kontrakt = shared().expect("kontraktet ska finnas i repot");

        assert_eq!(kontrakt.phases, crate::models::agent::PHASES);
        assert_eq!(kontrakt.statuses, crate::models::agent::STATUSES);
    }
}
