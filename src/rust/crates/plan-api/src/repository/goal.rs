//! Mål och statusar.
//!
//! Port av `src/api/internal/repository/goal_repository.go`. Frågorna är
//! identiska med Go-tjänstens - samma JOIN, samma sortering, samma
//! felmeddelanden.

use sqlx::{PgPool, Row};

use crate::error::AppError;
use crate::models::status::{Status, StatusUpdate};

/// Har användaren tillgång till målet?
///
/// Tillgången ges via `user_goals` - tabellen som binder samman en
/// användare med de mål hen ser.
pub async fn user_has_access(pool: &PgPool, user_id: i64, goal_id: i64) -> Result<bool, AppError> {
    let finns: bool = sqlx::query_scalar(
        "SELECT EXISTS(SELECT 1 FROM user_goals WHERE user_id = $1 AND goal_id = $2)",
    )
    .bind(user_id)
    .bind(goal_id)
    .fetch_one(pool)
    .await
    .map_err(|err| AppError::Internal(format!("failed to check access: {err}")))?;

    Ok(finns)
}

/// Alla statusar, i id-ordning.
pub async fn get_all_statuses(pool: &PgPool) -> Result<Vec<Status>, AppError> {
    let rader = sqlx::query("SELECT id, name, label FROM statuses ORDER BY id")
        .fetch_all(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to get statuses: {err}")))?;

    Ok(rader
        .into_iter()
        .map(|rad| Status {
            id: rad.get::<i32, _>(0) as i64,
            name: rad.get::<String, _>(1),
            label: rad.get::<String, _>(2),
            created: String::new(),
        })
        .collect())
}

/// Statusen på ett mål.
pub async fn get_goal_status(pool: &PgPool, goal_id: i64) -> Result<Option<Status>, AppError> {
    let rad = sqlx::query(
        r#"
        SELECT s.id, s.name, s.label
        FROM goals g
        INNER JOIN statuses s ON s.id = g.status_id
        WHERE g.id = $1
        "#,
    )
    .bind(goal_id)
    .fetch_optional(pool)
    .await
    .map_err(|err| AppError::Internal(format!("failed to get goal status: {err}")))?;

    Ok(rad.map(|rad| Status {
        id: rad.get::<i32, _>(0) as i64,
        name: rad.get::<String, _>(1),
        label: rad.get::<String, _>(2),
        created: String::new(),
    }))
}

/// Byter status på ett mål och loggar ändringen.
///
/// Statusbytet och loggen sker i samma transaktion - annars kan historiken
/// och målets status glida isär.
pub async fn update_goal_status(
    pool: &PgPool,
    goal_id: i64,
    new_status_id: i64,
) -> Result<StatusUpdate, AppError> {
    let mut tx = pool
        .begin()
        .await
        .map_err(|err| AppError::Internal(format!("failed to begin transaction: {err}")))?;

    // Statusen målet hade innan.
    let nuvarande = sqlx::query("SELECT status_id FROM goals WHERE id = $1")
        .bind(goal_id)
        .fetch_optional(&mut *tx)
        .await
        .map_err(|err| AppError::Internal(format!("failed to get current status: {err}")))?;

    let current_status_id: Option<i64> = nuvarande.as_ref().map(|rad| rad.get::<i32, _>(0) as i64);

    let from_status_name = match current_status_id {
        Some(id) => sqlx::query("SELECT name FROM statuses WHERE id = $1")
            .bind(id)
            .fetch_optional(&mut *tx)
            .await
            .ok()
            .flatten()
            .map(|rad| rad.get::<String, _>(0))
            .unwrap_or_default(),
        None => String::new(),
    };

    sqlx::query("UPDATE goals SET status_id = $1 WHERE id = $2")
        .bind(new_status_id)
        .bind(goal_id)
        .execute(&mut *tx)
        .await
        .map_err(|err| AppError::Internal(format!("failed to update goal status: {err}")))?;

    let to_status_name = sqlx::query("SELECT name FROM statuses WHERE id = $1")
        .bind(new_status_id)
        .fetch_optional(&mut *tx)
        .await
        .ok()
        .flatten()
        .map(|rad| rad.get::<String, _>(0))
        .unwrap_or_default();

    let rad = sqlx::query(
        r#"
        INSERT INTO status_updates (goal_id, from_status_id, to_status_id)
        VALUES ($1, $2, $3)
        RETURNING id, changed_at
        "#,
    )
    .bind(goal_id)
    .bind(current_status_id)
    .bind(new_status_id)
    .fetch_one(&mut *tx)
    .await
    .map_err(|err| AppError::Internal(format!("failed to log status update: {err}")))?;

    let id = rad.get::<i32, _>(0) as i64;
    let changed_at = rad.get::<chrono::DateTime<chrono::Utc>, _>(1);

    tx.commit()
        .await
        .map_err(|err| AppError::Internal(format!("failed to commit transaction: {err}")))?;

    Ok(StatusUpdate {
        id,
        goal_id,
        from_status_id: current_status_id,
        to_status_id: new_status_id,
        changed_at,
        from_status_name,
        to_status_name,
    })
}

/// Statushistoriken för ett mål, nyast först.
pub async fn get_status_history(
    pool: &PgPool,
    goal_id: i64,
) -> Result<Vec<StatusUpdate>, AppError> {
    let rader = sqlx::query(
        r#"
        SELECT
            su.id,
            su.goal_id,
            su.from_status_id,
            su.to_status_id,
            su.changed_at,
            COALESCE(fs.name, '') as from_status_name,
            ts.name as to_status_name
        FROM status_updates su
        LEFT JOIN statuses fs ON fs.id = su.from_status_id
        INNER JOIN statuses ts ON ts.id = su.to_status_id
        WHERE su.goal_id = $1
        ORDER BY su.changed_at DESC
        "#,
    )
    .bind(goal_id)
    .fetch_all(pool)
    .await
    .map_err(|err| AppError::Internal(format!("failed to get status history: {err}")))?;

    Ok(rader
        .into_iter()
        .map(|rad| StatusUpdate {
            id: rad.get::<i32, _>(0) as i64,
            goal_id: rad.get::<i32, _>(1) as i64,
            from_status_id: rad.get::<Option<i32>, _>(2).map(|id| id as i64),
            to_status_id: rad.get::<i32, _>(3) as i64,
            changed_at: rad.get::<chrono::DateTime<chrono::Utc>, _>(4),
            from_status_name: rad.get::<String, _>(5),
            to_status_name: rad.get::<String, _>(6),
        })
        .collect())
}

/// Användarens mål, filtrerade på agent-flagga och status.
///
/// Ett tomt filter ger alla mål. Agent-fältet är valfritt, så att false kan
/// skiljas från "ingen filtrering".
pub async fn get_goals_filtered(
    pool: &PgPool,
    user_id: i64,
    filter: &crate::models::goal::GoalFilter,
) -> Result<Vec<crate::models::goal::Goal>, AppError> {
    let query = r#"
        SELECT g.id, g.title, COALESCE(g.description, ''), g.icon, g.status_id, g.created, g.started, g.finished, g.due, g.start_date, g.estimated_minutes, g.inbox, g.eternal, g.agent, g.now, g."order",
            EXISTS (SELECT 1 FROM goal_relations gr WHERE gr.child_id = g.id) AS has_parent,
            g.recurrence_rule, g.recurrence_mode, g.recurrence_next,
            (SELECT COUNT(*) FROM goal_completions gc WHERE gc.goal_id = g.id) AS recurrence_count
        FROM goals g
        INNER JOIN user_goals ug ON g.id = ug.goal_id
        WHERE ug.user_id =
        "#;

    let mut byggare = sqlx::QueryBuilder::<sqlx::Postgres>::new(query);
    byggare.push_bind(user_id);

    if let Some(agent) = filter.agent {
        byggare.push(" AND g.agent = ").push_bind(agent);
    }

    if let Some(status_id) = filter.status_id {
        byggare.push(" AND g.status_id = ").push_bind(status_id);
    }

    byggare.push(" ORDER BY g.created DESC");

    let rader = byggare
        .build()
        .fetch_all(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to query goals: {err}")))?;

    Ok(rader.iter().map(scanna_mal).collect())
}

/// Ett mål på sitt id.
pub async fn get_goal_by_id(
    pool: &PgPool,
    goal_id: i64,
) -> Result<Option<crate::models::goal::Goal>, AppError> {
    let rad = sqlx::query(
        r#"
        SELECT id, title, COALESCE(description, ''), icon, status_id, created, started, finished, due, start_date, estimated_minutes, inbox, eternal, agent, now, "order",
            EXISTS (SELECT 1 FROM goal_relations gr WHERE gr.child_id = goals.id) AS has_parent,
            recurrence_rule, recurrence_mode, recurrence_next,
            (SELECT COUNT(*) FROM goal_completions gc WHERE gc.goal_id = goals.id) AS recurrence_count
        FROM goals WHERE id = $1
        "#,
    )
    .bind(goal_id)
    .fetch_optional(pool)
    .await
    .map_err(|err| AppError::Internal(format!("failed to get goal: {err}")))?;

    Ok(rad.as_ref().map(scanna_mal))
}

/// Läser en målrad.
fn scanna_mal(rad: &sqlx::postgres::PgRow) -> crate::models::goal::Goal {
    use sqlx::Row;

    crate::models::goal::Goal {
        id: rad.get::<i32, _>(0) as i64,
        title: rad.get::<String, _>(1),
        description: rad.get::<String, _>(2),
        icon: rad.get::<String, _>(3),
        status_id: rad.get::<i32, _>(4),
        created: rad.get::<chrono::DateTime<chrono::Utc>, _>(5),
        started: rad.get::<Option<chrono::DateTime<chrono::Utc>>, _>(6),
        finished: rad.get::<Option<chrono::DateTime<chrono::Utc>>, _>(7),
        due: rad.get::<Option<chrono::NaiveDate>, _>(8),
        start_date: rad.get::<Option<chrono::NaiveDate>, _>(9),
        estimated_minutes: rad.get::<Option<i32>, _>(10),
        inbox: rad.get::<i32, _>(11),
        eternal: rad.get::<bool, _>(12),
        agent: rad.get::<bool, _>(13),
        now: rad.get::<bool, _>(14),
        order: rad.get::<i32, _>(15),
        has_parent: rad.get::<bool, _>(16),
        recurrence_rule: rad.get::<Option<String>, _>(17),
        recurrence_mode: rad.get::<Option<String>, _>(18),
        recurrence_next: rad.get::<Option<chrono::NaiveDate>, _>(19),
        recurrence_count: rad.get::<i64, _>(20) as i32,
        places: Vec::new(),
        tags: Vec::new(),
    }
}

/// Skapar ett mål och kopplar det till användaren.
///
/// Nya grundmål läggs sist i ordningen, så att de inte hamnar ovanför
/// mål som redan har en sparad placering.
pub async fn create_goal(
    pool: &PgPool,
    user_id: i64,
    title: &str,
    icon: &str,
) -> Result<crate::models::goal::Goal, AppError> {
    let mut tx = pool
        .begin()
        .await
        .map_err(|err| AppError::Internal(format!("failed to begin transaction: {err}")))?;

    let rad = sqlx::query(
        r#"
        INSERT INTO goals (title, icon, "order")
        SELECT $1, $2, COALESCE((
            SELECT MAX(g2."order") + 1
            FROM goals g2
            INNER JOIN user_goals ug2 ON g2.id = ug2.goal_id
            WHERE ug2.user_id = $3), 0)
        RETURNING id, title, COALESCE(description, ''), icon, status_id, created, started, finished, due, estimated_minutes, inbox, eternal, agent, now, "order"
        "#,
    )
    .bind(title)
    .bind(icon)
    .bind(user_id)
    .fetch_one(&mut *tx)
    .await
    .map_err(|err| AppError::Internal(format!("failed to create goal: {err}")))?;

    let goal_id = rad.get::<i32, _>(0) as i64;

    sqlx::query("INSERT INTO user_goals (user_id, goal_id) VALUES ($1, $2)")
        .bind(user_id)
        .bind(goal_id)
        .execute(&mut *tx)
        .await
        .map_err(|err| AppError::Internal(format!("failed to associate goal with user: {err}")))?;

    tx.commit()
        .await
        .map_err(|err| AppError::Internal(format!("failed to commit transaction: {err}")))?;

    get_goal_by_id(pool, goal_id)
        .await?
        .ok_or_else(|| AppError::Internal("failed to read created goal".to_string()))
}

/// Tar bort ett mål.
pub async fn delete_goal(pool: &PgPool, goal_id: i64) -> Result<(), AppError> {
    sqlx::query("DELETE FROM goals WHERE id = $1")
        .bind(goal_id)
        .execute(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to delete goal: {err}")))?;

    Ok(())
}

/// Alla relationer mellan användarens mål.
///
/// Endast relationer vars barnmål tillhör användaren.
pub async fn get_all_relations(
    pool: &PgPool,
    user_id: i64,
) -> Result<Vec<crate::models::goal::GoalRelation>, AppError> {
    let rader = sqlx::query(
        r#"
        SELECT gr.parent_id, gr.child_id, gr."order", gr.weight
        FROM goal_relations gr
        INNER JOIN user_goals ug ON ug.goal_id = gr.child_id
        WHERE ug.user_id = $1
        ORDER BY gr."order" ASC
        "#,
    )
    .bind(user_id)
    .fetch_all(pool)
    .await
    .map_err(|err| AppError::Internal(format!("failed to get relations: {err}")))?;

    Ok(rader
        .iter()
        .map(|rad| crate::models::goal::GoalRelation {
            parent_id: rad.get::<i32, _>(0) as i64,
            child_id: rad.get::<i32, _>(1) as i64,
            order: rad.get::<i32, _>(2),
            weight: rad.get::<i32, _>(3),
        })
        .collect())
}

/// Alla beroenden mellan användarens mål.
pub async fn get_all_dependencies(
    pool: &PgPool,
    user_id: i64,
) -> Result<Vec<crate::models::goal::GoalDependency>, AppError> {
    let rader = sqlx::query(
        r#"
        SELECT gd.id, gd.goal_id, gd.depends_on_id, gd.created
        FROM goal_dependencies gd
        INNER JOIN user_goals ug ON ug.goal_id = gd.goal_id
        WHERE ug.user_id = $1
        "#,
    )
    .bind(user_id)
    .fetch_all(pool)
    .await
    .map_err(|err| AppError::Internal(format!("failed to get dependencies: {err}")))?;

    Ok(rader
        .iter()
        .map(|rad| crate::models::goal::GoalDependency {
            id: rad.get::<i64, _>(0),
            goal_id: rad.get::<i32, _>(1) as i64,
            depends_on_id: rad.get::<i32, _>(2) as i64,
            created: rad.get::<chrono::DateTime<chrono::Utc>, _>(3),
        })
        .collect())
}
