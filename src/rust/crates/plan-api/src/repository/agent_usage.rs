//! Användning och kostnad för agentkörningar.
//!
//! Port av `src/api/internal/repository/agent_usage_repository.go`.
//!
//! Kostnaden räknas fram i Plan, inte i agenten. Då sparas samma prislista
//! för alla körningar, och kostnaden kan reproduceras senare.

use sqlx::postgres::PgRow;
use sqlx::{PgPool, Postgres, Row};

use crate::error::AppError;
use crate::models::agent::{
    AgentRunUsage, AgentUsageByModel, AgentUsageSummary, CreateAgentRunUsageRequest,
};
use crate::pricing::{PriceList, Tokens};

/// Kolumnerna som läses för en användningsrad.
const KOLUMNER: &str = "u.id, u.agent_run_id, u.goal_id, u.agent_id, u.provider, u.model, \
     u.model_version, u.input_tokens, u.output_tokens, u.cache_read_tokens, u.tool_calls, \
     u.tool_names, u.cost_usd, u.currency, u.price_list_version, u.price_source, u.status, \
     u.started_at, u.finished_at, u.created_at";

/// Samma kolumner, utan tabellprefix - de används i RETURNING.
const RETURNING: &str = "id, agent_run_id, goal_id, agent_id, provider, model, \
     model_version, input_tokens, output_tokens, cache_read_tokens, tool_calls, \
     tool_names, cost_usd, currency, price_list_version, price_source, status, \
     started_at, finished_at, created_at";

/// Läser en användningsrad.
fn läs_användning(rad: &PgRow) -> AgentRunUsage {
    AgentRunUsage {
        id: rad.get::<i64, _>(0),
        agent_run_id: rad.get::<i32, _>(1) as i64,
        goal_id: rad.get::<i32, _>(2) as i64,
        agent_id: rad.get::<String, _>(3),
        provider: rad.get::<String, _>(4),
        model: rad.get::<String, _>(5),
        model_version: rad.get::<String, _>(6),
        input_tokens: rad.get::<i32, _>(7) as i64,
        output_tokens: rad.get::<i32, _>(8) as i64,
        cache_read_tokens: rad.get::<i32, _>(9) as i64,
        tool_calls: rad.get::<i32, _>(10) as i64,
        tool_names: rad.get::<Option<Vec<String>>, _>(11),
        cost_usd: rad.get::<f64, _>(12),
        currency: rad.get::<String, _>(13),
        price_list_version: rad.get::<String, _>(14),
        price_source: rad.get::<String, _>(15),
        status: rad.get::<String, _>(16),
        started_at: rad.get::<Option<chrono::DateTime<chrono::Utc>>, _>(17),
        finished_at: rad.get::<Option<chrono::DateTime<chrono::Utc>>, _>(18),
        created_at: rad.get::<chrono::DateTime<chrono::Utc>, _>(19),
    }
}

/// Sparar en användningsrad för en körning.
///
/// Kostnaden räknas fram ur den aktuella prislistan, och listans version
/// sparas med raden. Om modellen saknas i listan används schablonen, och
/// källan markeras som "fallback" - kostnaden ska inte läsas som exakt.
pub async fn record(
    pool: &PgPool,
    user_id: i64,
    run_id: i64,
    req: &CreateAgentRunUsageRequest,
) -> Result<Option<AgentRunUsage>, AppError> {
    let Some((goal_id, run_provider, run_model)) = körningsuppgifter(pool, user_id, run_id).await?
    else {
        return Ok(None);
    };

    // En tom lista ska sparas som en tom array, inte som null.
    let tool_names = req.tool_names.clone().unwrap_or_default();

    // Agentens uppgifter vinner över körningens - de kommer från den
    // som faktiskt körde. Körningens uppgifter används som reserv.
    let provider = if req.provider.is_empty() {
        run_provider
    } else {
        req.provider.clone()
    };

    let model = if req.model.is_empty() {
        run_model
    } else {
        req.model.clone()
    };

    let prislista = PriceList::current();
    let resultat = prislista.cost(
        &provider,
        &model,
        Tokens {
            input: req.input_tokens,
            output: req.output_tokens,
            cache_read: req.cache_read_tokens,
        },
    );

    let fråga = format!(
        "INSERT INTO agent_run_usage (agent_run_id, goal_id, agent_id, provider, model, model_version, \
         input_tokens, output_tokens, cache_read_tokens, tool_calls, tool_names, \
         cost_usd, currency, price_list_version, price_source, status, started_at, finished_at) \
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18) \
         ON CONFLICT (agent_run_id) DO UPDATE SET \
         agent_id = EXCLUDED.agent_id, \
         provider = EXCLUDED.provider, \
         model = EXCLUDED.model, \
         model_version = EXCLUDED.model_version, \
         input_tokens = EXCLUDED.input_tokens, \
         output_tokens = EXCLUDED.output_tokens, \
         cache_read_tokens = EXCLUDED.cache_read_tokens, \
         tool_calls = EXCLUDED.tool_calls, \
         tool_names = EXCLUDED.tool_names, \
         cost_usd = EXCLUDED.cost_usd, \
         currency = EXCLUDED.currency, \
         price_list_version = EXCLUDED.price_list_version, \
         price_source = EXCLUDED.price_source, \
         status = EXCLUDED.status, \
         started_at = EXCLUDED.started_at, \
         finished_at = EXCLUDED.finished_at \
         RETURNING {RETURNING}"
    );

    let rad = sqlx::query(&fråga)
        .bind(run_id)
        .bind(goal_id)
        .bind(&req.agent_id)
        .bind(&provider)
        .bind(&model)
        .bind(&req.model_version)
        .bind(req.input_tokens)
        .bind(req.output_tokens)
        .bind(req.cache_read_tokens)
        .bind(req.tool_calls)
        .bind(&tool_names)
        .bind(resultat.cost_usd)
        .bind(&resultat.currency)
        .bind(&resultat.price_list_version)
        .bind(&resultat.source)
        .bind(crate::models::agent::USAGE_STATUS_REPORTED)
        .bind(req.started_at)
        .bind(req.finished_at)
        .fetch_one(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to record agent usage: {err}")))?;

    // Körningens kostnad speglas i agent_runs, så att listan snabbt kan
    // visa vad en körning kostade.
    sqlx::query("UPDATE agent_runs SET cost_usd = $1, updated_at = NOW() WHERE id = $2")
        .bind(resultat.cost_usd)
        .bind(run_id)
        .execute(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to update run cost: {err}")))?;

    Ok(Some(läs_användning(&rad)))
}

/// Skapar en ofullständig rad om körningen saknar en.
///
/// En körning som avbryts innan agenten hann rapportera ska ändå ha en
/// kostnadsrad, så att det går att se att något saknas i stället för att
/// raden tyst försvinner. Raden markeras som "incomplete".
pub async fn ensure_row(pool: &PgPool, user_id: i64, run_id: i64) -> Result<(), AppError> {
    let prislista = PriceList::current();

    sqlx::query(
        "INSERT INTO agent_run_usage (agent_run_id, goal_id, status, currency, price_list_version, price_source) \
         SELECT ar.id, ar.goal_id, $3, $4, $5, $6 \
         FROM agent_runs ar \
         WHERE ar.id = $1 \
         AND ar.goal_id IN (SELECT goal_id FROM user_goals WHERE user_id = $2) \
         ON CONFLICT (agent_run_id) DO NOTHING",
    )
    .bind(run_id)
    .bind(user_id)
    .bind(crate::models::agent::USAGE_STATUS_INCOMPLETE)
    .bind(&prislista.currency)
    .bind(&prislista.version)
    .bind(crate::pricing::SOURCE_PRICE_LIST)
    .execute(pool)
    .await
    .map_err(|err| AppError::Internal(format!("failed to ensure usage row: {err}")))?;

    Ok(())
}

/// Användningen för en körning.
pub async fn get_by_run(
    pool: &PgPool,
    user_id: i64,
    run_id: i64,
) -> Result<Option<AgentRunUsage>, AppError> {
    let fråga = format!(
        "SELECT {KOLUMNER} FROM agent_run_usage u \
         INNER JOIN agent_runs ar ON ar.id = u.agent_run_id \
         WHERE u.agent_run_id = $1 \
         AND ar.goal_id IN (SELECT goal_id FROM user_goals WHERE user_id = $2)"
    );

    let rad = sqlx::query(&fråga)
        .bind(run_id)
        .bind(user_id)
        .fetch_optional(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to get agent usage: {err}")))?;

    Ok(rad.as_ref().map(läs_användning))
}

/// Användningsrader för användaren, med filter.
pub async fn list(
    pool: &PgPool,
    user_id: i64,
    filter: &UsageFilter,
) -> Result<Vec<AgentRunUsage>, AppError> {
    let mut byggare = sqlx::QueryBuilder::<Postgres>::new(format!(
        "SELECT {KOLUMNER} FROM agent_run_usage u \
         INNER JOIN user_goals ug ON ug.goal_id = u.goal_id \
         WHERE ug.user_id = "
    ));
    byggare.push_bind(user_id);

    if !filter.agent_id.is_empty() {
        byggare
            .push(" AND u.agent_id = ")
            .push_bind(filter.agent_id.clone());
    }
    if !filter.model.is_empty() {
        byggare
            .push(" AND u.model = ")
            .push_bind(filter.model.clone());
    }
    if !filter.tool.is_empty() {
        byggare
            .push(" AND ")
            .push_bind(filter.tool.clone())
            .push(" = ANY(u.tool_names)");
    }
    if let Some(from) = filter.from {
        byggare.push(" AND u.created_at >= ").push_bind(from);
    }
    if let Some(to) = filter.to {
        byggare.push(" AND u.created_at <= ").push_bind(to);
    }

    byggare.push(" ORDER BY u.created_at DESC");

    let rader = byggare
        .build()
        .fetch_all(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to list agent usage: {err}")))?;

    Ok(rader.iter().map(läs_användning).collect())
}

/// Filtret för kostnadslistan.
#[derive(Debug, Clone, Default)]
pub struct UsageFilter {
    /// Gäller en enskild agent.
    pub agent_id: String,
    /// Gäller en enskild modell.
    pub model: String,
    /// Gäller körningar som använt ett verktyg.
    pub tool: String,
    /// Periodens början.
    pub from: Option<chrono::DateTime<chrono::Utc>>,
    /// Periodens slut.
    pub to: Option<chrono::DateTime<chrono::Utc>>,
}

/// Summerar användningen för ett filter.
pub async fn summary(
    pool: &PgPool,
    user_id: i64,
    filter: &UsageFilter,
) -> Result<AgentUsageSummary, AppError> {
    let rader = list(pool, user_id, filter).await?;

    let mut summary = AgentUsageSummary {
        runs: rader.len() as i64,
        total_cost_usd: 0.0,
        input_tokens: 0,
        output_tokens: 0,
        cache_read_tokens: 0,
        tool_calls: 0,
        currency: PriceList::current().currency,
        by_model: None,
    };

    let mut per_modell: Vec<(String, AgentUsageByModel)> = Vec::new();

    for rad in &rader {
        summary.total_cost_usd += rad.cost_usd;
        summary.input_tokens += rad.input_tokens;
        summary.output_tokens += rad.output_tokens;
        summary.cache_read_tokens += rad.cache_read_tokens;
        summary.tool_calls += rad.tool_calls;

        let nyckel = format!("{}/{}", rad.provider, rad.model);
        match per_modell.iter_mut().find(|(k, _)| *k == nyckel) {
            Some((_, post)) => {
                post.runs += 1;
                post.cost_usd += rad.cost_usd;
            }
            None => {
                per_modell.push((
                    nyckel,
                    AgentUsageByModel {
                        provider: rad.provider.clone(),
                        model: rad.model.clone(),
                        runs: 1,
                        cost_usd: rad.cost_usd,
                    },
                ));
            }
        }
    }

    if !per_modell.is_empty() {
        summary.by_model = Some(per_modell.into_iter().map(|(_, post)| post).collect());
    }

    Ok(summary)
}

/// Körningens uppgifter som behövs för att prissätta.
///
/// `None` betyder att körningen inte finns eller tillhör annan.
async fn körningsuppgifter(
    pool: &PgPool,
    user_id: i64,
    run_id: i64,
) -> Result<Option<(i64, String, String)>, AppError> {
    let fråga = "SELECT ar.goal_id, COALESCE(ar.provider, ''), COALESCE(ar.model, '') \
         FROM agent_runs ar \
         WHERE ar.id = $1 \
         AND ar.goal_id IN (SELECT goal_id FROM user_goals WHERE user_id = $2)";

    let rad = sqlx::query(fråga)
        .bind(run_id)
        .bind(user_id)
        .fetch_optional(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to get agent run: {err}")))?;

    Ok(rad.map(|rad| {
        (
            rad.get::<i32, _>(0) as i64,
            rad.get::<String, _>(1),
            rad.get::<String, _>(2),
        )
    }))
}
