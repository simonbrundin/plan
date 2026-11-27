//! Typer agenten använder mot Plan.
//!
//! Fältena är samma som i Go-klienten (`src/agent/internal/planclient`),
//! så agenten och Plan pratar om samma saker.

use serde::{Deserialize, Serialize};

/// Ett mål från Plan API.
#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct Goal {
    pub id: i64,
    pub title: String,
    #[serde(default)]
    pub description: String,
    #[serde(default)]
    pub icon: String,
    pub status_id: i32,
    #[serde(default)]
    pub created: chrono::DateTime<chrono::Utc>,
    #[serde(default)]
    pub started: Option<chrono::DateTime<chrono::Utc>>,
    #[serde(default)]
    pub finished: Option<chrono::DateTime<chrono::Utc>>,
    #[serde(default)]
    pub inbox: i32,
    /// Sant om målet är markerat för agentarbete.
    #[serde(default)]
    pub agent: bool,
    /// Sant om målet ska göras nu.
    #[serde(default)]
    pub now: bool,
    /// Sant om målet aldrig slutar.
    #[serde(default)]
    pub eternal: bool,
    /// Det datum målet ska vara klart ("YYYY-MM-DD").
    #[serde(default)]
    pub due: Option<String>,
    /// Det datum målet får dyka upp i prioriteringslistan.
    #[serde(default)]
    pub start_date: Option<String>,
    /// Uppskattad tidsåtgång i minuter.
    #[serde(default)]
    pub estimated_minutes: Option<i32>,
    /// Upprepningsregeln (RRULE).
    #[serde(default)]
    pub recurrence_rule: Option<String>,
    /// Schemaläggningsstrategin: "schedule" eller "push".
    #[serde(default)]
    pub recurrence_mode: Option<String>,
    /// Målets taggar.
    #[serde(default)]
    pub tags: Vec<String>,
    /// Platser där målet kan slutföras.
    #[serde(default)]
    pub places: Vec<Place>,
}

/// En plats i Plan - till exempel "Hem" eller "Gymmet".
#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct Place {
    pub id: i64,
    pub name: String,
    #[serde(default)]
    pub icon: String,
}

/// Körningens tillstånd - de ägs av Plan, agenten läser dem.
pub const RUN_STATUS_QUEUED: &str = "queued";
pub const RUN_STATUS_RUNNING: &str = "running";
pub const RUN_STATUS_WAITING_APPROVAL: &str = "waiting_approval";
pub const RUN_STATUS_SUCCEEDED: &str = "succeeded";
pub const RUN_STATUS_FAILED: &str = "failed";
pub const RUN_STATUS_CANCELLED: &str = "cancelled";

/// Tillstånden i ordning - samma ordning som i kontraktet.
pub const RUN_STATUSES: &[&str] = &[
    RUN_STATUS_QUEUED,
    RUN_STATUS_RUNNING,
    RUN_STATUS_WAITING_APPROVAL,
    RUN_STATUS_SUCCEEDED,
    RUN_STATUS_FAILED,
    RUN_STATUS_CANCELLED,
];

/// En agentkörning registrerad i Plan.
#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct AgentRun {
    pub id: i64,
    pub goal_id: i64,
    pub phase: String,
    pub status: String,
    #[serde(default)]
    pub trigger: String,
    #[serde(default)]
    pub initiator: String,
    #[serde(default)]
    pub agent_profile: String,
    #[serde(default)]
    pub summary: String,
    #[serde(default)]
    pub result: String,
    #[serde(default)]
    pub artifact_url: Option<String>,
    #[serde(default)]
    pub cost_usd: Option<f64>,
    #[serde(default)]
    pub conversation_id: Option<String>,
    pub created_at: chrono::DateTime<chrono::Utc>,
    pub updated_at: chrono::DateTime<chrono::Utc>,
}

/// Skapar en agentkörning.
#[derive(Debug, Clone, Default, Serialize)]
pub struct CreateAgentRunRequest {
    pub goal_id: i64,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub phase: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub status: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub trigger: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub initiator: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub agent_profile: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub summary: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub result: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub artifact_url: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub conversation_id: Option<String>,
    /// Identifierar agenten - används för att filtrera kostnader per agent.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub agent_id: Option<String>,
    /// Provider och modell behövs för att Plan ska kunna räkna fram
    /// kostnaden ur prislistan.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub provider: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub model: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub model_version: Option<String>,
}

/// Rapporterar tokenanvändningen för en körning.
///
/// Bara användningen rapporteras - Plan räknar fram kostnaden ur sin
/// prislista, så varje körning använder samma priser.
#[derive(Debug, Clone, Default, Serialize)]
pub struct AgentRunUsageRequest {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub agent_id: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub provider: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub model: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub model_version: Option<String>,
    pub input_tokens: i64,
    pub output_tokens: i64,
    pub cache_read_tokens: i64,
    pub tool_calls: i64,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub tool_names: Option<Vec<String>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub started_at: Option<chrono::DateTime<chrono::Utc>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub finished_at: Option<chrono::DateTime<chrono::Utc>>,
}

/// Registrerar en strukturerad händelse i Plan.
#[derive(Debug, Clone, Default, Serialize)]
pub struct CreateAgentRunEventRequest {
    pub r#type: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub phase: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub payload: Option<serde_json::Value>,
}

/// Uppdaterar en agentkörning.
#[derive(Debug, Clone, Default, Serialize)]
pub struct UpdateAgentRunRequest {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub phase: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub status: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub agent_profile: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub summary: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub result: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub artifact_url: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub cost_usd: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub conversation_id: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub finished_at: Option<chrono::DateTime<chrono::Utc>>,
}

/// En status i Plan.
#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct Status {
    pub id: i64,
    pub name: String,
    pub label: String,
}

/// Uppdaterar ett måls fält.
///
/// Pekare (Option) skiljer "sätt värdet" från "lämna oförändrat": ett
/// fält som inte finns i requesten lämnas som det var.
#[derive(Debug, Clone, Default, Serialize)]
pub struct UpdateGoalRequest {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub title: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub description: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub icon: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub status_id: Option<i32>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub now: Option<bool>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub eternal: Option<bool>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub agent: Option<bool>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub due: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub start_date: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub estimated_minutes: Option<i32>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub recurrence_rule: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub recurrence_mode: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub recurrence_start_date: Option<String>,
}

/// Skapar ett mål i Plan.
#[derive(Debug, Clone, Default, Serialize)]
pub struct CreateGoalRequest {
    pub title: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub icon: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub status_id: Option<i32>,
}

/// Kopplar ett barnmål till ett föräldermål.
#[derive(Debug, Clone, Default, Serialize)]
pub struct CreateRelationRequest {
    #[serde(rename = "childId")]
    pub child_id: i64,
    #[serde(rename = "parentId")]
    pub parent_id: i64,
    pub order: i32,
    pub weight: i32,
}

/// Ett mänskligt beslut i en agentkörning.
#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct AgentApproval {
    pub id: i64,
    pub agent_run_id: i64,
    pub goal_id: i64,
    pub title: String,
    pub description: String,
    pub status: String,
    pub feedback: String,
    pub consequence: String,
    /// Beskriver hur resultatet ska kontrolleras.
    pub verification_plan: String,
    /// Innehåller exakt det som föreslogs.
    #[serde(default)]
    pub payload: Option<serde_json::Value>,
    pub decided_at: Option<chrono::DateTime<chrono::Utc>>,
}

/// Ett förslag på hur ett mål ska se ut.
pub const APPROVAL_KIND_GOAL_DEFINITION: &str = "goal_definition";
/// Ett förslag på en kodändring.
pub const APPROVAL_KIND_CODE_CHANGE: &str = "code_change";

/// Godkännandetyperna, i ordning - samma ordning som i kontraktet.
pub const APPROVAL_KINDS: &[&str] = &[APPROVAL_KIND_GOAL_DEFINITION, APPROVAL_KIND_CODE_CHANGE];

impl AgentApproval {
    /// Godkännandets typ.
    ///
    /// Typen ligger i godkännandets payload ("kind"), så att en människa ser
    /// vad som föreslogs och att agenten inte blandar ihop ett
    /// definitionsförslag med ett kodförslag i samma körning.
    ///
    /// Ett godkännande utan typ är ett äldre förslag och behandlas som ett
    /// kodförslag - så att förslag som skapats tidigare fortfarande fungerar.
    pub fn approval_kind(&self) -> &str {
        self.payload
            .as_ref()
            .and_then(|payload| payload.get("kind"))
            .and_then(|kind| kind.as_str())
            .filter(|kind| !kind.is_empty())
            .unwrap_or(APPROVAL_KIND_CODE_CHANGE)
    }
}

/// Skapar ett godkännande som väntar på beslut.
#[derive(Debug, Clone, Default, Serialize)]
pub struct CreateAgentApprovalRequest {
    pub title: String,
    pub description: String,
    pub consequence: String,
    pub verification_plan: String,
    /// Innehåller exakt det som föreslogs.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub payload: Option<serde_json::Value>,
}

/// Status-id:n från seed-data.
pub const STATUS_DEFINE: i32 = 1;
pub const STATUS_PLAN: i32 = 2;
pub const STATUS_EXECUTE: i32 = 3;
pub const STATUS_VERIFY: i32 = 4;
pub const STATUS_COMPLETED: i32 = 5;

/// Fasnamnen som målen går igenom.
pub const GOAL_PHASES: &[&str] = &["define", "plan", "execute", "verify", "completed"];

/// Händelsetyper som agenten skriver till Plan.
///
/// Go: `internal/planclient/events.go`.
pub const EVENT_RUN_STARTED: &str = "run_started";
/// Agenten har definierat målet.
pub const EVENT_DEFINITION_CREATED: &str = "definition_created";
/// Körningen har bytt fas.
pub const EVENT_PHASE_CHANGED: &str = "phase_changed";
/// Ett verktyg anropades.
pub const EVENT_TOOL_CALL: &str = "tool_call";
/// Ett verktyg gav ett resultat.
pub const EVENT_TOOL_RESULT: &str = "tool_result";
/// Agenten skapade ett undermål.
pub const EVENT_SUBGOAL_CREATED: &str = "subgoal_created";
/// Agenten begär ett mänskligt beslut.
pub const EVENT_APPROVAL_REQUESTED: &str = "approval_requested";
/// En människa fattade beslutet.
pub const EVENT_APPROVAL_DECIDED: &str = "approval_decided";
/// En verifiering utfördes.
pub const EVENT_VERIFICATION: &str = "verification";
/// En artefakt skapades.
pub const EVENT_ARTIFACT: &str = "artifact";
/// Något gick fel under körningen.
pub const EVENT_ERROR: &str = "error";
/// Körningen misslyckades.
pub const EVENT_RUN_FAILED: &str = "run_failed";
/// Körningen var klar.
pub const EVENT_RUN_COMPLETED: &str = "run_completed";
