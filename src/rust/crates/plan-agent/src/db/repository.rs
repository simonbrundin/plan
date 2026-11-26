//! Agentens databasfrågor.
//!
//! SQL:en är densamma som Go-agentens `internal/api/db/repository.go`, så
//! att båda kan köra mot samma databas.

use sqlx::PgPool;
use uuid::Uuid;

use super::models::{ApiKey, Conversation, MaskedApiKey, Message};

/// Kolumnerna för ett samtal.
const SAMTALSKOLUMNER: &str = "id, created_at, updated_at, title, model, provider";

/// Skapar ett nytt samtal.
pub async fn create_conversation(
    pool: &PgPool,
    model: &str,
    provider: &str,
) -> Result<Conversation, sqlx::Error> {
    sqlx::query_as::<_, Conversation>(&format!(
        "INSERT INTO conversations (model, provider) VALUES ($1, $2) RETURNING {SAMTALSKOLUMNER}"
    ))
    .bind(model)
    .bind(provider)
    .fetch_one(pool)
    .await
}

/// Hämtar ett samtal.
pub async fn get_conversation(
    pool: &PgPool,
    id: Uuid,
) -> Result<Option<Conversation>, sqlx::Error> {
    sqlx::query_as::<_, Conversation>(&format!(
        "SELECT {SAMTALSKOLUMNER} FROM conversations WHERE id = $1"
    ))
    .bind(id)
    .fetch_optional(pool)
    .await
}

/// Listar samtalen, nyast först.
pub async fn list_conversations(
    pool: &PgPool,
    limit: i64,
    offset: i64,
) -> Result<Vec<Conversation>, sqlx::Error> {
    sqlx::query_as::<_, Conversation>(&format!(
        "SELECT {SAMTALSKOLUMNER} FROM conversations ORDER BY created_at DESC LIMIT $1 OFFSET $2"
    ))
    .bind(limit)
    .bind(offset)
    .fetch_all(pool)
    .await
}

/// Sätter samtalets titel.
pub async fn update_conversation_title(
    pool: &PgPool,
    id: Uuid,
    title: &str,
) -> Result<(), sqlx::Error> {
    sqlx::query("UPDATE conversations SET title = $1, updated_at = NOW() WHERE id = $2")
        .bind(title)
        .bind(id)
        .execute(pool)
        .await?;

    Ok(())
}

/// Sparar ett meddelande i ett samtal.
pub async fn create_message(
    pool: &PgPool,
    conversation_id: Uuid,
    role: &str,
    content: &str,
    tokens_used: i32,
    latency_ms: i32,
) -> Result<Message, sqlx::Error> {
    let meddelande = sqlx::query_as::<_, Message>(&format!(
        "INSERT INTO messages (conversation_id, role, content, tokens_used, latency_ms) \
         VALUES ($1, $2, $3, $4, $5) RETURNING {MEDDELANDEKOLUMNER}"
    ))
    .bind(conversation_id)
    .bind(role)
    .bind(content)
    .bind(tokens_used)
    .bind(latency_ms)
    .fetch_one(pool)
    .await?;

    // Samtalets tidsstämpel uppdateras. Felet ignoreras med vetande - ett
    // meddelande som redan sparats ska inte försvinna för att stämpeln
    // inte gick att uppdatera.
    let _ = sqlx::query("UPDATE conversations SET updated_at = NOW() WHERE id = $1")
        .bind(conversation_id)
        .execute(pool)
        .await;

    Ok(meddelande)
}

/// Hämtar meddelandena i ett samtal, äldst först.
pub async fn get_messages(
    pool: &PgPool,
    conversation_id: Uuid,
) -> Result<Vec<Message>, sqlx::Error> {
    sqlx::query_as::<_, Message>(&format!(
        "SELECT {MEDDELANDEKOLUMNER} FROM messages WHERE conversation_id = $1 ORDER BY created_at ASC"
    ))
    .bind(conversation_id)
    .fetch_all(pool)
    .await
}

/// Kolumnerna för ett meddelande.
///
/// `tool_args` hämtas som text - Go läser JSONB-kolumnen som en sträng,
/// och svaret ska se likadant ut i båda tjänsterna.
const MEDDELANDEKOLUMNER: &str = "id, conversation_id, role, content, tool_name, \
     tool_args::text AS tool_args, tool_result, is_error, thinking, tokens_used, \
     latency_ms, created_at";

/// Listar de sparade API-nycklarna.
pub async fn get_api_keys(pool: &PgPool) -> Result<Vec<ApiKey>, sqlx::Error> {
    sqlx::query_as::<_, ApiKey>(
        "SELECT id, provider, api_key, model, created_at, updated_at FROM api_keys",
    )
    .fetch_all(pool)
    .await
}

/// Sätter en API-nyckel - den skapas eller uppdateras.
pub async fn set_api_key(
    pool: &PgPool,
    provider: &str,
    api_key: &str,
    model: &str,
) -> Result<(), sqlx::Error> {
    sqlx::query(
        "INSERT INTO api_keys (provider, api_key, model) VALUES ($1, $2, $3) \
         ON CONFLICT (provider) DO UPDATE SET api_key = $2, model = $3, updated_at = NOW()",
    )
    .bind(provider)
    .bind(api_key)
    .bind(model)
    .execute(pool)
    .await?;

    Ok(())
}

/// Tar bort en API-nyckel.
pub async fn delete_api_key(pool: &PgPool, provider: &str) -> Result<(), sqlx::Error> {
    sqlx::query("DELETE FROM api_keys WHERE provider = $1")
        .bind(provider)
        .execute(pool)
        .await?;

    Ok(())
}

/// Listar API-nycklarna utan att visa själva nycklarna.
pub async fn list_masked_api_keys(pool: &PgPool) -> Result<Vec<MaskedApiKey>, sqlx::Error> {
    let nycklar = get_api_keys(pool).await?;

    Ok(nycklar
        .into_iter()
        .map(|nyckel| MaskedApiKey {
            id: nyckel.id,
            provider: nyckel.provider,
            has_key: !nyckel.api_key.is_empty(),
            model: nyckel.model,
        })
        .collect())
}
