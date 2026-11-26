//! Kontraktet mellan Plan och agenten.
//!
//! Filen ligger i `src/contracts/plan-agent.yaml` och beskriver de delar
//! agenten använder: faser, tillstånd, triggrar, profiler och API-anropen.
//! Båda sidor testar mot samma fil - om en sida ändras utan den andra
//! faller testerna.

use std::path::{Path, PathBuf};

use serde::Deserialize;

/// Ett API-anrop agenten gör mot Plan.
#[derive(Debug, Clone, Deserialize)]
pub struct Endpoint {
    pub method: String,
    pub path: String,
    /// Sant om agenten själv gör anropet. Plan ska tillåta det för
    /// agent-scope - annars får agenten 403 när den använder sin egen nyckel.
    #[serde(default)]
    pub agent: bool,
}

/// Kontraktet mellan Plan och agenten.
#[derive(Debug, Clone, Deserialize)]
pub struct Contract {
    pub version: i32,
    #[serde(default)]
    pub phases: Vec<String>,
    #[serde(default)]
    pub statuses: Vec<String>,
    #[serde(default)]
    pub terminal_statuses: Vec<String>,
    #[serde(default)]
    pub status_needing_human: Vec<String>,
    #[serde(default)]
    pub triggers: Vec<String>,
    #[serde(default)]
    pub profiles: Vec<String>,
    /// Typer av godkännanden agenten kan skapa.
    #[serde(default)]
    pub approval_kinds: Vec<String>,
    #[serde(default)]
    pub endpoints: Vec<Endpoint>,
}

/// Filnamnet på kontraktet.
pub const FILENAME: &str = "plan-agent.yaml";

impl Contract {
    /// Läser kontraktet från disk.
    pub fn load() -> Result<Self, String> {
        let path =
            Self::path().ok_or_else(|| format!("hittade inte kontraktet contracts/{FILENAME}"))?;

        let raw = std::fs::read_to_string(&path)
            .map_err(|err| format!("kan inte läsa kontraktet: {err}"))?;

        serde_yaml::from_str(&raw).map_err(|err| format!("kan inte tolka kontraktet: {err}"))
    }

    /// Sökvägen till kontraktet.
    ///
    /// Sökningen börjar i crate-katalogen och går uppåt tills filen hittas
    /// - så fungerar den oavsett varifrån testet körs.
    pub fn path() -> Option<PathBuf> {
        let mut katalog: &Path = Path::new(env!("CARGO_MANIFEST_DIR"));

        loop {
            let kandidat = katalog.join("contracts").join(FILENAME);
            if kandidat.exists() {
                return Some(kandidat);
            }

            katalog = katalog.parent()?;
        }
    }

    /// Har kontraktet ett anrop med den metoden och sökvägen?
    pub fn has_endpoint(&self, method: &str, path: &str) -> bool {
        self.endpoints
            .iter()
            .any(|endpoint| endpoint.method == method && endpoint.path == path)
    }

    /// Är sökvägen markerad som agentens eget anrop?
    pub fn is_agent_endpoint(&self, path: &str) -> bool {
        self.endpoints
            .iter()
            .any(|endpoint| endpoint.agent && endpoint.path == path)
    }
}
