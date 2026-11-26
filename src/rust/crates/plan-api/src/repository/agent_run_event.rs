//! Händelser i agentkörningar.
//!
//! Port av `src/api/internal/repository/agent_run_event_repository.go`.

use sqlx::postgres::PgRow;
use sqlx::{PgPool, Row};

use crate::error::AppError;
use crate::models::agent::{AgentRunEvent, CreateAgentRunEventRequest};

/// Kolumnerna som läses för en händelse.
const KOLUMNER: &str = "id, agent_run_id, type, phase, payload, created_at";

/// Läser en händelserad.
fn läs_händelse(rad: &PgRow) -> AgentRunEvent {
    let payload: Option<serde_json::Value> = rad.get::<Option<serde_json::Value>, _>(4);

    AgentRunEvent {
        id: rad.get::<i64, _>(0),
        agent_run_id: rad.get::<i32, _>(1) as i64,
        r#type: rad.get::<String, _>(2),
        phase: rad.get::<String, _>(3),
        payload,
        created_at: rad.get::<chrono::DateTime<chrono::Utc>, _>(5),
    }
}

/// Händelserna i en körning, i kronologisk ordning.
pub async fn list(
    pool: &PgPool,
    user_id: i64,
    run_id: i64,
) -> Result<Vec<AgentRunEvent>, AppError> {
    let fråga = format!(
        "SELECT {KOLUMNER} FROM agent_run_events \
         WHERE agent_run_id = $1 \
         AND agent_run_id IN ( \
           SELECT ar.id FROM agent_runs ar \
           JOIN user_goals ug ON ug.goal_id = ar.goal_id \
           WHERE ug.user_id = $2 \
         ) \
         ORDER BY created_at ASC, id ASC"
    );

    let rader = sqlx::query(&fråga)
        .bind(run_id)
        .bind(user_id)
        .fetch_all(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to query agent run events: {err}")))?;

    Ok(rader.iter().map(läs_händelse).collect())
}

/// Sparar en händelse om körningen tillhör användaren.
///
/// `None` betyder att körningen inte finns eller tillhör annan.
pub async fn create(
    pool: &PgPool,
    user_id: i64,
    run_id: i64,
    req: &CreateAgentRunEventRequest,
) -> Result<Option<AgentRunEvent>, AppError> {
    let payload = serde_json::to_value(req.payload.clone().unwrap_or_default())
        .unwrap_or_else(|_| serde_json::json!({}));

    let fråga = format!(
        "INSERT INTO agent_run_events (agent_run_id, type, phase, payload) \
         SELECT $1, $2, $3, $4 \
         WHERE EXISTS ( \
           SELECT 1 FROM agent_runs ar \
           JOIN user_goals ug ON ug.goal_id = ar.goal_id \
           WHERE ar.id = $1 AND ug.user_id = $5 \
         ) \
         RETURNING {KOLUMNER}"
    );

    let rad = sqlx::query(&fråga)
        .bind(run_id)
        .bind(&req.r#type)
        .bind(&req.phase)
        .bind(payload)
        .bind(user_id)
        .fetch_optional(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to create agent run event: {err}")))?;

    Ok(rad.as_ref().map(läs_händelse))
}
