//! Sökvägarna agenten använder mot Plan.
//!
//! De finns i kontraktet (`src/contracts/plan-agent.yaml`) - ett test
//! kontrollerar att de stämmer. Här skrivs parametrar som `{id}`, i
//! kontraktet på samma sätt.

/// Alla mål.
pub const PATH_GOALS: &str = "/api/v1/goals";
/// Ett mål.
pub const PATH_GOAL: &str = "/api/v1/goals/{id}";
/// Ett måls status.
pub const PATH_GOAL_STATUS: &str = "/api/v1/goals/{id}/status";
/// Ett måls taggar.
pub const PATH_GOAL_TAGS: &str = "/api/v1/goals/{id}/tags";
/// Ett måls platser.
pub const PATH_GOAL_PLACES: &str = "/api/v1/goals/{id}/places";
/// Relationer mellan mål.
pub const PATH_GOAL_RELATIONS: &str = "/api/v1/goals/relations";
/// Statusar.
pub const PATH_STATUSES: &str = "/api/v1/statuses";
/// Platser.
pub const PATH_PLACES: &str = "/api/v1/places";
/// Agentkörningar.
pub const PATH_AGENT_RUNS: &str = "/api/v1/agent-runs";
/// En agentkörning.
pub const PATH_AGENT_RUN: &str = "/api/v1/agent-runs/{id}";
/// Användning för en agentkörning.
pub const PATH_AGENT_RUN_USAGE: &str = "/api/v1/agent-runs/{id}/usage";
/// Händelser för en agentkörning.
pub const PATH_AGENT_RUN_EVENTS: &str = "/api/v1/agent-runs/{id}/events";
/// Godkännanden för en agentkörning.
pub const PATH_AGENT_RUN_APPROVALS: &str = "/api/v1/agent-runs/{id}/approvals";

/// Alla sökvägar agenten använder. Testet mot kontraktet går igenom
/// listan.
pub const ALLA_SOKVAGAR: &[&str] = &[
    PATH_GOALS,
    PATH_GOAL,
    PATH_GOAL_STATUS,
    PATH_GOAL_TAGS,
    PATH_GOAL_PLACES,
    PATH_GOAL_RELATIONS,
    PATH_STATUSES,
    PATH_PLACES,
    PATH_AGENT_RUNS,
    PATH_AGENT_RUN,
    PATH_AGENT_RUN_USAGE,
    PATH_AGENT_RUN_EVENTS,
    PATH_AGENT_RUN_APPROVALS,
];

/// Bygger en sökväg med substituerad id-parameter.
///
/// Parametern skrivs `{id}` i koden och i kontraktet - det är Plan som
/// binder värdet.
pub fn med_id(sokvag: &str, id: i64) -> String {
    sokvag.replace("{id}", &id.to_string())
}
