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

/// Användarens taggar med antalet mål som har dem.
///
/// Taggar som inte sitter på något mål finns med, men med count 0 - de
/// kan tas bort genom att skriva om målets tagglista.
pub async fn list(pool: &PgPool, user_id: i64) -> Result<Vec<crate::models::tag::Tag>, AppError> {
    let rader = sqlx::query(
        r#"
        SELECT t.name, COUNT(gt.goal_id)
        FROM tags t
        LEFT JOIN goal_tags gt ON gt.tag_id = t.id
        WHERE t.user_id = $1
        GROUP BY t.name
        ORDER BY t.name
        "#,
    )
    .bind(user_id)
    .fetch_all(pool)
    .await
    .map_err(|err| AppError::Internal(format!("failed to query tags: {err}")))?;

    Ok(rader
        .iter()
        .map(|rad| crate::models::tag::Tag {
            name: rad.get::<String, _>(0),
            goal_count: rad.get::<i64, _>(1) as i32,
        })
        .collect())
}

/// Taggarna på ett mål.
pub async fn get_goal_tags(pool: &PgPool, goal_id: i64) -> Result<Vec<String>, AppError> {
    let per_mål = tags_for_goals(pool, &[goal_id]).await?;
    Ok(per_mål.get(&goal_id).cloned().unwrap_or_default())
}

/// Ersätter taggarna på ett mål.
///
/// Taggar som inte finns ännu skapas. Hela bytet sker i en transaktion -
/// antingen får målet alla taggar eller inga.
pub async fn set_goal_tags(
    pool: &PgPool,
    user_id: i64,
    goal_id: i64,
    names: &[String],
) -> Result<Vec<String>, AppError> {
    let normaliserade =
        crate::models::tag::normalize_tag_names(names).map_err(AppError::BadRequest)?;

    let mut tx = pool
        .begin()
        .await
        .map_err(|err| AppError::Internal(format!("failed to start transaction: {err}")))?;

    sqlx::query("DELETE FROM goal_tags WHERE goal_id = $1")
        .bind(goal_id)
        .execute(&mut *tx)
        .await
        .map_err(|err| AppError::Internal(format!("failed to clear goal tags: {err}")))?;

    for namn in &normaliserade {
        let tag_id: i64 = sqlx::query_scalar(
            r#"
            INSERT INTO tags (user_id, name) VALUES ($1, $2)
            ON CONFLICT (user_id, name) DO UPDATE SET name = EXCLUDED.name
            RETURNING id
            "#,
        )
        .bind(user_id)
        .bind(namn)
        .fetch_one(&mut *tx)
        .await
        .map_err(|err| AppError::Internal(format!("failed to create tag {namn:?}: {err}")))?;

        sqlx::query(
            r#"
            INSERT INTO goal_tags (goal_id, tag_id) VALUES ($1, $2)
            ON CONFLICT DO NOTHING
            "#,
        )
        .bind(goal_id)
        .bind(tag_id)
        .execute(&mut *tx)
        .await
        .map_err(|err| AppError::Internal(format!("failed to link tag {namn:?}: {err}")))?;
    }

    tx.commit()
        .await
        .map_err(|err| AppError::Internal(format!("failed to commit goal tags: {err}")))?;

    Ok(normaliserade)
}

/// Tar bort en tagg från ett mål, utan att radera taggen för andra mål.
pub async fn delete_goal_tag(
    pool: &PgPool,
    user_id: i64,
    goal_id: i64,
    name: &str,
) -> Result<Vec<String>, AppError> {
    let normaliserat =
        crate::models::tag::normalize_tag_name(name).map_err(AppError::BadRequest)?;

    let tag_id: Option<i64> =
        sqlx::query_scalar("SELECT id FROM tags WHERE user_id = $1 AND name = $2")
            .bind(user_id)
            .bind(&normaliserat)
            .fetch_optional(pool)
            .await
            .map_err(|err| AppError::Internal(format!("failed to query tag: {err}")))?;

    let Some(tag_id) = tag_id else {
        // Taggen finns inte - inget att ta bort.
        return get_goal_tags(pool, goal_id).await;
    };

    sqlx::query("DELETE FROM goal_tags WHERE goal_id = $1 AND tag_id = $2")
        .bind(goal_id)
        .bind(tag_id)
        .execute(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to delete goal tag: {err}")))?;

    get_goal_tags(pool, goal_id).await
}
