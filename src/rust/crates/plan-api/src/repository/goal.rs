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
        places: None,
        tags: None,
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

/// Bygger UPDATE-frågan för ett mål.
///
/// Returnerar även om något alls ska uppdateras - en request utan fält
/// lämnar målet orört.
/// Bygger UPDATE-frågan för ett mål.
///
/// Returnerar även om något alls ska uppdateras - en request utan fält
/// lämnar målet orört.
fn bygg_update_query(
    req: &crate::models::goal::UpdateGoalRequest,
) -> Result<(sqlx::QueryBuilder<'_, sqlx::Postgres>, bool), AppError> {
    // QueryBuilder sätter ihop argumenten utan mellanrum, så varje
    // textstycke börjar med ett blanksteg, och fälten skiljs med komma.
    let mut byggare = sqlx::QueryBuilder::<sqlx::Postgres>::new("UPDATE goals SET");
    let mut har_uppdatering = false;
    let mut forsta = true;

    // Sätter ett fält med ett värde.
    macro_rules! satt {
        ($text:expr, $varde:expr) => {{
            if !forsta {
                byggare.push(",");
            }
            byggare.push($text).push_bind($varde);
            forsta = false;
            har_uppdatering = true;
        }};
    }

    // Sätter ett fält till null.
    macro_rules! rensa {
        ($text:expr) => {{
            if !forsta {
                byggare.push(",");
            }
            byggare.push($text);
            forsta = false;
            har_uppdatering = true;
        }};
    }

    if let Some(titel) = &req.title {
        satt!(" title = ", titel);
    }

    if let Some(beskrivning) = &req.description {
        satt!(" description = ", beskrivning);
    }

    if let Some(ikon) = &req.icon {
        satt!(" icon = ", ikon);
    }

    if req.started.set {
        match tidsstämpel(req.started.value.as_deref())? {
            Some(tid) => satt!(" started = ", tid),
            None => rensa!(" started = NULL"),
        }
    }

    // Ett evigt mål kan aldrig vara färdigt. När målet markeras som evigt
    // nollställs därför finished, oavsett vad requesten samtidigt säger.
    let blir_evigt = req.eternal == Some(true);

    if req.finished.set && !blir_evigt {
        match tidsstämpel(req.finished.value.as_deref())? {
            Some(tid) => satt!(" finished = ", tid),
            None => rensa!(" finished = NULL"),
        }
    }

    if let Some(evigt) = req.eternal {
        satt!(" eternal = ", evigt);
        if blir_evigt {
            rensa!(" finished = NULL");
        }
    }

    if let Some(agent) = req.agent {
        satt!(" agent = ", agent);
    }

    if let Some(now) = req.now {
        satt!(" now = ", now);
    }

    if req.due.set {
        match datum(req.due.value.as_deref(), || AppError::InvalidDueDate)? {
            Some(datum) => satt!(" due = ", datum),
            None => rensa!(" due = NULL"),
        }
    }

    // Startdatumet anger från vilket datum målet får dyka upp i
    // prioriteringslistan. Null = målet får synas direkt.
    if req.start_date.set {
        match datum(req.start_date.value.as_deref(), || {
            AppError::InvalidStartDate
        })? {
            Some(datum) => satt!(" start_date = ", datum),
            None => rensa!(" start_date = NULL"),
        }
    }

    // Den uppskattade tidsåtgången sätts i minuter. Null rensar den,
    // och negativa värden avvisas - en uppskattning kan inte vara negativ.
    if req.estimated_minutes.set {
        match req.estimated_minutes.value {
            Some(minuter) if minuter < 0 => {
                return Err(AppError::InvalidEstimate);
            }
            Some(minuter) => satt!(" estimated_minutes = ", minuter),
            None => rensa!(" estimated_minutes = NULL"),
        }
    }

    // Flaggan läses här - annars varnar rustc om den sista tilldelningen.
    let _ = forsta;

    Ok((byggare, har_uppdatering))
}

/// Uppdaterar ett mål.
///
/// Endast de fält som skickas ändras. Fält som skickas som null rensas,
/// och fält som saknas lämnas oförändrade - det är därför fälten är
/// `Nullable`.
pub async fn update_goal(
    pool: &PgPool,
    goal_id: i64,
    req: &crate::models::goal::UpdateGoalRequest,
) -> Result<crate::models::goal::Goal, AppError> {
    let (mut byggare, har_uppdatering) = bygg_update_query(req)?;

    if !har_uppdatering {
        // Inget att uppdatera - målet returneras som det är.
        return get_goal_by_id(pool, goal_id)
            .await?
            .ok_or_else(|| AppError::NotFound("Goal not found".to_string()));
    }

    byggare.push(" WHERE id = ").push_bind(goal_id);

    byggare
        .build()
        .execute(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to update goal: {err}")))?;

    get_goal_by_id(pool, goal_id)
        .await?
        .ok_or_else(|| AppError::NotFound("Goal not found".to_string()))
}

fn tidsstämpel(värde: Option<&str>) -> Result<Option<chrono::DateTime<chrono::Utc>>, AppError> {
    let Some(rått) = värde.filter(|v| !v.is_empty()) else {
        return Ok(None);
    };

    let tid = chrono::DateTime::parse_from_rfc3339(rått)
        .map_err(|_| AppError::InvalidTimestamp(format!("invalid timestamp: {rått:?}")))?;

    Ok(Some(tid.with_timezone(&chrono::Utc)))
}

/// Tolkar ett datum på formatet "YYYY-MM-DD".
///
/// `None` betyder "inget datum" - fältet rensas. `fel` väljer vilket fel
/// som ges - due-datum och startdatum har olika meddelanden i Go.
fn datum(
    värde: Option<&str>,
    fel: fn() -> AppError,
) -> Result<Option<chrono::NaiveDate>, AppError> {
    let Some(rått) = värde.filter(|v| !v.is_empty()) else {
        return Ok(None);
    };

    let datum = chrono::NaiveDate::parse_from_str(rått, "%Y-%m-%d").map_err(|_| fel())?;

    Ok(Some(datum))
}

/// Användarens prioriteringslista.
///
/// Ett mål är inte görbart förrän alla direkta beroenden är klara, och
/// mål med ett startdatum i framtiden syns inte ännu.
pub async fn get_prioritized_goals(
    pool: &PgPool,
    user_id: i64,
) -> Result<Vec<crate::models::goal::PrioritizedGoal>, AppError> {
    let rader = sqlx::query(
        r#"
        SELECT
            g.id, g.title, COALESCE(g.description, ''), g.icon, g.status_id, g.created, g.started, g.finished, g.due, g.start_date, g.estimated_minutes, g.inbox, g.eternal, g.agent, g.now,
            g.priority_order,
            COALESCE((SELECT MAX(weight) FROM goal_relations WHERE child_id = g.id), 10) AS weight,
            (SELECT g2.title FROM goal_relations gr INNER JOIN goals g2 ON g2.id = gr.parent_id WHERE gr.child_id = g.id ORDER BY gr.weight DESC LIMIT 1) AS parent_title,
            (SELECT gr.parent_id FROM goal_relations gr WHERE gr.child_id = g.id ORDER BY gr.weight DESC LIMIT 1) AS parent_id,
            -- Alla föräldrar som målet ligger under, sorterade på vikt.
            COALESCE(
                (
                    SELECT json_agg(json_build_object('id', p.id, 'title', p.title) ORDER BY gr.weight DESC)
                    FROM goal_relations gr
                    INNER JOIN goals p ON p.id = gr.parent_id
                    WHERE gr.child_id = g.id
                ),
                '[]'::json
            ) AS parents,
            g.recurrence_rule, g.recurrence_mode, g.recurrence_next,
            (SELECT COUNT(*) FROM goal_completions gc WHERE gc.goal_id = g.id) AS recurrence_count
        FROM goals g
        INNER JOIN user_goals ug ON g.id = ug.goal_id
        WHERE ug.user_id = $1
          AND g.finished IS NULL
          AND g.id <> 1
          AND (g.start_date IS NULL OR g.start_date <= CURRENT_DATE)
          -- Ett mål är inte görbart förrän alla direkta beroenden är klara.
          AND NOT EXISTS (
            SELECT 1
            FROM goal_dependencies gd
            INNER JOIN goals dependency ON dependency.id = gd.depends_on_id
            WHERE gd.goal_id = g.id
              AND dependency.finished IS NULL
          )
        ORDER BY g.priority_order ASC, weight DESC, g.created DESC
        "#,
    )
    .bind(user_id)
    .fetch_all(pool)
    .await
    .map_err(|err| AppError::Internal(err.to_string()))?;

    Ok(rader
        .iter()
        .map(|rad| {
            let mål = crate::models::goal::Goal {
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
                has_parent: false,
                order: 0,
                places: None,
                tags: None,
                recurrence_rule: rad.get::<Option<String>, _>(20),
                recurrence_mode: rad.get::<Option<String>, _>(21),
                recurrence_next: rad.get::<Option<chrono::NaiveDate>, _>(22),
                recurrence_count: rad.get::<i64, _>(23) as i32,
            };

            let föräldrar: Vec<crate::models::goal::GoalParent> =
                serde_json::from_value(rad.get::<serde_json::Value, _>(19)).unwrap_or_default();

            crate::models::goal::PrioritizedGoal {
                goal: mål,
                weight: rad.get::<i32, _>(16),
                parent_title: rad.get::<Option<String>, _>(17),
                parent_id: rad.get::<Option<i32>, _>(18).map(|id| id as i64),
                priority_order: rad.get::<i32, _>(15),
                parents: föräldrar,
            }
        })
        .collect())
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::models::goal::UpdateGoalRequest;
    use sqlx::Execute;

    /// SQL:en som byggs ska gå att köra - varje fält ska ha sitt utrymme
    /// och sina bindningar på rätt plats.
    #[test]
    fn update_sqlen_byggs_ratt() {
        let req: UpdateGoalRequest =
            serde_json::from_str(r#"{"title":"Ny","due":"2026-12-24","now":true}"#).unwrap();

        let (mut byggare, har_uppdatering) = bygg_update_query(&req).unwrap();
        assert!(har_uppdatering);

        let sql = byggare.build().sql().to_string();
        assert_eq!(sql, "UPDATE goals SET title = $1, now = $2, due = $3");
    }

    /// En request utan fält uppdaterar ingenting.
    #[test]
    fn tom_request_uppdaterar_inget() {
        let req: UpdateGoalRequest = serde_json::from_str("{}").unwrap();

        let (_, har_uppdatering) = bygg_update_query(&req).unwrap();
        assert!(!har_uppdatering);
    }

    /// Ett nullat fält rensas i stället för att sättas.
    #[test]
    fn nullade_falt_rensas() {
        let req: UpdateGoalRequest =
            serde_json::from_str(r#"{"due":null,"finished":null}"#).unwrap();

        let (mut byggare, har_uppdatering) = bygg_update_query(&req).unwrap();
        assert!(har_uppdatering);

        let sql = byggare.build().sql().to_string();
        assert_eq!(sql, "UPDATE goals SET finished = NULL, due = NULL");
    }

    /// Negativ tidsåtgång avvisas.
    #[test]
    fn negativ_tidsatgang_avvisas() {
        let req: UpdateGoalRequest = serde_json::from_str(r#"{"estimated_minutes":-5}"#).unwrap();

        match bygg_update_query(&req) {
            Err(AppError::InvalidEstimate) => {}
            _ => panic!("ett negativt antal minuter ska avvisas"),
        }
    }

    /// Ogiltiga datum avvisas med rätt feltyp.
    #[test]
    fn ogiltiga_datum_avvisas() {
        let req: UpdateGoalRequest = serde_json::from_str(r#"{"due":"2026-13-45"}"#).unwrap();
        match bygg_update_query(&req) {
            Err(AppError::InvalidDueDate) => {}
            _ => panic!("ett ogiltigt due-datum ska avvisas"),
        }

        let req: UpdateGoalRequest = serde_json::from_str(r#"{"start_date":"gammalt"}"#).unwrap();
        match bygg_update_query(&req) {
            Err(AppError::InvalidStartDate) => {}
            _ => panic!("ett ogiltigt startdatum ska avvisas"),
        }
    }
}
