//! Handlers för agent-API:et.

use axum::Json;
use serde_json::json;

/// Svar på ett anrop mot en endpoint som ännu inte är portad.
///
/// 501 i stället för 404 - en förbise endpoint ska vara tydlig i loggen och
/// i testerna, inte se ut att fungera.
pub async fn not_implemented() -> Result<Json<serde_json::Value>, crate::error::AgentError> {
    Err(crate::error::AgentError::NotImplemented(
        "Endpointen är inte portad till Rust ännu".to_string(),
    ))
}

/// Hälsokontroll - samma svar som Go-agenten.
pub async fn health() -> Json<serde_json::Value> {
    Json(json!({
        "status": "ok",
        "service": "agent-api",
        // Build-sha gör att en rollout kan verifieras utifrån: en
        // smoke-test kan kolla att det är just den nya versionen
        // som svarar. APP_VERSION har högre prioritet.
        "version": std::env::var("APP_VERSION")
            .ok()
            .filter(|v| !v.is_empty())
            .or_else(|| std::env::var("PLAN_BUILD_SHA").ok())
            .unwrap_or_default(),
        // Go formaterar tiden med RFC3339 i sekundprecision - Rusts
        // `to_rfc3339` ger nanosekunder och "+00:00". Klienten jämför
        // texterna, så formaten ska vara densamma.
        "timestamp": chrono::Utc::now()
            .format("%Y-%m-%dT%H:%M:%SZ")
            .to_string(),
    }))
}
