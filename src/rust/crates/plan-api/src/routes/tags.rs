//! Taggar.
//!
//! Port av `src/api/internal/handlers/tags.go`.

use axum::extract::{Extension, Path, State};
use axum::Json;

use crate::auth::Caller;
use crate::error::AppError;
use crate::models::tag::{SetGoalTagsRequest, Tag};
use crate::repository::{goal, tag};
use crate::routes::state::AppState;

/// Databasen är inte ansluten - samma svar som Go-tjänsten.
fn no_database() -> AppError {
    AppError::ServiceUnavailable(
        "Database not connected. Set DATABASE_URL environment variable.".to_string(),
    )
}

/// Användarens id ur anropet.
fn user_id(caller: &Caller) -> Result<i64, AppError> {
    if caller.user_id == 0 {
        return Err(AppError::Unauthorized("Unauthorized".to_string()));
    }
    Ok(caller.user_id)
}

/// Tolkar ett mål-id ur sökvägen.
fn goal_id(id: &str) -> Result<i64, AppError> {
    id.parse::<i64>()
        .map_err(|_| AppError::BadRequest("Invalid goal ID".to_string()))
}

/// Listar användarens taggar.
#[utoipa::path(
    get,
    path = "/api/v1/tags",
    tag = "tags",
    responses(
        (status = 200, description = "Taggarna", content_type = "application/json", body = Vec<Tag>),
        (status = 401, description = "Ingen giltig session")
    )
)]
pub async fn list(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
) -> Result<Json<Vec<Tag>>, AppError> {
    let user_id = user_id(&caller)?;

    let Some(pool) = state.pool.as_ref() else {
        return Err(no_database());
    };

    let taggar = tag::list(pool, user_id).await?;

    Ok(Json(taggar))
}

/// Listar ett måls taggar.
#[utoipa::path(
    get,
    path = "/api/v1/goals/{id}/tags",
    tag = "tags",
    params(("id" = i64, Path, description = "Målets id")),
    responses(
        (status = 200, description = "Taggarna", content_type = "application/json",
            body = Vec<String>),
        (status = 400, description = "Ogiltigt mål-id"),
        (status = 403, description = "Ingen åtkomst till målet")
    )
)]
pub async fn get_goal_tags(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
) -> Result<Json<Vec<String>>, AppError> {
    let user_id = user_id(&caller)?;
    let goal_id = goal_id(&id)?;

    let Some(pool) = state.pool.as_ref() else {
        return Err(no_database());
    };

    if !goal::user_has_access(pool, user_id, goal_id).await? {
        return Err(AppError::Forbidden("Access denied".to_string()));
    }

    let taggar = tag::get_goal_tags(pool, goal_id).await?;

    Ok(Json(taggar))
}

/// Ersätter ett måls taggar.
#[utoipa::path(
    put,
    path = "/api/v1/goals/{id}/tags",
    tag = "tags",
    params(("id" = i64, Path, description = "Målets id")),
    request_body = SetGoalTagsRequest,
    responses(
        (status = 200, description = "Taggarna", content_type = "application/json",
            body = Vec<String>),
        (status = 400, description = "Ogiltigt taggnamn"),
        (status = 403, description = "Ingen åtkomst till målet")
    )
)]
pub async fn set_goal_tags(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
    Json(req): Json<SetGoalTagsRequest>,
) -> Result<Json<Vec<String>>, AppError> {
    let user_id = user_id(&caller)?;
    let goal_id = goal_id(&id)?;

    let Some(pool) = state.pool.as_ref() else {
        return Err(no_database());
    };

    if !goal::user_has_access(pool, user_id, goal_id).await? {
        return Err(AppError::Forbidden("Access denied".to_string()));
    }

    let taggar = tag::set_goal_tags(pool, user_id, goal_id, &req.tags).await?;

    Ok(Json(taggar))
}

/// Tar bort en tagg från ett mål.
#[utoipa::path(
    delete,
    path = "/api/v1/goals/{id}/tags/{name}",
    tag = "tags",
    params(
        ("id" = i64, Path, description = "Målets id"),
        ("name" = String, Path, description = "Taggnamnet")
    ),
    responses(
        (status = 200, description = "Taggarna", content_type = "application/json",
            body = Vec<String>),
        (status = 400, description = "Ogiltigt taggnamn"),
        (status = 403, description = "Ingen åtkomst till målet")
    )
)]
pub async fn delete_goal_tag(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path((id, name)): Path<(String, String)>,
) -> Result<Json<Vec<String>>, AppError> {
    let user_id = user_id(&caller)?;
    let goal_id = goal_id(&id)?;

    let Some(pool) = state.pool.as_ref() else {
        return Err(no_database());
    };

    if !goal::user_has_access(pool, user_id, goal_id).await? {
        return Err(AppError::Forbidden("Access denied".to_string()));
    }

    let taggar = tag::delete_goal_tag(pool, user_id, goal_id, &name).await?;

    Ok(Json(taggar))
}
