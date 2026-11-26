//! Agentkörningar.
//!
//! Port av `src/api/internal/repository/agent_run_repository.go`.
//!
//! En körning tillhör en användare via sitt mål. Agenten autentiserar med
//! en API-nyckel som är scopad till samma användare, så agenten ser bara
//! sina egna körningar.

use sqlx::postgres::{PgConnection, PgRow};
use sqlx::{PgPool, Postgres, QueryBuilder, Row, Transaction};

use crate::error::AppError;
use crate::models::agent::{
    self, AgentBudget, AgentRun, BudgetStatus, CreateAgentRunRequest, UpdateAgentRunRequest,
};
use crate::pricing::PriceList;

/// Kolumnerna som läses för en agentkörning.
const KOLUMNER: &str = "ar.id, ar.goal_id, ar.phase, ar.status, COALESCE(ar.trigger, 'user'), \
     COALESCE(ar.initiator, ''), COALESCE(ar.agent_profile, 'coding'), COALESCE(ar.summary, ''), \
     COALESCE(ar.result, ''), ar.artifact_url, ar.cost_usd, ar.conversation_id, ar.started_at, \
     ar.finished_at, ar.created_at, ar.updated_at, COALESCE(ar.agent_id, ''), COALESCE(ar.provider, ''), \
     COALESCE(ar.model, ''), COALESCE(ar.model_version, ''), COALESCE(ar.reserved_cost_usd, 0)";

/// Läser en körningsrad.
fn läs_körning(rad: &PgRow) -> AgentRun {
    AgentRun {
        // Körningens id är INTEGER i schemat.
        id: rad.get::<i32, _>(0) as i64,
        // Målets id är INTEGER i schemat.
        goal_id: rad.get::<i32, _>(1) as i64,
        phase: rad.get::<String, _>(2),
        status: rad.get::<String, _>(3),
        trigger: rad.get::<String, _>(4),
        initiator: rad.get::<String, _>(5),
        agent_profile: rad.get::<String, _>(6),
        summary: rad.get::<String, _>(7),
        result: rad.get::<String, _>(8),
        artifact_url: rad.get::<Option<String>, _>(9),
        cost_usd: rad.get::<Option<f64>, _>(10),
        conversation_id: rad.get::<Option<String>, _>(11),
        started_at: rad.get::<Option<chrono::DateTime<chrono::Utc>>, _>(12),
        finished_at: rad.get::<Option<chrono::DateTime<chrono::Utc>>, _>(13),
        created_at: rad.get::<chrono::DateTime<chrono::Utc>, _>(14),
        updated_at: rad.get::<chrono::DateTime<chrono::Utc>, _>(15),
        agent_id: rad.get::<String, _>(16),
        provider: rad.get::<String, _>(17),
        model: rad.get::<String, _>(18),
        model_version: rad.get::<String, _>(19),
        reserved_cost_usd: rad.get::<f64, _>(20),
    }
}

/// Användarens agentkörningar, nyast först.
///
/// `goal_id` begränsar listan till ett mål.
pub async fn list(
    pool: &PgPool,
    user_id: i64,
    goal_id: Option<i64>,
) -> Result<Vec<AgentRun>, AppError> {
    let mut byggare = QueryBuilder::<Postgres>::new(format!(
        "SELECT {KOLUMNER} FROM agent_runs ar \
         INNER JOIN user_goals ug ON ug.goal_id = ar.goal_id \
         WHERE ug.user_id = "
    ));
    byggare.push_bind(user_id);

    if let Some(goal_id) = goal_id {
        byggare.push(" AND ar.goal_id = ").push_bind(goal_id);
    }

    byggare.push(" ORDER BY ar.created_at DESC");

    let rader = byggare
        .build()
        .fetch_all(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to query agent runs: {err}")))?;

    Ok(rader.iter().map(läs_körning).collect())
}

/// En enskild körning. `None` om den inte finns eller tillhör annan.
pub async fn get(pool: &PgPool, user_id: i64, run_id: i64) -> Result<Option<AgentRun>, AppError> {
    let fråga = format!(
        "SELECT {KOLUMNER} FROM agent_runs ar \
         INNER JOIN user_goals ug ON ug.goal_id = ar.goal_id \
         WHERE ug.user_id = $1 AND ar.id = $2"
    );

    let rad = sqlx::query(&fråga)
        .bind(user_id)
        .bind(run_id)
        .fetch_optional(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to get agent run: {err}")))?;

    Ok(rad.as_ref().map(läs_körning))
}

/// Skapar en agentkörning för ett mål.
///
/// Målet måste tillhöra användaren - annars avbryts skapandet.
///
/// Budgetkontrollen sker i samma transaktion som sparningen. Därför kan
/// två körningar som startar samtidigt inte tillsammans överskrida en
/// budget - den andra ser den första körningens reservation.
pub async fn create(
    pool: &PgPool,
    user_id: i64,
    req: &CreateAgentRunRequest,
) -> Result<(AgentRun, Option<BudgetStatus>), AppError> {
    let goal_id = req
        .goal_id
        .ok_or_else(|| AppError::BadRequest("goal_id is required".to_string()))?;

    if !målet_tillhör(pool, user_id, goal_id).await? {
        return Err(AppError::NotFound(format!(
            "goal {goal_id} does not belong to user"
        )));
    }

    let phase = if req.phase.is_empty() {
        agent::PHASE_DEFINE.to_string()
    } else {
        req.phase.clone()
    };

    let status = if req.status.is_empty() {
        agent::STATUS_QUEUED.to_string()
    } else {
        req.status.clone()
    };

    let trigger = if req.trigger.is_empty() {
        "user".to_string()
    } else {
        req.trigger.clone()
    };

    let agent_profile = if req.agent_profile.is_empty() {
        agent::PROFILE_CODING.to_string()
    } else {
        req.agent_profile.clone()
    };

    let mut transaktion = pool
        .begin()
        .await
        .map_err(|err| AppError::Internal(format!("failed to start transaction: {err}")))?;

    let (reservation, warning) = kontrollera_budgetar(
        &mut transaktion,
        user_id,
        req,
        &agent::normalisera_agent_id(&req.agent_id),
    )
    .await?;

    let fråga = "INSERT INTO agent_runs (goal_id, phase, status, trigger, initiator, agent_profile, summary, result, \
         artifact_url, conversation_id, started_at, agent_id, provider, model, model_version, reserved_cost_usd) \
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, COALESCE($11, NOW()), $12, $13, $14, $15, $16) \
         RETURNING id, goal_id, phase, status, COALESCE(trigger, 'user'), COALESCE(initiator, ''), \
         COALESCE(agent_profile, 'coding'), COALESCE(summary, ''), COALESCE(result, ''), \
         artifact_url, cost_usd, conversation_id, started_at, finished_at, created_at, updated_at, \
         COALESCE(agent_id, ''), COALESCE(provider, ''), COALESCE(model, ''), COALESCE(model_version, ''), \
         COALESCE(reserved_cost_usd, 0)".to_string();

    let rad = sqlx::query(&fråga)
        .bind(goal_id)
        .bind(&phase)
        .bind(&status)
        .bind(&trigger)
        .bind(&req.initiator)
        .bind(&agent_profile)
        .bind(&req.summary)
        .bind(&req.result)
        .bind(&req.artifact_url)
        .bind(&req.conversation_id)
        .bind(req.started_at)
        .bind(agent::normalisera_agent_id(&req.agent_id))
        .bind(&req.provider)
        .bind(&req.model)
        .bind(&req.model_version)
        .bind(reservation)
        .fetch_one(&mut *transaktion)
        .await
        .map_err(|err| AppError::Internal(format!("failed to create agent run: {err}")))?;

    transaktion
        .commit()
        .await
        .map_err(|err| AppError::Internal(format!("failed to commit agent run: {err}")))?;

    Ok((läs_körning(&rad), warning))
}

/// Kontrollerar budgeten och reserverar plats för en körning.
///
/// Budgetraderna låses, så att en parallell körning väntar. Den som får
/// låsen först reserverar sin plats - den andra ser reservationen.
///
/// Returnerar reservationen och en varning om varningsgränsen är nådd.
async fn kontrollera_budgetar(
    transaktion: &mut Transaction<'_, Postgres>,
    user_id: i64,
    req: &CreateAgentRunRequest,
    agent_id: &str,
) -> Result<(f64, Option<BudgetStatus>), AppError> {
    let budgetar = lås_budgetar(transaktion, user_id, agent_id).await?;

    if budgetar.is_empty() {
        return Ok((0.0, None));
    }

    // Reservationen är den största av budgetarnas - den körning som
    // startar reserverar så mycket som den tightaste budgeten kräver.
    let reservation = budgetar
        .iter()
        .map(|budget| budget.reservation_per_run_usd)
        .fold(0.0_f64, f64::max);

    let mut varning: Option<BudgetStatus> = None;

    for budget in &budgetar {
        let mut status = budget_status(&mut *transaktion, user_id, budget, reservation).await?;

        // En hård budget kräver att modellens pris är känt. Annars går
        // det inte att veta om körningen ryms - då stoppas den hellre
        // än att köra mot en okänd kostnad.
        if budget.hard_limit_usd > 0.0 {
            let (_, känd) = PriceList::current().lookup(&req.provider, &req.model);
            if !känd {
                status.reason = "unknown_price".to_string();
                return Err(budgetfel(status));
            }
        }

        if status.exceeded {
            status.reason = "limit_reached".to_string();
            return Err(budgetfel(status));
        }

        if status.warning {
            varning = Some(status);
        }
    }

    Ok((reservation, varning))
}

/// Ett budgetfel - samma format som Go-tjänstens svar.
fn budgetfel(status: BudgetStatus) -> AppError {
    let message = agent::BudgetExceeded {
        status: status.clone(),
    }
    .to_string();

    AppError::Conflict {
        message,
        details: serde_json::to_value(&status).unwrap_or_default(),
    }
}

/// Låser användarens aktiva budgetar.
async fn lås_budgetar(
    transaktion: &mut Transaction<'_, Postgres>,
    user_id: i64,
    agent_id: &str,
) -> Result<Vec<AgentBudget>, AppError> {
    let rader = sqlx::query(
        "SELECT id, name, period, hard_limit_usd, warning_limit_usd, \
         reservation_per_run_usd, agent_id, active, created_at \
         FROM agent_budgets \
         WHERE user_id = $1 AND active AND (agent_id = '' OR agent_id = $2) \
         FOR UPDATE",
    )
    .bind(user_id)
    .bind(agent_id)
    .fetch_all(&mut **transaktion)
    .await
    .map_err(|err| AppError::Internal(format!("failed to lock agent budgets: {err}")))?;

    Ok(rader.iter().map(läs_budget).collect())
}

/// Uppdaterar en agentkörning.
///
/// Endast de fält som sätts i `req` ändras - resten lämnas som de var.
pub async fn update(
    pool: &PgPool,
    user_id: i64,
    run_id: i64,
    req: &UpdateAgentRunRequest,
) -> Result<Option<AgentRun>, AppError> {
    // QueryBuilder sätter ihop argumenten utan mellanrum, så varje
    // textstycke börjar med ett blankstug, och fälten skiljs med komma.
    let mut byggare = QueryBuilder::<Postgres>::new("UPDATE agent_runs SET");
    let mut antal = 0_usize;

    // Sätter ett fält med ett värde.
    macro_rules! satt {
        ($text:expr, $varde:expr) => {{
            if antal > 0 {
                byggare.push(",");
            }
            byggare.push($text).push_bind($varde);
            antal += 1;
        }};
    }

    // Sätter ett fält utan värde - till exempel NOW().
    macro_rules! satt_utan {
        ($text:expr) => {{
            if antal > 0 {
                byggare.push(",");
            }
            byggare.push($text);
        }};
    }

    if let Some(phase) = &req.phase {
        satt!(" phase = ", phase.clone());
    }
    if let Some(status) = &req.status {
        satt!(" status = ", status.clone());
        // En terminal körning får en sluttid - annars går den inte att
        // mäta hur länge den tog.
        if req.finished_at.is_none() && agent::är_terminal(status) {
            satt_utan!(" finished_at = NOW()");
        }
    }
    if let Some(trigger) = &req.trigger {
        satt!(" trigger = ", trigger.clone());
    }
    if let Some(initiator) = &req.initiator {
        satt!(" initiator = ", initiator.clone());
    }
    if let Some(agent_profile) = &req.agent_profile {
        satt!(" agent_profile = ", agent_profile.clone());
    }
    if let Some(result) = &req.result {
        satt!(" result = ", result.clone());
    }
    if let Some(started_at) = &req.started_at {
        satt!(" started_at = ", *started_at);
    }
    if let Some(finished_at) = &req.finished_at {
        satt!(" finished_at = ", *finished_at);
    }
    if let Some(summary) = &req.summary {
        satt!(" summary = ", summary.clone());
    }
    if let Some(artifact_url) = &req.artifact_url {
        satt!(" artifact_url = ", artifact_url.clone());
    }
    if let Some(cost_usd) = &req.cost_usd {
        satt!(" cost_usd = ", *cost_usd);
    }
    if let Some(conversation_id) = &req.conversation_id {
        satt!(" conversation_id = ", conversation_id.clone());
    }
    if let Some(agent_id) = &req.agent_id {
        satt!(" agent_id = ", agent_id.clone());
    }
    if let Some(provider) = &req.provider {
        satt!(" provider = ", provider.clone());
    }
    if let Some(model) = &req.model {
        satt!(" model = ", model.clone());
    }
    if let Some(model_version) = &req.model_version {
        satt!(" model_version = ", model_version.clone());
    }

    if antal == 0 {
        return get(pool, user_id, run_id).await;
    }

    satt_utan!(" updated_at = NOW()");
    byggare
        .push(" WHERE id = ")
        .push_bind(run_id)
        .push(" AND goal_id IN (SELECT goal_id FROM user_goals WHERE user_id = ")
        .push_bind(user_id)
        .push(") RETURNING id, goal_id, phase, status, COALESCE(trigger, 'user'), COALESCE(initiator, ''), \
         COALESCE(agent_profile, 'coding'), COALESCE(summary, ''), COALESCE(result, ''), \
         artifact_url, cost_usd, conversation_id, started_at, finished_at, created_at, updated_at, \
         COALESCE(agent_id, ''), COALESCE(provider, ''), COALESCE(model, ''), COALESCE(model_version, ''), \
         COALESCE(reserved_cost_usd, 0)");

    let rad = byggare
        .build()
        .fetch_optional(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to update agent run: {err}")))?;

    Ok(rad.as_ref().map(läs_körning))
}

/// Om ett mål tillhör användaren.
async fn målet_tillhör(pool: &PgPool, user_id: i64, goal_id: i64) -> Result<bool, AppError> {
    let finns: bool = sqlx::query_scalar(
        "SELECT EXISTS(SELECT 1 FROM user_goals WHERE user_id = $1 AND goal_id = $2)",
    )
    .bind(user_id)
    .bind(goal_id)
    .fetch_one(pool)
    .await
    .map_err(|err| AppError::Internal(format!("failed to check goal access: {err}")))?;

    Ok(finns)
}

/// Läser en budgetrad.
pub fn läs_budget(rad: &PgRow) -> AgentBudget {
    AgentBudget {
        id: rad.get::<i64, _>(0),
        name: rad.get::<String, _>(1),
        period: rad.get::<String, _>(2),
        hard_limit_usd: rad.get::<f64, _>(3),
        warning_limit_usd: rad.get::<Option<f64>, _>(4),
        reservation_per_run_usd: rad.get::<f64, _>(5),
        agent_id: rad.get::<String, _>(6),
        active: rad.get::<bool, _>(7),
        created_at: rad.get::<chrono::DateTime<chrono::Utc>, _>(8),
    }
}

/// Räknar ut en budgets läge.
///
/// Spent är den kostnad som redan bokförts i perioden. Reserved är den
/// kostnad som reserverats för körningar som fortfarande pågår.
pub async fn budget_status(
    anslutning: &mut PgConnection,
    user_id: i64,
    budget: &AgentBudget,
    pending: f64,
) -> Result<BudgetStatus, AppError> {
    let start = period_start(&budget.period, chrono::Utc::now());
    let nu = chrono::Utc::now();

    let spent: f64 = sqlx::query_scalar(
        "SELECT COALESCE(SUM(u.cost_usd), 0) \
         FROM agent_run_usage u \
         INNER JOIN user_goals ug ON ug.goal_id = u.goal_id \
         WHERE ug.user_id = $1 AND u.created_at >= $2 AND u.created_at <= $3",
    )
    .bind(user_id)
    .bind(start)
    .bind(nu)
    .fetch_one(&mut *anslutning)
    .await
    .map_err(|err| AppError::Internal(format!("failed to sum spent cost: {err}")))?;

    let reserved: f64 = sqlx::query_scalar(
        "SELECT COALESCE(SUM(ar.reserved_cost_usd), 0) \
         FROM agent_runs ar \
         INNER JOIN user_goals ug ON ug.goal_id = ar.goal_id \
         WHERE ug.user_id = $1 \
         AND ar.status IN ('queued', 'running', 'waiting_approval') \
         AND ar.reserved_cost_usd > 0 \
         AND ($2 = '' OR ar.agent_id = $2 OR ar.agent_id = '')",
    )
    .bind(user_id)
    .bind(&budget.agent_id)
    .fetch_one(&mut *anslutning)
    .await
    .map_err(|err| AppError::Internal(format!("failed to sum reserved cost: {err}")))?;

    let mut status = BudgetStatus {
        budget_id: budget.id,
        name: budget.name.clone(),
        period: budget.period.clone(),
        hard_limit_usd: budget.hard_limit_usd,
        spent_usd: spent,
        reserved_usd: reserved,
        exceeded: false,
        warning: false,
        reason: String::new(),
    };

    // Den väntande körningens reservation räknas med - annars kan två
    // samtidiga körningar tillsammans överskrida gränsen.
    status.exceeded = spent + reserved + pending >= budget.hard_limit_usd;
    if let Some(warning_limit) = budget.warning_limit_usd {
        status.warning = spent + reserved + pending >= warning_limit;
    }

    Ok(status)
}

/// Periodens början för en budget.
pub fn period_start(
    period: &str,
    now: chrono::DateTime<chrono::Utc>,
) -> chrono::DateTime<chrono::Utc> {
    use chrono::Datelike;

    match period {
        agent::BUDGET_PERIOD_DAY => now
            .date_naive()
            .and_hms_opt(0, 0, 0)
            .map(|naive| naive.and_utc())
            .unwrap_or(now),
        agent::BUDGET_PERIOD_WEEK => {
            // Veckans början är måndagen.
            let veckodag = now.weekday().num_days_from_monday() as i64;
            let start = now.date_naive() - chrono::Duration::days(veckodag);
            start
                .and_hms_opt(0, 0, 0)
                .map(|naive| naive.and_utc())
                .unwrap_or(now)
        }
        _ => chrono::NaiveDate::from_ymd_opt(now.year(), now.month(), 1)
            .and_then(|första| första.and_hms_opt(0, 0, 0))
            .map(|naive| naive.and_utc())
            .unwrap_or(now),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn dagsperioden_börjar_på_midnatt() {
        let nu = chrono::DateTime::parse_from_rfc3339("2026-11-24T13:45:12Z")
            .unwrap()
            .with_timezone(&chrono::Utc);

        let start = period_start(agent::BUDGET_PERIOD_DAY, nu);

        assert_eq!(start.to_rfc3339(), "2026-11-24T00:00:00+00:00");
    }

    #[test]
    fn veckoperioden_börjar_pa_mandagen() {
        // 2026-11-24 är en tisdag.
        let nu = chrono::DateTime::parse_from_rfc3339("2026-11-24T13:45:12Z")
            .unwrap()
            .with_timezone(&chrono::Utc);

        let start = period_start(agent::BUDGET_PERIOD_WEEK, nu);

        assert_eq!(start.to_rfc3339(), "2026-11-23T00:00:00+00:00");
    }

    #[test]
    fn manadsperioden_borjar_forsta_dagen() {
        let nu = chrono::DateTime::parse_from_rfc3339("2026-11-24T13:45:12Z")
            .unwrap()
            .with_timezone(&chrono::Utc);

        let start = period_start(agent::BUDGET_PERIOD_MONTH, nu);

        assert_eq!(start.to_rfc3339(), "2026-11-01T00:00:00+00:00");
    }
}
