//! Godkännanden - mänskliga beslut i agentkörningar.
//!
//! Port av `src/api/internal/repository/agent_approval_repository.go`.
//!
//! Plan äger godkännandena så att besluten finns kvar även när agenttjänsten
//! är nere - och så att agenten aldrig kan godkänna sig själv.

use sqlx::postgres::PgRow;
use sqlx::{PgPool, Row};

use crate::error::AppError;
use crate::models::agent::{
    AgentApproval, CreateAgentApprovalRequest, APPROVAL_APPROVED, APPROVAL_PENDING,
    APPROVAL_REJECTED,
};

/// Kolumnerna som läses för ett godkännande.
const KOLUMNER: &str = "aa.id, aa.agent_run_id, aa.goal_id, aa.title, aa.description, aa.status, \
     aa.decided_by, aa.feedback, aa.consequence, aa.verification_plan, aa.payload, aa.created_at, \
     aa.decided_at";

/// Läser en godkännanderad.
fn läs_godkännande(rad: &PgRow) -> AgentApproval {
    AgentApproval {
        id: rad.get::<i64, _>(0),
        agent_run_id: rad.get::<i32, _>(1) as i64,
        goal_id: rad.get::<i32, _>(2) as i64,
        title: rad.get::<String, _>(3),
        description: rad.get::<String, _>(4),
        status: rad.get::<String, _>(5),
        decided_by: rad.get::<Option<i32>, _>(6).map(|id| id as i64),
        feedback: rad.get::<String, _>(7),
        consequence: rad.get::<String, _>(8),
        verification_plan: rad.get::<String, _>(9),
        payload: rad.get::<Option<serde_json::Value>, _>(10),
        created_at: rad.get::<chrono::DateTime<chrono::Utc>, _>(11),
        decided_at: rad.get::<Option<chrono::DateTime<chrono::Utc>>, _>(12),
    }
}

/// Godkännanden för en körning, nyast först.
pub async fn list(
    pool: &PgPool,
    user_id: i64,
    run_id: i64,
) -> Result<Vec<AgentApproval>, AppError> {
    let fråga = format!(
        "SELECT {KOLUMNER} FROM agent_approvals aa \
         JOIN agent_runs ar ON ar.id = aa.agent_run_id \
         JOIN user_goals ug ON ug.goal_id = ar.goal_id \
         WHERE aa.agent_run_id = $1 AND ug.user_id = $2 \
         ORDER BY aa.created_at DESC, aa.id DESC"
    );

    let rader = sqlx::query(&fråga)
        .bind(run_id)
        .bind(user_id)
        .fetch_all(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to query agent approvals: {err}")))?;

    Ok(rader.iter().map(läs_godkännande).collect())
}

/// De godkännanden som väntar på ett beslut.
///
/// Det är listan en människa behöver se - alla beslut som väntar, oavsett
/// vilket mål de gäller.
pub async fn list_pending(pool: &PgPool, user_id: i64) -> Result<Vec<AgentApproval>, AppError> {
    let fråga = format!(
        "SELECT {KOLUMNER} FROM agent_approvals aa \
         JOIN agent_runs ar ON ar.id = aa.agent_run_id \
         JOIN user_goals ug ON ug.goal_id = ar.goal_id \
         WHERE ug.user_id = $1 AND aa.status = $2 \
         ORDER BY aa.created_at ASC, aa.id ASC"
    );

    let rader = sqlx::query(&fråga)
        .bind(user_id)
        .bind(APPROVAL_PENDING)
        .fetch_all(pool)
        .await
        .map_err(|err| {
            AppError::Internal(format!("failed to query pending agent approvals: {err}"))
        })?;

    Ok(rader.iter().map(läs_godkännande).collect())
}

/// De godkännanden som väntar på ett beslut för ett visst mål.
pub async fn list_by_goal(
    pool: &PgPool,
    user_id: i64,
    goal_id: i64,
) -> Result<Vec<AgentApproval>, AppError> {
    let fråga = format!(
        "SELECT {KOLUMNER} FROM agent_approvals aa \
         JOIN agent_runs ar ON ar.id = aa.agent_run_id \
         JOIN user_goals ug ON ug.goal_id = ar.goal_id \
         WHERE ug.user_id = $1 AND aa.goal_id = $2 \
         ORDER BY aa.created_at ASC, aa.id ASC"
    );

    let rader = sqlx::query(&fråga)
        .bind(user_id)
        .bind(goal_id)
        .fetch_all(pool)
        .await
        .map_err(|err| {
            AppError::Internal(format!("failed to query pending agent approvals: {err}"))
        })?;

    Ok(rader.iter().map(läs_godkännande).collect())
}

/// Sparar ett godkännande som väntar på ett mänskligt beslut.
pub async fn create(
    pool: &PgPool,
    user_id: i64,
    req: &CreateAgentApprovalRequest,
    run_id: i64,
) -> Result<Option<AgentApproval>, AppError> {
    let payload = serde_json::to_value(req.payload.clone().unwrap_or_default())
        .unwrap_or_else(|_| serde_json::json!({}));

    let fråga = "INSERT INTO agent_approvals (agent_run_id, goal_id, title, description, status, consequence, \
         verification_plan, payload) \
         SELECT ar.id, ar.goal_id, $2, $3, $4, $5, $6, $7 \
         FROM agent_runs ar \
         JOIN user_goals ug ON ug.goal_id = ar.goal_id \
         WHERE ar.id = $1 AND ug.user_id = $8 \
         RETURNING id, agent_run_id, goal_id, title, description, status, \
         decided_by, feedback, consequence, verification_plan, payload, \
         created_at, decided_at".to_string();

    let rad = sqlx::query(&fråga)
        .bind(run_id)
        .bind(req.title.clone().unwrap_or_default())
        .bind(&req.description)
        .bind(APPROVAL_PENDING)
        .bind(&req.consequence)
        .bind(&req.verification_plan)
        .bind(payload)
        .bind(user_id)
        .fetch_optional(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to create agent approval: {err}")))?;

    Ok(rad.as_ref().map(läs_godkännande))
}

/// Registrerar ett mänskligt beslut.
///
/// Ett redan avgjort godkännande kan inte ändras - beslutet ska vara
/// entydigt i historiken.
pub async fn decide(
    pool: &PgPool,
    user_id: i64,
    approval_id: i64,
    decision: &str,
    feedback: &str,
) -> Result<Option<AgentApproval>, AppError> {
    let status = match decision {
        "approve" => APPROVAL_APPROVED,
        "reject" => APPROVAL_REJECTED,
        other => {
            return Err(AppError::BadRequest(format!(
                "invalid approval: okänt beslut {other:?}"
            )))
        }
    };

    let fråga = "UPDATE agent_approvals aa \
         SET status = $3, feedback = $4, decided_by = $5, decided_at = NOW() \
         FROM agent_runs ar, user_goals ug \
         WHERE aa.id = $1 AND aa.agent_run_id = ar.id AND ar.goal_id = ug.goal_id \
         AND ug.user_id = $5 AND aa.status = $2 \
         RETURNING aa.id, aa.agent_run_id, aa.goal_id, aa.title, aa.description, aa.status, \
         aa.decided_by, aa.feedback, aa.consequence, aa.verification_plan, aa.payload, \
         aa.created_at, aa.decided_at"
        .to_string();

    let rad = sqlx::query(&fråga)
        .bind(approval_id)
        .bind(APPROVAL_PENDING)
        .bind(status)
        .bind(feedback)
        .bind(user_id)
        .fetch_optional(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to decide agent approval: {err}")))?;

    Ok(rad.as_ref().map(läs_godkännande))
}

/// Hur många beslut som väntar på en människa.
pub async fn pending_count(pool: &PgPool, user_id: i64) -> Result<i64, AppError> {
    let antal: i64 = sqlx::query_scalar(
        "SELECT COUNT(*) FROM agent_approvals aa \
         JOIN agent_runs ar ON ar.id = aa.agent_run_id \
         JOIN user_goals ug ON ug.goal_id = ar.goal_id \
         WHERE ug.user_id = $1 AND aa.status = $2",
    )
    .bind(user_id)
    .bind(APPROVAL_PENDING)
    .fetch_one(pool)
    .await
    .map_err(|err| AppError::Internal(format!("failed to count pending approvals: {err}")))?;

    Ok(antal)
}
