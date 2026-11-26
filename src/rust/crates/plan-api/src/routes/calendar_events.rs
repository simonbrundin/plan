//! Kalenderhändelser.
//!
//! Port av `src/api/internal/handlers/calendar_events.go`.

use axum::body::Bytes;
use axum::extract::{Path, Query, State};
use axum::http::StatusCode;
use axum::response::{IntoResponse, Response};
use axum::Extension;
use axum::Json;
use chrono::{DateTime, Utc};

use crate::auth::Caller;
use crate::error::AppError;
use crate::models::calendar_event::{
    CalendarEvent, CreateCalendarEventRequest, CreateCalendarOccurrenceRequest,
    SplitCalendarEventRequest, UpdateCalendarEventRequest,
};
use crate::repository::calendar_event;
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

/// Händelsens id ur sökvägen.
fn event_id(id: &str) -> Result<i64, AppError> {
    let värde = id
        .parse::<i64>()
        .map_err(|_| AppError::BadRequest("Invalid calendar event ID".to_string()))?;

    if värde <= 0 {
        return Err(AppError::BadRequest(
            "Invalid calendar event ID".to_string(),
        ));
    }

    Ok(värde)
}

/// Tolkar en tid i RFC3339-format.
fn tolka_tid(värde: &str) -> Result<DateTime<Utc>, AppError> {
    DateTime::parse_from_rfc3339(värde)
        .map(|tid| tid.with_timezone(&Utc))
        .map_err(|_| AppError::BadRequest("must be an RFC3339 timestamp".to_string()))
}

/// Väljer vilken ikon en händelse ska få.
///
/// Om begäran inte innehåller någon ikon ärvs förälderns ikon, vilket
/// gör att en override eller en delad serie inte tappar ikon av sig.
/// Ett medvetet tomt värde däremot tömmer ikonen.
pub fn resolve_icon(förälder_ikon: &str, begär_ikon: Option<&str>) -> Option<String> {
    if let Some(ikon) = begär_ikon {
        return Some(ikon.to_string());
    }

    if förälder_ikon.is_empty() {
        return None;
    }

    Some(förälder_ikon.to_string())
}

/// Kontrollerar att ikonen är användbar.
///
/// `None` betyder "sätt inte ikonen".
pub fn validate_icon(icon: Option<&str>) -> Result<Option<String>, AppError> {
    let Some(ikon) = icon else {
        return Ok(None);
    };

    let trimmed = ikon.trim();
    if trimmed.chars().count() > 100 {
        return Err(AppError::BadRequest(
            "icon must be at most 100 characters".to_string(),
        ));
    }

    Ok(Some(trimmed.to_string()))
}

/// Kontrollerar att upprepningsregeln är giltig.
fn validate_recurrence(regel: &str) -> Result<(), AppError> {
    if plan_recurrence::Regel::validera(regel).is_ok() {
        Ok(())
    } else {
        Err(AppError::BadRequest("Invalid recurrence_rule".to_string()))
    }
}

/// Kontrollerar att statusen är en av de tillåtna.
fn validate_status(status: &str) -> Result<(), AppError> {
    match status {
        "" | "confirmed" | "tentative" | "cancelled" => Ok(()),
        _ => Err(AppError::BadRequest("Invalid status".to_string())),
    }
}

/// Kontrollerar att användaren har målet.
async fn check_goal(
    pool: &sqlx::PgPool,
    user_id: i64,
    goal_id: Option<i64>,
) -> Result<(), AppError> {
    let Some(goal_id) = goal_id else {
        return Ok(());
    };

    if !calendar_event::user_has_goal(pool, user_id, goal_id).await? {
        return Err(AppError::Forbidden("Goal access denied".to_string()));
    }

    Ok(())
}

/// Kontrollerar att angiven kalender finns och tillhör användaren.
///
/// Värdet 0 betyder "ingen kalender" och är tillåtet, eftersom det
/// används för att ta bort en kalenderkoppling.
async fn check_calendar(
    pool: &sqlx::PgPool,
    user_id: i64,
    calendar_id: Option<i64>,
) -> Result<(), AppError> {
    let Some(calendar_id) = calendar_id else {
        return Ok(());
    };

    if calendar_id == 0 {
        return Ok(());
    }

    if !calendar_event::user_has_calendar(pool, user_id, calendar_id).await? {
        return Err(AppError::Forbidden("Calendar access denied".to_string()));
    }

    Ok(())
}

/// Listar händelser.
///
/// Utan tidsfönster returneras serier oexpanderade - klienten expanderar
/// dem själv. Med tidsfönster expanderas serierna till förekomster.
#[utoipa::path(
    get,
    path = "/api/v1/calendar-events",
    tag = "calendar-events",
    params(
        ("scope" = Option<String>, Query, description = "all returns every event without expanding recurring series"),
        ("from" = Option<String>, Query, description = "Start of the window (RFC3339)"),
        ("to" = Option<String>, Query, description = "End of the window (RFC3339)")
    ),
    responses(
        (status = 200, description = "Händelserna", body = Vec<CalendarEvent>),
        (status = 400, description = "Ogiltigt tidsfönster"),
        (status = 401, description = "Ingen giltig session")
    )
)]
pub async fn list(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Query(params): Query<std::collections::HashMap<String, String>>,
) -> Result<Json<Vec<CalendarEvent>>, AppError> {
    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;

    let scope = params.get("scope").map(String::as_str).unwrap_or("");

    let mut händelser = match scope {
        "all" => calendar_event::list_all(pool, user_id).await?,
        "" => {
            let från = tolka_tid(params.get("from").map(String::as_str).ok_or_else(|| {
                AppError::BadRequest("from must be an RFC3339 timestamp".to_string())
            })?)
            .map_err(|err| AppError::BadRequest(format!("from {err}")))?;

            let till = tolka_tid(params.get("to").map(String::as_str).ok_or_else(|| {
                AppError::BadRequest("to must be an RFC3339 timestamp after from".to_string())
            })?)
            .map_err(|err| AppError::BadRequest(format!("to {err}")))?;

            if till <= från {
                return Err(AppError::BadRequest(
                    "to must be an RFC3339 timestamp after from".to_string(),
                ));
            }

            calendar_event::list(pool, user_id, från, till).await?
        }
        _ => {
            return Err(AppError::BadRequest("scope must be 'all'".to_string()));
        }
    };

    calendar_event::fyll_platser(pool, &mut händelser).await?;

    Ok(Json(händelser))
}

/// Hämtar en händelse.
#[utoipa::path(
    get,
    path = "/api/v1/calendar-events/{id}",
    tag = "calendar-events",
    params(("id" = i64, Path, description = "Händelsens id")),
    responses(
        (status = 200, description = "Händelsen", body = CalendarEvent),
        (status = 400, description = "Ogiltigt id"),
        (status = 401, description = "Ingen giltig session"),
        (status = 404, description = "Calendar event not found")
    )
)]
pub async fn get(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
) -> Result<Json<CalendarEvent>, AppError> {
    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;
    let event_id = event_id(&id)?;

    let Some(mut händelse) = calendar_event::get(pool, user_id, event_id).await? else {
        return Err(AppError::NotFound("Calendar event not found".to_string()));
    };

    let mut händelser = vec![händelse.clone()];
    calendar_event::fyll_platser(pool, &mut händelser).await?;
    händelse = händelser.remove(0);

    Ok(Json(händelse))
}

/// Skapar en händelse.
#[utoipa::path(
    post,
    path = "/api/v1/calendar-events",
    tag = "calendar-events",
    request_body = CreateCalendarEventRequest,
    responses(
        (status = 201, description = "Händelsen", body = CalendarEvent),
        (status = 400, description = "Ogiltig händelse"),
        (status = 401, description = "Ingen giltig session"),
        (status = 403, description = "Ke tillgång till målet eller kalendern")
    )
)]
pub async fn create(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    body: Bytes,
) -> Result<Response, AppError> {
    let req: CreateCalendarEventRequest =
        crate::routes::parse_body(&body, "title, start_at and end_at are required")?;

    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;

    let titel = req
        .title
        .as_deref()
        .filter(|titel| !titel.is_empty())
        .ok_or_else(|| {
            AppError::BadRequest("title, start_at and end_at are required".to_string())
        })?;
    let start_at_text = req
        .start_at
        .as_deref()
        .filter(|värde| !värde.is_empty())
        .ok_or_else(|| {
            AppError::BadRequest("title, start_at and end_at are required".to_string())
        })?;
    let end_at_text = req
        .end_at
        .as_deref()
        .filter(|värde| !värde.is_empty())
        .ok_or_else(|| {
            AppError::BadRequest("title, start_at and end_at are required".to_string())
        })?;

    let start_at = DateTime::parse_from_rfc3339(start_at_text)
        .map(|tid| tid.with_timezone(&Utc))
        .map_err(|_| AppError::BadRequest("start_at must be an RFC3339 timestamp".to_string()))?;
    let end_at = DateTime::parse_from_rfc3339(end_at_text)
        .map(|tid| tid.with_timezone(&Utc))
        .map_err(|_| {
            AppError::BadRequest("end_at must be an RFC3339 timestamp after start_at".to_string())
        })?;

    if end_at <= start_at {
        return Err(AppError::BadRequest(
            "end_at must be an RFC3339 timestamp after start_at".to_string(),
        ));
    }

    check_goal(pool, user_id, req.goal_id).await?;
    check_calendar(pool, user_id, req.calendar_id).await?;

    if let Some(regel) = req.recurrence_rule.as_deref() {
        validate_recurrence(regel)?;
    }

    let ikon = validate_icon(req.icon.as_deref())?;

    // Kalendern väljs här - "ingen kalender" betyder noll, och att ingen
    // kalender anges betyder standardkalendern.
    let kalender_id = match req.calendar_id {
        Some(0) => None,
        Some(id) => Some(id),
        None => crate::repository::calendar::default_calendar_id(pool, user_id).await?,
    };

    let mut händelse = calendar_event::create(
        pool,
        user_id,
        &crate::models::calendar_event::CreateCalendarEventRequest {
            title: Some(titel.to_string()),
            description: req.description.clone(),
            icon: ikon.clone(),
            start_at: req.start_at.clone(),
            end_at: req.end_at.clone(),
            time_zone: req.time_zone.clone(),
            all_day: req.all_day,
            status: req.status.clone(),
            recurrence_rule: req.recurrence_rule.clone(),
            goal_id: req.goal_id,
            calendar_id: kalender_id,
        },
        start_at,
        end_at,
    )
    .await?;

    let mut händelser = vec![händelse.clone()];
    calendar_event::fyll_platser(pool, &mut händelser).await?;
    händelse = händelser.remove(0);

    Ok((StatusCode::CREATED, Json(händelse)).into_response())
}

/// Uppdaterar en händelse.
#[utoipa::path(
    patch,
    path = "/api/v1/calendar-events/{id}",
    tag = "calendar-events",
    params(("id" = i64, Path, description = "Händelsens id")),
    request_body = UpdateCalendarEventRequest,
    responses(
        (status = 200, description = "Händelsen", body = CalendarEvent),
        (status = 400, description = "Ogiltig händelse"),
        (status = 401, description = "Ingen giltig session"),
        (status = 403, description = "Ke tillgång till målet eller kalendern"),
        (status = 404, description = "Calendar event not found")
    )
)]
pub async fn update(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
    body: Bytes,
) -> Result<Json<CalendarEvent>, AppError> {
    let mut req: UpdateCalendarEventRequest =
        crate::routes::parse_body(&body, "Invalid calendar event")?;

    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;
    let event_id = event_id(&id)?;

    let mut start_at = None;
    if let Some(värde) = req.start_at.as_deref() {
        start_at = Some(
            DateTime::parse_from_rfc3339(värde)
                .map(|tid| tid.with_timezone(&Utc))
                .map_err(|_| {
                    AppError::BadRequest("start_at must be an RFC3339 timestamp".to_string())
                })?,
        );
    }

    let mut end_at = None;
    if let Some(värde) = req.end_at.as_deref() {
        end_at = Some(
            DateTime::parse_from_rfc3339(värde)
                .map(|tid| tid.with_timezone(&Utc))
                .map_err(|_| {
                    AppError::BadRequest("end_at must be an RFC3339 timestamp".to_string())
                })?,
        );
    }

    if let (Some(start), Some(end)) = (start_at, end_at) {
        if end <= start {
            return Err(AppError::BadRequest(
                "end_at must be after start_at".to_string(),
            ));
        }
    }

    // Bara ena tiden ändrad - den andra hämtas från den befintliga
    // händelsen, så att tidslängden kan kontrolleras.
    if start_at.is_some() || end_at.is_some() {
        let Some(existerande) = calendar_event::get(pool, user_id, event_id).await? else {
            return Err(AppError::NotFound("Calendar event not found".to_string()));
        };

        let start = start_at.unwrap_or(existerande.start_at);
        let end = end_at.unwrap_or(existerande.end_at);
        if end <= start {
            return Err(AppError::BadRequest(
                "end_at must be after start_at".to_string(),
            ));
        }
    }

    check_goal(pool, user_id, req.goal_id).await?;
    check_calendar(pool, user_id, req.calendar_id).await?;

    if let Some(regel) = req.recurrence_rule.as_deref() {
        validate_recurrence(regel)?;
    }

    // Ikonen valideras och sätts tillbaka - den kan vara en tom sträng,
    // vilket betyder "töm ikonen".
    req.icon = validate_icon(req.icon.as_deref())?;

    let Some(mut händelse) =
        calendar_event::update(pool, user_id, event_id, &req, start_at, end_at).await?
    else {
        return Err(AppError::NotFound("Calendar event not found".to_string()));
    };

    let mut händelser = vec![händelse.clone()];
    calendar_event::fyll_platser(pool, &mut händelser).await?;
    händelse = händelser.remove(0);

    Ok(Json(händelse))
}

/// Tar bort en händelse.
#[utoipa::path(
    delete,
    path = "/api/v1/calendar-events/{id}",
    tag = "calendar-events",
    params(("id" = i64, Path, description = "Händelsens id")),
    responses(
        (status = 204, description = "Borttagen"),
        (status = 400, description = "Ogiltigt id"),
        (status = 401, description = "Ingen giltig session"),
        (status = 404, description = "Calendar event not found")
    )
)]
pub async fn delete(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
) -> Result<StatusCode, AppError> {
    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;
    let event_id = event_id(&id)?;

    if !calendar_event::delete(pool, user_id, event_id).await? {
        return Err(AppError::NotFound("Calendar event not found".to_string()));
    }

    Ok(StatusCode::NO_CONTENT)
}

/// Skapar en override för en enskild förekomst i en serie.
///
/// Fält som inte anges ärvs från förälderserien.
#[utoipa::path(
    post,
    path = "/api/v1/calendar-events/{id}/occurrences",
    tag = "calendar-events",
    params(("id" = i64, Path, description = "Förälderseriens id")),
    request_body = CreateCalendarOccurrenceRequest,
    responses(
        (status = 201, description = "Förekomsten", body = CalendarEvent),
        (status = 400, description = "Ogiltig förekomst"),
        (status = 401, description = "Ingen giltig session"),
        (status = 403, description = "Ke tillgång till målet eller kalendern"),
        (status = 404, description = "Recurring calendar event not found")
    )
)]
pub async fn create_occurrence_override(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
    body: Bytes,
) -> Result<Response, AppError> {
    let req: CreateCalendarOccurrenceRequest =
        crate::routes::parse_body(&body, "original_start_at, start_at and end_at are required")?;

    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;
    let parent_id = event_id(&id)?;

    let original_start_text = req
        .original_start_at
        .as_deref()
        .filter(|värde| !värde.is_empty())
        .ok_or_else(|| {
            AppError::BadRequest("original_start_at, start_at and end_at are required".to_string())
        })?;
    let start_at_text = req
        .start_at
        .as_deref()
        .filter(|värde| !värde.is_empty())
        .ok_or_else(|| {
            AppError::BadRequest("original_start_at, start_at and end_at are required".to_string())
        })?;
    let end_at_text = req
        .end_at
        .as_deref()
        .filter(|värde| !värde.is_empty())
        .ok_or_else(|| {
            AppError::BadRequest("original_start_at, start_at and end_at are required".to_string())
        })?;

    let original_start = DateTime::parse_from_rfc3339(original_start_text)
        .map(|tid| tid.with_timezone(&Utc))
        .map_err(|_| {
            AppError::BadRequest("original_start_at must be an RFC3339 timestamp".to_string())
        })?;
    let start_at = DateTime::parse_from_rfc3339(start_at_text)
        .map(|tid| tid.with_timezone(&Utc))
        .map_err(|_| AppError::BadRequest("start_at must be an RFC3339 timestamp".to_string()))?;
    let end_at = DateTime::parse_from_rfc3339(end_at_text)
        .map(|tid| tid.with_timezone(&Utc))
        .map_err(|_| AppError::BadRequest("end_at must be after start_at".to_string()))?;

    if end_at <= start_at {
        return Err(AppError::BadRequest(
            "end_at must be after start_at".to_string(),
        ));
    }

    if let Some(status) = req.status.as_deref() {
        validate_status(status)?;
    }

    check_goal(pool, user_id, req.goal_id).await?;
    check_calendar(pool, user_id, req.calendar_id).await?;

    let ikon = validate_icon(req.icon.as_deref())?;

    let Some(händelse) = calendar_event::create_override(
        pool,
        user_id,
        parent_id,
        &CreateCalendarOccurrenceRequest {
            original_start_at: req.original_start_at.clone(),
            start_at: req.start_at.clone(),
            end_at: req.end_at.clone(),
            title: req.title.clone(),
            description: req.description.clone(),
            icon: ikon,
            time_zone: req.time_zone.clone(),
            all_day: req.all_day,
            status: req.status.clone(),
            goal_id: req.goal_id,
            calendar_id: req.calendar_id,
        },
        original_start,
        start_at,
        end_at,
    )
    .await?
    else {
        return Err(AppError::NotFound(
            "Recurring calendar event not found".to_string(),
        ));
    };

    Ok((StatusCode::CREATED, Json(händelse)).into_response())
}

/// Delar en återkommande serie.
///
/// Den gamla serien avslutas strax före den valda förekomsten och en ny
/// serie fortsätter därifrån.
#[utoipa::path(
    post,
    path = "/api/v1/calendar-events/{id}/split",
    tag = "calendar-events",
    params(("id" = i64, Path, description = "Förälderseriens id")),
    request_body = SplitCalendarEventRequest,
    responses(
        (status = 201, description = "Den nya serien", body = CalendarEvent),
        (status = 400, description = "Ogiltig delning"),
        (status = 401, description = "Ingen giltig session"),
        (status = 403, description = "Ke tillgång till målet eller kalendern"),
        (status = 404, description = "Recurring calendar event not found")
    )
)]
pub async fn split_series(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
    body: Bytes,
) -> Result<Response, AppError> {
    let req: SplitCalendarEventRequest = crate::routes::parse_body(
        &body,
        "original_start_at, start_at, end_at and recurrence_rule are required",
    )?;

    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;
    let parent_id = event_id(&id)?;

    let original_start_text = req
        .original_start_at
        .as_deref()
        .filter(|värde| !värde.is_empty())
        .ok_or_else(|| {
            AppError::BadRequest(
                "original_start_at, start_at, end_at and recurrence_rule are required".to_string(),
            )
        })?;
    let start_at_text = req
        .start_at
        .as_deref()
        .filter(|värde| !värde.is_empty())
        .ok_or_else(|| {
            AppError::BadRequest(
                "original_start_at, start_at, end_at and recurrence_rule are required".to_string(),
            )
        })?;
    let end_at_text = req
        .end_at
        .as_deref()
        .filter(|värde| !värde.is_empty())
        .ok_or_else(|| {
            AppError::BadRequest(
                "original_start_at, start_at, end_at and recurrence_rule are required".to_string(),
            )
        })?;
    let regel_text = req
        .recurrence_rule
        .as_deref()
        .filter(|värde| !värde.is_empty())
        .ok_or_else(|| {
            AppError::BadRequest(
                "original_start_at, start_at, end_at and recurrence_rule are required".to_string(),
            )
        })?;

    let original_start = DateTime::parse_from_rfc3339(original_start_text)
        .map(|tid| tid.with_timezone(&Utc))
        .map_err(|_| {
            AppError::BadRequest("original_start_at must be an RFC3339 timestamp".to_string())
        })?;
    let start_at = DateTime::parse_from_rfc3339(start_at_text)
        .map(|tid| tid.with_timezone(&Utc))
        .map_err(|_| AppError::BadRequest("start_at must be an RFC3339 timestamp".to_string()))?;
    let end_at = DateTime::parse_from_rfc3339(end_at_text)
        .map(|tid| tid.with_timezone(&Utc))
        .map_err(|_| AppError::BadRequest("end_at must be after start_at".to_string()))?;

    if end_at <= start_at {
        return Err(AppError::BadRequest(
            "end_at must be after start_at".to_string(),
        ));
    }

    validate_recurrence(regel_text)?;

    check_goal(pool, user_id, req.goal_id).await?;
    check_calendar(pool, user_id, req.calendar_id).await?;

    let ikon = validate_icon(req.icon.as_deref())?;

    let Some(händelse) = calendar_event::split_series(
        pool,
        user_id,
        parent_id,
        &SplitCalendarEventRequest {
            original_start_at: req.original_start_at.clone(),
            start_at: req.start_at.clone(),
            end_at: req.end_at.clone(),
            recurrence_rule: Some(regel_text.to_string()),
            title: req.title.clone(),
            description: req.description.clone(),
            icon: ikon,
            time_zone: req.time_zone.clone(),
            all_day: req.all_day,
            goal_id: req.goal_id,
            calendar_id: req.calendar_id,
        },
        original_start,
        start_at,
        end_at,
    )
    .await?
    else {
        return Err(AppError::NotFound(
            "Recurring calendar event not found".to_string(),
        ));
    };

    Ok((StatusCode::CREATED, Json(händelse)).into_response())
}
