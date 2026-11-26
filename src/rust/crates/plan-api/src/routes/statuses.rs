//! Statusar.
//!
//! Port av statusdelen i `src/api/internal/handlers/goals.go`.

use axum::extract::{Extension, Path, State};
use axum::Json;

use crate::auth::Caller;
use crate::error::AppError;
use crate::models::status::{Status, StatusUpdate, UpdateStatusRequest};
use crate::repository::goal;
use crate::routes::state::AppState;

/// Databasen är inte ansluten - samma svar som Go-tjänsten.
fn no_database() -> AppError {
    AppError::ServiceUnavailable(
        "Database not connected. Set DATABASE_URL environment variable.".to_string(),
    )
}

/// Användarens id ur anropet.
fn user_id(caller: Option<&Caller>) -> Result<i64, AppError> {
    match caller {
        Some(caller) if caller.user_id != 0 => Ok(caller.user_id),
        _ => Err(AppError::Unauthorized("Unauthorized".to_string())),
    }
}

/// Tolkar ett mål-id ur sökvägen.
fn goal_id(id: &str) -> Result<i64, AppError> {
    id.parse::<i64>()
        .map_err(|_| AppError::BadRequest("Invalid goal ID".to_string()))
}

/// Alla statusar.
#[utoipa::path(
    get,
    path = "/api/v1/statuses",
    tag = "statuses",
    responses(
        (status = 200, description = "Statusarna", content_type = "application/json",
            body = Vec<Status>),
        (status = 401, description = "Ingen giltig session"),
        (status = 503, description = "Databasen är inte ansluten")
    )
)]
pub async fn get_statuses(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
) -> Result<Json<Vec<Status>>, AppError> {
    let _ = user_id(Some(&caller))?;

    let Some(pool) = state.pool.as_ref() else {
        return Err(no_database());
    };

    let statusar = goal::get_all_statuses(pool).await?;

    Ok(Json(statusar))
}

/// Statusen på ett mål.
#[utoipa::path(
    get,
    path = "/api/v1/goals/{id}/status",
    tag = "statuses",
    params(("id" = i64, Path, description = "Målets id")),
    responses(
        (status = 200, description = "Statusen", content_type = "application/json", body = Status),
        (status = 400, description = "Ogiltigt mål-id"),
        (status = 403, description = "Ingen åtkomst till målet"),
        (status = 404, description = "Målet finns inte")
    )
)]
pub async fn get_goal_status(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
) -> Result<Json<Status>, AppError> {
    let user_id = user_id(Some(&caller))?;
    let goal_id = goal_id(&id)?;

    let Some(pool) = state.pool.as_ref() else {
        return Err(no_database());
    };

    if !goal::user_has_access(pool, user_id, goal_id).await? {
        return Err(AppError::Forbidden("Access denied".to_string()));
    }

    let status = goal::get_goal_status(pool, goal_id)
        .await?
        .ok_or_else(|| AppError::NotFound("Goal not found".to_string()))?;

    Ok(Json(status))
}

/// Byter status på ett mål.
#[utoipa::path(
    patch,
    path = "/api/v1/goals/{id}/status",
    tag = "statuses",
    params(("id" = i64, Path, description = "Målets id")),
    request_body = UpdateStatusRequest,
    responses(
        (status = 200, description = "Statusändringen", content_type = "application/json",
            body = StatusUpdate),
        (status = 400, description = "Ogiltig request eller mål-id"),
        (status = 403, description = "Ingen åtkomst till målet")
    )
)]
pub async fn update_goal_status(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
    Json(req): Json<UpdateStatusRequest>,
) -> Result<Json<StatusUpdate>, AppError> {
    let user_id = user_id(Some(&caller))?;
    let goal_id = goal_id(&id)?;

    let Some(pool) = state.pool.as_ref() else {
        return Err(no_database());
    };

    if !goal::user_has_access(pool, user_id, goal_id).await? {
        return Err(AppError::Forbidden("Access denied".to_string()));
    }

    let update = goal::update_goal_status(pool, goal_id, req.status_id).await?;

    Ok(Json(update))
}

/// Statushistoriken för ett mål.
#[utoipa::path(
    get,
    path = "/api/v1/goals/{id}/status/history",
    tag = "statuses",
    params(("id" = i64, Path, description = "Målets id")),
    responses(
        (status = 200, description = "Historiken", content_type = "application/json",
            body = Vec<StatusUpdate>),
        (status = 400, description = "Ogiltigt mål-id"),
        (status = 403, description = "Ingen åtkomst till målet")
    )
)]
pub async fn get_status_history(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
) -> Result<Json<Vec<StatusUpdate>>, AppError> {
    let user_id = user_id(Some(&caller))?;
    let goal_id = goal_id(&id)?;

    let Some(pool) = state.pool.as_ref() else {
        return Err(no_database());
    };

    if !goal::user_has_access(pool, user_id, goal_id).await? {
        return Err(AppError::Forbidden("Access denied".to_string()));
    }

    let historik = goal::get_status_history(pool, goal_id).await?;

    Ok(Json(historik))
}
