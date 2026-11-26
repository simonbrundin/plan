//! Platser.
//!
//! Port av `src/api/internal/handlers/places.go`. Felmeddelandena är
//! desamma som i Go - gränssnittet visar dem direkt.

use axum::body::Bytes;
use axum::extract::{Path, Query, State};
use axum::http::StatusCode;
use axum::response::{IntoResponse, Response};
use axum::Extension;
use axum::Json;
use serde::Deserialize;
use serde_json::{json, Value};

use crate::auth::Caller;
use crate::error::AppError;
use crate::models::place::{
    CreatePlaceRequest, Place, SetCurrentPlaceRequest, SetPlacesRequest, UpdatePlaceRequest,
    DEFAULT_PLACE_RADIUS_METERS,
};
use crate::repository::{goal, place};
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

/// Målets id ur sökvägen.
fn goal_id(id: &str) -> Result<i64, AppError> {
    id.parse::<i64>()
        .map_err(|_| AppError::BadRequest("Invalid goal ID".to_string()))
}

/// Platsens id ur sökvägen.
fn place_id(id: &str) -> Result<i64, AppError> {
    id.parse::<i64>()
        .map_err(|_| AppError::BadRequest("Invalid place ID".to_string()))
}

/// Svaret för "jag är här"-platsen.
///
/// Används en struct i stället för `json!` - annars sorterar serde_json
/// nycklarna, medan Go behåller fältens ordning.
#[derive(serde::Serialize)]
pub struct CurrentPlaceResponse {
    /// Den valda platsen, eller null.
    pub place: Option<Place>,
}

/// Listar användarens platser.
#[utoipa::path(
    get,
    path = "/api/v1/places",
    tag = "places",
    responses(
        (status = 200, description = "Platserna", body = Vec<Place>),
        (status = 401, description = "Ingen giltig session")
    )
)]
pub async fn list(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
) -> Result<Json<Vec<Place>>, AppError> {
    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;

    Ok(Json(place::list(pool, user_id).await?))
}

/// Skapar en plats.
#[utoipa::path(
    post,
    path = "/api/v1/places",
    tag = "places",
    request_body = CreatePlaceRequest,
    responses(
        (status = 201, description = "Platsen", body = Place),
        (status = 400, description = "Ogiltiga värden"),
        (status = 401, description = "Ingen giltig session")
    )
)]
pub async fn create(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    body: Bytes,
) -> Result<Response, AppError> {
    let req: CreatePlaceRequest =
        crate::routes::parse_body(&body, "Name, latitude and longitude are required")?;

    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;

    // Standardvärdena sätts här, som i Go - annars skulle en plats utan
    // angiven radie få radien noll och avvisas i valideringen.
    let radie = req.radius_meters.unwrap_or(DEFAULT_PLACE_RADIUS_METERS);
    let ikon = req.icon.clone().unwrap_or_default();

    let plats = Place {
        id: 0,
        name: req.name.clone(),
        latitude: req.latitude,
        longitude: req.longitude,
        radius_meters: radie,
        icon: ikon.clone(),
        created_at: chrono::Utc::now(),
        updated_at: chrono::Utc::now(),
    };
    plats.validate().map_err(AppError::BadRequest)?;

    let skapad = place::create(pool, user_id, &req).await?;

    Ok((StatusCode::CREATED, Json(skapad)).into_response())
}

/// Uppdaterar en plats.
#[utoipa::path(
    patch,
    path = "/api/v1/places/{id}",
    tag = "places",
    params(("id" = i64, Path, description = "Platsens id")),
    request_body = UpdatePlaceRequest,
    responses(
        (status = 200, description = "Platsen", body = Place),
        (status = 400, description = "Ogiltiga värden"),
        (status = 401, description = "Ingen giltig session"),
        (status = 404, description = "Place not found")
    )
)]
pub async fn update(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
    body: Bytes,
) -> Result<Json<Place>, AppError> {
    let req: UpdatePlaceRequest = crate::routes::parse_body(&body, "No fields to update")?;

    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;
    let place_id = place_id(&id)?;

    let Some(existerande) = place::get(pool, user_id, place_id).await? else {
        return Err(AppError::NotFound("Place not found".to_string()));
    };

    // Bara det som ändras valideras - oförändrade värden tas från den
    // befintliga platsen.
    let mut uppdaterad = existerande.clone();
    if let Some(namn) = &req.name {
        uppdaterad.name = namn.clone();
    }
    if let Some(latitud) = req.latitude {
        uppdaterad.latitude = latitud;
    }
    if let Some(longitud) = req.longitude {
        uppdaterad.longitude = longitud;
    }
    if let Some(radie) = req.radius_meters {
        uppdaterad.radius_meters = radie;
    }
    if let Some(ikon) = &req.icon {
        uppdaterad.icon = ikon.clone();
    }
    uppdaterad.validate().map_err(AppError::BadRequest)?;

    let plats = place::update(pool, user_id, place_id, &req)
        .await?
        .ok_or_else(|| AppError::NotFound("Place not found".to_string()))?;

    Ok(Json(plats))
}

/// Tar bort en plats.
#[utoipa::path(
    delete,
    path = "/api/v1/places/{id}",
    tag = "places",
    params(("id" = i64, Path, description = "Platsens id")),
    responses(
        (status = 200, description = "Borttagen"),
        (status = 401, description = "Ingen giltig session"),
        (status = 404, description = "Place not found")
    )
)]
pub async fn delete(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
) -> Result<Json<Value>, AppError> {
    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;
    let place_id = place_id(&id)?;

    if !place::delete(pool, user_id, place_id).await? {
        return Err(AppError::NotFound("Place not found".to_string()));
    }

    Ok(Json(json!({ "success": true, "deleted": place_id })))
}

/// Parametrarna för "platser nära mig".
#[derive(Debug, Deserialize)]
pub struct NearbyQuery {
    /// Latituden som söks.
    pub lat: Option<String>,
    /// Longituden som söks.
    pub lon: Option<String>,
}

/// Listar de platser vars radie innehåller en position.
#[utoipa::path(
    get,
    path = "/api/v1/places/nearby",
    tag = "places",
    params(
        ("lat" = String, Query, description = "Latitud"),
        ("lon" = String, Query, description = "Longitud")
    ),
    responses(
        (status = 200, description = "Platserna nära positionen", body = Vec<Place>),
        (status = 400, description = "lat eller lon är inte ett tal"),
        (status = 401, description = "Ingen giltig session")
    )
)]
pub async fn nearby(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Query(query): Query<NearbyQuery>,
) -> Result<Json<Vec<Place>>, AppError> {
    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;

    // Go svarar med ett eget felmeddelande per koordinat.
    let lat = query
        .lat
        .as_deref()
        .unwrap_or_default()
        .parse::<f64>()
        .map_err(|_| AppError::BadRequest("lat must be a number".to_string()))?;
    let lon = query
        .lon
        .as_deref()
        .unwrap_or_default()
        .parse::<f64>()
        .map_err(|_| AppError::BadRequest("lon must be a number".to_string()))?;

    Ok(Json(
        place::get_nearby_places(pool, user_id, lat, lon).await?,
    ))
}

/// Listar ett måls platser.
#[utoipa::path(
    get,
    path = "/api/v1/goals/{id}/places",
    tag = "places",
    params(("id" = i64, Path, description = "Målets id")),
    responses(
        (status = 200, description = "Platserna", body = Vec<Place>),
        (status = 400, description = "Invalid goal ID"),
        (status = 401, description = "Ingen giltig session"),
        (status = 403, description = "Access denied")
    )
)]
pub async fn get_goal_places(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
) -> Result<Json<Vec<Place>>, AppError> {
    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;
    let goal_id = goal_id(&id)?;

    if !goal::user_has_access(pool, user_id, goal_id).await? {
        return Err(AppError::Forbidden("Access denied".to_string()));
    }

    Ok(Json(place::get_goal_places(pool, goal_id).await?))
}

/// Ersätter ett måls platser.
#[utoipa::path(
    put,
    path = "/api/v1/goals/{id}/places",
    tag = "places",
    params(("id" = i64, Path, description = "Målets id")),
    request_body = SetPlacesRequest,
    responses(
        (status = 200, description = "Platserna", body = Vec<Place>),
        (status = 400, description = "place_ids is required"),
        (status = 401, description = "Ingen giltig session"),
        (status = 403, description = "Access denied")
    )
)]
pub async fn set_goal_places(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
    body: Bytes,
) -> Result<Json<Vec<Place>>, AppError> {
    let req: SetPlacesRequest = crate::routes::parse_body(&body, "place_ids is required")?;

    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;
    let goal_id = goal_id(&id)?;

    if !goal::user_has_access(pool, user_id, goal_id).await? {
        return Err(AppError::Forbidden("Access denied".to_string()));
    }

    Ok(Json(
        place::set_places(
            pool,
            user_id,
            goal_id,
            &req.place_ids,
            "goal_places",
            "goal_id",
        )
        .await?,
    ))
}

/// Listar en kalenderhändelses platser.
#[utoipa::path(
    get,
    path = "/api/v1/calendar-events/{id}/places",
    tag = "places",
    params(("id" = i64, Path, description = "Händelsens id")),
    responses(
        (status = 200, description = "Platserna", body = Vec<Place>),
        (status = 400, description = "Invalid event ID"),
        (status = 401, description = "Ingen giltig session"),
        (status = 404, description = "Event not found")
    )
)]
pub async fn get_event_places(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
) -> Result<Json<Vec<Place>>, AppError> {
    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;
    let event_id = event_id(&id)?;

    if !event_exists(pool, user_id, event_id).await? {
        return Err(AppError::NotFound("Event not found".to_string()));
    }

    Ok(Json(place::get_event_places(pool, event_id).await?))
}

/// Ersätter en kalenderhändelses platser.
#[utoipa::path(
    put,
    path = "/api/v1/calendar-events/{id}/places",
    tag = "places",
    params(("id" = i64, Path, description = "Händelsens id")),
    request_body = SetPlacesRequest,
    responses(
        (status = 200, description = "Platserna", body = Vec<Place>),
        (status = 400, description = "place_ids is required"),
        (status = 401, description = "Ingen giltig session"),
        (status = 404, description = "Event not found")
    )
)]
pub async fn set_event_places(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
    body: Bytes,
) -> Result<Json<Vec<Place>>, AppError> {
    let req: SetPlacesRequest = crate::routes::parse_body(&body, "place_ids is required")?;

    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;
    let event_id = event_id(&id)?;

    if !event_exists(pool, user_id, event_id).await? {
        return Err(AppError::NotFound("Event not found".to_string()));
    }

    Ok(Json(
        place::set_places(
            pool,
            user_id,
            event_id,
            &req.place_ids,
            "calendar_event_places",
            "event_id",
        )
        .await?,
    ))
}

/// Händelsens id ur sökvägen.
fn event_id(id: &str) -> Result<i64, AppError> {
    id.parse::<i64>()
        .map_err(|_| AppError::BadRequest("Invalid event ID".to_string()))
}

/// Finns händelsen, och tillhör den användaren?
async fn event_exists(pool: &sqlx::PgPool, user_id: i64, event_id: i64) -> Result<bool, AppError> {
    let finns: bool = sqlx::query_scalar(
        "SELECT EXISTS(SELECT 1 FROM calendar_events WHERE id = $1 AND user_id = $2)",
    )
    .bind(event_id)
    .bind(user_id)
    .fetch_one(pool)
    .await
    .map_err(|err| AppError::Internal(err.to_string()))?;

    Ok(finns)
}

/// Användarens valda "jag är här"-plats.
#[utoipa::path(
    get,
    path = "/api/v1/me/current-place",
    tag = "places",
    responses(
        (status = 200, description = "Den valda platsen, eller null"),
        (status = 401, description = "Ingen giltig session")
    )
)]
pub async fn get_current_place(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
) -> Result<Json<CurrentPlaceResponse>, AppError> {
    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;

    let plats = place::get_current_place(pool, user_id).await?;

    Ok(Json(CurrentPlaceResponse { place: plats }))
}

/// Sätter användarens "jag är här"-plats.
#[utoipa::path(
    put,
    path = "/api/v1/me/current-place",
    tag = "places",
    request_body = SetCurrentPlaceRequest,
    responses(
        (status = 200, description = "Den valda platsen, eller null"),
        (status = 400, description = "place_id is required eller platsen finns inte"),
        (status = 401, description = "Ingen giltig session")
    )
)]
pub async fn set_current_place(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    body: Bytes,
) -> Result<Json<CurrentPlaceResponse>, AppError> {
    let req: SetCurrentPlaceRequest = crate::routes::parse_body(&body, "place_id is required")?;

    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;

    if let Some(place_id) = req.place_id {
        if place::get(pool, user_id, place_id).await?.is_none() {
            return Err(AppError::BadRequest("Place not found".to_string()));
        }
    }

    place::set_current_place(pool, user_id, req.place_id).await?;

    let plats = place::get_current_place(pool, user_id).await?;

    Ok(Json(CurrentPlaceResponse { place: plats }))
}
