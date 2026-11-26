//! Samtal och meddelanden.
//!
//! Go-agentens `internal/api/handlers/conversations.go` - samma vägar,
//! samma statuskoder och samma JSON-format.

use axum::extract::{Path, Query, State};
use axum::Json;
use serde::{Deserialize, Serialize};
use uuid::Uuid;

use crate::db::models::{Conversation, Message};
use crate::error::AgentError;
use crate::state::AppState;

/// Listparametrar för samtal.
#[derive(Debug, Default, Deserialize)]
pub struct ListQuery {
    /// Hur många samtal som ska returneras.
    pub limit: Option<i64>,
    /// Hur många samtal som ska hoppas över.
    pub offset: Option<i64>,
}

/// Request för att skapa ett samtal.
#[derive(Debug, Default, Deserialize)]
pub struct CreateConversationRequest {
    /// Modellen som ska användas.
    pub model: String,
    /// Leverantören som ska användas.
    pub provider: String,
    /// Samtalets titel.
    pub title: String,
}

/// Request för att ändra ett samtal.
#[derive(Debug, Default, Deserialize)]
pub struct UpdateConversationRequest {
    /// Den nya titeln.
    pub title: String,
}

/// Agentens egen databas - annars kan inga samtal läsas.
fn databas(state: &AppState) -> Result<&sqlx::PgPool, AgentError> {
    state
        .db
        .as_ref()
        .map(|db| db.pool())
        .ok_or_else(|| AgentError::Internal("Agentens databas är inte ansluten".to_string()))
}

/// Svar för listan av samtal.
///
/// Wrappers används i stället för `json!` - `json!` bygger en
/// `serde_json::Map`, ochMaps fält sorteras. Go behåller structens
/// fältordning, så svaret måste göras detsamma.
#[derive(Debug, Serialize)]
pub struct SamtalSvar {
    /// Samtalen, nyast först.
    conversations: Option<Vec<Conversation>>,
}

/// Svar för listan av meddelanden.
#[derive(Debug, Serialize)]
pub struct MeddelandenSvar {
    /// Meddelandena i samtalet.
    messages: Option<Vec<Message>>,
}

/// Listar alla samtal.
pub async fn list(
    State(state): State<AppState>,
    Query(query): Query<ListQuery>,
) -> Result<Json<SamtalSvar>, AgentError> {
    let pool = databas(&state)?;

    let samtal = crate::db::repository::list_conversations(
        pool,
        query.limit.unwrap_or(50),
        query.offset.unwrap_or(0),
    )
    .await?;

    // Go lämnar en tom lista som null - inte som en tom array.
    let samtal = if samtal.is_empty() {
        None
    } else {
        Some(samtal)
    };

    Ok(Json(SamtalSvar {
        conversations: samtal,
    }))
}

/// Skapar ett nytt samtal.
pub async fn create(
    State(state): State<AppState>,
    body: axum::body::Bytes,
) -> Result<(axum::http::StatusCode, Json<Conversation>), AgentError> {
    let pool = databas(&state)?;

    let req: CreateConversationRequest =
        serde_json::from_slice(&body).map_err(|err| AgentError::BadRequest(err.to_string()))?;

    // Standardvärdena är desamma som i Go.
    let modell = if req.model.is_empty() {
        "MiniMax-M2.7".to_string()
    } else {
        req.model.clone()
    };
    let leverantör = if req.provider.is_empty() {
        "minimax".to_string()
    } else {
        req.provider.clone()
    };
    let titel = if req.title.is_empty() {
        "New conversation".to_string()
    } else {
        req.title.clone()
    };

    let samtal = crate::db::repository::create_conversation(pool, &modell, &leverantör).await?;

    // Titeln sparas efteråt. Går det fel svarar agenten ändå ok - samtalet
    // finns, bara titeln saknas.
    if let Err(err) =
        crate::db::repository::update_conversation_title(pool, samtal.id, &titel).await
    {
        tracing::warn!("Kunde inte spara samtalets titel: {err}");
    }

    Ok((
        axum::http::StatusCode::CREATED,
        Json(
            crate::db::repository::get_conversation(pool, samtal.id)
                .await?
                .unwrap_or(samtal),
        ),
    ))
}

/// Hämtar ett samtal.
pub async fn get(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> Result<Json<Conversation>, AgentError> {
    let pool = databas(&state)?;
    let id = tolka_id(&id)?;

    let samtal = crate::db::repository::get_conversation(pool, id)
        .await?
        .ok_or_else(|| AgentError::NotFound("conversation not found".to_string()))?;

    Ok(Json(samtal))
}

/// Ändrar ett samtal.
pub async fn update(
    State(state): State<AppState>,
    Path(id): Path<String>,
    body: axum::body::Bytes,
) -> Result<Json<Conversation>, AgentError> {
    let pool = databas(&state)?;
    let id = tolka_id(&id)?;

    let req: UpdateConversationRequest =
        serde_json::from_slice(&body).map_err(|err| AgentError::BadRequest(err.to_string()))?;

    if !req.title.is_empty() {
        crate::db::repository::update_conversation_title(pool, id, &req.title).await?;
    }

    let samtal = crate::db::repository::get_conversation(pool, id)
        .await?
        .ok_or_else(|| AgentError::NotFound("conversation not found".to_string()))?;

    Ok(Json(samtal))
}

/// Hämtar meddelandena i ett samtal.
pub async fn messages(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> Result<Json<MeddelandenSvar>, AgentError> {
    let pool = databas(&state)?;
    let id = tolka_id(&id)?;

    let meddelanden = crate::db::repository::get_messages(pool, id).await?;

    // Go lämnar en tom lista som null - inte som en tom array.
    let meddelanden = if meddelanden.is_empty() {
        None
    } else {
        Some(meddelanden)
    };

    Ok(Json(MeddelandenSvar {
        messages: meddelanden,
    }))
}

/// Tolkar ett conversations-id.
///
/// Go svarar 400 med "invalid conversation ID" när id:t inte är ett UUID.
fn tolka_id(id: &str) -> Result<Uuid, AgentError> {
    Uuid::parse_str(id).map_err(|_| AgentError::BadRequest("invalid conversation ID".to_string()))
}
