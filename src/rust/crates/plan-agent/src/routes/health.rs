//! Agentens hälsokontroll och status.
//!
//! Go-agentens `healthHandler` och `statusHandler` i
//! `internal/api/server.go`. Orchestratorns tillstånd syns här - ett
//! initieringsfel ska vara synligt utan att tjänsten dör.

use axum::Json;
use serde_json::json;

/// Agent-API:ets hälsokontroll.
///
/// Health svarar 200 även när orchestratorn inte startade - tjänsten
/// lever, men en del av den saknas. Det visas i `/status`.
pub async fn health() -> Json<serde_json::Value> {
    Json(json!({
        "status": "ok",
        "service": "agent-api",
        "orchestrator": orchestrator_state(),
    }))
}

/// Agent-API:ets status.
pub async fn status() -> Json<serde_json::Value> {
    Json(json!({
        "status": "healthy",
        "version": "1.0.0",
        "uptime": "running",
        "orchestrator": orchestrator_state(),
    }))
}

/// Orchestratorns tillstånd.
///
/// Felet sparas här så att det syns i både health och status. Ett fel som
/// bara finns i loggen är lätt att missa - och det är just det fel som gör
/// att endpointarna saknas.
pub fn orchestrator_state() -> serde_json::Value {
    match crate::orchestrator::init_error() {
        Some(err) => json!({
            "initialized": false,
            "error": err,
        }),
        None => json!({
            "initialized": crate::orchestrator::is_ready(),
        }),
    }
}
