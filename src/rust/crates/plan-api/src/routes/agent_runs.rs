//! Agentkörningar.
//!
//! Port av `src/api/internal/handlers/agent_runs.go` och
//! `agent_run_lifecycle.go`.

use axum::body::Bytes;
use axum::extract::{Path, Query, State};
use axum::http::StatusCode;
use axum::response::{IntoResponse, Response};
use axum::{Extension, Json};
use serde::Deserialize;
use std::collections::BTreeMap;

use crate::auth::Caller;
use crate::error::AppError;
use crate::models::agent::{
    self, AgentRun, AgentRunEvent, AgentRunLifecycle, CreateAgentRunEventRequest,
    CreateAgentRunRequest, UpdateAgentRunRequest,
};
use crate::repository::{agent_run, agent_run_event, agent_usage};
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

/// Listar agentkörningar.
#[utoipa::path(
    get,
    path = "/api/v1/agent-runs",
    tag = "agent-runs",
    params(("goal_id" = Option<i64>, Query, description = "Filtrera på mål")),
    responses(
        (status = 200, description = "Körningarna", body = Vec<AgentRun>),
        (status = 401, description = "Ingen giltig session")
    )
)]
pub async fn list(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Query(params): Query<ListParams>,
) -> Result<Json<Vec<AgentRun>>, AppError> {
    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;

    let goal_id = match params.goal_id {
        Some(raw) => Some(
            raw.parse::<i64>()
                .map_err(|_| AppError::BadRequest("Invalid goal_id".to_string()))?,
        ),
        None => None,
    };

    Ok(Json(agent_run::list(pool, user_id, goal_id).await?))
}

/// Parametern för listan.
#[derive(Debug, Deserialize)]
pub struct ListParams {
    /// Filtrera på mål.
    #[serde(default)]
    pub goal_id: Option<String>,
}

/// Startar en agentkörning.
#[utoipa::path(
    post,
    path = "/api/v1/agent-runs",
    tag = "agent-runs",
    request_body = CreateAgentRunRequest,
    responses(
        (status = 201, description = "Körningen", body = AgentRun),
        (status = 400, description = "Ogiltig fas, status, trigger eller profil"),
        (status = 401, description = "Ingen giltig session"),
        (status = 404, description = "Målet tillhör inte användaren"),
        (status = 409, description = "En budget stoppar körningen")
    )
)]
pub async fn create(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    body: Bytes,
) -> Result<Response, AppError> {
    let req: CreateAgentRunRequest = crate::routes::parse_body(&body, "goal_id is required")?;

    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;

    validera_skapa(&req)?;

    let (körning, varning) = agent_run::create(pool, user_id, &req).await?;

    // Budgetvarningen finns med när varningsgränsen är nådd - körningen
    // får starta, men människan behöver veta att budgeten är nära slut.
    let mut svar = serde_json::to_value(&körning).unwrap_or_default();
    if let (serde_json::Value::Object(fält), Some(varning)) = (&mut svar, &varning) {
        fält.insert(
            "budget_warning".to_string(),
            serde_json::to_value(varning).unwrap_or_default(),
        );
    }

    Ok((StatusCode::CREATED, Json(svar)).into_response())
}

/// Kontrollerar fälten i en ny körning.
fn validera_skapa(req: &CreateAgentRunRequest) -> Result<(), AppError> {
    if !req.phase.is_empty() && !agent::giltig_fas(&req.phase) {
        return Err(AppError::BadRequest(format!(
            "invalid agent run: okänd fas {:?}",
            req.phase
        )));
    }
    if !req.status.is_empty() && !agent::giltig_status(&req.status) {
        return Err(AppError::BadRequest(format!(
            "invalid agent run: okänt tillstånd {:?}",
            req.status
        )));
    }
    if !req.trigger.is_empty() && !agent::giltig_trigger(&req.trigger) {
        return Err(AppError::BadRequest(format!(
            "invalid agent run: okänd trigger {:?}",
            req.trigger
        )));
    }
    if !req.agent_profile.is_empty() && !agent::giltig_profil(&req.agent_profile) {
        return Err(AppError::BadRequest(format!(
            "invalid agent run: okänd agentprofil {:?}",
            req.agent_profile
        )));
    }

    Ok(())
}

/// Uppdaterar en agentkörning.
#[utoipa::path(
    patch,
    path = "/api/v1/agent-runs/{id}",
    tag = "agent-runs",
    params(("id" = i64, Path, description = "Körningens id")),
    request_body = UpdateAgentRunRequest,
    responses(
        (status = 200, description = "Körningen", body = AgentRun),
        (status = 400, description = "Ogiltig övergång eller okänt värde"),
        (status = 401, description = "Ingen giltig session"),
        (status = 404, description = "Körningen hittades inte")
    )
)]
pub async fn update(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
    body: Bytes,
) -> Result<Json<AgentRun>, AppError> {
    let req: UpdateAgentRunRequest = crate::routes::parse_body(&body, "Invalid request body")?;

    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;
    let run_id = run_id(&id)?;

    validera_uppdatera(&req)?;

    let nuvarande = agent_run::get(pool, user_id, run_id)
        .await?
        .ok_or_else(|| AppError::NotFound("Agent run not found".to_string()))?;

    kontrollera_övergång(&nuvarande, &req)?;

    let körning = agent_run::update(pool, user_id, run_id, &req)
        .await?
        .ok_or_else(|| AppError::NotFound("Agent run not found".to_string()))?;

    if agent::är_terminal(&körning.status) {
        säkerställ_kostnadsrad(pool, user_id, run_id).await;
    }

    Ok(Json(körning))
}

/// Kontrollerar fälten i en uppdatering.
fn validera_uppdatera(req: &UpdateAgentRunRequest) -> Result<(), AppError> {
    if let Some(phase) = &req.phase {
        if !agent::giltig_fas(phase) {
            return Err(AppError::BadRequest(format!(
                "invalid agent run: okänd fas {phase:?}"
            )));
        }
    }
    if let Some(status) = &req.status {
        if !agent::giltig_status(status) {
            return Err(AppError::BadRequest(format!(
                "invalid agent run: okänt tillstånd {status:?}"
            )));
        }
    }
    if let Some(trigger) = &req.trigger {
        if !agent::giltig_trigger(trigger) {
            return Err(AppError::BadRequest(format!(
                "invalid agent run: okänd trigger {trigger:?}"
            )));
        }
    }
    if let Some(agent_profile) = &req.agent_profile {
        if !agent::giltig_profil(agent_profile) {
            return Err(AppError::BadRequest(format!(
                "invalid agent run: okänd agentprofil {agent_profile:?}"
            )));
        }
    }

    Ok(())
}

/// Kontrollerar att körningen får gå framåt i livscykeln.
///
/// En terminal körning får inte återupplivas - avbrytning är ett sätt att
/// stanna, inte ett sätt att skriva om historiken.
fn kontrollera_övergång(
    nuvarande: &AgentRun,
    req: &UpdateAgentRunRequest,
) -> Result<(), AppError> {
    if let Some(fas) = &req.phase {
        if agent::fasrang(fas) < agent::fasrang(&nuvarande.phase) {
            return Err(AppError::BadRequest(format!(
                "invalid agent run: fasen kan inte gå bakåt från {:?} till {:?}",
                nuvarande.phase, fas
            )));
        }
    }

    if let Some(status) = &req.status {
        if !giltig_övergång(&nuvarande.status, status) {
            return Err(AppError::BadRequest(format!(
                "invalid agent run: tillståndet kan inte gå från {:?} till {:?}",
                nuvarande.status, status
            )));
        }
    }

    Ok(())
}

/// Om ett tillstånd får övergå till ett annat.
fn giltig_övergång(nuvarande: &str, begärt: &str) -> bool {
    if nuvarande == begärt {
        return true;
    }
    if agent::är_terminal(nuvarande) {
        return false;
    }

    agent::övergångar()
        .into_iter()
        .find(|(från, _)| *från == nuvarande)
        .map(|(_, till)| till.contains(&begärt))
        .unwrap_or(false)
}

/// Avbryter en körning.
#[utoipa::path(
    post,
    path = "/api/v1/agent-runs/{id}/cancel",
    tag = "agent-runs",
    params(("id" = i64, Path, description = "Körningens id")),
    responses(
        (status = 200, description = "Körningen", body = AgentRun),
        (status = 400, description = "Körningen är redan avslutad"),
        (status = 401, description = "Ingen giltig session"),
        (status = 404, description = "Körningen hittades inte")
    )
)]
pub async fn cancel(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
) -> Result<Json<AgentRun>, AppError> {
    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;
    let run_id = run_id(&id)?;

    let nuvarande = agent_run::get(pool, user_id, run_id)
        .await?
        .ok_or_else(|| AppError::NotFound("Agent run not found".to_string()))?;

    let status = agent::STATUS_CANCELLED.to_string();
    let summary = "Körningen avbröts av en människa".to_string();
    let result = "Körningen avbröts innan den blev klar".to_string();
    let finished_at = chrono::Utc::now();

    let req = UpdateAgentRunRequest {
        status: Some(status),
        summary: Some(summary),
        result: Some(result),
        finished_at: Some(finished_at),
        ..Default::default()
    };

    kontrollera_övergång(&nuvarande, &req)?;

    let körning = agent_run::update(pool, user_id, run_id, &req)
        .await?
        .ok_or_else(|| AppError::NotFound("Agent run not found".to_string()))?;

    säkerställ_kostnadsrad(pool, user_id, run_id).await;

    // Händelsen gör det synligt i historiken att det var en människa som
    // avbröt - inte ett fel.
    agent_run_event::create(
        pool,
        user_id,
        run_id,
        &CreateAgentRunEventRequest {
            r#type: "run_cancelled".to_string(),
            phase: körning.phase.clone(),
            payload: Some(serde_json::json!({ "cancelled_by": "user" })),
        },
    )
    .await?;

    Ok(Json(körning))
}

/// Skapar en kostnadsrad om körningen saknar en.
///
/// Varje avslutad körning ska ha en rad - annars ser det ut som att
/// agenten aldrig kostade något. Raden markeras som ofullständig när
/// agenten inte hann rapportera användningen.
async fn säkerställ_kostnadsrad(pool: &sqlx::PgPool, user_id: i64, run_id: i64) {
    if let Err(err) = agent_usage::ensure_row(pool, user_id, run_id).await {
        // Körningen är sparad. Att kostnadsraden saknas får inte stoppa
        // svaret - det loggas i stället.
        tracing::warn!("Kunde inte skapa kostnadsrad för körning {run_id}: {err}");
    }
}

/// Listar händelserna i en körning.
#[utoipa::path(
    get,
    path = "/api/v1/agent-runs/{id}/events",
    tag = "agent-runs",
    params(("id" = i64, Path, description = "Körningens id")),
    responses(
        (status = 200, description = "Händelserna", body = Vec<AgentRunEvent>),
        (status = 401, description = "Ingen giltig session")
    )
)]
pub async fn list_events(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
) -> Result<Json<Vec<AgentRunEvent>>, AppError> {
    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;
    let run_id = run_id(&id)?;

    Ok(Json(agent_run_event::list(pool, user_id, run_id).await?))
}

/// Sparar en händelse i en körning.
#[utoipa::path(
    post,
    path = "/api/v1/agent-runs/{id}/events",
    tag = "agent-runs",
    params(("id" = i64, Path, description = "Körningens id")),
    request_body = CreateAgentRunEventRequest,
    responses(
        (status = 201, description = "Händelsen", body = AgentRunEvent),
        (status = 400, description = "Typen saknas eller är okänd"),
        (status = 401, description = "Ingen giltig session"),
        (status = 404, description = "Körningen hittades inte")
    )
)]
pub async fn create_event(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
    body: Bytes,
) -> Result<Response, AppError> {
    let req: CreateAgentRunEventRequest = crate::routes::parse_body(&body, "type is required")?;

    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;
    let run_id = run_id(&id)?;

    if req.r#type.is_empty() {
        return Err(AppError::BadRequest("type is required".to_string()));
    }
    if !agent::giltig_händelsetyp(&req.r#type) {
        return Err(AppError::BadRequest(format!(
            "invalid agent run event: okänd händelsetyp {:?}",
            req.r#type
        )));
    }
    if !req.phase.is_empty() && !agent::giltig_fas(&req.phase) {
        return Err(AppError::BadRequest(format!(
            "invalid agent run: okänd fas {:?}",
            req.phase
        )));
    }

    let händelse = agent_run_event::create(pool, user_id, run_id, &req)
        .await?
        .ok_or_else(|| AppError::NotFound("Agent run not found".to_string()))?;

    Ok((StatusCode::CREATED, Json(händelse)).into_response())
}

/// Beskriver agentkörningens livscykel.
///
/// Gränssnittet och andra klienter ska inte hålla egna listor över faser
/// och tillstånd - de läser dem härifrån.
#[utoipa::path(
    get,
    path = "/api/v1/agent-run-lifecycle",
    tag = "agent-runs",
    responses(
        (status = 200, description = "Livscykeln", body = AgentRunLifecycle),
        (status = 401, description = "Ingen giltig session")
    )
)]
pub async fn lifecycle(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
) -> Result<Json<AgentRunLifecycle>, AppError> {
    let _ = user_id(&caller)?;
    let _ = &state;

    let terminal: Vec<String> = agent::STATUSES
        .iter()
        .filter(|status| agent::är_terminal(status))
        .map(|status| status.to_string())
        .collect();

    let needs_human: Vec<String> = agent::STATUSES
        .iter()
        .filter(|status| agent::kräver_människa(status))
        .map(|status| status.to_string())
        .collect();

    let transitions: BTreeMap<String, Option<Vec<String>>> = agent::övergångar()
        .into_iter()
        .map(|(från, till)| {
            let till: Vec<String> = till.iter().map(|t| t.to_string()).collect();
            // Go lämnar en tom lista som null - därför gör Rust det också.
            (från.to_string(), (!till.is_empty()).then_some(till))
        })
        .collect();

    Ok(Json(AgentRunLifecycle {
        phases: agent::PHASES.iter().map(|f| f.to_string()).collect(),
        status: agent::STATUSES.iter().map(|s| s.to_string()).collect(),
        terminal,
        needs_human,
        transitions,
    }))
}
