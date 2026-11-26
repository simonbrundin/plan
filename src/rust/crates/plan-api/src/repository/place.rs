//! Platser.
//!
//! Port av `src/api/internal/repository/place_repository.go`. Frågorna är
//! identiska med Go-tjänstens - samma JOIN, samma sortering.
//!
//! Observera att `places.id` är BIGINT i schemat, till skillnad från
//! målens id:n som är INTEGER. Blanda inte ihop dem - en i64 som läses
//! ur en int4-kolumn ger ett tyst avkodningsfel.

use std::collections::HashMap;

use sqlx::{PgPool, Row};

use crate::error::AppError;
use crate::models::place::Place;

/// Kolumnerna som läses för en plats.
const PLATSKOLUMNER: &str = "p.id, p.name, p.latitude, p.longitude, p.radius_meters, \
     COALESCE(p.icon, ''), p.created_at, p.updated_at";

/// Läser en platsrad.
fn läs_plats(rad: &sqlx::postgres::PgRow) -> Place {
    Place {
        id: rad.get::<i64, _>(0),
        name: rad.get::<String, _>(1),
        latitude: rad.get::<f64, _>(2),
        longitude: rad.get::<f64, _>(3),
        radius_meters: rad.get::<i32, _>(4),
        icon: rad.get::<String, _>(5),
        created_at: rad.get::<chrono::DateTime<chrono::Utc>, _>(6),
        updated_at: rad.get::<chrono::DateTime<chrono::Utc>, _>(7),
    }
}

/// Alla platser för en användare, sorterade på namn.
pub async fn list(pool: &PgPool, user_id: i64) -> Result<Vec<Place>, AppError> {
    let fråga =
        format!("SELECT {PLATSKOLUMNER} FROM places p WHERE p.user_id = $1 ORDER BY p.name");

    let rader = sqlx::query(&fråga)
        .bind(user_id)
        .fetch_all(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to query places: {err}")))?;

    Ok(rader.iter().map(läs_plats).collect())
}

/// En enskild plats.
///
/// En plats som inte tillhör användaren är `None` - inte ett fel. Samma
/// mönster som övriga repositories.
pub async fn get(pool: &PgPool, user_id: i64, place_id: i64) -> Result<Option<Place>, AppError> {
    let fråga = format!("SELECT {PLATSKOLUMNER} FROM places p WHERE p.user_id = $1 AND p.id = $2");

    let rad = sqlx::query(&fråga)
        .bind(user_id)
        .bind(place_id)
        .fetch_optional(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to scan place: {err}")))?;

    Ok(rad.as_ref().map(läs_plats))
}

/// Skapar en ny plats.
///
/// Radien och ikonen får standardvärden om de inte anges - samma
/// hantering som i Go.
pub async fn create(
    pool: &PgPool,
    user_id: i64,
    req: &crate::models::place::CreatePlaceRequest,
) -> Result<Place, AppError> {
    let radie = req
        .radius_meters
        .unwrap_or(crate::models::place::DEFAULT_PLACE_RADIUS_METERS);
    let ikon = req.icon.clone().unwrap_or_default();

    let rad = sqlx::query(
        r#"
        INSERT INTO places (user_id, name, latitude, longitude, radius_meters, icon)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, name, latitude, longitude, radius_meters, COALESCE(icon, ''), created_at, updated_at
        "#,
    )
    .bind(user_id)
    .bind(&req.name)
    .bind(req.latitude)
    .bind(req.longitude)
    .bind(radie)
    .bind(ikon)
    .fetch_one(pool)
    .await
    .map_err(|err| AppError::Internal(format!("failed to create place: {err}")))?;

    Ok(läs_plats(&rad))
}

/// Uppdaterar en plats. Endast fält som sätts i requesten ändras.
pub async fn update(
    pool: &PgPool,
    user_id: i64,
    place_id: i64,
    req: &crate::models::place::UpdatePlaceRequest,
) -> Result<Option<Place>, AppError> {
    let rad = sqlx::query(
        r#"
        UPDATE places SET
            name = COALESCE($3, name),
            latitude = COALESCE($4, latitude),
            longitude = COALESCE($5, longitude),
            radius_meters = COALESCE($6, radius_meters),
            icon = COALESCE($7, icon),
            updated_at = CURRENT_TIMESTAMP
        WHERE user_id = $1 AND id = $2
        RETURNING id, name, latitude, longitude, radius_meters, COALESCE(icon, ''), created_at, updated_at
        "#,
    )
    .bind(user_id)
    .bind(place_id)
    .bind(req.name.as_deref())
    .bind(req.latitude)
    .bind(req.longitude)
    .bind(req.radius_meters)
    .bind(req.icon.as_deref())
    .fetch_optional(pool)
    .await
    .map_err(|err| AppError::Internal(format!("failed to update place: {err}")))?;

    Ok(rad.as_ref().map(läs_plats))
}

/// Tar bort en plats.
///
/// Kopplingarna till mål och händelser försvinner med platsen (kaskad i
/// databasen).
pub async fn delete(pool: &PgPool, user_id: i64, place_id: i64) -> Result<bool, AppError> {
    let resultat = sqlx::query("DELETE FROM places WHERE user_id = $1 AND id = $2")
        .bind(user_id)
        .bind(place_id)
        .execute(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to delete place: {err}")))?;

    Ok(resultat.rows_affected() > 0)
}

/// Platserna kopplade till ett mål.
pub async fn get_goal_places(pool: &PgPool, goal_id: i64) -> Result<Vec<Place>, AppError> {
    let fråga = format!(
        "SELECT {PLATSKOLUMNER} FROM places p \
         INNER JOIN goal_places gp ON gp.place_id = p.id \
         WHERE gp.goal_id = $1 ORDER BY p.name"
    );

    let rader = sqlx::query(&fråga)
        .bind(goal_id)
        .fetch_all(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to query goal places: {err}")))?;

    Ok(rader.iter().map(läs_plats).collect())
}

/// Platserna kopplade till en kalenderhändelse.
pub async fn get_event_places(pool: &PgPool, event_id: i64) -> Result<Vec<Place>, AppError> {
    let fråga = format!(
        "SELECT {PLATSKOLUMNER} FROM places p \
         INNER JOIN calendar_event_places ep ON ep.place_id = p.id \
         WHERE ep.event_id = $1 ORDER BY p.name"
    );

    let rader = sqlx::query(&fråga)
        .bind(event_id)
        .fetch_all(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to query event places: {err}")))?;

    Ok(rader.iter().map(läs_plats).collect())
}

/// Ersätter platserna på ett mål eller en händelse.
///
/// Endast platser som tillhör användaren kopplas - en plats från någon
/// annan ignoreras.
pub async fn set_places(
    pool: &PgPool,
    user_id: i64,
    owner_id: i64,
    place_ids: &[i64],
    table: &str,
    owner_column: &str,
) -> Result<Vec<Place>, AppError> {
    let mut tx = pool
        .begin()
        .await
        .map_err(|err| AppError::Internal(err.to_string()))?;

    sqlx::query(&format!("DELETE FROM {table} WHERE {owner_column} = $1"))
        .bind(owner_id)
        .execute(&mut *tx)
        .await
        .map_err(|err| AppError::Internal(format!("failed to clear places: {err}")))?;

    for place_id in unique_ids(place_ids) {
        sqlx::query(&format!(
            "INSERT INTO {table} ({owner_column}, place_id) \
             SELECT $1, p.id FROM places p WHERE p.user_id = $2 AND p.id = $3 \
             ON CONFLICT DO NOTHING"
        ))
        .bind(owner_id)
        .bind(user_id)
        .bind(place_id)
        .execute(&mut *tx)
        .await
        .map_err(|err| AppError::Internal(format!("failed to link place: {err}")))?;
    }

    tx.commit()
        .await
        .map_err(|err| AppError::Internal(format!("failed to commit places: {err}")))?;

    let fråga = format!(
        "SELECT {PLATSKOLUMNER} FROM places p \
         INNER JOIN {table} l ON l.place_id = p.id \
         WHERE l.{owner_column} = $1 ORDER BY p.name"
    );

    let rader = sqlx::query(&fråga)
        .bind(owner_id)
        .fetch_all(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to query places: {err}")))?;

    Ok(rader.iter().map(läs_plats).collect())
}

/// Tar bort dubbletter ur en lista med id:n.
fn unique_ids(ids: &[i64]) -> Vec<i64> {
    let mut sedda: Vec<i64> = Vec::new();
    let mut unika: Vec<i64> = Vec::new();

    for id in ids {
        if *id == 0 || sedda.contains(id) {
            continue;
        }
        sedda.push(*id);
        unika.push(*id);
    }

    unika
}

/// Användarens valda "jag är här"-plats.
///
/// `None` betyder att ingen plats är vald.
pub async fn get_current_place(pool: &PgPool, user_id: i64) -> Result<Option<Place>, AppError> {
    let fråga = format!(
        "SELECT {PLATSKOLUMNER} FROM places p \
         INNER JOIN users u ON u.current_place_id = p.id WHERE u.id = $1"
    );

    let rad = sqlx::query(&fråga)
        .bind(user_id)
        .fetch_optional(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to query current place: {err}")))?;

    Ok(rad.as_ref().map(läs_plats))
}

/// Sätter (eller rensar) användarens "jag är här"-plats.
///
/// `Ok(false)` betyder att användaren inte finns - Go svarar då med
/// "Place not found".
pub async fn set_current_place(
    pool: &PgPool,
    user_id: i64,
    place_id: Option<i64>,
) -> Result<(), AppError> {
    match place_id {
        None => {
            sqlx::query("UPDATE users SET current_place_id = NULL WHERE id = $1")
                .bind(user_id)
                .execute(pool)
                .await
                .map_err(|err| {
                    AppError::Internal(format!("failed to clear current place: {err}"))
                })?;
        }
        Some(place_id) => {
            let resultat = sqlx::query("UPDATE users SET current_place_id = $1 WHERE id = $2")
                .bind(place_id)
                .bind(user_id)
                .execute(pool)
                .await
                .map_err(|err| AppError::Internal(format!("failed to set current place: {err}")))?;

            if resultat.rows_affected() == 0 {
                return Err(AppError::BadRequest("Place not found".to_string()));
            }
        }
    }

    Ok(())
}

/// De platser vars radie innehåller positionen.
///
/// Matchningen sker i Rust med haversine-formeln, eftersom radien är per
/// plats och inte kan uttryckas som ett gemensamt WHERE-villkor.
pub async fn get_nearby_places(
    pool: &PgPool,
    user_id: i64,
    latitude: f64,
    longitude: f64,
) -> Result<Vec<Place>, AppError> {
    let platser = list(pool, user_id).await?;

    let mut nearby: Vec<Place> = platser
        .into_iter()
        .filter(|plats| plats.contains_position(latitude, longitude))
        .collect();

    nearby.sort_by(|a, b| {
        a.distance_meters(latitude, longitude)
            .partial_cmp(&b.distance_meters(latitude, longitude))
            .unwrap_or(std::cmp::Ordering::Equal)
    });

    Ok(nearby)
}

/// Platserna för flera mål samtidigt.
///
/// Nyckeln är målets id. Mål utan platser finns inte i kartan.
pub async fn get_places_for_goals(
    pool: &PgPool,
    goal_ids: &[i64],
) -> Result<HashMap<i64, Vec<Place>>, AppError> {
    let mut resultat: HashMap<i64, Vec<Place>> = HashMap::new();

    if goal_ids.is_empty() {
        return Ok(resultat);
    }

    let rader = sqlx::query(
        r#"
        SELECT gp.goal_id, p.id, p.name, p.latitude, p.longitude, p.radius_meters,
            COALESCE(p.icon, ''), p.created_at, p.updated_at
        FROM places p
        INNER JOIN goal_places gp ON gp.place_id = p.id
        WHERE gp.goal_id = ANY($1)
        ORDER BY p.name
        "#,
    )
    .bind(goal_ids)
    .fetch_all(pool)
    .await
    .map_err(|err| AppError::Internal(format!("failed to query goal places: {err}")))?;

    for rad in rader {
        let goal_id = rad.get::<i64, _>(0);
        let plats = Place {
            id: rad.get::<i64, _>(1),
            name: rad.get::<String, _>(2),
            latitude: rad.get::<f64, _>(3),
            longitude: rad.get::<f64, _>(4),
            radius_meters: rad.get::<i32, _>(5),
            icon: rad.get::<String, _>(6),
            created_at: rad.get::<chrono::DateTime<chrono::Utc>, _>(7),
            updated_at: rad.get::<chrono::DateTime<chrono::Utc>, _>(8),
        };

        resultat.entry(goal_id).or_default().push(plats);
    }

    Ok(resultat)
}
