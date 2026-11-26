//! Handlers.
//!
//! Varje handler som ännu inte är portad från Go svarar med 501, så att
//! en förbise endpoint är tydlig i loggen och i testerna - inte en tyst
//! 404 som ser ut att fungera.

use axum::Json;

use crate::error::AppError;

/// Hälsokontroll. Samma svar som Go-tjänsten: `{"status": "ok"}`.
#[utoipa::path(
    get,
    path = "/health",
    tag = "health",
    responses(
        (status = 200, description = "API:t är igång", content_type = "application/json",
            body = inline(HealthResponse))
    )
)]
pub async fn health() -> Json<HealthResponse> {
    Json(HealthResponse {
        status: "ok".to_string(),
    })
}

/// Svaret från hälsokontrollen.
#[derive(serde::Serialize, utoipa::ToSchema)]
pub struct HealthResponse {
    pub status: String,
}

/// Placeholder för endpoints som ännu inte är portade.
pub async fn not_implemented() -> Result<Json<serde_json::Value>, AppError> {
    Err(AppError::NotImplemented(
        "Endpointen är inte portad till Rust ännu".to_string(),
    ))
}
