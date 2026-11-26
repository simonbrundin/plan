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
        "version": std::env::var("APP_VERSION").unwrap_or_default(),
        "timestamp": chrono::Utc::now().to_rfc3339(),
    }))
}
