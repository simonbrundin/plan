//! Kalendrar.
//!
//! Port av `src/api/internal/repository/calendar_repository.go`.

use sqlx::{PgPool, Row};

use crate::error::AppError;
use crate::models::calendar::Calendar;

/// Kolumnerna som läses för en kalender.
const KALENDERKOLUMNER: &str = "id, user_id, name, color, COALESCE(icon, ''), \
     is_default, sort_order, created_at, updated_at";

/// Läser en kalenderrad.
fn läs_kalender(rad: &sqlx::postgres::PgRow) -> Calendar {
    Calendar {
        id: rad.get::<i64, _>(0),
        // Användarens id är INTEGER i schemat, kalenderns id BIGINT.
        user_id: rad.get::<i32, _>(1) as i64,
        name: rad.get::<String, _>(2),
        color: rad.get::<String, _>(3),
        icon: rad.get::<String, _>(4),
        is_default: rad.get::<bool, _>(5),
        sort_order: rad.get::<i32, _>(6),
        created_at: rad.get::<chrono::DateTime<chrono::Utc>, _>(7),
        updated_at: rad.get::<chrono::DateTime<chrono::Utc>, _>(8),
    }
}

/// Användarens kalendrar i den ordning de visas i gränssnittet.
pub async fn list(pool: &PgPool, user_id: i64) -> Result<Vec<Calendar>, AppError> {
    let fråga = format!(
        "SELECT {KALENDERKOLUMNER} FROM calendars WHERE user_id = $1 ORDER BY sort_order, id"
    );

    let rader = sqlx::query(&fråga)
        .bind(user_id)
        .fetch_all(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to list calendars: {err}")))?;

    Ok(rader.iter().map(läs_kalender).collect())
}

/// En enskild kalender. `None` om den inte finns eller tillhör annan.
pub async fn get(
    pool: &PgPool,
    user_id: i64,
    calendar_id: i64,
) -> Result<Option<Calendar>, AppError> {
    let fråga = format!("SELECT {KALENDERKOLUMNER} FROM calendars WHERE id = $1 AND user_id = $2");

    let rad = sqlx::query(&fråga)
        .bind(calendar_id)
        .bind(user_id)
        .fetch_optional(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to get calendar: {err}")))?;

    Ok(rad.as_ref().map(läs_kalender))
}

/// Skapar en kalender.
///
/// En användare kan bara ha en standardkalender. När en ny kalender skapas
/// som standard avmarkeras därför de gamla i samma transaktion.
pub async fn create(
    pool: &PgPool,
    user_id: i64,
    req: &crate::models::calendar::CreateCalendarRequest,
) -> Result<Calendar, AppError> {
    let färg = match req.color.as_deref() {
        Some(färg) if !färg.is_empty() => färg.to_string(),
        _ => crate::models::calendar::DEFAULT_CALENDAR_COLOR.to_string(),
    };
    let ordning = req.sort_order.unwrap_or(0);

    let fråga = format!(
        "INSERT INTO calendars (user_id, name, color, icon, is_default, sort_order) \
         VALUES ($1, $2, $3, NULLIF($4::text, ''), $5, $6) RETURNING {KALENDERKOLUMNER}"
    );

    // Utan "standard" behövs ingen transaktion - då kan inga andra
    // kalendrar påverkas.
    if !req.is_default.unwrap_or(false) {
        let rad = sqlx::query(&fråga)
            .bind(user_id)
            .bind(&req.name)
            .bind(&färg)
            .bind(req.icon.as_deref())
            .bind(false)
            .bind(ordning)
            .fetch_one(pool)
            .await
            .map_err(|err| AppError::Internal(format!("failed to create calendar: {err}")))?;

        return Ok(läs_kalender(&rad));
    }

    let mut tx = pool
        .begin()
        .await
        .map_err(|err| AppError::Internal(err.to_string()))?;

    sqlx::query("UPDATE calendars SET is_default = FALSE WHERE user_id = $1 AND is_default")
        .bind(user_id)
        .execute(&mut *tx)
        .await
        .map_err(|err| AppError::Internal(format!("failed to clear default calendar: {err}")))?;

    let rad = sqlx::query(&fråga)
        .bind(user_id)
        .bind(&req.name)
        .bind(&färg)
        .bind(req.icon.as_deref())
        .bind(true)
        .bind(ordning)
        .fetch_one(&mut *tx)
        .await
        .map_err(|err| AppError::Internal(format!("failed to create calendar: {err}")))?;

    tx.commit()
        .await
        .map_err(|err| AppError::Internal(err.to_string()))?;

    Ok(läs_kalender(&rad))
}

/// Uppdaterar en kalender.
///
/// Färgen och ikonen har tre tillstånd: inget betyder "behåll nuvarande
/// värde", en tom sträng "töm" och ett värde "sätt".
pub async fn update(
    pool: &PgPool,
    user_id: i64,
    calendar_id: i64,
    req: &crate::models::calendar::UpdateCalendarRequest,
) -> Result<Option<Calendar>, AppError> {
    let fråga = format!(
        "UPDATE calendars SET
            name = COALESCE($3, name),
            color = CASE WHEN $4::text IS NULL THEN color WHEN $4::text = '' THEN $5 ELSE $4::text END,
            icon = CASE WHEN $6::text IS NULL THEN icon WHEN $6::text = '' THEN NULL ELSE $6::text END,
            is_default = COALESCE($8::boolean, is_default),
            sort_order = COALESCE($7, sort_order),
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $1 AND user_id = $2
        RETURNING {KALENDERKOLUMNER}"
    );

    // Att sätta en kalender som standard avmarkerar den tidigare
    // standardkalendern. Det sker i en transaktion så att det aldrig kan
    // finnas två standardkalendrar samtidigt.
    if req.is_default == Some(true) {
        let mut tx = pool
            .begin()
            .await
            .map_err(|err| AppError::Internal(err.to_string()))?;

        sqlx::query(
            "UPDATE calendars SET is_default = FALSE WHERE user_id = $1 AND is_default AND id <> $2",
        )
        .bind(user_id)
        .bind(calendar_id)
        .execute(&mut *tx)
        .await
        .map_err(|err| AppError::Internal(format!("failed to clear default calendar: {err}")))?;

        let rad = sqlx::query(&fråga)
            .bind(calendar_id)
            .bind(user_id)
            .bind(req.name.as_deref())
            .bind(req.color.as_deref())
            .bind(crate::models::calendar::DEFAULT_CALENDAR_COLOR)
            .bind(req.icon.as_deref())
            .bind(req.sort_order)
            .bind(true)
            .fetch_optional(&mut *tx)
            .await
            .map_err(|err| AppError::Internal(format!("failed to update calendar: {err}")))?;

        let kalender = rad.as_ref().map(läs_kalender);

        tx.commit()
            .await
            .map_err(|err| AppError::Internal(err.to_string()))?;

        return Ok(kalender);
    }

    let rad = sqlx::query(&fråga)
        .bind(calendar_id)
        .bind(user_id)
        .bind(req.name.as_deref())
        .bind(req.color.as_deref())
        .bind(crate::models::calendar::DEFAULT_CALENDAR_COLOR)
        .bind(req.icon.as_deref())
        .bind(req.sort_order)
        .bind(req.is_default)
        .fetch_optional(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to update calendar: {err}")))?;

    Ok(rad.as_ref().map(läs_kalender))
}

/// Tar bort en kalender.
///
/// Händelserna som legat i den behålls men blir kalenderlösa, eftersom
/// kopplingen är ON DELETE SET NULL.
pub async fn delete(pool: &PgPool, user_id: i64, calendar_id: i64) -> Result<bool, AppError> {
    let resultat = sqlx::query("DELETE FROM calendars WHERE id = $1 AND user_id = $2")
        .bind(calendar_id)
        .bind(user_id)
        .execute(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to delete calendar: {err}")))?;

    Ok(resultat.rows_affected() > 0)
}

/// Id:t för användarens standardkalender, eller `None` om den inte har
/// satt någon.
pub async fn default_calendar_id(pool: &PgPool, user_id: i64) -> Result<Option<i64>, AppError> {
    let id =
        sqlx::query_scalar("SELECT id FROM calendars WHERE user_id = $1 AND is_default LIMIT 1")
            .bind(user_id)
            .fetch_optional(pool)
            .await
            .map_err(|err| AppError::Internal(format!("failed to find default calendar: {err}")))?;

    Ok(id)
}
