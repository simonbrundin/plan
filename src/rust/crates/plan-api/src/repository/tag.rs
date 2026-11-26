//! Taggar.
//!
//! Port av `src/api/internal/repository/tag_repository.go`.

use std::collections::HashMap;

use sqlx::{PgPool, Row};

use crate::error::AppError;

/// Taggarna för flera mål, grupperade per mål.
///
/// Taggarna hämtas samlat för alla mål - en fråga i stället för en per mål.
pub async fn tags_for_goals(
    pool: &PgPool,
    goal_ids: &[i64],
) -> Result<HashMap<i64, Vec<String>>, AppError> {
    let mut resultat: HashMap<i64, Vec<String>> = HashMap::new();

    if goal_ids.is_empty() {
        return Ok(resultat);
    }

    let rader = sqlx::query(
        r#"
        SELECT gt.goal_id, t.name
        FROM goal_tags gt
        INNER JOIN tags t ON t.id = gt.tag_id
        WHERE gt.goal_id = ANY($1)
        ORDER BY t.name
        "#,
    )
    .bind(goal_ids)
    .fetch_all(pool)
    .await
    .map_err(|err| AppError::Internal(format!("failed to query goal tags: {err}")))?;

    for rad in rader {
        let goal_id = rad.get::<i32, _>(0) as i64;
        let namn = rad.get::<String, _>(1);
        resultat.entry(goal_id).or_default().push(namn);
    }

    Ok(resultat)
}
