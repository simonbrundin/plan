//! Kalendrar.
//!
//! Port av `src/api/internal/handlers/calendars.go`.

use axum::body::Bytes;
use axum::extract::{Path, State};
use axum::http::StatusCode;
use axum::response::{IntoResponse, Response};
use axum::Extension;
use axum::Json;

use crate::auth::Caller;
use crate::error::AppError;
use crate::models::calendar::{Calendar, CreateCalendarRequest, UpdateCalendarRequest};
use crate::repository::calendar;
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

/// Kalenderns id ur sökvägen.
fn calendar_id(id: &str) -> Result<i64, AppError> {
    // Go avvisar also noll och negativa id:n.
    let värde = id
        .parse::<i64>()
        .map_err(|_| AppError::BadRequest("Invalid calendar ID".to_string()))?;

    if värde <= 0 {
        return Err(AppError::BadRequest("Invalid calendar ID".to_string()));
    }

    Ok(värde)
}

/// Ogiltigt värde i en validering.
///
/// Egen typ i stället för `()` - så att anropstället vet vilket fält det
/// var och kan ge rätt meddelande.
#[derive(Debug, Clone, Copy)]
pub struct OgiltigtVärde;

/// Kontrollerar att namnet är användbart.
///
/// `None` betyder "sätt inte namnet".
pub fn validate_calendar_name(name: Option<&str>) -> Result<Option<String>, OgiltigtVärde> {
    let Some(name) = name else {
        return Ok(None);
    };

    let trimmed = name.trim();
    if trimmed.is_empty() || trimmed.chars().count() > 64 {
        return Err(OgiltigtVärde);
    }

    Ok(Some(trimmed.to_string()))
}

/// Kontrollerar att färgen är en hexkod och normaliserar den.
///
/// "#abc" skrivs ut som "#aabbcc", så att alltid sex siffror lagras.
/// `None` betyder "behåll nuvarande värde".
pub fn validate_color(color: Option<&str>) -> Result<Option<String>, OgiltigtVärde> {
    let Some(color) = color else {
        return Ok(None);
    };

    let trimmed = color.trim();
    if trimmed.is_empty() {
        return Ok(Some(String::new()));
    }

    let Some(digits) = trimmed.strip_prefix('#') else {
        return Err(OgiltigtVärde);
    };

    let digits = digits.to_lowercase();
    if !digits.chars().all(|c| c.is_ascii_hexdigit()) {
        return Err(OgiltigtVärde);
    }

    match digits.len() {
        3 => {
            let mut utvidgad = String::with_capacity(6);
            for c in digits.chars() {
                utvidgad.push(c);
                utvidgad.push(c);
            }
            Ok(Some(format!("#{utvidgad}")))
        }
        6 => Ok(Some(format!("#{digits}"))),
        _ => Err(OgiltigtVärde),
    }
}

/// Kontrollerar att ikonen är användbar.
///
/// `None` betyder "behåll nuvarande värde".
pub fn validate_icon(icon: Option<&str>) -> Result<Option<String>, OgiltigtVärde> {
    let Some(icon) = icon else {
        return Ok(None);
    };

    let trimmed = icon.trim();
    if trimmed.chars().count() > 100 {
        return Err(OgiltigtVärde);
    }

    Ok(Some(trimmed.to_string()))
}

/// Listar användarens kalendrar.
#[utoipa::path(
    get,
    path = "/api/v1/calendars",
    tag = "calendars",
    responses(
        (status = 200, description = "Kalendrarna", body = Vec<Calendar>),
        (status = 401, description = "Ingen giltig session")
    )
)]
pub async fn list(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
) -> Result<Json<Vec<Calendar>>, AppError> {
    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;

    Ok(Json(calendar::list(pool, user_id).await?))
}

/// Skapar en kalender.
#[utoipa::path(
    post,
    path = "/api/v1/calendars",
    tag = "calendars",
    request_body = CreateCalendarRequest,
    responses(
        (status = 201, description = "Kalendern", body = Calendar),
        (status = 400, description = "Ogiltigt namn, färg eller ikon"),
        (status = 401, description = "Ingen giltig session")
    )
)]
pub async fn create(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    body: Bytes,
) -> Result<Response, AppError> {
    let mut req: CreateCalendarRequest = crate::routes::parse_body(&body, "name is required")?;

    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;

    let namn = validate_calendar_name(Some(&req.name)).map_err(|_| {
        AppError::BadRequest("Calendar name must be between 1 and 64 characters".to_string())
    })?;
    req.name = namn.unwrap_or_default();

    let färg = validate_color(req.color.as_deref())
        .map_err(|_| AppError::BadRequest("Invalid color".to_string()))?;
    req.color = färg;

    let ikon = validate_icon(req.icon.as_deref())
        .map_err(|_| AppError::BadRequest("Invalid icon".to_string()))?;
    req.icon = ikon;

    let kalender = calendar::create(pool, user_id, &req).await?;

    Ok((StatusCode::CREATED, Json(kalender)).into_response())
}

/// Uppdaterar en kalender.
#[utoipa::path(
    patch,
    path = "/api/v1/calendars/{id}",
    tag = "calendars",
    params(("id" = i64, Path, description = "Kalenderns id")),
    request_body = UpdateCalendarRequest,
    responses(
        (status = 200, description = "Kalendern", body = Calendar),
        (status = 400, description = "Ogiltigt namn, färg eller ikon"),
        (status = 401, description = "Ingen giltig session"),
        (status = 404, description = "Calendar not found")
    )
)]
pub async fn update(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
    body: Bytes,
) -> Result<Json<Calendar>, AppError> {
    let mut req: UpdateCalendarRequest = crate::routes::parse_body(&body, "No fields to update")?;

    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;
    let calendar_id = calendar_id(&id)?;

    let namn = validate_calendar_name(req.name.as_deref()).map_err(|_| {
        AppError::BadRequest("Calendar name must be between 1 and 64 characters".to_string())
    })?;
    req.name = namn;

    let färg = validate_color(req.color.as_deref())
        .map_err(|_| AppError::BadRequest("Invalid color".to_string()))?;
    req.color = färg;

    let ikon = validate_icon(req.icon.as_deref())
        .map_err(|_| AppError::BadRequest("Invalid icon".to_string()))?;
    req.icon = ikon;

    let kalender = calendar::update(pool, user_id, calendar_id, &req)
        .await?
        .ok_or_else(|| AppError::NotFound("Calendar not found".to_string()))?;

    Ok(Json(kalender))
}

/// Tar bort en kalender.
///
/// Händelserna som legat i den behålls men blir kalenderlösa.
#[utoipa::path(
    delete,
    path = "/api/v1/calendars/{id}",
    tag = "calendars",
    params(("id" = i64, Path, description = "Kalenderns id")),
    responses(
        (status = 204, description = "Borttagen"),
        (status = 401, description = "Ingen giltig session"),
        (status = 404, description = "Calendar not found")
    )
)]
pub async fn delete(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
) -> Result<StatusCode, AppError> {
    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;
    let calendar_id = calendar_id(&id)?;

    if !calendar::delete(pool, user_id, calendar_id).await? {
        return Err(AppError::NotFound("Calendar not found".to_string()));
    }

    Ok(StatusCode::NO_CONTENT)
}
