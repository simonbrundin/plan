//! Checkpointplaner - agentens arbetsplan.
//!
//! Go-agentens `internal/core/checkpoint`. En plan beskriver en funktion
//! som ska byggas, uppdelad i checkpoints som var för sig går igenom
//! gates (granskningar) innan de räknas som klara.

use std::collections::BTreeMap;

use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};

/// Planens status.
pub mod status {
    /// Planen väntar på granskning.
    pub const PENDING: &str = "pending";
    /// Planen är godkänd.
    pub const APPROVED: &str = "approved";
    /// Planen är avvisad.
    pub const REJECTED: &str = "rejected";
    /// Planen är klar.
    pub const COMPLETED: &str = "completed";
}

/// Checkpointens status.
pub mod checkpoint_status {
    /// Checkpointen väntar på sin tur.
    pub const PENDING: &str = "pending";
    /// Checkpointen körs just nu.
    pub const IN_PROGRESS: &str = "in_progress";
    /// Checkpointen klarade sina gates.
    pub const PASSED: &str = "passed";
    /// Checkpointen misslyckades.
    pub const FAILED: &str = "failed";
    /// Checkpointen hoppades över.
    pub const SKIPPED: &str = "skipped";
}

/// Gate-status.
pub mod gate_status {
    /// Gaten har inte kört ännu.
    pub const PENDING: &str = "pending";
    /// Gaten kör just nu.
    pub const RUNNING: &str = "running";
    /// Gaten godkände.
    pub const PASSED: &str = "passed";
    /// Gaten underkände.
    pub const FAILED: &str = "failed";
    /// Gaten hoppades över.
    pub const SKIPPED: &str = "skipped";
}

/// Komplexitetsnivå.
pub mod complexity {
    /// Enkel uppgift.
    pub const LOW: &str = "low";
    /// Mellansvår uppgift.
    pub const MEDIUM: &str = "medium";
    /// Svår uppgift.
    pub const HIGH: &str = "high";
}

/// En plan för att implementera en funktion.
///
/// Fältens ordning är Go:s - svaret jämförs rad för rad.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct CheckpointPlan {
    /// Planens id.
    pub id: String,
    /// Funktionen som ska byggas.
    pub feature: String,
    /// Beskrivningen av funktionen.
    pub description: String,
    /// Checkpoints i den ordning de ska göras.
    ///
    /// `None` blir `null` i JSON - Go gör så när planen läses från
    /// databasen och inga checkpoints finns.
    pub checkpoints: Option<Vec<Checkpoint>>,
    /// Planens status.
    pub status: String,
    /// Vem som godkände planen.
    #[serde(skip_serializing_if = "String::is_empty")]
    pub approved_by: String,
    /// När planen godkändes.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub approved_at: Option<DateTime<Utc>>,
    /// När planen avvisades.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub rejected_at: Option<DateTime<Utc>>,
    /// När planen skapades.
    pub created_at: DateTime<Utc>,
    /// När planen ändrades sist.
    pub updated_at: DateTime<Utc>,
}

impl CheckpointPlan {
    /// Skapar en ny plan.
    pub fn new(
        id: impl Into<String>,
        feature: impl Into<String>,
        description: impl Into<String>,
    ) -> Self {
        let nu = Utc::now();

        Self {
            id: id.into(),
            feature: feature.into(),
            description: description.into(),
            checkpoints: Some(Vec::new()),
            status: status::PENDING.to_string(),
            approved_by: String::new(),
            approved_at: None,
            rejected_at: None,
            created_at: nu,
            updated_at: nu,
        }
    }

    /// Lägger till en checkpoint i ordning.
    pub fn add_checkpoint(&mut self, checkpoint: Checkpoint) {
        self.checkpoints
            .get_or_insert_with(Vec::new)
            .push(checkpoint);
        self.updated_at = Utc::now();
    }

    /// Checkpointsen i planen.
    pub fn checkpoints(&self) -> &[Checkpoint] {
        self.checkpoints.as_deref().unwrap_or(&[])
    }

    /// Hämtar en checkpoint på id.
    pub fn get_checkpoint(&self, id: &str) -> Option<&Checkpoint> {
        self.checkpoints().iter().find(|cp| cp.id == id)
    }

    /// Nästa checkpoint som inte är klar.
    pub fn next_pending(&self) -> Option<&Checkpoint> {
        self.checkpoints().iter().find(|cp| {
            cp.status == checkpoint_status::PENDING || cp.status == checkpoint_status::IN_PROGRESS
        })
    }

    /// Checkpointen som körs just nu.
    pub fn current(&self) -> Option<&Checkpoint> {
        self.checkpoints()
            .iter()
            .find(|cp| cp.status == checkpoint_status::IN_PROGRESS)
    }

    /// Om alla checkpoints har klarat sina gates.
    pub fn all_passed(&self) -> bool {
        if self.checkpoints().is_empty() {
            return false;
        }

        self.checkpoints().iter().all(|cp| {
            cp.status == checkpoint_status::PASSED || cp.status == checkpoint_status::SKIPPED
        })
    }

    /// Framsteget - hur många checkpoints som klarat sina gates.
    pub fn progress(&self) -> (usize, usize) {
        let total = self.checkpoints().len();
        let passed = self
            .checkpoints()
            .iter()
            .filter(|cp| {
                cp.status == checkpoint_status::PASSED || cp.status == checkpoint_status::SKIPPED
            })
            .count();

        (passed, total)
    }

    /// Godkänner planen.
    pub fn approve(&mut self, approved_by: impl Into<String>) {
        let nu = Utc::now();

        self.status = status::APPROVED.to_string();
        self.approved_by = approved_by.into();
        self.approved_at = Some(nu);
        self.updated_at = nu;
    }

    /// Avvisar planen.
    pub fn reject(&mut self) {
        let nu = Utc::now();

        self.status = status::REJECTED.to_string();
        self.rejected_at = Some(nu);
        self.updated_at = nu;
    }

    /// Markerar planen som klar.
    pub fn complete(&mut self) {
        self.status = status::COMPLETED.to_string();
        self.updated_at = Utc::now();
    }
}

/// En enhet av arbete i en plan.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Checkpoint {
    /// Checkpointens id.
    pub id: String,
    /// Namnet på checkpointen.
    pub name: String,
    /// Beskrivningen av checkpointen.
    pub description: String,
    /// Hur svår checkpointen är.
    pub complexity: String,
    /// Vilka gates checkpointen ska passera.
    pub gates: Vec<String>,
    /// Uppgifterna i checkpointen.
    pub tasks: Vec<String>,
    /// Checkpointens status.
    pub status: String,
    /// Resultaten av varje gate.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub gate_results: Option<BTreeMap<String, GateResult>>,
    /// Extra metadata.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub metadata: Option<BTreeMap<String, String>>,
    /// När checkpointen skapades.
    pub created_at: DateTime<Utc>,
    /// När checkpointen ändrades sist.
    pub updated_at: DateTime<Utc>,
    /// När checkpointen klarades.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub completed_at: Option<DateTime<Utc>>,
}

impl Checkpoint {
    /// Skapar en ny checkpoint.
    pub fn new(
        id: impl Into<String>,
        name: impl Into<String>,
        description: impl Into<String>,
        complexity: impl Into<String>,
    ) -> Self {
        let nu = Utc::now();

        Self {
            id: id.into(),
            name: name.into(),
            description: description.into(),
            complexity: complexity.into(),
            gates: Vec::new(),
            tasks: Vec::new(),
            status: checkpoint_status::PENDING.to_string(),
            gate_results: None,
            metadata: None,
            created_at: nu,
            updated_at: nu,
            completed_at: None,
        }
    }

    /// Lägger till en uppgift.
    pub fn add_task(&mut self, task: impl Into<String>) {
        self.tasks.push(task.into());
        self.updated_at = Utc::now();
    }

    /// Lägger till en gate.
    pub fn add_gate(&mut self, gate_type: impl Into<String>) {
        self.gates.push(gate_type.into());
        self.updated_at = Utc::now();
    }

    /// Sparar resultatet av en gate.
    pub fn set_gate_result(&mut self, gate_type: impl Into<String>, result: GateResult) {
        self.gate_results
            .get_or_insert_with(BTreeMap::new)
            .insert(gate_type.into(), result);
        self.updated_at = Utc::now();
    }

    /// Om alla gates har godkänt.
    pub fn all_gates_passed(&self) -> bool {
        if self.gates.is_empty() {
            return false;
        }

        let results = self.gate_results.as_ref();

        self.gates.iter().all(|gate| {
            results
                .and_then(|resultat| resultat.get(gate))
                .map(|result| result.status == gate_status::PASSED)
                .unwrap_or(false)
        })
    }

    /// De gates som inte godkänt ännu.
    pub fn pending_gates(&self) -> Vec<String> {
        let results = self.gate_results.as_ref();

        self.gates
            .iter()
            .filter(|gate| {
                !results
                    .and_then(|resultat| resultat.get(*gate))
                    .map(|result| result.status == gate_status::PASSED)
                    .unwrap_or(false)
            })
            .cloned()
            .collect()
    }

    /// Markerar checkpointen som pågående.
    pub fn mark_in_progress(&mut self) {
        self.status = checkpoint_status::IN_PROGRESS.to_string();
        self.updated_at = Utc::now();
    }

    /// Markerar checkpointen som klar.
    pub fn mark_passed(&mut self) {
        let nu = Utc::now();

        self.status = checkpoint_status::PASSED.to_string();
        self.completed_at = Some(nu);
        self.updated_at = nu;
    }

    /// Markerar checkpointen som misslyckad.
    pub fn mark_failed(&mut self) {
        self.status = checkpoint_status::FAILED.to_string();
        self.updated_at = Utc::now();
    }

    /// Läser en metadatat.
    pub fn metadata_value(&self, key: &str) -> String {
        self.metadata
            .as_ref()
            .and_then(|metadata| metadata.get(key))
            .cloned()
            .unwrap_or_default()
    }

    /// Sätter en metadatat.
    pub fn set_metadata(&mut self, key: impl Into<String>, value: impl Into<String>) {
        self.metadata
            .get_or_insert_with(BTreeMap::new)
            .insert(key.into(), value.into());
        self.updated_at = Utc::now();
    }
}

/// Resultatet av en gate.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct GateResult {
    /// Vilken gate det är.
    #[serde(rename = "type")]
    pub gate_type: String,
    /// Gatens status.
    pub status: String,
    /// När gaten startade.
    pub started_at: DateTime<Utc>,
    /// När gaten slutade.
    ///
    /// Go har `omitempty` på tidsstämpeln, men `omitempty` har ingen
    /// effekt på en struct - fältet skrivs därför alltid ut, även när
    /// det är nolltiden.
    pub ended_at: DateTime<Utc>,
    /// Hur länge gaten tog.
    #[serde(skip_serializing_if = "String::is_empty")]
    pub duration: String,
    /// Återkopplingen från gaten.
    #[serde(skip_serializing_if = "String::is_empty")]
    pub feedback: String,
    /// Problemen gaten hittade.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub issues: Option<Vec<Issue>>,
    /// Sammanfattningen av gaten.
    #[serde(skip_serializing_if = "String::is_empty")]
    pub summary: String,
}

impl GateResult {
    /// Skapar ett gate-resultat i status "pending".
    pub fn new(gate_type: impl Into<String>) -> Self {
        Self {
            gate_type: gate_type.into(),
            status: gate_status::PENDING.to_string(),
            started_at: Utc::now(),
            ended_at: nolltid(),
            duration: String::new(),
            feedback: String::new(),
            issues: None,
            summary: String::new(),
        }
    }

    /// Gaten godkände.
    pub fn pass(&mut self, feedback: impl Into<String>) {
        self.avsluta(gate_status::PASSED);
        self.feedback = feedback.into();
    }

    /// Gaten underkände.
    pub fn fail(&mut self, feedback: impl Into<String>, issues: Vec<Issue>) {
        self.avsluta(gate_status::FAILED);
        self.feedback = feedback.into();
        self.issues = Some(issues);
    }

    /// Gaten hoppades över.
    pub fn skip(&mut self, reason: impl Into<String>) {
        self.avsluta(gate_status::SKIPPED);
        self.feedback = reason.into();
    }

    /// Avslutar gaten och räknar ut tiden.
    fn avsluta(&mut self, status: &str) {
        let nu = Utc::now();

        self.status = status.to_string();
        self.ended_at = nu;
        self.duration = (nu - self.started_at).to_string();
    }
}

/// Nolltiden - som Go:s `time.Time{}`.
pub fn nolltid() -> DateTime<Utc> {
    use chrono::TimeZone;

    Utc.with_ymd_and_hms(1, 1, 1, 0, 0, 0)
        .single()
        .expect("år 1 är ett giltigt datum")
}

/// Ett problem som en gate hittade.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Issue {
    /// Hur allvarligt problemet är.
    pub severity: String,
    /// Vilken sorts problem det är.
    pub category: String,
    /// Rubriken.
    pub title: String,
    /// Beskrivningen.
    pub description: String,
    /// Var problemet finns - "fil:rad".
    #[serde(skip_serializing_if = "String::is_empty")]
    pub location: String,
    /// Förslag på fix.
    #[serde(skip_serializing_if = "String::is_empty")]
    pub suggestion: String,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn tom_plan_har_inga_checkpoints() {
        let plan = CheckpointPlan::new("id-1", "funktion", "beskrivning");

        assert!(plan.checkpoints().is_empty());
        assert_eq!(plan.status, status::PENDING);
        assert!(!plan.all_passed());
        assert_eq!(plan.progress(), (0, 0));
    }

    #[test]
    fn en_tom_plan_json_har_inga_tomma_fält() {
        let plan = CheckpointPlan::new("id-1", "funktion", "beskrivning");
        let json = serde_json::to_string(&plan).unwrap();

        // Go hoppar över tomma fält - samma här.
        assert!(!json.contains("approvedBy"));
        assert!(!json.contains("approvedAt"));
        assert!(!json.contains("rejectedAt"));
        assert!(json.contains("\"checkpoints\":[]"));
    }

    #[test]
    fn godkannande_satter_tidpunkter() {
        let mut plan = CheckpointPlan::new("id-1", "funktion", "beskrivning");

        plan.approve("api");

        assert_eq!(plan.status, status::APPROVED);
        assert_eq!(plan.approved_by, "api");
        assert!(plan.approved_at.is_some());
    }

    #[test]
    fn framsteg_raknar_klarade_checkpoints() {
        let mut plan = CheckpointPlan::new("id-1", "funktion", "beskrivning");

        let mut cp = Checkpoint::new("cp-1", "namn", "beskrivning", complexity::MEDIUM);
        cp.mark_passed();
        plan.add_checkpoint(cp);

        let mut cp = Checkpoint::new("cp-2", "namn", "beskrivning", complexity::LOW);
        cp.mark_passed();
        plan.add_checkpoint(cp);

        assert_eq!(plan.progress(), (2, 2));
        assert!(plan.all_passed());
    }
}
