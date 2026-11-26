//! API-nycklar för leverantörerna.
//!
//! Go-agentens `internal/api/handlers/apikeys.go`.

use axum::extract::{Path, State};
use axum::Json;
use serde::{Deserialize, Serialize};

use crate::error::AgentError;
use crate::state::AppState;

/// Svar med bara en status.
#[derive(Debug, Serialize)]
pub struct StatusSvar {
    /// Vad som hände.
    status: &'static str,
}

/// Request för att sätta en API-nyckel.
#[derive(Debug, Default, Deserialize)]
pub struct SetApiKeyRequest {
    /// Leverantören nyckeln gäller.
    pub provider: String,
    /// Själva nyckeln.
    pub api_key: String,
    /// Modellen nyckeln gäller.
    pub model: String,
}

/// Agentens egen databas - annars kan inga nycklar läsas.
fn databas(state: &AppState) -> Result<&sqlx::PgPool, AgentError> {
    state
        .db
        .as_ref()
        .map(|db| db.pool())
        .ok_or_else(|| AgentError::Internal("Agentens databas är inte ansluten".to_string()))
}

/// Svar för listan av API-nycklar.
///
/// Wrappers används i stället för `json!` - `json!` bygger en
/// `serde_json::Map`, och Maps fält sorteras. Go behåller structens
/// fältordning, så svaret måste göras detsamma.
#[derive(Debug, Serialize)]
pub struct ApiNycklarSvar {
    /// Nycklarna, utan själva nyckelvärdena.
    api_keys: Vec<crate::db::models::MaskedApiKey>,
}

/// Listar API-nycklarna - utan att visa själva nycklarna.
pub async fn list(State(state): State<AppState>) -> Result<Json<ApiNycklarSvar>, AgentError> {
    let pool = databas(&state)?;

    let nycklar = crate::db::repository::list_masked_api_keys(pool).await?;

    // Go gör en lista med `make` - en tom lista är därför `[]`, inte null.
    Ok(Json(ApiNycklarSvar { api_keys: nycklar }))
}

/// Sätter en API-nyckel.
pub async fn set(
    State(state): State<AppState>,
    body: axum::body::Bytes,
) -> Result<Json<StatusSvar>, AgentError> {
    let pool = databas(&state)?;

    let req: SetApiKeyRequest =
        serde_json::from_slice(&body).map_err(|err| AgentError::BadRequest(err.to_string()))?;

    crate::db::repository::set_api_key(pool, &req.provider, &req.api_key, &req.model).await?;

    Ok(Json(StatusSvar { status: "ok" }))
}

/// Tar bort en API-nyckel.
pub async fn delete(
    State(state): State<AppState>,
    Path(provider): Path<String>,
) -> Result<Json<StatusSvar>, AgentError> {
    if provider.is_empty() {
        return Err(AgentError::BadRequest("provider is required".to_string()));
    }

    let pool = databas(&state)?;

    crate::db::repository::delete_api_key(pool, &provider).await?;

    Ok(Json(StatusSvar { status: "deleted" }))
}
