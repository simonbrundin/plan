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
        // Vilken build som körs - sätts av Dockerfilen (ARG BUILD_SHA).
        //
        // Smoke-testen efter en rollout jämför detta värde mot den sha
        // som image-tagen pekar på. Utan det går det inte att veta om
        // det är den nya eller den gamla versionen som svarar.
        build_sha: std::env::var("PLAN_BUILD_SHA").unwrap_or_default(),
    })
}

/// Svaret från hälsokontrollen.
#[derive(serde::Serialize, utoipa::ToSchema)]
pub struct HealthResponse {
    pub status: String,
    /// Tom sträng när bilden är byggd utan sha (t.ex. lokalt).
    ///
    /// Fältet hoppas över när det är tomt - då är svaret exakt
    /// `{"status":"ok"}`, som Go-agenten svarar. Paritetstestet
    /// jämför svaret, så fältet får inte dyka upp i onödan.
    #[serde(skip_serializing_if = "String::is_empty")]
    pub build_sha: String,
}

/// Placeholder för endpoints som ännu inte är portade.
pub async fn not_implemented() -> Result<Json<serde_json::Value>, AppError> {
    Err(AppError::NotImplemented(
        "Endpointen är inte portad till Rust ännu".to_string(),
    ))
}
