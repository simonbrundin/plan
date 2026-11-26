//! Kalenderhändelser.
//!
//! Port av `src/api/internal/repository/calendar_event_repository.go`.
//! Återkommande serier expanderas med `plan-recurrence`, som är en port av
//! Go-tjänstens `rrule-go`-användning.

use std::collections::HashMap;

use chrono::{DateTime, Duration, Utc};
use sqlx::{PgPool, Row};

use crate::error::AppError;
use crate::models::calendar_event::{
    CalendarEvent, CreateCalendarEventRequest, CreateCalendarOccurrenceRequest,
    SplitCalendarEventRequest, UpdateCalendarEventRequest,
};
use crate::models::place::Place;
use plan_recurrence::Regel;

/// Kolumnerna som läses för en händelse.
const HÄNDELSEKOLUMNER: &str = "id, user_id, goal_id, calendar_id, title, \
     COALESCE(description, ''), COALESCE(icon, ''), start_at, end_at, time_zone, \
     all_day, status, recurrence_rule, parent_event_id, original_start_at, \
     created_at, updated_at";

/// Maximalt antal förekomster som expanderas för en serie.
const MAX_FÖREKOMSTER: usize = 1000;

/// Läser en händelserad.
fn läs_händelse(rad: &sqlx::postgres::PgRow) -> CalendarEvent {
    CalendarEvent {
        id: rad.get::<i64, _>(0),
        occurrence_id: String::new(),
        // Användarens id är INTEGER i schemat, händelsens id BIGINT.
        user_id: rad.get::<i32, _>(1) as i64,
        goal_id: rad.get::<Option<i64>, _>(2),
        calendar_id: rad.get::<Option<i64>, _>(3),
        title: rad.get::<String, _>(4),
        description: rad.get::<String, _>(5),
        icon: rad.get::<String, _>(6),
        start_at: rad.get::<DateTime<Utc>, _>(7),
        end_at: rad.get::<DateTime<Utc>, _>(8),
        time_zone: rad.get::<String, _>(9),
        all_day: rad.get::<bool, _>(10),
        status: rad.get::<String, _>(11),
        recurrence_rule: rad.get::<Option<String>, _>(12),
        parent_event_id: rad.get::<Option<i64>, _>(13),
        original_start_at: rad.get::<Option<DateTime<Utc>>, _>(14),
        created_at: rad.get::<DateTime<Utc>, _>(15),
        updated_at: rad.get::<DateTime<Utc>, _>(16),
        places: None,
    }
}

/// Nyckeln som identifierar en förekomst i en serie.
///
/// Nyckeln bygger på föräldreid och förekomstens start, precis som i Go.
fn förekomstnyckel(förälder: i64, start: DateTime<Utc>) -> String {
    format!("{förälder}:{}", start.format("%Y-%m-%dT%H:%M:%S%.9fZ"))
}

/// Expanderar en återkommande serie till sina förekomster.
///
/// Serier utan upprepningsregel returneras som de är. Förekomsterna
/// filtreras sedan bort på samma sätt som i Go: en förekomst som slutar
/// före fönstrets början eller börjar vid fönstrets slut räknas inte med.
fn expandera(
    händelse: &CalendarEvent,
    från: DateTime<Utc>,
    till: DateTime<Utc>,
) -> Result<Vec<CalendarEvent>, AppError> {
    let Some(regel_text) = händelse.recurrence_rule.as_deref() else {
        return Ok(vec![händelse.clone()]);
    };

    if regel_text.is_empty() {
        return Ok(vec![händelse.clone()]);
    }

    let regel = Regel::tolka(regel_text).map_err(|err| {
        AppError::Internal(format!(
            "invalid recurrence rule for event {}: {}",
            händelse.id,
            felmeddelande(&err)
        ))
    })?;

    let varaktighet = händelse.end_at - händelse.start_at;
    let förekomster = regel.mellan(händelse.start_at, från - varaktighet, till, true);

    let mut expanderade = Vec::new();
    for start in förekomster.into_iter().take(MAX_FÖREKOMSTER) {
        let slut = start + varaktighet;
        if start >= till || slut <= från {
            continue;
        }

        let mut instans = händelse.clone();
        instans.occurrence_id = format!(
            "{}-{}",
            händelse.id,
            start.timestamp_nanos_opt().unwrap_or(0)
        );
        instans.original_start_at = Some(start);
        instans.start_at = start;
        instans.end_at = slut;
        expanderade.push(instans);
    }

    Ok(expanderade)
}

/// Gör om ett fel från RRULE-motorn till en text.
fn felmeddelande(fel: &plan_recurrence::Fel) -> String {
    match fel {
        plan_recurrence::Fel::OgiltigRegel(detalj) => format!("invalid rule: {detalj}"),
        plan_recurrence::Fel::IngenFörekomst(detalj) => format!("no occurrence: {detalj}"),
        plan_recurrence::Fel::OgiltigLäge(läge) => format!("invalid mode: {läge}"),
    }
}

/// Fyller i händelsernas platser.
///
/// Platserna är extra information - ett fel får inte förstöra svaret.
pub async fn fyll_platser(pool: &PgPool, händelser: &mut [CalendarEvent]) -> Result<(), AppError> {
    if händelser.is_empty() {
        return Ok(());
    }

    let händelse_id: Vec<i64> = händelser.iter().map(|h| h.id).collect();
    let platser = platser_för_händelser(pool, &händelse_id).await?;

    for händelse in händelser.iter_mut() {
        händelse.places = Some(platser.get(&händelse.id).cloned().unwrap_or_default());
    }

    Ok(())
}

/// Platserna för flera händelser samtidigt.
async fn platser_för_händelser(
    pool: &PgPool,
    händelse_id: &[i64],
) -> Result<HashMap<i64, Vec<Place>>, AppError> {
    let mut resultat: HashMap<i64, Vec<Place>> = HashMap::new();

    if händelse_id.is_empty() {
        return Ok(resultat);
    }

    let rader = sqlx::query(
        r#"
        SELECT ep.event_id, p.id, p.name, p.latitude, p.longitude, p.radius_meters,
            COALESCE(p.icon, ''), p.created_at, p.updated_at
        FROM places p
        INNER JOIN calendar_event_places ep ON ep.place_id = p.id
        WHERE ep.event_id = ANY($1)
        ORDER BY p.name
        "#,
    )
    .bind(händelse_id)
    .fetch_all(pool)
    .await
    .map_err(|err| AppError::Internal(format!("failed to query event places: {err}")))?;

    for rad in rader {
        let händelse_id = rad.get::<i64, _>(0);
        let plats = Place {
            id: rad.get::<i64, _>(1),
            name: rad.get::<String, _>(2),
            latitude: rad.get::<f64, _>(3),
            longitude: rad.get::<f64, _>(4),
            radius_meters: rad.get::<i32, _>(5),
            icon: rad.get::<String, _>(6),
            created_at: rad.get::<DateTime<Utc>, _>(7),
            updated_at: rad.get::<DateTime<Utc>, _>(8),
        };

        resultat.entry(händelse_id).or_default().push(plats);
    }

    Ok(resultat)
}

/// Alla händelser, där återkommande serier inte expanderas.
///
/// Klienten expanderar dem själv.
pub async fn list_all(pool: &PgPool, user_id: i64) -> Result<Vec<CalendarEvent>, AppError> {
    let fråga = format!(
        "SELECT {HÄNDELSEKOLUMNER} FROM calendar_events \
         WHERE user_id = $1 AND (status <> 'cancelled' OR parent_event_id IS NOT NULL) \
         ORDER BY start_at, id"
    );

    let rader = sqlx::query(&fråga)
        .bind(user_id)
        .fetch_all(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to list calendar events: {err}")))?;

    Ok(rader.iter().map(läs_händelse).collect())
}

/// Händelserna i ett tidsfönster, där återkommande serier expanderas.
///
/// Enskilda förekomster - overrides - ersätter den expanderade förekomsten
/// de hör till.
pub async fn list(
    pool: &PgPool,
    user_id: i64,
    från: DateTime<Utc>,
    till: DateTime<Utc>,
) -> Result<Vec<CalendarEvent>, AppError> {
    let fråga = format!(
        "SELECT {HÄNDELSEKOLUMNER} FROM calendar_events \
         WHERE user_id = $1 \
           AND (status <> 'cancelled' OR parent_event_id IS NOT NULL) \
           AND ((start_at < $3 AND end_at > $2) \
             OR recurrence_rule IS NOT NULL \
             OR (original_start_at >= $2 AND original_start_at < $3)) \
         ORDER BY start_at, id"
    );

    let rader = sqlx::query(&fråga)
        .bind(user_id)
        .bind(från)
        .bind(till)
        .fetch_all(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to list calendar events: {err}")))?;

    let mut serier = Vec::new();
    let mut overrides: HashMap<String, CalendarEvent> = HashMap::new();

    for rad in &rader {
        let händelse = läs_händelse(rad);

        if let (Some(förälder), Some(original)) =
            (händelse.parent_event_id, händelse.original_start_at)
        {
            overrides.insert(förekomstnyckel(förälder, original), händelse);
            continue;
        }

        serier.push(händelse);
    }

    let mut händelser = Vec::new();
    for serie in &serier {
        for förekomst in expandera(serie, från, till)? {
            let nyckel = förekomstnyckel(serie.id, förekomst.start_at);

            if let Some(mut override_rad) = overrides.remove(&nyckel) {
                override_rad.occurrence_id = format!(
                    "{}-{}",
                    serie.id,
                    förekomst.start_at.timestamp_nanos_opt().unwrap_or(0)
                );
                if override_rad.status != "cancelled" {
                    händelser.push(override_rad);
                }
                continue;
            }

            händelser.push(förekomst);
        }
    }

    // Overrides som inte matchar någon expanderad förekomst - de kan ligga
    // utanför fönstret, eller höra till en serie som inte expanderades.
    for (_, mut override_rad) in overrides {
        if override_rad.start_at < till
            && override_rad.end_at > från
            && override_rad.status != "cancelled"
        {
            override_rad.occurrence_id = format!(
                "{}-{}",
                override_rad.id,
                override_rad.start_at.timestamp_nanos_opt().unwrap_or(0)
            );
            händelser.push(override_rad);
        }
    }

    Ok(händelser)
}

/// En enskild händelse. `None` om den inte finns eller tillhör annan.
pub async fn get(
    pool: &PgPool,
    user_id: i64,
    event_id: i64,
) -> Result<Option<CalendarEvent>, AppError> {
    let fråga =
        format!("SELECT {HÄNDELSEKOLUMNER} FROM calendar_events WHERE id = $1 AND user_id = $2");

    let rad = sqlx::query(&fråga)
        .bind(event_id)
        .bind(user_id)
        .fetch_optional(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to get calendar event: {err}")))?;

    Ok(rad.as_ref().map(läs_händelse))
}

/// Skapar en händelse.
pub async fn create(
    pool: &PgPool,
    user_id: i64,
    req: &CreateCalendarEventRequest,
    start_at: DateTime<Utc>,
    end_at: DateTime<Utc>,
) -> Result<CalendarEvent, AppError> {
    let status = match req.status.as_deref() {
        Some(status) if !status.is_empty() => status.to_string(),
        _ => "confirmed".to_string(),
    };
    let tidszon = match req.time_zone.as_deref() {
        Some(tidszon) if !tidszon.is_empty() => tidszon.to_string(),
        _ => "UTC".to_string(),
    };

    let fråga = format!(
        "INSERT INTO calendar_events \
         (user_id, goal_id, calendar_id, title, description, icon, start_at, end_at, \
          time_zone, all_day, status, recurrence_rule) \
         VALUES ($1, $2, $3, $4, $5, NULLIF($6::text, ''), $7, $8, $9, $10, $11, $12) \
         RETURNING {HÄNDELSEKOLUMNER}"
    );

    let rad = sqlx::query(&fråga)
        .bind(user_id)
        .bind(req.goal_id)
        .bind(req.calendar_id)
        .bind(&req.title)
        .bind(req.description.as_deref().unwrap_or_default())
        .bind(req.icon.as_deref())
        .bind(start_at)
        .bind(end_at)
        .bind(&tidszon)
        // "all_day: null" betyder false - som i Go, där null lämnar
        // fältets standardvärde.
        .bind(req.all_day.unwrap_or(false))
        .bind(&status)
        .bind(req.recurrence_rule.as_deref())
        .fetch_one(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to create calendar event: {err}")))?;

    Ok(läs_händelse(&rad))
}

/// Uppdaterar en händelse.
///
/// Ikonen och kalendern har tre tillstånd: inget betyder "behåll nuvarande
/// värde", en tom sträng "töm" och ett värde "sätt".
pub async fn update(
    pool: &PgPool,
    user_id: i64,
    event_id: i64,
    req: &UpdateCalendarEventRequest,
    start_at: Option<DateTime<Utc>>,
    end_at: Option<DateTime<Utc>>,
) -> Result<Option<CalendarEvent>, AppError> {
    let fråga = format!(
        "UPDATE calendar_events SET \
            goal_id = COALESCE($3, goal_id), \
            title = COALESCE($4, title), \
            description = COALESCE($5, description), \
            icon = CASE WHEN $12::text IS NULL THEN icon WHEN $12::text = '' THEN NULL ELSE $12::text END, \
            start_at = COALESCE($6, start_at), \
            end_at = COALESCE($7, end_at), \
            time_zone = COALESCE($8, time_zone), \
            all_day = COALESCE($9, all_day), \
            status = COALESCE($10, status), \
            recurrence_rule = COALESCE($11, recurrence_rule), \
            calendar_id = CASE WHEN $13::bigint IS NULL THEN calendar_id ELSE NULLIF($13::bigint, 0) END, \
            updated_at = CURRENT_TIMESTAMP \
         WHERE id = $1 AND user_id = $2 \
         RETURNING {HÄNDELSEKOLUMNER}"
    );

    let rad = sqlx::query(&fråga)
        .bind(event_id)
        .bind(user_id)
        .bind(req.goal_id)
        .bind(req.title.as_deref())
        .bind(req.description.as_deref())
        .bind(start_at)
        .bind(end_at)
        .bind(req.time_zone.as_deref())
        .bind(req.all_day)
        .bind(req.status.as_deref())
        .bind(req.recurrence_rule.as_deref())
        .bind(req.icon.as_deref())
        .bind(req.calendar_id)
        .fetch_optional(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to update calendar event: {err}")))?;

    Ok(rad.as_ref().map(läs_händelse))
}

/// Skapar en override för en förekomst i en serie.
///
/// Fält som inte anges ärvs från förälderserien.
pub async fn create_override(
    pool: &PgPool,
    user_id: i64,
    parent_event_id: i64,
    req: &CreateCalendarOccurrenceRequest,
    original_start: DateTime<Utc>,
    start_at: DateTime<Utc>,
    end_at: DateTime<Utc>,
) -> Result<Option<CalendarEvent>, AppError> {
    let Some(förälder) = get(pool, user_id, parent_event_id).await? else {
        return Ok(None);
    };

    if förälder.recurrence_rule.is_none() {
        return Ok(None);
    }

    let titel = match req.title.as_deref() {
        Some(titel) if !titel.is_empty() => titel.to_string(),
        _ => förälder.title.clone(),
    };
    let beskrivning = match req.description.as_deref() {
        Some(beskrivning) if !beskrivning.is_empty() => beskrivning.to_string(),
        _ => förälder.description.clone(),
    };
    let tidszon = match req.time_zone.as_deref() {
        Some(tidszon) if !tidszon.is_empty() => tidszon.to_string(),
        _ => förälder.time_zone.clone(),
    };
    let hela_dagen = req.all_day.unwrap_or(förälder.all_day);
    let status = match req.status.as_deref() {
        Some(status) if !status.is_empty() => status.to_string(),
        _ => "confirmed".to_string(),
    };
    let mål_id = req.goal_id.or(förälder.goal_id);
    let kalender_id = match req.calendar_id {
        Some(0) => None,
        Some(id) => Some(id),
        None => förälder.calendar_id,
    };
    let ikon = crate::routes::calendar_events::resolve_icon(&förälder.icon, req.icon.as_deref());

    let fråga = format!(
        "INSERT INTO calendar_events \
         (user_id, goal_id, calendar_id, title, description, icon, start_at, end_at, \
          time_zone, all_day, status, parent_event_id, original_start_at) \
         VALUES ($1, $2, $3, $4, $5, NULLIF($6::text, ''), $7, $8, $9, $10, $11, $12, $13) \
         ON CONFLICT DO NOTHING \
         RETURNING {HÄNDELSEKOLUMNER}"
    );

    let rad = sqlx::query(&fråga)
        .bind(user_id)
        .bind(mål_id)
        .bind(kalender_id)
        .bind(&titel)
        .bind(&beskrivning)
        .bind(ikon.as_deref())
        .bind(start_at)
        .bind(end_at)
        .bind(&tidszon)
        .bind(hela_dagen)
        .bind(&status)
        .bind(parent_event_id)
        .bind(original_start)
        .fetch_optional(pool)
        .await
        .map_err(|err| {
            AppError::Internal(format!("failed to create calendar occurrence: {err}"))
        })?;

    Ok(rad.as_ref().map(läs_händelse))
}

/// Delar en återkommande serie.
///
/// Den gamla serien avslutas strax före den valda förekomsten och en ny
/// serie fortsätter därifrån.
pub async fn split_series(
    pool: &PgPool,
    user_id: i64,
    parent_event_id: i64,
    req: &SplitCalendarEventRequest,
    original_start: DateTime<Utc>,
    start_at: DateTime<Utc>,
    end_at: DateTime<Utc>,
) -> Result<Option<CalendarEvent>, AppError> {
    let Some(förälder) = get(pool, user_id, parent_event_id).await? else {
        return Ok(None);
    };

    let Some(föräldelsregel) = förälder.recurrence_rule.as_deref() else {
        return Ok(None);
    };

    let titel = req.title.clone().unwrap_or_else(|| förälder.title.clone());
    let beskrivning = req
        .description
        .clone()
        .unwrap_or_else(|| förälder.description.clone());
    let tidszon = req
        .time_zone
        .clone()
        .unwrap_or_else(|| förälder.time_zone.clone());
    let hela_dagen = req.all_day.unwrap_or(förälder.all_day);
    let mål_id = req.goal_id.or(förälder.goal_id);
    let kalender_id = match req.calendar_id {
        Some(0) => None,
        Some(id) => Some(id),
        None => förälder.calendar_id,
    };
    let ikon = crate::routes::calendar_events::resolve_icon(&förälder.icon, req.icon.as_deref());

    // Den gamla serien får ett UNTIL strax före den valda förekomsten.
    let trunkerad = trunkera_regel(föräldelsregel, original_start - Duration::nanoseconds(1))
        .map_err(AppError::BadRequest)?;

    let mut tx = pool
        .begin()
        .await
        .map_err(|err| AppError::Internal(err.to_string()))?;

    sqlx::query(
        "UPDATE calendar_events SET recurrence_rule = $1, updated_at = CURRENT_TIMESTAMP \
         WHERE id = $2 AND user_id = $3",
    )
    .bind(&trunkerad)
    .bind(parent_event_id)
    .bind(user_id)
    .execute(&mut *tx)
    .await
    .map_err(|err| AppError::Internal(format!("failed to truncate calendar series: {err}")))?;

    let fråga = format!(
        "INSERT INTO calendar_events \
         (user_id, goal_id, calendar_id, title, description, icon, start_at, end_at, \
          time_zone, all_day, status, recurrence_rule) \
         VALUES ($1, $2, $3, $4, $5, NULLIF($6::text, ''), $7, $8, $9, $10, $11, $12) \
         RETURNING {HÄNDELSEKOLUMNER}"
    );

    let rad = sqlx::query(&fråga)
        .bind(user_id)
        .bind(mål_id)
        .bind(kalender_id)
        .bind(&titel)
        .bind(&beskrivning)
        .bind(ikon.as_deref())
        .bind(start_at)
        .bind(end_at)
        .bind(&tidszon)
        .bind(hela_dagen)
        .bind(&förälder.status)
        .bind(&req.recurrence_rule)
        .fetch_one(&mut *tx)
        .await
        .map_err(|err| AppError::Internal(format!("failed to create calendar series: {err}")))?;

    let skapad = läs_händelse(&rad);

    tx.commit()
        .await
        .map_err(|err| AppError::Internal(err.to_string()))?;

    Ok(Some(skapad))
}

/// Trunkerar en upprepningsregel vid en tidpunkt.
///
/// COUNT tas bort och UNTIL sätts, så att serien slutar strax före
/// tidpunkten. Formatet är detsamma som Go:s `rrule-go` producerar.
fn trunkera_regel(regel: &str, till: DateTime<Utc>) -> Result<String, String> {
    let tolkad = Regel::tolka(regel).map_err(|err| felmeddelande(&err))?;

    let mut delar = vec![format!("FREQ={}", frekvensnamn(tolkad.frekvens))];

    if tolkad.intervall != 1 {
        delar.push(format!("INTERVAL={}", tolkad.intervall));
    }

    delar.push(format!("UNTIL={}", till.format("%Y%m%dT%H%M%SZ")));

    if !tolkad.by_day.is_empty() {
        let dagar: Vec<&str> = tolkad.by_day.iter().map(|d| vardagsnamn(*d)).collect();
        delar.push(format!("BYDAY={}", dagar.join(",")));
    }

    Ok(delar.join(";"))
}

/// Frekvensens namn i RRULE-format.
fn frekvensnamn(frekvens: plan_recurrence::Frekvens) -> &'static str {
    match frekvens {
        plan_recurrence::Frekvens::Daglig => "DAILY",
        plan_recurrence::Frekvens::Vecklig => "WEEKLY",
        plan_recurrence::Frekvens::Månadslig => "MONTHLY",
        plan_recurrence::Frekvens::Årlig => "YEARLY",
    }
}

/// Veckodagens namn i RRULE-format.
fn vardagsnamn(dag: chrono::Weekday) -> &'static str {
    match dag {
        chrono::Weekday::Mon => "MO",
        chrono::Weekday::Tue => "TU",
        chrono::Weekday::Wed => "WE",
        chrono::Weekday::Thu => "TH",
        chrono::Weekday::Fri => "FR",
        chrono::Weekday::Sat => "SA",
        chrono::Weekday::Sun => "SU",
    }
}

/// Om användaren har målet.
pub async fn user_has_goal(pool: &PgPool, user_id: i64, goal_id: i64) -> Result<bool, AppError> {
    let finns = sqlx::query_scalar(
        "SELECT EXISTS (SELECT 1 FROM user_goals WHERE user_id = $1 AND goal_id = $2)",
    )
    .bind(user_id)
    .bind(goal_id)
    .fetch_one(pool)
    .await
    .map_err(|err| AppError::Internal(format!("failed to check goal access: {err}")))?;

    Ok(finns)
}

/// Om användaren har kalendern.
pub async fn user_has_calendar(
    pool: &PgPool,
    user_id: i64,
    calendar_id: i64,
) -> Result<bool, AppError> {
    let finns = sqlx::query_scalar(
        "SELECT EXISTS (SELECT 1 FROM calendars WHERE id = $1 AND user_id = $2)",
    )
    .bind(calendar_id)
    .bind(user_id)
    .fetch_one(pool)
    .await
    .map_err(|err| AppError::Internal(format!("failed to check calendar access: {err}")))?;

    Ok(finns)
}

/// Tar bort en händelse.
pub async fn delete(pool: &PgPool, user_id: i64, event_id: i64) -> Result<bool, AppError> {
    let resultat = sqlx::query("DELETE FROM calendar_events WHERE id = $1 AND user_id = $2")
        .bind(event_id)
        .bind(user_id)
        .execute(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to delete calendar event: {err}")))?;

    Ok(resultat.rows_affected() > 0)
}

#[cfg(test)]
mod tests {
    use chrono::TimeZone;

    use super::*;

    #[test]
    fn en_regel_trunkeras_med_untill() {
        // Serien avslutas en nanosekund före den valda förekomsten.
        let till = Utc.with_ymd_and_hms(2026, 1, 1, 0, 0, 0).unwrap() - Duration::nanoseconds(1);
        let trunkerad = trunkera_regel("FREQ=DAILY", till).unwrap();

        assert_eq!(trunkerad, "FREQ=DAILY;UNTIL=20251231T235959Z");
    }

    #[test]
    fn intervall_och_veckodagar_overlever_trunkering() {
        let till = Utc.with_ymd_and_hms(2026, 1, 1, 0, 0, 0).unwrap() - Duration::nanoseconds(1);
        let trunkerad = trunkera_regel("FREQ=WEEKLY;BYDAY=MO", till).unwrap();

        assert_eq!(trunkerad, "FREQ=WEEKLY;UNTIL=20251231T235959Z;BYDAY=MO");
    }

    #[test]
    fn en_ogiltig_regel_kunde_inte_trunkeras() {
        let till = Utc.with_ymd_and_hms(2026, 1, 1, 0, 0, 0).unwrap() - Duration::nanoseconds(1);
        assert!(trunkera_regel("FREQ=BOGUS", till).is_err());
    }
}
