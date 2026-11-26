//! Kostnad och budgetar för agentkörningar.
//!
//! Port av `src/api/internal/handlers/agent_usage.go`.
//!
//! Plan äger kostnaderna. Agenten rapporterar bara användningen - räknandet
//! görs här, med en versionsstyrd prislista, så att kostnaden kan
//! reproduceras.

use axum::body::Bytes;
use axum::extract::{Path, Query, State};
use axum::http::StatusCode;
use axum::response::IntoResponse;
use axum::{Extension, Json};
use serde::Deserialize;
use serde_json::json;

use crate::auth::Caller;
use crate::error::AppError;
use crate::models::agent::{
    AgentBudget, AgentRunUsage, CreateAgentBudgetRequest, CreateAgentRunUsageRequest,
    UpdateAgentBudgetRequest,
};
use crate::repository::{agent_budget, agent_run, agent_usage};
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

/// Budgetens id ur sökvägen.
fn budget_id(id: &str) -> Result<i64, AppError> {
    id.parse::<i64>()
        .map_err(|_| AppError::BadRequest("Invalid budget ID".to_string()))
}

/// Sparar en användningsrad för en körning.
#[utoipa::path(
    post,
    path = "/api/v1/agent-runs/{id}/usage",
    tag = "agent-usage",
    params(("id" = i64, Path, description = "Körningens id")),
    request_body = CreateAgentRunUsageRequest,
    responses(
        (status = 200, description = "Kostnadsraden", body = AgentRunUsage),
        (status = 400, description = "Ogiltig body"),
        (status = 401, description = "Ingen giltig session"),
        (status = 404, description = "Körningen hittades inte")
    )
)]
pub async fn record_usage(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
    body: Bytes,
) -> Result<Json<AgentRunUsage>, AppError> {
    let req: CreateAgentRunUsageRequest = crate::routes::parse_body(&body, "Invalid request body")?;

    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;
    let run_id = run_id(&id)?;

    let användning = agent_usage::record(pool, user_id, run_id, &req)
        .await?
        .ok_or_else(|| AppError::NotFound("Agent run not found".to_string()))?;

    Ok(Json(användning))
}

/// Läser användningen för en körning.
#[utoipa::path(
    get,
    path = "/api/v1/agent-runs/{id}/usage",
    tag = "agent-usage",
    params(("id" = i64, Path, description = "Körningens id")),
    responses(
        (status = 200, description = "Kostnadsraden", body = AgentRunUsage),
        (status = 401, description = "Ingen giltig session"),
        (status = 404, description = "Ingen kostnad registrerad för körningen")
    )
)]
pub async fn get_usage(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
) -> Result<Json<AgentRunUsage>, AppError> {
    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;
    let run_id = run_id(&id)?;

    let användning = agent_usage::get_by_run(pool, user_id, run_id)
        .await?
        .ok_or_else(|| AppError::NotFound("No usage recorded for this run".to_string()))?;

    Ok(Json(användning))
}

/// Listar och summerar användningen.
#[utoipa::path(
    get,
    path = "/api/v1/agent-usage",
    tag = "agent-usage",
    params(
        ("agent_id" = Option<String>, Query, description = "Filtrera på agent"),
        ("model" = Option<String>, Query, description = "Filtrera på modell"),
        ("tool" = Option<String>, Query, description = "Filtrera på verktyg"),
        ("from" = Option<String>, Query, description = "Periodens början (RFC3339)"),
        ("to" = Option<String>, Query, description = "Periodens slut (RFC3339)")
    ),
    responses(
        (status = 200, description = "Raderna och summeringen"),
        (status = 400, description = "Ogiltig period"),
        (status = 401, description = "Ingen giltig session")
    )
)]
pub async fn list_usage(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Query(params): Query<UsageParams>,
) -> Result<Json<serde_json::Value>, AppError> {
    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;

    let filter = agent_usage::UsageFilter {
        agent_id: params.agent_id.unwrap_or_default(),
        model: params.model.unwrap_or_default(),
        tool: params.tool.unwrap_or_default(),
        from: params.from.map(läs_tid).transpose()?,
        to: params.to.map(läs_tid).transpose()?,
    };

    let rader = agent_usage::list(pool, user_id, &filter).await?;
    let summary = agent_usage::summary(pool, user_id, &filter).await?;

    Ok(Json(json!({
        "summary": summary,
        "rows": rader,
    })))
}

/// Parametrarna för kostnadslistan.
#[derive(Debug, Default, Deserialize)]
pub struct UsageParams {
    /// Filtrera på agent.
    #[serde(default)]
    pub agent_id: Option<String>,
    /// Filtrera på modell.
    #[serde(default)]
    pub model: Option<String>,
    /// Filtrera på verktyg.
    #[serde(default)]
    pub tool: Option<String>,
    /// Periodens början.
    #[serde(default)]
    pub from: Option<String>,
    /// Periodens slut.
    #[serde(default)]
    pub to: Option<String>,
}

/// Läser en tidpunkt ur en sträng.
///
/// Go använder `time.Parse(time.RFC3339, ...)` - felmeddelandet från den
/// används här också, så att klienten ser samma text från båda tjänsterna.
fn läs_tid(raw: String) -> Result<chrono::DateTime<chrono::Utc>, AppError> {
    chrono::DateTime::parse_from_rfc3339(&raw)
        .map(|tid| tid.with_timezone(&chrono::Utc))
        .map_err(|_| {
            AppError::BadRequest(format!(
                "parsing time {:?} as \"2006-01-02T15:04:05Z07:00\": cannot parse {:?} as \"2006\"",
                raw, raw
            ))
        })
}

/// Listar budgetarna.
#[utoipa::path(
    get,
    path = "/api/v1/agent-budgets",
    tag = "agent-budgets",
    responses(
        (status = 200, description = "Budgetarna", body = Vec<AgentBudget>),
        (status = 401, description = "Ingen giltig session")
    )
)]
pub async fn list_budgets(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
) -> Result<Json<Vec<AgentBudget>>, AppError> {
    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;

    Ok(Json(agent_budget::list(pool, user_id).await?))
}

/// Skapar en budget.
#[utoipa::path(
    post,
    path = "/api/v1/agent-budgets",
    tag = "agent-budgets",
    request_body = CreateAgentBudgetRequest,
    responses(
        (status = 201, description = "Budgeten", body = AgentBudget),
        (status = 400, description = "Ogiltig period eller gräns"),
        (status = 401, description = "Ingen giltig session")
    )
)]
pub async fn create_budget(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    body: Bytes,
) -> Result<impl IntoResponse, AppError> {
    let mut req: CreateAgentBudgetRequest =
        crate::routes::parse_body(&body, "Invalid request body")?;

    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;

    if req.period.is_empty() {
        req.period = crate::models::agent::BUDGET_PERIOD_MONTH.to_string();
    }
    if !crate::models::agent::giltig_budgetperiod(&req.period) {
        return Err(AppError::BadRequest(format!(
            "ogiltig budgetperiod {:?}",
            req.period
        )));
    }
    if req.hard_limit_usd <= 0.0 {
        return Err(AppError::BadRequest(
            "hard_limit_usd måste vara större än noll".to_string(),
        ));
    }

    let budget = agent_budget::create(pool, user_id, &req).await?;

    Ok((StatusCode::CREATED, Json(budget)))
}

/// Ändrar en budget.
#[utoipa::path(
    patch,
    path = "/api/v1/agent-budgets/{id}",
    tag = "agent-budgets",
    params(("id" = i64, Path, description = "Budgetens id")),
    request_body = UpdateAgentBudgetRequest,
    responses(
        (status = 200, description = "Budgeten", body = AgentBudget),
        (status = 400, description = "Ogiltig period eller gräns"),
        (status = 401, description = "Ingen giltig session"),
        (status = 404, description = "Budgeten hittades inte")
    )
)]
pub async fn update_budget(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
    body: Bytes,
) -> Result<Json<AgentBudget>, AppError> {
    let req: UpdateAgentBudgetRequest = crate::routes::parse_body(&body, "Invalid request body")?;

    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;
    let budget_id = budget_id(&id)?;

    if let Some(period) = &req.period {
        if !crate::models::agent::giltig_budgetperiod(period) {
            return Err(AppError::BadRequest(format!(
                "ogiltig budgetperiod {period:?}"
            )));
        }
    }
    if let Some(gräns) = req.hard_limit_usd {
        if gräns <= 0.0 {
            return Err(AppError::BadRequest(
                "hard_limit_usd måste vara större än noll".to_string(),
            ));
        }
    }

    let budget = agent_budget::update(pool, user_id, budget_id, &req).await?;

    Ok(Json(budget))
}

/// Tar bort en budget.
#[utoipa::path(
    delete,
    path = "/api/v1/agent-budgets/{id}",
    tag = "agent-budgets",
    params(("id" = i64, Path, description = "Budgetens id")),
    responses(
        (status = 204, description = "Borttagen"),
        (status = 401, description = "Ingen giltig session"),
        (status = 404, description = "Budgeten hittades inte")
    )
)]
pub async fn delete_budget(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
) -> Result<StatusCode, AppError> {
    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;
    let budget_id = budget_id(&id)?;

    agent_budget::delete(pool, user_id, budget_id).await?;

    Ok(StatusCode::NO_CONTENT)
}

/// Visar hur en budget ligger.
#[utoipa::path(
    get,
    path = "/api/v1/agent-budgets/{id}/status",
    tag = "agent-budgets",
    params(("id" = i64, Path, description = "Budgetens id")),
    responses(
        (status = 200, description = "Budgetens läge", body = crate::models::agent::BudgetStatus),
        (status = 401, description = "Ingen giltig session"),
        (status = 404, description = "Budgeten hittades inte")
    )
)]
pub async fn budget_status(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
) -> Result<Json<crate::models::agent::BudgetStatus>, AppError> {
    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;
    let budget_id = budget_id(&id)?;

    let budgetar = agent_budget::list(pool, user_id).await?;

    let budget = budgetar
        .into_iter()
        .find(|budget| budget.id == budget_id)
        .ok_or_else(|| AppError::NotFound("Budget not found".to_string()))?;

    let mut anslutning = pool
        .acquire()
        .await
        .map_err(|err| AppError::Internal(format!("failed to acquire connection: {err}")))?;

    let status = agent_run::budget_status(&mut anslutning, user_id, &budget, 0.0).await?;

    Ok(Json(status))
}
