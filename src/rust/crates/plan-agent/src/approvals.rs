//! Agentens godkännandestore.
//!
//! Go-agentens `internal/approvals/store.go` - förslag som väntar på ett
//! mänskligt beslut ligger som JSON-filer under `AGENT_DATA_DIR`.

use std::path::PathBuf;

use serde::{Deserialize, Serialize};

/// Ett förslag som väntar på ett beslut.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PendingApproval {
    /// Målets id.
    pub goal_id: i64,
    /// Föreslagen titel.
    pub title: String,
    /// Föreslagen beskrivning.
    pub description: String,
    /// När förslaget skapades.
    pub created_at: String,
    /// Antal försök som gjorts.
    pub attempts: i32,
    /// Återkopplingen på förslaget.
    #[serde(skip_serializing_if = "String::is_empty")]
    pub feedback: String,
    /// Förslagets status.
    #[serde(skip_serializing_if = "String::is_empty")]
    pub status: String,
    /// Samtalet där förslaget diskuterades.
    #[serde(skip_serializing_if = "String::is_empty")]
    pub conversation_id: String,
}

/// Förslaget väntar på ett beslut.
pub const STATUS_PENDING: &str = "pending";
/// Förslaget avvisades - agenten får göra om definitionen.
pub const STATUS_REJECTED: &str = "rejected";

/// Store för väntande förslag.
#[derive(Debug, Clone)]
pub struct Store {
    bas: PathBuf,
}

impl Store {
    /// Skapar store:n och katalogen den skriver i.
    pub fn new() -> Result<Self, std::io::Error> {
        let bas = std::env::var("AGENT_DATA_DIR")
            .ok()
            .filter(|katalog| !katalog.is_empty())
            .unwrap_or_else(|| "agent-data".to_string());

        let bas = PathBuf::from(bas);
        std::fs::create_dir_all(bas.join("pending-approvals"))?;

        Ok(Self { bas })
    }

    /// Sökvägen till ett förslags fil.
    fn sökväg(&self, goal_id: i64) -> PathBuf {
        self.bas
            .join("pending-approvals")
            .join(format!("goal-{goal_id}.json"))
    }

    /// Sparar ett förslag.
    pub fn save(&self, godkännande: &PendingApproval) -> Result<(), std::io::Error> {
        let data = serde_json::to_string_pretty(godkännande)
            .map_err(|err| std::io::Error::new(std::io::ErrorKind::InvalidData, err))?;

        std::fs::write(self.sökväg(godkännande.goal_id), data)
    }

    /// Hämtar ett förslag.
    ///
    /// `None` betyder att det inte finns - inte att något gick fel.
    pub fn get(&self, goal_id: i64) -> Result<Option<PendingApproval>, std::io::Error> {
        let sökväg = self.sökväg(goal_id);

        let Ok(data) = std::fs::read_to_string(&sökväg) else {
            return Ok(None);
        };

        let godkännande = serde_json::from_str(&data)
            .map_err(|err| std::io::Error::new(std::io::ErrorKind::InvalidData, err))?;

        Ok(Some(godkännande))
    }

    /// Tar bort ett förslag.
    pub fn delete(&self, goal_id: i64) -> Result<(), std::io::Error> {
        let sökväg = self.sökväg(goal_id);

        match std::fs::remove_file(&sökväg) {
            Ok(()) => Ok(()),
            Err(err) if err.kind() == std::io::ErrorKind::NotFound => Ok(()),
            Err(err) => Err(err),
        }
    }

    /// Listar alla förslag.
    pub fn list(&self) -> Result<Vec<PendingApproval>, std::io::Error> {
        let katalog = self.bas.join("pending-approvals");

        let Ok(poster) = std::fs::read_dir(&katalog) else {
            return Ok(Vec::new());
        };

        let mut förslag = Vec::new();

        for post in poster {
            let Ok(post) = post else {
                continue;
            };

            let sökväg = post.path();
            if sökväg.extension().and_then(|e| e.to_str()) != Some("json") {
                continue;
            }

            let Ok(data) = std::fs::read_to_string(&sökväg) else {
                // Läsbara filer hoppas över - en trasig fil ska inte
                // stoppa listan.
                continue;
            };

            let Ok(godkännande) = serde_json::from_str::<PendingApproval>(&data) else {
                continue;
            };

            förslag.push(godkännande);
        }

        Ok(förslag)
    }

    /// Uppdaterar återkopplingen på ett förslag.
    pub fn update_feedback(&self, goal_id: i64, feedback: &str) -> Result<(), std::io::Error> {
        let Some(mut godkännande) = self.get(goal_id)? else {
            return Err(std::io::Error::new(
                std::io::ErrorKind::NotFound,
                format!("no pending approval for goal {goal_id}"),
            ));
        };

        godkännande.feedback = feedback.to_string();

        self.save(&godkännande)
    }
}
