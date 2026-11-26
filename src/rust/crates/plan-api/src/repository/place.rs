//! Platser.
//!
//! Port av `src/api/internal/repository/place_repository.go`.

use std::collections::HashMap;

use sqlx::{PgPool, Row};

use crate::error::AppError;
use crate::models::place::Place;

/// Platserna för flera mål, grupperade per mål.
///
/// Målen sorteras på namn, som i Go.
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
        let goal_id = rad.get::<i32, _>(0) as i64;
        let plats = Place {
            id: rad.get::<i32, _>(1) as i64,
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
