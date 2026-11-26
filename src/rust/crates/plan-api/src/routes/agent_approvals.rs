//! Godkännanden - mänskliga beslut i agentkörningar.
//!
//! Port av `src/api/internal/handlers/agent_approvals.go`.
//!
//! Agenten får föreslå, men beslutet fattas alltid av en människa. Därför
//! ligger beslutet i Plan och inte i agentens eget minne.

use axum::body::Bytes;
use axum::extract::{Path, Query, State};
use axum::http::StatusCode;
use axum::response::IntoResponse;
use axum::{Extension, Json};
use serde::Deserialize;

use crate::auth::Caller;
use crate::error::AppError;
use crate::models::agent::{
    AgentApproval, CreateAgentApprovalRequest, CreateAgentRunEventRequest,
    DecideAgentApprovalRequest,
};
use crate::repository::{agent_approval, agent_run_event};
use crate::routes::state::AppState;

/// Användarens id ur anropet.
fn user_id(caller: &Caller) -> Result<i64, AppError> {
    if caller.user_id == 0 {
        return Err(AppError::Unauthorized("Unauthorized".to_string()));
    }
    Ok(caller.user_id)
}

/// Databasen, eller ett fel som säger att den saknas.
fn pool(state: &AppState) -> Result<&sqlx::PgPool, AppError> {
    state.pool.as_ref().ok_or_else(|| {
        AppError::ServiceUnavailable(
            "Database not connected. Set DATABASE_URL environment variable.".to_string(),
        )
    })
}

/// Körningens id ur sökvägen.
fn run_id(id: &str) -> Result<i64, AppError> {
    id.parse::<i64>()
        .map_err(|_| AppError::BadRequest("Invalid agent run ID".to_string()))
}

/// Godkännandets id ur sökvägen.
fn approval_id(id: &str) -> Result<i64, AppError> {
    id.parse::<i64>()
        .map_err(|_| AppError::BadRequest("Invalid approval ID".to_string()))
}

/// Listar godkännanden för en körning.
#[utoipa::path(
    get,
    path = "/api/v1/agent-runs/{id}/approvals",
    tag = "agent-runs",
    params(("id" = i64, Path, description = "Körningens id")),
    responses(
        (status = 200, description = "Godkännandena", body = Vec<AgentApproval>),
        (status = 401, description = "Ingen giltig session")
    )
)]
pub async fn list_for_run(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
) -> Result<Json<Vec<AgentApproval>>, AppError> {
    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;
    let run_id = run_id(&id)?;

    Ok(Json(agent_approval::list(pool, user_id, run_id).await?))
}

/// Listar de godkännanden som väntar på ett beslut.
#[utoipa::path(
    get,
    path = "/api/v1/agent-approvals",
    tag = "agent-runs",
    params(("goal_id" = Option<i64>, Query, description = "Bara godkännanden för detta mål")),
    responses(
        (status = 200, description = "Godkännandena", body = Vec<AgentApproval>),
        (status = 400, description = "Ogiltigt mål-id"),
        (status = 401, description = "Ingen giltig session")
    )
)]
pub async fn list_pending(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Query(params): Query<PendingParams>,
) -> Result<Json<Vec<AgentApproval>>, AppError> {
    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;

    // Ett mål kan ha flera godkännanden - då visas just det målets.
    if let Some(raw) = params.goal_id.filter(|raw| !raw.is_empty()) {
        let goal_id = raw
            .parse::<i64>()
            .map_err(|_| AppError::BadRequest("Invalid goal ID".to_string()))?;

        return Ok(Json(
            agent_approval::list_by_goal(pool, user_id, goal_id).await?,
        ));
    }

    Ok(Json(agent_approval::list_pending(pool, user_id).await?))
}

/// Parametern för listan med väntande beslut.
#[derive(Debug, Default, Deserialize)]
pub struct PendingParams {
    /// Bara godkännanden för detta mål.
    #[serde(default)]
    pub goal_id: Option<String>,
}

/// Föreslår en ändring som behöver godkännande.
#[utoipa::path(
    post,
    path = "/api/v1/agent-runs/{id}/approvals",
    tag = "agent-runs",
    params(("id" = i64, Path, description = "Körningens id")),
    request_body = CreateAgentApprovalRequest,
    responses(
        (status = 201, description = "Godkännandet", body = AgentApproval),
        (status = 400, description = "Titeln saknas eller typen är okänd"),
        (status = 401, description = "Ingen giltig session"),
        (status = 404, description = "Körningen hittades inte")
    )
)]
pub async fn create(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
    body: Bytes,
) -> Result<impl IntoResponse, AppError> {
    let req: CreateAgentApprovalRequest = crate::routes::parse_body(&body, "title is required")?;

    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;
    let run_id = run_id(&id)?;

    if req.title.as_deref().unwrap_or_default().is_empty() {
        return Err(AppError::BadRequest("title is required".to_string()));
    }

    kontrollera_typ(&req)?;

    let godkännande = agent_approval::create(pool, user_id, &req, run_id)
        .await?
        .ok_or_else(|| AppError::NotFound("Agent run not found".to_string()))?;

    Ok((StatusCode::CREATED, Json(godkännande)))
}

/// Kontrollerar att godkännandets typ finns i kontraktet mellan Plan och
/// agenten.
///
/// Ett förslag utan typ godtas - det är ett äldre förslag, före typerna
/// introducerades, och behandlas som ett kodförslag.
fn kontrollera_typ(req: &CreateAgentApprovalRequest) -> Result<(), AppError> {
    let Some(payload) = &req.payload else {
        return Ok(());
    };

    let Some(kind) = payload.get("kind").and_then(|värde| värde.as_str()) else {
        return Ok(());
    };

    if kind.is_empty() {
        return Ok(());
    }

    match crate::contracts::shared() {
        Some(kontrakt) if !kontrakt.har_godkännandetyp(kind) => Err(AppError::BadRequest(
            format!("okänd godkännandetyp {kind:?}"),
        )),
        // Kontraktet kunde inte läsas. Då går förslaget igenom - ett
        // förslag ska inte stoppas av att en fil saknas.
        _ => Ok(()),
    }
}

/// Fattar beslutet om ett godkännande.
///
/// Bara en människa kan göra detta - agent-scopet stoppar anropet innan
/// det når hit.
#[utoipa::path(
    post,
    path = "/api/v1/agent-approvals/{id}/decision",
    tag = "agent-runs",
    params(("id" = i64, Path, description = "Godkännandets id")),
    request_body = DecideAgentApprovalRequest,
    responses(
        (status = 200, description = "Godkännandet", body = AgentApproval),
        (status = 400, description = "Beslutet saknas eller är okänt"),
        (status = 401, description = "Ingen giltig session"),
        (status = 404, description = "Godkännandet hittades inte eller är redan avgjort")
    )
)]
pub async fn decide(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
    body: Bytes,
) -> Result<Json<AgentApproval>, AppError> {
    let req: DecideAgentApprovalRequest = crate::routes::parse_body(&body, "decision is required")?;

    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;
    let approval_id = approval_id(&id)?;

    if req.decision.is_empty() {
        return Err(AppError::BadRequest("decision is required".to_string()));
    }
    if !crate::models::agent::giltigt_beslut(&req.decision) {
        return Err(AppError::BadRequest(format!(
            "invalid approval: okänt beslut {:?}",
            req.decision
        )));
    }

    let godkännande =
        agent_approval::decide(pool, user_id, approval_id, &req.decision, &req.feedback)
            .await?
            .ok_or_else(|| {
                AppError::NotFound("Approval not found or already decided".to_string())
            })?;

    // Beslutet skrivs till körningens historik så att det går att se
    // att en människa fattade det.
    let mut payload = serde_json::json!({
        "decision": godkännande.status,
        "approval_id": godkännande.id,
    });
    if !godkännande.feedback.is_empty() {
        payload["feedback"] = serde_json::Value::String(godkännande.feedback.clone());
    }

    agent_run_event::create(
        pool,
        user_id,
        godkännande.agent_run_id,
        &CreateAgentRunEventRequest {
            r#type: "approval_decided".to_string(),
            phase: String::new(),
            payload: Some(payload),
        },
    )
    .await?;

    Ok(Json(godkännande))
}
