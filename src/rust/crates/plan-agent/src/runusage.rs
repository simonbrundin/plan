//! Samlar tokenanvändningen för en agentkörning.
//!
//! Go: `internal/runusage`. Agenten rapporterar bara användningen.
//! Kostnaden räknas fram i Plan, ur en versionsstyrd prislista - så att
//! samma prislista gäller för alla körningar och kostnaden kan
//! reproduceras.

use std::collections::BTreeMap;
use std::sync::Mutex;

use chrono::{DateTime, Utc};

use crate::messages::Usage;

/// Den samlade användningen för en körning.
#[derive(Debug, Clone, Default)]
pub struct Snapshot {
    /// Leverantören som användes.
    pub provider: String,
    /// Modellen som användes.
    pub model: String,
    /// Modellens version, om providern uppger den.
    pub model_version: String,
    /// Antal indata-token.
    pub input_tokens: i64,
    /// Antal utdata-token.
    pub output_tokens: i64,
    /// Antal cache-lästa token.
    pub cache_read_tokens: i64,
    /// Antalet verktygsanrop.
    pub tool_calls: i64,
    /// De verktyg som anropats.
    pub tool_names: Vec<String>,
    /// När körningen startade.
    pub started_at: DateTime<Utc>,
    /// När körningen slutade.
    pub finished_at: DateTime<Utc>,
}

/// Samlar användningen för en körning.
///
/// Den är trådsäker, eftersom agentens verktyg kan köras parallellt.
pub struct Recorder {
    /// Leverantören.
    leverantör: String,
    /// Modellen.
    modell: String,
    /// Insamlingen.
    insamling: Mutex<Insamling>,
}

/// Den del som skyddas av låset.
#[derive(Debug, Clone)]
struct Insamling {
    /// Leverantören.
    leverantör: String,
    /// Modellen.
    modell: String,
    /// Modellens version.
    modellversion: String,
    /// Antal indata-token.
    input_tokens: i64,
    /// Antal utdata-token.
    output_tokens: i64,
    /// Antal cache-lästa token.
    cache_read_tokens: i64,
    /// Antalet verktygsanrop.
    tool_calls: i64,
    /// Verktygen som anropats.
    verktyg: BTreeMap<String, i64>,
    /// När körningen startade.
    started_at: DateTime<Utc>,
}

impl Recorder {
    /// Skapar en recorder för en leverantör och modell.
    ///
    /// Go: `runusage.New`.
    pub fn ny(leverantör: &str, modell: &str) -> Self {
        Self {
            leverantör: leverantör.to_string(),
            modell: modell.to_string(),
            insamling: Mutex::new(Insamling {
                leverantör: leverantör.to_string(),
                modell: modell.to_string(),
                modellversion: String::new(),
                input_tokens: 0,
                output_tokens: 0,
                cache_read_tokens: 0,
                tool_calls: 0,
                verktyg: BTreeMap::new(),
                started_at: Utc::now(),
            }),
        }
    }

    /// Användningen från ett modellsvar.
    ///
    /// Go: `Recorder.Record`.
    pub fn registrera(&self, användning: &Usage) {
        let mut insamling = match self.insamling.lock() {
            Ok(lås) => lås,
            Err(fel) => fel.into_inner(),
        };

        insamling.input_tokens += användning.input_tokens;
        insamling.output_tokens += användning.output_tokens;
        insamling.cache_read_tokens += användning.cache_tokens;
    }

    /// Registrerar ett verktygsanrop.
    ///
    /// Go: `Recorder.RecordTool`.
    pub fn registrera_verktyg(&self, namn: &str) {
        if namn.is_empty() {
            return;
        }

        let mut insamling = match self.insamling.lock() {
            Ok(lås) => lås,
            Err(fel) => fel.into_inner(),
        };

        *insamling.verktyg.entry(namn.to_string()).or_insert(0) += 1;
        insamling.tool_calls += 1;
    }

    /// Sparar modellens version.
    ///
    /// Go: `Recorder.SetModelVersion`.
    pub fn sätt_modellversion(&self, version: &str) {
        let mut insamling = match self.insamling.lock() {
            Ok(lås) => lås,
            Err(fel) => fel.into_inner(),
        };

        insamling.modellversion = version.to_string();
    }

    /// Den samlade användningen.
    ///
    /// Verktygsnamnen sorteras, så att samma körning alltid ger samma
    /// svar.
    ///
    /// Go: `Recorder.Snapshot`.
    pub fn snapshot(&self) -> Snapshot {
        let insamling = match self.insamling.lock() {
            Ok(lås) => lås,
            Err(fel) => fel.into_inner(),
        };

        Snapshot {
            provider: insamling.leverantör.clone(),
            model: insamling.modell.clone(),
            model_version: insamling.modellversion.clone(),
            input_tokens: insamling.input_tokens,
            output_tokens: insamling.output_tokens,
            cache_read_tokens: insamling.cache_read_tokens,
            tool_calls: insamling.tool_calls,
            tool_names: insamling.verktyg.keys().cloned().collect(),
            started_at: insamling.started_at,
            finished_at: Utc::now(),
        }
    }

    /// Tömmer insamlingen, så att samma recorder kan användas för nästa
    /// körning.
    ///
    /// Go: `Recorder.Reset`.
    pub fn reset(&self) {
        let mut insamling = match self.insamling.lock() {
            Ok(lås) => lås,
            Err(fel) => fel.into_inner(),
        };

        insamling.leverantör = self.leverantör.clone();
        insamling.modell = self.modell.clone();
        insamling.modellversion = String::new();
        insamling.input_tokens = 0;
        insamling.output_tokens = 0;
        insamling.cache_read_tokens = 0;
        insamling.tool_calls = 0;
        insamling.verktyg.clear();
        insamling.started_at = Utc::now();
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn användning_summeras() {
        let recorder = Recorder::ny("minimax", "MiniMax-M2.7");

        recorder.registrera(&Usage {
            input_tokens: 10,
            output_tokens: 5,
            cache_tokens: 2,
            ..Default::default()
        });
        recorder.registrera(&Usage {
            input_tokens: 3,
            output_tokens: 1,
            ..Default::default()
        });
        recorder.registrera_verktyg("bash");
        recorder.registrera_verktyg("bash");
        recorder.registrera_verktyg("read");

        let snapshot = recorder.snapshot();

        assert_eq!(snapshot.input_tokens, 13);
        assert_eq!(snapshot.output_tokens, 6);
        assert_eq!(snapshot.cache_read_tokens, 2);
        assert_eq!(snapshot.tool_calls, 3);
        assert_eq!(snapshot.tool_names, vec!["bash", "read"]);
        assert_eq!(snapshot.provider, "minimax");
    }

    #[test]
    fn reset_tömmer_insamlingen() {
        let recorder = Recorder::ny("minimax", "MiniMax-M2.7");

        recorder.registrera(&Usage {
            input_tokens: 10,
            ..Default::default()
        });
        recorder.reset();

        let snapshot = recorder.snapshot();

        assert_eq!(snapshot.input_tokens, 0);
        assert!(snapshot.tool_names.is_empty());
    }
}
