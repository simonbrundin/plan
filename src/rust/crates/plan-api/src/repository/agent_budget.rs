//! Budgetgränser för agentkostnad.
//!
//! Port av `src/api/internal/repository/agent_budget_repository.go`.
//!
//! En budget har en hård gräns som stoppar nya körningar och en
//! varningsgräns som bara varnar. Kontrollen och reservationen sker i
//! samma transaktion som körningens skapande, så att samtidiga körningar
//! inte tillsammans kan överskrida gränsen.

use sqlx::postgres::PgRow;
use sqlx::PgPool;

use crate::error::AppError;
use crate::models::agent::{AgentBudget, CreateAgentBudgetRequest, UpdateAgentBudgetRequest};

/// Kolumnerna som läses för en budget.
const KOLUMNER: &str = "id, name, period, hard_limit_usd, warning_limit_usd, \
     reservation_per_run_usd, agent_id, active, created_at";

/// Läser en budgetrad.
fn läs_rad(rad: &PgRow) -> AgentBudget {
    crate::repository::agent_run::läs_budget(rad)
}

/// Användarens budgetar.
pub async fn list(pool: &PgPool, user_id: i64) -> Result<Vec<AgentBudget>, AppError> {
    let fråga =
        format!("SELECT {KOLUMNER} FROM agent_budgets WHERE user_id = $1 ORDER BY created_at DESC");

    let rader = sqlx::query(&fråga)
        .bind(user_id)
        .fetch_all(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to list agent budgets: {err}")))?;

    Ok(rader.iter().map(läs_rad).collect())
}

/// Sparar en ny budget.
pub async fn create(
    pool: &PgPool,
    user_id: i64,
    req: &CreateAgentBudgetRequest,
) -> Result<AgentBudget, AppError> {
    let reservation = if req.reservation_per_run_usd <= 0.0 {
        1.0
    } else {
        req.reservation_per_run_usd
    };

    let fråga = format!(
        "INSERT INTO agent_budgets (user_id, name, period, hard_limit_usd, warning_limit_usd, \
         reservation_per_run_usd, agent_id, active) \
         VALUES ($1, $2, $3, $4, $5, $6, $7, TRUE) \
         RETURNING {KOLUMNER}"
    );

    let rad = sqlx::query(&fråga)
        .bind(user_id)
        .bind(&req.name)
        .bind(&req.period)
        .bind(req.hard_limit_usd)
        .bind(req.warning_limit_usd)
        .bind(reservation)
        .bind(crate::models::agent::normalisera_agent_id(&req.agent_id))
        .fetch_one(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to create agent budget: {err}")))?;

    Ok(läs_rad(&rad))
}

/// Ändrar en budget. Bara de fält som skickas ändras.
///
/// En budget som nått sin gräns måste kunna pausas, annars kan agenten
/// inte köra igen förrän perioden går ut.
pub async fn update(
    pool: &PgPool,
    user_id: i64,
    budget_id: i64,
    req: &UpdateAgentBudgetRequest,
) -> Result<AgentBudget, AppError> {
    let fråga = format!(
        "UPDATE agent_budgets SET \
         name = COALESCE($3, name), \
         period = COALESCE($4, period), \
         hard_limit_usd = COALESCE($5, hard_limit_usd), \
         warning_limit_usd = CASE WHEN $6::boolean THEN $7 ELSE warning_limit_usd END, \
         reservation_per_run_usd = COALESCE($8, reservation_per_run_usd), \
         active = COALESCE($9, active) \
         WHERE id = $1 AND user_id = $2 \
         RETURNING {KOLUMNER}"
    );

    let rad = sqlx::query(&fråga)
        .bind(budget_id)
        .bind(user_id)
        .bind(req.name.clone())
        .bind(req.period.clone())
        .bind(req.hard_limit_usd)
        .bind(req.warning_limit_usd.is_some())
        .bind(req.warning_limit_usd)
        .bind(req.reservation_per_run_usd)
        .bind(req.active)
        .fetch_optional(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to update agent budget: {err}")))?;

    rad.map(|rad| läs_rad(&rad))
        .ok_or_else(|| AppError::NotFound("budgeten hittades inte".to_string()))
}

/// Tar bort en budget.
pub async fn delete(pool: &PgPool, user_id: i64, budget_id: i64) -> Result<(), AppError> {
    let resultat = sqlx::query("DELETE FROM agent_budgets WHERE id = $1 AND user_id = $2")
        .bind(budget_id)
        .bind(user_id)
        .execute(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to delete agent budget: {err}")))?;

    if resultat.rows_affected() == 0 {
        return Err(AppError::NotFound("budgeten hittades inte".to_string()));
    }

    Ok(())
}
