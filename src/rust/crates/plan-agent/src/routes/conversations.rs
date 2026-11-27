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
    /// Bara samtal för detta Plan-mål.
    pub goal_id: Option<i64>,
}

/// Request för att skapa ett samtal.
///
/// Fälten är valfria - Go-agenten accepterar ett samtal utan dem,
/// så det gör Rust-agenten också.
#[derive(Debug, Default, Deserialize)]
#[serde(default)]
pub struct CreateConversationRequest {
    /// Modellen som ska användas.
    pub model: String,
    /// Leverantören som ska användas.
    pub provider: String,
    /// Samtalets titel.
    pub title: String,
    /// Målet i Plan som samtalet gäller.
    ///
    /// Utan det hittas samtalet inte tillbaka från målsidan.
    pub goal_id: Option<i64>,
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
        query.goal_id,
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

    let samtal =
        crate::db::repository::create_conversation(pool, &modell, &leverantör, req.goal_id).await?;

    // Titeln sparas efteråt. Går det fel svarar agenten ändå ok - samtalet
    // finns, bara titeln saknas.
    if let Err(err) =
        crate::db::repository::update_conversation_title(pool, samtal.id, &titel).await
    {
        tracing::warn!("Kunde inte spara samtalets titel: {err}");
    }

    // Go svarar med samtalet som det såg ut vid skapandet - titeln
    // uppdateras i databasen men syns inte i svaret.
    Ok((axum::http::StatusCode::CREATED, Json(samtal)))
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

/// Skickar ett meddelande till agenten och strömmar svaret som NDJSON.
///
/// Go: `internal/api/handlers/conversations.go` - `CreateMessage`.
pub async fn create_message(
    State(state): State<AppState>,
    Path(id): Path<String>,
    body: axum::body::Bytes,
) -> Result<axum::response::Response, AgentError> {
    let pool = databas(&state)?.clone();
    let id = tolka_id(&id)?;

    let req: crate::routes::chat::CreateMessageRequest =
        crate::routes::json_fel::tolka(&body, "handlers.CreateMessageRequest")
            .map_err(AgentError::BadRequest)?;

    let samtal = crate::db::repository::get_conversation(&pool, id)
        .await?
        .ok_or_else(|| AgentError::NotFound("conversation not found".to_string()))?;

    // Användarmeddelandet sparas först - som i Go.
    let användarmeddelande =
        crate::db::repository::create_message(&pool, id, "user", &req.content, 0, 0).await?;

    // Med målkontext kopplas chatten till Plan - annars saknas historik.
    let (plan_client, recorder) = match req.goal_context.as_ref() {
        Some(mål) => {
            let client = crate::planclient::PlanClient::new()
                .map_err(|_| AgentError::Internal("Plan client is not configured".to_string()))?;

            let recorder = crate::routes::chat::RunRecorder::hitta(&client, mål.id).await;

            (Some(client), recorder)
        }
        None => (None, None),
    };

    let agent = crate::routes::chat::bygg_agent(
        &pool,
        &samtal,
        req.goal_context.as_ref(),
        plan_client.as_ref(),
        recorder.clone(),
        state.orchestrator.as_ref(),
    )
    .await
    .map_err(AgentError::Internal)?;

    let historik = crate::db::repository::get_messages(&pool, id).await?;

    let (sändare, mottagare) =
        tokio::sync::mpsc::channel::<Result<axum::body::Bytes, std::io::Error>>(16);

    // Användarmeddelandet skickas direkt - så att klienten ser det
    // utan att vänta på agenten.
    crate::routes::chat::skicka_chunk(
        &sändare,
        &crate::routes::chat::MeddelandeStycke {
            message: användarmeddelande,
            r#type: "user_message".to_string(),
        },
    )
    .await;

    let prompt = match req.goal_context.as_ref() {
        Some(mål) => format!(
            "{}\n\nUSER REQUEST:\n{}",
            crate::routes::chat::formatera_målkontext(mål),
            req.content
        ),
        None => req.content.clone(),
    };

    let agentmeddelanden = crate::routes::chat::databasmeddelanden_till_agent(&historik);

    let recorder_för_agenten = recorder.clone();

    tokio::spawn(async move {
        crate::routes::chat::kör_agenten(
            agent,
            agentmeddelanden,
            prompt,
            id,
            pool.clone(),
            recorder_för_agenten,
            sändare,
        )
        .await;
    });

    let body =
        axum::body::Body::from_stream(tokio_stream::wrappers::ReceiverStream::new(mottagare));

    let mut svar = axum::response::Response::new(body);

    svar.headers_mut()
        .insert("Content-Type", "application/x-ndjson".parse().unwrap());
    svar.headers_mut()
        .insert("Transfer-Encoding", "chunked".parse().unwrap());
    svar.headers_mut()
        .insert("Cache-Control", "no-cache".parse().unwrap());

    Ok(svar)
}

/// Strömmar nya meddelanden i ett samtal med SSE.
///
/// Go: `internal/api/handlers/stream.go` - `StreamMessages`.
pub async fn stream(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> Result<axum::response::Response, AgentError> {
    let pool = databas(&state)?.clone();
    let id = tolka_id(&id)?;

    let (sändare, mottagare) =
        tokio::sync::mpsc::channel::<Result<axum::body::Bytes, std::io::Error>>(16);

    let pool = pool.clone();

    tokio::spawn(async move {
        let mut senaste_kontroll = chrono::Utc::now();

        loop {
            tokio::time::sleep(std::time::Duration::from_secs(1)).await;

            let meddelanden = match crate::db::repository::get_messages_since(
                &pool,
                id,
                senaste_kontroll,
            )
            .await
            {
                Ok(meddelanden) => meddelanden,
                Err(fel) => {
                    tracing::warn!("Kunde inte läsa meddelandena: {fel}");
                    continue;
                }
            };

            for meddelande in meddelanden {
                let händelse = MeddelandeHändelse {
                    r#type: "message".to_string(),
                    id: meddelande.id.to_string(),
                    role: meddelande.role.clone(),
                    content: meddelande.content.clone(),
                    created_at: meddelande.created_at,
                };

                let Some(json) = serde_json::to_string(&händelse).ok() else {
                    continue;
                };

                if sändare
                    .send(Ok(axum::body::Bytes::from(format!("data: {json}\n\n"))))
                    .await
                    .is_err()
                {
                    return;
                }

                senaste_kontroll = meddelande.created_at;
            }
        }
    });

    let body =
        axum::body::Body::from_stream(tokio_stream::wrappers::ReceiverStream::new(mottagare));

    let mut svar = axum::response::Response::new(body);

    svar.headers_mut()
        .insert("Content-Type", "text/event-stream".parse().unwrap());
    svar.headers_mut()
        .insert("Cache-Control", "no-cache".parse().unwrap());
    svar.headers_mut()
        .insert("Connection", "keep-alive".parse().unwrap());
    svar.headers_mut()
        .insert("Transfer-Encoding", "chunked".parse().unwrap());
    svar.headers_mut().insert(
        "Access-Control-Allow-Origin",
        "http://localhost:3000".parse().unwrap(),
    );

    Ok(svar)
}

/// En händelse i meddelandeströmmen.
///
/// Go: `internal/api/handlers/stream.go` - `MessageEvent`. Fältens
/// ordning är densamma som i Go - de skrivs i structens ordning.
#[derive(Debug, Serialize)]
pub struct MeddelandeHändelse {
    /// Händelsens typ.
    pub r#type: String,
    /// Meddelandets id.
    pub id: String,
    /// Meddelandets roll.
    pub role: String,
    /// Meddelandets innehåll.
    pub content: String,
    /// När meddelandet skapades.
    pub created_at: chrono::DateTime<chrono::Utc>,
}
