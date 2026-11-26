//! Agentkörningar, godkännanden, kostnad och budgetar.
//!
//! Port av `src/api/internal/models/agent_run.go`, `agent_run_event.go`,
//! `agent_approval.go` och `agent_usage.go`.

use serde::{Deserialize, Serialize};
use utoipa::ToSchema;

/// Fasen "define" - agenten specificerar målet.
pub const PHASE_DEFINE: &str = "define";
/// Fasen "plan" - agenten planerar arbetet.
pub const PHASE_PLAN: &str = "plan";
/// Fasen "execute" - agenten utför arbetet.
pub const PHASE_EXECUTE: &str = "execute";
/// Fasen "verify" - agenten verifierar resultatet.
pub const PHASE_VERIFY: &str = "verify";
/// Fasen "completed" - målet är klart.
pub const PHASE_COMPLETE: &str = "completed";

/// Tillståndet "queued" - körningen väntar på att starta.
pub const STATUS_QUEUED: &str = "queued";
/// Tillståndet "running" - agenten arbetar.
pub const STATUS_RUNNING: &str = "running";
/// Tillståndet "waiting_approval" - agenten väntar på en människa.
pub const STATUS_WAITING_APPROVAL: &str = "waiting_approval";
/// Tillståndet "succeeded" - körningen lyckades.
pub const STATUS_SUCCEEDED: &str = "succeeded";
/// Tillståndet "failed" - körningen misslyckades.
pub const STATUS_FAILED: &str = "failed";
/// Tillståndet "cancelled" - körningen avbröts.
pub const STATUS_CANCELLED: &str = "cancelled";

/// Agentprofilen "coding" - kodande agent.
pub const PROFILE_CODING: &str = "coding";
/// Agentprofilen "life" - livsagent.
pub const PROFILE_LIFE: &str = "life";

/// Kostnadsradens status "reported" - agenten har rapporterat användningen.
pub const USAGE_STATUS_REPORTED: &str = "reported";
/// Kostnadsradens status "incomplete" - körningen avslutades utan rapport.
pub const USAGE_STATUS_INCOMPLETE: &str = "incomplete";

/// Budgetperioden "day" - ett kalenderdygn.
pub const BUDGET_PERIOD_DAY: &str = "day";
/// Budgetperioden "week" - sju dagar.
pub const BUDGET_PERIOD_WEEK: &str = "week";
/// Budgetperioden "month" - en kalendermånad.
pub const BUDGET_PERIOD_MONTH: &str = "month";

/// Godkännandets status "pending" - beslutet väntar på en människa.
pub const APPROVAL_PENDING: &str = "pending";
/// Godkännandets status "approved" - en människa godkände.
pub const APPROVAL_APPROVED: &str = "approved";
/// Godkännandets status "rejected" - en människa avvisade.
pub const APPROVAL_REJECTED: &str = "rejected";

/// Faserna i den ordning de sker.
pub const PHASES: &[&str] = &[
    PHASE_DEFINE,
    PHASE_PLAN,
    PHASE_EXECUTE,
    PHASE_VERIFY,
    PHASE_COMPLETE,
];

/// De tillstånd en körning kan ha.
pub const STATUSES: &[&str] = &[
    STATUS_QUEUED,
    STATUS_RUNNING,
    STATUS_WAITING_APPROVAL,
    STATUS_SUCCEEDED,
    STATUS_FAILED,
    STATUS_CANCELLED,
];

/// Övergångarna mellan tillstånd.
///
/// Från ett terminalt tillstånd går det inte vidare, och en körning får
/// alltid avbrytas.
pub fn övergångar() -> Vec<(&'static str, Vec<&'static str>)> {
    vec![
        (
            STATUS_QUEUED,
            vec![
                STATUS_RUNNING,
                STATUS_WAITING_APPROVAL,
                STATUS_SUCCEEDED,
                STATUS_FAILED,
                STATUS_CANCELLED,
            ],
        ),
        (
            STATUS_RUNNING,
            vec![
                STATUS_WAITING_APPROVAL,
                STATUS_SUCCEEDED,
                STATUS_FAILED,
                STATUS_CANCELLED,
            ],
        ),
        (
            STATUS_WAITING_APPROVAL,
            vec![
                STATUS_QUEUED,
                STATUS_RUNNING,
                STATUS_SUCCEEDED,
                STATUS_FAILED,
                STATUS_CANCELLED,
            ],
        ),
        (STATUS_SUCCEEDED, vec![]),
        (STATUS_FAILED, vec![]),
        (STATUS_CANCELLED, vec![]),
    ]
}

/// Om ett tillstånd är terminalt - då kan körningen inte återupplivas.
pub fn är_terminal(status: &str) -> bool {
    matches!(status, STATUS_SUCCEEDED | STATUS_FAILED | STATUS_CANCELLED)
}

/// Om ett tillstånd kräver att en människa agerar.
pub fn kräver_människa(status: &str) -> bool {
    status == STATUS_WAITING_APPROVAL
}

/// Fasens ordning i livscykeln.
pub fn fasrang(fas: &str) -> i32 {
    match fas {
        PHASE_DEFINE => 0,
        PHASE_PLAN => 1,
        PHASE_EXECUTE => 2,
        PHASE_VERIFY => 3,
        PHASE_COMPLETE => 4,
        _ => -1,
    }
}

/// Ett godkänt agent-id - mellanslag runt är borttaget.
pub fn normalisera_agent_id(agent_id: &str) -> String {
    agent_id.trim().to_string()
}

/// Kontrollerar att fasen är en känd agentfas.
pub fn giltig_fas(fas: &str) -> bool {
    PHASES.contains(&fas)
}

/// Kontrollerar att tillståndet är ett känt tillstånd.
pub fn giltig_status(status: &str) -> bool {
    STATUSES.contains(&status)
}

/// Kontrollerar att triggern är känd.
pub fn giltig_trigger(trigger: &str) -> bool {
    matches!(trigger, "user" | "schedule" | "webhook" | "agent")
}

/// Kontrollerar att agentprofilen är känd.
pub fn giltig_profil(profil: &str) -> bool {
    matches!(profil, PROFILE_CODING | PROFILE_LIFE)
}

/// Kontrollerar att budgetperioden är känd.
pub fn giltig_budgetperiod(period: &str) -> bool {
    matches!(
        period,
        BUDGET_PERIOD_DAY | BUDGET_PERIOD_WEEK | BUDGET_PERIOD_MONTH
    )
}

/// Kontrollerar att beslutet är känt.
pub fn giltigt_beslut(beslut: &str) -> bool {
    matches!(beslut, "approve" | "reject")
}

/// En agentkörning kopplad till ett mål.
///
/// Plan äger körningarna, så gränssnittet visar agentens arbete även när
/// agent-tjänsten är nere. Flera körningar kan finnas för samma mål - en
/// körning per försök.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, ToSchema)]
pub struct AgentRun {
    /// Körningens id.
    pub id: i64,
    /// Målets id.
    pub goal_id: i64,
    /// Fasen körningen befinner sig i.
    pub phase: String,
    /// Tillståndet.
    pub status: String,
    /// Vad som startade körningen - user, schedule, webhook eller agent.
    pub trigger: String,
    /// Vem som startade körningen.
    #[serde(skip_serializing_if = "String::is_empty")]
    pub initiator: String,
    /// Vilken sorts agent som körde - coding eller life.
    pub agent_profile: String,
    /// Kort sammanfattning.
    pub summary: String,
    /// Vad körningen ledde till.
    #[serde(skip_serializing_if = "String::is_empty")]
    pub result: String,
    /// Länken till resultatet - en PR-länk eller ett bokningsutkast.
    pub artifact_url: Option<String>,
    /// Kostnaden för körningen hittills.
    pub cost_usd: Option<f64>,
    /// Körningens konversation i agent-tjänsten.
    pub conversation_id: Option<String>,
    /// Agenten som gjorde körningen.
    #[serde(skip_serializing_if = "String::is_empty")]
    pub agent_id: String,
    /// Leverantören av modellen.
    #[serde(skip_serializing_if = "String::is_empty")]
    pub provider: String,
    /// Modellen som användes.
    #[serde(skip_serializing_if = "String::is_empty")]
    pub model: String,
    /// Modellens version.
    #[serde(skip_serializing_if = "String::is_empty")]
    pub model_version: String,
    /// Kostnaden som reserverats för körningen så länge den pågår.
    pub reserved_cost_usd: f64,
    /// När körningen började.
    pub started_at: Option<chrono::DateTime<chrono::Utc>>,
    /// När körningen blev klar.
    pub finished_at: Option<chrono::DateTime<chrono::Utc>>,
    /// När körningen skapades.
    pub created_at: chrono::DateTime<chrono::Utc>,
    /// När körningen senast ändrades.
    pub updated_at: chrono::DateTime<chrono::Utc>,
}

/// Request-body för att skapa en agentkörning.
#[derive(Debug, Clone, Deserialize, ToSchema)]
pub struct CreateAgentRunRequest {
    /// Målets id - krävs.
    pub goal_id: Option<i64>,
    /// Fasen.
    #[serde(default)]
    pub phase: String,
    /// Tillståndet.
    #[serde(default)]
    pub status: String,
    /// Triggern.
    #[serde(default)]
    pub trigger: String,
    /// Initiatorn.
    #[serde(default)]
    pub initiator: String,
    /// Agentprofilen.
    #[serde(default)]
    pub agent_profile: String,
    /// Sammanfattningen.
    #[serde(default)]
    pub summary: String,
    /// Resultatet.
    #[serde(default)]
    pub result: String,
    /// Länken till artefakten.
    #[serde(default)]
    pub artifact_url: Option<String>,
    /// Konversationen.
    #[serde(default)]
    pub conversation_id: Option<String>,
    /// Agenten.
    #[serde(default)]
    pub agent_id: String,
    /// Leverantören av modellen.
    #[serde(default)]
    pub provider: String,
    /// Modellen.
    #[serde(default)]
    pub model: String,
    /// Modellens version.
    #[serde(default)]
    pub model_version: String,
    /// När körningen börjar. Inget ger "nu".
    #[serde(default)]
    pub started_at: Option<chrono::DateTime<chrono::Utc>>,
}

/// Request-body för att uppdatera en agentkörning.
///
/// Alla fält är valfria - inget betyder "behåll nuvarande värde".
#[derive(Debug, Clone, Default, Deserialize, ToSchema)]
pub struct UpdateAgentRunRequest {
    /// Ny fas.
    #[serde(default)]
    pub phase: Option<String>,
    /// Nytt tillstånd.
    #[serde(default)]
    pub status: Option<String>,
    /// Ny trigger.
    #[serde(default)]
    pub trigger: Option<String>,
    /// Ny initiator.
    #[serde(default)]
    pub initiator: Option<String>,
    /// Ny agentprofil.
    #[serde(default)]
    pub agent_profile: Option<String>,
    /// Ny sammanfattning.
    #[serde(default)]
    pub summary: Option<String>,
    /// Nytt resultat.
    #[serde(default)]
    pub result: Option<String>,
    /// Ny artefaktlänk.
    #[serde(default)]
    pub artifact_url: Option<String>,
    /// Ny kostnad.
    #[serde(default)]
    pub cost_usd: Option<f64>,
    /// Ny konversation.
    #[serde(default)]
    pub conversation_id: Option<String>,
    /// Ny agent.
    #[serde(default)]
    pub agent_id: Option<String>,
    /// Ny leverantör.
    #[serde(default)]
    pub provider: Option<String>,
    /// Ny modell.
    #[serde(default)]
    pub model: Option<String>,
    /// Ny modellversion.
    #[serde(default)]
    pub model_version: Option<String>,
    /// Ny starttid.
    #[serde(default)]
    pub started_at: Option<chrono::DateTime<chrono::Utc>>,
    /// Ny sluttid.
    #[serde(default)]
    pub finished_at: Option<chrono::DateTime<chrono::Utc>>,
}

/// En spårbar händelse i en agentkörning.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, ToSchema)]
pub struct AgentRunEvent {
    /// Händelsens id.
    pub id: i64,
    /// Körningens id.
    pub agent_run_id: i64,
    /// Händelsetypen.
    pub r#type: String,
    /// Fasen händelsen inträffade i.
    #[serde(skip_serializing_if = "String::is_empty")]
    pub phase: String,
    /// Nyttolasten.
    #[serde(skip_serializing_if = "är_tom_payload")]
    pub payload: Option<serde_json::Value>,
    /// När händelsen skapades.
    pub created_at: chrono::DateTime<chrono::Utc>,
}

/// Request-body för att skapa en händelse.
#[derive(Debug, Clone, Default, Deserialize, ToSchema)]
pub struct CreateAgentRunEventRequest {
    /// Händelsetypen - krävs.
    pub r#type: String,
    /// Fasen.
    #[serde(default)]
    pub phase: String,
    /// Nyttolasten.
    #[serde(default)]
    pub payload: Option<serde_json::Value>,
}

/// Händelsetyperna som är giltiga.
///
/// Typerna är ett slutet set - det gör att historiken blir läsbar och att
/// en trasig agent inte kan skriva godtycklig text i historiken.
pub const EVENT_TYPES: &[&str] = &[
    "run_started",
    "definition_created",
    "phase_changed",
    "tool_call",
    "tool_result",
    "subgoal_created",
    "approval_requested",
    "approval_decided",
    "verification",
    "artifact",
    "error",
    "run_failed",
    "run_completed",
    "run_cancelled",
    "verification_checkpoint",
];

/// Om en lista ska utelämnas i svaret.
///
/// Go använder `omitempty` på slicefält - en tom lista skickas inte med.
pub fn är_tom_lista<T>(lista: &Option<Vec<T>>) -> bool {
    lista.as_ref().is_none_or(|lista| lista.is_empty())
}

/// Om ett JSON-objekt ska utelämnas i svaret.
///
/// Go använder `omitempty` på mapfält - en tom map skickas inte med.
pub fn är_tom_payload(payload: &Option<serde_json::Value>) -> bool {
    match payload {
        None => true,
        Some(serde_json::Value::Object(fält)) => fält.is_empty(),
        Some(_) => false,
    }
}

/// Kontrollerar att händelsetypen är en känd typ.
pub fn giltig_händelsetyp(händelsetyp: &str) -> bool {
    EVENT_TYPES.contains(&händelsetyp)
}

/// Ett mänskligt beslut i en agentkörning.
///
/// Plan äger godkännandena. Agenten får föreslå, men bara en människa kan
/// godkänna - därför är beslutet en egen entitet med sin aktör.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, ToSchema)]
pub struct AgentApproval {
    /// Godkännandets id.
    pub id: i64,
    /// Körningens id.
    pub agent_run_id: i64,
    /// Målets id.
    pub goal_id: i64,
    /// Titeln på förslaget.
    pub title: String,
    /// Varför förslaget finns.
    pub description: String,
    /// Status - pending, approved eller rejected.
    pub status: String,
    /// Användaren som fattade beslutet.
    pub decided_by: Option<i64>,
    /// Motiveringen.
    #[serde(skip_serializing_if = "String::is_empty")]
    pub feedback: String,
    /// Vad som händer om beslutet tas.
    #[serde(skip_serializing_if = "String::is_empty")]
    pub consequence: String,
    /// Hur resultatet ska kontrolleras.
    #[serde(skip_serializing_if = "String::is_empty")]
    pub verification_plan: String,
    /// Exakt det som föreslogs.
    #[serde(skip_serializing_if = "är_tom_payload")]
    pub payload: Option<serde_json::Value>,
    /// När godkännandet skapades.
    pub created_at: chrono::DateTime<chrono::Utc>,
    /// När beslutet fattades.
    pub decided_at: Option<chrono::DateTime<chrono::Utc>>,
}

/// Request-body för att skapa ett godkännande.
#[derive(Debug, Clone, Default, Deserialize, ToSchema)]
pub struct CreateAgentApprovalRequest {
    /// Titeln - krävs.
    pub title: Option<String>,
    /// Varför förslaget finns.
    #[serde(default)]
    pub description: String,
    /// Vad som händer om beslutet tas.
    #[serde(default)]
    pub consequence: String,
    /// Hur resultatet ska kontrolleras.
    #[serde(default)]
    pub verification_plan: String,
    /// Exakt det som föreslogs.
    #[serde(default)]
    pub payload: Option<serde_json::Value>,
}

/// Request-body för att fatta ett beslut.
#[derive(Debug, Clone, Default, Deserialize, ToSchema)]
pub struct DecideAgentApprovalRequest {
    /// Beslutet - "approve" eller "reject".
    pub decision: String,
    /// Motiveringen.
    #[serde(default)]
    pub feedback: String,
}

/// Kostnaden för en agentkörning.
///
/// Raden sparas när körningen avslutas. Den innehåller allt som behövs
/// för att reproducera kostnaden: användningen, modellen, och vilken
/// prislista som räknades fram den.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, ToSchema)]
pub struct AgentRunUsage {
    /// Radens id.
    pub id: i64,
    /// Körningens id.
    pub agent_run_id: i64,
    /// Målets id.
    pub goal_id: i64,
    /// Agenten som gjorde körningen.
    pub agent_id: String,
    /// Leverantören av modellen.
    pub provider: String,
    /// Modellen som användes.
    pub model: String,
    /// Modellens version.
    #[serde(skip_serializing_if = "String::is_empty")]
    pub model_version: String,
    /// Antalet indata-tokens.
    pub input_tokens: i64,
    /// Antalet utdata-tokens.
    pub output_tokens: i64,
    /// Antalet tokens som lästs från cachen.
    pub cache_read_tokens: i64,
    /// Antalet verktygsanrop.
    pub tool_calls: i64,
    /// De anropade verktygens namn.
    #[serde(skip_serializing_if = "är_tom_lista")]
    pub tool_names: Option<Vec<String>>,
    /// Den beräknade kostnaden.
    pub cost_usd: f64,
    /// Valutan kostnaden anges i.
    pub currency: String,
    /// Versionen av prislistan som användes.
    pub price_list_version: String,
    /// Om priset kom från prislistan eller från en schablon.
    pub price_source: String,
    /// Status - "reported" eller "incomplete".
    pub status: String,
    /// När körningen startade.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub started_at: Option<chrono::DateTime<chrono::Utc>>,
    /// När körningen blev klar.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub finished_at: Option<chrono::DateTime<chrono::Utc>>,
    /// När raden skapades.
    pub created_at: chrono::DateTime<chrono::Utc>,
}

/// Request-body för att rapportera användning.
#[derive(Debug, Clone, Default, Deserialize, ToSchema)]
pub struct CreateAgentRunUsageRequest {
    /// Agenten.
    #[serde(default)]
    pub agent_id: String,
    /// Leverantören av modellen.
    #[serde(default)]
    pub provider: String,
    /// Modellen.
    #[serde(default)]
    pub model: String,
    /// Modellens version.
    #[serde(default)]
    pub model_version: String,
    /// Antalet indata-tokens.
    #[serde(default)]
    pub input_tokens: i64,
    /// Antalet utdata-tokens.
    #[serde(default)]
    pub output_tokens: i64,
    /// Antalet tokens som lästs från cachen.
    #[serde(default)]
    pub cache_read_tokens: i64,
    /// Antalet verktygsanrop.
    #[serde(default)]
    pub tool_calls: i64,
    /// De anropade verktygens namn.
    #[serde(default)]
    pub tool_names: Option<Vec<String>>,
    /// När körningen startade.
    #[serde(default)]
    pub started_at: Option<chrono::DateTime<chrono::Utc>>,
    /// När körningen blev klar.
    #[serde(default)]
    pub finished_at: Option<chrono::DateTime<chrono::Utc>>,
}

/// En gräns för agentkostnad under en period.
///
/// En budget gäller antingen en viss agent eller alla agenter. `agent_id`
/// är tom när budgeten gäller alla.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, ToSchema)]
pub struct AgentBudget {
    /// Budgetens id.
    pub id: i64,
    /// Namnet.
    pub name: String,
    /// Perioden - day, week eller month.
    pub period: String,
    /// Gränsen som stoppar nya körningar.
    pub hard_limit_usd: f64,
    /// Gränsen som bara varnar.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub warning_limit_usd: Option<f64>,
    /// Hur mycket en ny körning reserverar direkt.
    pub reservation_per_run_usd: f64,
    /// Agenten budgeten gäller. Tom gäller alla.
    #[serde(skip_serializing_if = "String::is_empty")]
    pub agent_id: String,
    /// Om budgeten gäller.
    pub active: bool,
    /// När budgeten skapades.
    pub created_at: chrono::DateTime<chrono::Utc>,
}

/// Request-body för att skapa en budget.
#[derive(Debug, Clone, Default, Deserialize, ToSchema)]
pub struct CreateAgentBudgetRequest {
    /// Namnet.
    #[serde(default)]
    pub name: String,
    /// Perioden.
    #[serde(default)]
    pub period: String,
    /// Den hårda gränsen.
    #[serde(default)]
    pub hard_limit_usd: f64,
    /// Varningsgränsen.
    #[serde(default)]
    pub warning_limit_usd: Option<f64>,
    /// Reservationen per körning.
    #[serde(default)]
    pub reservation_per_run_usd: f64,
    /// Agenten budgeten gäller.
    #[serde(default)]
    pub agent_id: String,
}

/// Request-body för att ändra en budget.
///
/// Varje fält är valfritt - det som inte skickas lämnas som det är.
#[derive(Debug, Clone, Default, Deserialize, ToSchema)]
pub struct UpdateAgentBudgetRequest {
    /// Nytt namn.
    #[serde(default)]
    pub name: Option<String>,
    /// Ny period.
    #[serde(default)]
    pub period: Option<String>,
    /// Ny hård gräns.
    #[serde(default)]
    pub hard_limit_usd: Option<f64>,
    /// Ny varningsgräns.
    #[serde(default)]
    pub warning_limit_usd: Option<f64>,
    /// Ny reservation.
    #[serde(default)]
    pub reservation_per_run_usd: Option<f64>,
    /// Om budgeten är aktiv.
    #[serde(default)]
    pub active: Option<bool>,
}

/// En budgets läge just nu.
///
/// Det är detta som visas när en körning stoppas, så att en människa kan
/// se varför den inte startade.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, ToSchema)]
pub struct BudgetStatus {
    /// Budgetens id.
    pub budget_id: i64,
    /// Namnet.
    pub name: String,
    /// Perioden.
    pub period: String,
    /// Den hårda gränsen.
    pub hard_limit_usd: f64,
    /// Vad som redan är använd.
    pub spent_usd: f64,
    /// Vad som är reserverat.
    pub reserved_usd: f64,
    /// Sant om gränsen är nådd.
    pub exceeded: bool,
    /// Sant om varningsgränsen är nådd.
    pub warning: bool,
    /// Varför körningen stoppades - "limit_reached" eller "unknown_price".
    #[serde(skip_serializing_if = "String::is_empty")]
    pub reason: String,
}

/// Ett fel som säger att en budget stoppade en körning.
#[derive(Debug, Clone, PartialEq)]
pub struct BudgetExceeded {
    /// Budgetens läge.
    pub status: BudgetStatus,
}

impl std::fmt::Display for BudgetExceeded {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        // Skälet styr budskapet - en okänd prisbild är inte samma sak som
        // en budget som är slut.
        if self.status.reason == "unknown_price" {
            return write!(
                f,
                "budget {:?} kräver ett känt modellpris, men priset för modellen är okänt",
                self.status.name
            );
        }

        write!(
            f,
            "budget {:?} är slut: {:.4} {} använd eller reserverad, gräns {:.4} {}",
            self.status.name,
            self.status.spent_usd + self.status.reserved_usd,
            crate::pricing::CURRENCY,
            self.status.hard_limit_usd,
            crate::pricing::CURRENCY
        )
    }
}

/// En aggregerad kostnad.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, ToSchema)]
pub struct AgentUsageSummary {
    /// Antalet körningar i urvalet.
    pub runs: i64,
    /// Summan av körningarnas kostnad.
    pub total_cost_usd: f64,
    /// Summan av indata-tokens.
    pub input_tokens: i64,
    /// Summan av utdata-tokens.
    pub output_tokens: i64,
    /// Summan av tokens lästa från cachen.
    pub cache_read_tokens: i64,
    /// Summan av verktygsanrop.
    pub tool_calls: i64,
    /// Valutan kostnaderna anges i.
    pub currency: String,
    /// Kostnaden per modell.
    #[serde(skip_serializing_if = "är_tom_lista")]
    pub by_model: Option<Vec<AgentUsageByModel>>,
}

/// Kostnaden för en modell.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, ToSchema)]
pub struct AgentUsageByModel {
    /// Leverantören.
    pub provider: String,
    /// Modellen.
    pub model: String,
    /// Antalet körningar.
    pub runs: i64,
    /// Kostnaden.
    pub cost_usd: f64,
}

/// Ett svar som innehåller en körning och eventuell budgetvarning.
///
/// Varningen finns med när varningsgränsen är nådd - körningen får starta,
/// men människan behöver veta att budgeten är nära slut.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, ToSchema)]
pub struct AgentRunResponse {
    /// Körningen.
    #[serde(flatten)]
    pub run: AgentRun,
    /// Budgetvarningen, om det finns en.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub budget_warning: Option<BudgetStatus>,
}

/// Livscykeln för agentkörningar.
///
/// Klienter ska inte hålla egna listor över faser och tillstånd - de läser
/// dem härifrån.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, ToSchema)]
pub struct AgentRunLifecycle {
    /// Faserna i den ordning de sker.
    pub phases: Vec<String>,
    /// Tillstånden en körning kan ha.
    pub status: Vec<String>,
    /// Tillstånden en körning inte kan lämna.
    pub terminal: Vec<String>,
    /// Tillstånden där en människa måste agera.
    pub needs_human: Vec<String>,
    /// Vilka tillstånd ett tillstånd får övergå till.
    ///
    /// Nycklarna sorteras, precis som Gons `encoding/json` gör - så svarar
    /// de två tjänsterna identiskt. Ett terminalt tillstånd har `null` -
    /// det går ingenstans vidare.
    pub transitions: std::collections::BTreeMap<String, Option<Vec<String>>>,
}
