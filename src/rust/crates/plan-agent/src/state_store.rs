//! Agentens tillstånd - checkpointplaner och deras resultat.
//!
//! Go-agentens `internal/core/orchestrator/state.go` och
//! `checkpoint_store.go`. Agenten sparar planerna i databasen så att de
//! överlever en omstart.

use std::collections::HashMap;
use std::sync::{Arc, RwLock};

use chrono::{DateTime, Utc};
use sqlx::PgPool;

use crate::checkpoint::{Checkpoint, CheckpointPlan, GateResult};

/// Fel från tillståndslagret.
#[derive(Debug, thiserror::Error)]
pub enum StateError {
    /// Planen eller checkpointen finns inte.
    #[error("plan not found: {0}")]
    NotFound(String),

    /// Ett fel med en egen meddelande.
    #[error("{0}")]
    Message(String),
    /// Databasen svarade med ett fel.
    #[error(transparent)]
    Database(#[from] sqlx::Error),
}

/// Sparar och läser checkpointplaner.
///
/// Go har ett gränssnitt för detta. Rust använder en enum i stället -
/// då går det att dela samma lager mellan flera utan att bryta om
/// asynkroniteten.
#[derive(Clone)]
pub enum StateStore {
    /// Tillståndet i minnet.
    Memory(InMemoryStateStore),
    /// Tillståndet i databasen.
    Database(CheckpointStore),
}

impl StateStore {
    /// Sparar en plan och dess checkpoints.
    pub async fn save_plan(&self, plan: &CheckpointPlan) -> Result<(), StateError> {
        match self {
            StateStore::Memory(lager) => lager.save_plan(plan),
            StateStore::Database(lager) => lager.save_plan(plan).await,
        }
    }

    /// Läser en plan med sina checkpoints.
    pub async fn get_plan(&self, id: &str) -> Result<CheckpointPlan, StateError> {
        match self {
            StateStore::Memory(lager) => lager.get_plan(id),
            StateStore::Database(lager) => lager.get_plan(id).await,
        }
    }

    /// Läser alla planer, nyast först.
    pub async fn get_all_plans(&self) -> Result<Vec<CheckpointPlan>, StateError> {
        match self {
            StateStore::Memory(lager) => lager.get_all_plans(),
            StateStore::Database(lager) => lager.get_all_plans().await,
        }
    }

    /// Tar bort en plan.
    pub async fn delete_plan(&self, id: &str) -> Result<(), StateError> {
        match self {
            StateStore::Memory(lager) => lager.delete_plan(id),
            StateStore::Database(lager) => lager.delete_plan(id).await,
        }
    }

    /// Sparar en checkpoint.
    pub async fn save_checkpoint(&self, checkpoint: &Checkpoint) -> Result<(), StateError> {
        match self {
            StateStore::Memory(lager) => lager.save_checkpoint(checkpoint),
            StateStore::Database(lager) => lager.save_checkpoint(checkpoint).await,
        }
    }

    /// Läser en checkpoint.
    pub async fn get_checkpoint(&self, id: &str) -> Result<Checkpoint, StateError> {
        match self {
            StateStore::Memory(lager) => lager.get_checkpoint(id),
            StateStore::Database(lager) => lager.get_checkpoint(id).await,
        }
    }

    /// Läser checkpointsen i en plan.
    pub async fn get_checkpoints_by_plan(
        &self,
        plan_id: &str,
    ) -> Result<Option<Vec<Checkpoint>>, StateError> {
        match self {
            StateStore::Memory(lager) => lager.get_checkpoints_by_plan(plan_id),
            StateStore::Database(lager) => lager.get_checkpoints_by_plan(plan_id).await,
        }
    }

    /// Sparar resultatet av en gate.
    pub async fn save_gate_result(
        &self,
        checkpoint_id: &str,
        result: &GateResult,
    ) -> Result<(), StateError> {
        match self {
            StateStore::Memory(lager) => lager.save_gate_result(checkpoint_id, result),
            StateStore::Database(lager) => lager.save_gate_result(checkpoint_id, result).await,
        }
    }

    /// Läser resultatet av en gate.
    pub async fn get_gate_result(
        &self,
        checkpoint_id: &str,
        gate_type: &str,
    ) -> Result<GateResult, StateError> {
        match self {
            StateStore::Memory(lager) => lager.get_gate_result(checkpoint_id, gate_type),
            StateStore::Database(lager) => lager.get_gate_result(checkpoint_id, gate_type).await,
        }
    }
}

/// Tillståndet i minnet.
///
/// Används när agenten inte har någon databas - och i testerna.
#[derive(Default)]
pub struct InMemoryStateStore {
    planer: RwLock<HashMap<String, CheckpointPlan>>,
    checkpoints: RwLock<HashMap<String, Checkpoint>>,
    gate_results: RwLock<HashMap<String, HashMap<String, GateResult>>>,
}

impl Clone for InMemoryStateStore {
    /// Delar kartorna - de är kopior, precis som Go:s deepcopy.
    fn clone(&self) -> Self {
        Self {
            planer: RwLock::new(self.planer().clone()),
            checkpoints: RwLock::new(self.checkpoints().clone()),
            gate_results: RwLock::new(self.gate_results().clone()),
        }
    }
}

impl InMemoryStateStore {
    /// Skapar ett tomt lager.
    pub fn new() -> Self {
        Self::default()
    }

    /// Låser planerna.
    fn planer(&self) -> std::sync::RwLockReadGuard<'_, HashMap<String, CheckpointPlan>> {
        self.planer
            .read()
            .unwrap_or_else(|poisoned| poisoned.into_inner())
    }

    /// Låser checkpointsen.
    fn checkpoints(&self) -> std::sync::RwLockReadGuard<'_, HashMap<String, Checkpoint>> {
        self.checkpoints
            .read()
            .unwrap_or_else(|poisoned| poisoned.into_inner())
    }

    /// Låser gate-resultaten.
    fn gate_results(
        &self,
    ) -> std::sync::RwLockReadGuard<'_, HashMap<String, HashMap<String, GateResult>>> {
        self.gate_results
            .read()
            .unwrap_or_else(|poisoned| poisoned.into_inner())
    }
}

impl InMemoryStateStore {
    /// Sparar en plan och dess checkpoints.
    pub fn save_plan(&self, plan: &CheckpointPlan) -> Result<(), StateError> {
        let mut planer = self
            .planer
            .write()
            .unwrap_or_else(|poisoned| poisoned.into_inner());
        let mut checkpoints = self
            .checkpoints
            .write()
            .unwrap_or_else(|poisoned| poisoned.into_inner());

        for checkpoint in plan.checkpoints() {
            checkpoints.insert(checkpoint.id.clone(), checkpoint.clone());
        }

        planer.insert(plan.id.clone(), plan.clone());

        Ok(())
    }

    /// Läser en plan med sina checkpoints.
    pub fn get_plan(&self, id: &str) -> Result<CheckpointPlan, StateError> {
        let plan = self
            .planer()
            .get(id)
            .cloned()
            .ok_or_else(|| StateError::NotFound(id.to_string()))?;

        let checkpoints = self.get_checkpoints_by_plan(id).unwrap_or_default();

        Ok(CheckpointPlan {
            checkpoints,
            ..plan
        })
    }

    /// Läser alla planer.
    pub fn get_all_plans(&self) -> Result<Vec<CheckpointPlan>, StateError> {
        let planer: Vec<CheckpointPlan> = self.planer().values().cloned().collect();

        Ok(planer)
    }

    /// Tar bort en plan.
    pub fn delete_plan(&self, id: &str) -> Result<(), StateError> {
        self.planer
            .write()
            .unwrap_or_else(|poisoned| poisoned.into_inner())
            .remove(id);

        Ok(())
    }

    /// Sparar en checkpoint.
    pub fn save_checkpoint(&self, checkpoint: &Checkpoint) -> Result<(), StateError> {
        self.checkpoints
            .write()
            .unwrap_or_else(|poisoned| poisoned.into_inner())
            .insert(checkpoint.id.clone(), checkpoint.clone());

        Ok(())
    }

    /// Läser en checkpoint.
    pub fn get_checkpoint(&self, id: &str) -> Result<Checkpoint, StateError> {
        self.checkpoints()
            .get(id)
            .cloned()
            .ok_or_else(|| StateError::NotFound(id.to_string()))
    }

    /// Läser checkpointsen i en plan.
    pub fn get_checkpoints_by_plan(
        &self,
        plan_id: &str,
    ) -> Result<Option<Vec<Checkpoint>>, StateError> {
        let plan = self
            .planer()
            .get(plan_id)
            .cloned()
            .ok_or_else(|| StateError::NotFound(plan_id.to_string()))?;

        Ok(plan.checkpoints)
    }

    /// Sparar resultatet av en gate.
    pub fn save_gate_result(
        &self,
        checkpoint_id: &str,
        result: &GateResult,
    ) -> Result<(), StateError> {
        let mut resultat = self
            .gate_results
            .write()
            .unwrap_or_else(|poisoned| poisoned.into_inner());

        resultat
            .entry(checkpoint_id.to_string())
            .or_default()
            .insert(result.gate_type.clone(), result.clone());

        Ok(())
    }

    /// Läser resultatet av en gate.
    pub fn get_gate_result(
        &self,
        checkpoint_id: &str,
        gate_type: &str,
    ) -> Result<GateResult, StateError> {
        self.gate_results()
            .get(checkpoint_id)
            .and_then(|resultat| resultat.get(gate_type))
            .cloned()
            .ok_or_else(|| StateError::NotFound(format!("{checkpoint_id}/{gate_type}")))
    }
}

/// Tillståndet i databasen.
///
/// SQL:en är densamma som Go-agentens - samma tabeller, samma kolumner.
#[derive(Clone)]
pub struct CheckpointStore {
    pool: PgPool,
}

impl CheckpointStore {
    /// Skapar lagret mot en pool.
    pub fn new(pool: PgPool) -> Self {
        Self { pool }
    }

    /// Läser en rad som en plan.
    async fn rad_till_plan(
        rad: &sqlx::postgres::PgRow,
    ) -> Result<(CheckpointPlan, Option<DateTime<Utc>>, Option<DateTime<Utc>>), sqlx::Error> {
        use sqlx::Row;

        // Kolumnen är UUID - Go läser den som sträng, men Rust måste
        // veta vilken typ den är.
        let id: uuid::Uuid = rad.try_get("id")?;
        let feature: String = rad.try_get("feature")?;
        let description: String = rad.try_get("description")?;
        let status: String = rad.try_get("status")?;
        let approved_by: Option<String> = rad.try_get("approved_by")?;
        let approved_at: Option<DateTime<Utc>> = rad.try_get("approved_at")?;
        let rejected_at: Option<DateTime<Utc>> = rad.try_get("rejected_at")?;
        let created_at: DateTime<Utc> = rad.try_get("created_at")?;
        let updated_at: DateTime<Utc> = rad.try_get("updated_at")?;

        let plan = CheckpointPlan {
            id: id.to_string(),
            feature,
            description,
            checkpoints: None,
            status,
            approved_by: approved_by.unwrap_or_default(),
            approved_at,
            rejected_at,
            created_at,
            updated_at,
        };

        Ok((plan, approved_at, rejected_at))
    }
}

/// Gör en sträng till ett UUID.
///
/// Go skickar planens id som sträng och låter drivaren omvandla det.
/// Rust binder typer strikt - därför görs omvandlingen här.
fn minska_uuid(id: &str) -> Result<uuid::Uuid, sqlx::Error> {
    uuid::Uuid::parse_str(id).map_err(|err| sqlx::Error::Protocol(format!("ogiltigt id: {err}")))
}

/// Sparar en checkpoint-rad.
async fn save_checkpoint_rad(
    pool: &PgPool,
    checkpoint: &Checkpoint,
    goal_id: &str,
) -> Result<(), sqlx::Error> {
    let gates = serde_json::to_string(&checkpoint.gates).unwrap_or_else(|_| "[]".to_string());
    let tasks = serde_json::to_string(&checkpoint.tasks).unwrap_or_else(|_| "[]".to_string());

    sqlx::query(
        r#"
        INSERT INTO checkpoints (id, goal_id, name, description, complexity, gates, tasks, status, position, created_at, updated_at, completed_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
        ON CONFLICT (id) DO UPDATE SET
            name = EXCLUDED.name,
            description = EXCLUDED.description,
            complexity = EXCLUDED.complexity,
            gates = EXCLUDED.gates,
            tasks = EXCLUDED.tasks,
            status = EXCLUDED.status,
            position = EXCLUDED.position,
            updated_at = EXCLUDED.updated_at,
            completed_at = EXCLUDED.completed_at
        "#,
    )
    .bind(uuid::Uuid::parse_str(&checkpoint.id).unwrap_or_else(|_| uuid::Uuid::new_v4()))
    .bind(if goal_id.is_empty() {
        uuid::Uuid::nil()
    } else {
        uuid::Uuid::parse_str(goal_id).unwrap_or_else(|_| uuid::Uuid::nil())
    })
    .bind(&checkpoint.name)
    .bind(&checkpoint.description)
    .bind(&checkpoint.complexity)
    .bind(gates)
    .bind(tasks)
    .bind(&checkpoint.status)
    .bind(0i32)
    .bind(checkpoint.created_at)
    .bind(checkpoint.updated_at)
    .bind(checkpoint.completed_at)
    .execute(pool)
    .await?;

    if let Some(resultat) = checkpoint.gate_results.as_ref() {
        for result in resultat.values() {
            let issues = serde_json::to_string(result.issues.as_deref().unwrap_or(&[]))
                .unwrap_or_else(|_| "[]".to_string());

            sqlx::query(
                r#"
                INSERT INTO checkpoint_results (id, checkpoint_id, gate_type, status, started_at, ended_at, duration, feedback, issues, summary)
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
                ON CONFLICT (checkpoint_id, gate_type) DO UPDATE SET
                    status = EXCLUDED.status,
                    started_at = EXCLUDED.started_at,
                    ended_at = EXCLUDED.ended_at,
                    duration = EXCLUDED.duration,
                    feedback = EXCLUDED.feedback,
                    issues = EXCLUDED.issues,
                    summary = EXCLUDED.summary
                "#,
            )
            .bind(uuid::Uuid::new_v4())
            .bind(uuid::Uuid::parse_str(&checkpoint.id).unwrap_or_else(|_| uuid::Uuid::new_v4()))
            .bind(&result.gate_type)
            .bind(&result.status)
            .bind(result.started_at)
            .bind(None::<DateTime<Utc>>)
            .bind(&result.duration)
            .bind(&result.feedback)
            .bind(issues)
            .bind(&result.summary)
            .execute(pool)
            .await?;
        }
    }

    Ok(())
}

/// Läser checkpointsen i en plan.
///
/// En plan utan checkpoints ger `None` - det blir `null` i JSON, precis
/// som i Go.
async fn lasa_checkpoints(
    pool: &PgPool,
    plan_id: &str,
) -> Result<Option<Vec<Checkpoint>>, sqlx::Error> {
    let rader = sqlx::query(
        r#"
        SELECT id, goal_id, name, description, complexity, gates, tasks, status, position, created_at, updated_at, completed_at
        FROM checkpoints WHERE goal_id = $1
        ORDER BY position ASC, created_at ASC
        "#,
    )
    .bind(minska_uuid(plan_id)?)
    .fetch_all(pool)
    .await?;

    let mut checkpoints = Vec::new();

    for rad in &rader {
        let mut checkpoint = rad_till_checkpoint(rad).await?;
        checkpoint.gate_results = lasa_gate_results(pool, &checkpoint.id).await.ok();
        checkpoints.push(checkpoint);
    }

    if checkpoints.is_empty() {
        return Ok(None);
    }

    Ok(Some(checkpoints))
}

/// Läser gate-resultaten för en checkpoint.
async fn lasa_gate_results(
    pool: &PgPool,
    checkpoint_id: &str,
) -> Result<std::collections::BTreeMap<String, GateResult>, sqlx::Error> {
    let rader = sqlx::query(
        r#"
        SELECT gate_type, status, started_at, ended_at, duration, feedback, issues, summary
        FROM checkpoint_results WHERE checkpoint_id = $1
        "#,
    )
    .bind(checkpoint_id)
    .fetch_all(pool)
    .await?;

    let mut resultat = std::collections::BTreeMap::new();

    for rad in &rader {
        let gate = rad_till_gate_result(rad).await?;
        resultat.insert(gate.gate_type.clone(), gate);
    }

    Ok(resultat)
}

/// Gör en rad till en checkpoint.
async fn rad_till_checkpoint(rad: &sqlx::postgres::PgRow) -> Result<Checkpoint, sqlx::Error> {
    use sqlx::Row;

    let id: uuid::Uuid = rad.try_get("id")?;
    let name: String = rad.try_get("name")?;
    let description: String = rad.try_get("description")?;
    let complexity: String = rad.try_get("complexity")?;
    let gates_json: serde_json::Value = rad.try_get("gates")?;
    let tasks_json: serde_json::Value = rad.try_get("tasks")?;
    let status: String = rad.try_get("status")?;
    let created_at: DateTime<Utc> = rad.try_get("created_at")?;
    let updated_at: DateTime<Utc> = rad.try_get("updated_at")?;
    let completed_at: Option<DateTime<Utc>> = rad.try_get("completed_at")?;

    // Trasig JSON ger en tom lista - hellre än ett fel.
    let gates: Vec<String> = serde_json::from_value(gates_json).unwrap_or_default();
    let tasks: Vec<String> = serde_json::from_value(tasks_json).unwrap_or_default();

    Ok(Checkpoint {
        id: id.to_string(),
        name,
        description,
        complexity,
        gates,
        tasks,
        status,
        gate_results: None,
        metadata: None,
        created_at,
        updated_at,
        completed_at,
    })
}

/// Gör en rad till ett gate-resultat.
async fn rad_till_gate_result(rad: &sqlx::postgres::PgRow) -> Result<GateResult, sqlx::Error> {
    use sqlx::Row;

    let gate_type: String = rad.try_get("gate_type")?;
    let status: String = rad.try_get("status")?;
    let started_at: DateTime<Utc> = rad.try_get("started_at")?;
    let ended_at: Option<DateTime<Utc>> = rad.try_get("ended_at")?;
    let duration: Option<String> = rad.try_get("duration")?;
    let feedback: Option<String> = rad.try_get("feedback")?;
    let issues_json: Option<serde_json::Value> = rad.try_get("issues")?;
    let summary: Option<String> = rad.try_get("summary")?;

    let issues: Option<Vec<crate::checkpoint::Issue>> = issues_json
        .and_then(|json| serde_json::from_value(json).ok())
        .filter(|issues: &Vec<crate::checkpoint::Issue>| !issues.is_empty());

    Ok(GateResult {
        gate_type,
        status,
        started_at,
        ended_at: ended_at.unwrap_or_else(crate::checkpoint::nolltid),
        duration: duration.unwrap_or_default(),
        feedback: feedback.unwrap_or_default(),
        issues,
        summary: summary.unwrap_or_default(),
    })
}

/// Delar ett lager mellan flera.
pub type SharedStore = Arc<StateStore>;

impl CheckpointStore {
    /// Sparar en plan och dess checkpoints.
    pub async fn save_plan(&self, plan: &CheckpointPlan) -> Result<(), StateError> {
        sqlx::query(
            r#"
            INSERT INTO goals (id, feature, description, status, approved_by, approved_at, rejected_at, created_at, updated_at)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
            ON CONFLICT (id) DO UPDATE SET
                feature = EXCLUDED.feature,
                description = EXCLUDED.description,
                status = EXCLUDED.status,
                approved_by = EXCLUDED.approved_by,
                approved_at = EXCLUDED.approved_at,
                rejected_at = EXCLUDED.rejected_at,
                updated_at = EXCLUDED.updated_at
            "#,
        )
        .bind(uuid::Uuid::parse_str(&plan.id).map_err(|err| {
            sqlx::Error::Protocol(format!("ogiltigt plan-id: {err}"))
        })?)
        .bind(&plan.feature)
        .bind(&plan.description)
        .bind(&plan.status)
        .bind(&plan.approved_by)
        .bind(plan.approved_at)
        .bind(plan.rejected_at)
        .bind(plan.created_at)
        .bind(plan.updated_at)
        .execute(&self.pool)
        .await?;

        for checkpoint in plan.checkpoints() {
            save_checkpoint_rad(&self.pool, checkpoint, &plan.id).await?;
        }

        Ok(())
    }

    /// Läser en plan med sina checkpoints.
    pub async fn get_plan(&self, id: &str) -> Result<CheckpointPlan, StateError> {
        // Go svarar "plan not found" även när frågan misslyckas - ett
        // ogiltigt id räknas som att planen inte finns.
        let rad = sqlx::query(
            r#"
            SELECT id, feature, description, status, approved_by, approved_at, rejected_at, created_at, updated_at
            FROM goals WHERE id = $1
            "#,
        )
        .bind(minska_uuid(id)?)
        .fetch_optional(&self.pool)
        .await
        .map_err(|_| StateError::NotFound(id.to_string()))?;

        let Some(rad) = rad else {
            return Err(StateError::NotFound(id.to_string()));
        };

        let (mut plan, _, _) = CheckpointStore::rad_till_plan(&rad).await?;
        plan.checkpoints = lasa_checkpoints(&self.pool, id).await?;

        Ok(plan)
    }

    /// Läser alla planer, nyast först.
    pub async fn get_all_plans(&self) -> Result<Vec<CheckpointPlan>, StateError> {
        let rader = sqlx::query(
            r#"
            SELECT id, feature, description, status, approved_by, approved_at, rejected_at, created_at, updated_at
            FROM goals ORDER BY created_at DESC
            "#,
        )
        .fetch_all(&self.pool)
        .await?;

        let mut planer = Vec::new();

        for rad in &rader {
            let (mut plan, _, _) = CheckpointStore::rad_till_plan(rad).await?;
            plan.checkpoints = lasa_checkpoints(&self.pool, &plan.id).await?;
            planer.push(plan);
        }

        Ok(planer)
    }

    /// Tar bort en plan.
    pub async fn delete_plan(&self, id: &str) -> Result<(), StateError> {
        sqlx::query("DELETE FROM goals WHERE id = $1")
            .bind(id)
            .execute(&self.pool)
            .await?;

        Ok(())
    }

    /// Sparar en checkpoint.
    pub async fn save_checkpoint(&self, checkpoint: &Checkpoint) -> Result<(), StateError> {
        save_checkpoint_rad(&self.pool, checkpoint, "").await?;

        Ok(())
    }

    /// Läser en checkpoint.
    pub async fn get_checkpoint(&self, id: &str) -> Result<Checkpoint, StateError> {
        let rad = sqlx::query(
            r#"
            SELECT id, goal_id, name, description, complexity, gates, tasks, status, position, created_at, updated_at, completed_at
            FROM checkpoints WHERE id = $1
            "#,
        )
        .bind(id)
        .fetch_optional(&self.pool)
        .await?;

        let Some(rad) = rad else {
            return Err(StateError::NotFound(id.to_string()));
        };

        Ok(rad_till_checkpoint(&rad).await?)
    }

    /// Läser checkpointsen i en plan.
    pub async fn get_checkpoints_by_plan(
        &self,
        plan_id: &str,
    ) -> Result<Option<Vec<Checkpoint>>, StateError> {
        Ok(lasa_checkpoints(&self.pool, plan_id).await?)
    }

    /// Sparar resultatet av en gate.
    pub async fn save_gate_result(
        &self,
        checkpoint_id: &str,
        result: &GateResult,
    ) -> Result<(), StateError> {
        let issues = serde_json::to_string(result.issues.as_deref().unwrap_or(&[]))
            .unwrap_or_else(|_| "[]".to_string());

        sqlx::query(
            r#"
            INSERT INTO checkpoint_results (id, checkpoint_id, gate_type, status, started_at, ended_at, duration, feedback, issues, summary)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
            ON CONFLICT (checkpoint_id, gate_type) DO UPDATE SET
                status = EXCLUDED.status,
                started_at = EXCLUDED.started_at,
                ended_at = EXCLUDED.ended_at,
                duration = EXCLUDED.duration,
                feedback = EXCLUDED.feedback,
                issues = EXCLUDED.issues,
                summary = EXCLUDED.summary
            "#,
        )
        .bind(uuid::Uuid::new_v4())
        .bind(checkpoint_id)
        .bind(&result.gate_type)
        .bind(&result.status)
        .bind(result.started_at)
        .bind(if result.ended_at == crate::checkpoint::nolltid() {
            None
        } else {
            Some(result.ended_at)
        })
        .bind(&result.duration)
        .bind(&result.feedback)
        .bind(issues)
        .bind(&result.summary)
        .execute(&self.pool)
        .await?;

        Ok(())
    }

    /// Läser resultatet av en gate.
    pub async fn get_gate_result(
        &self,
        checkpoint_id: &str,
        gate_type: &str,
    ) -> Result<GateResult, StateError> {
        let rad = sqlx::query(
            r#"
            SELECT gate_type, status, started_at, ended_at, duration, feedback, issues, summary
            FROM checkpoint_results WHERE checkpoint_id = $1 AND gate_type = $2
            "#,
        )
        .bind(checkpoint_id)
        .bind(gate_type)
        .fetch_optional(&self.pool)
        .await?;

        let Some(rad) = rad else {
            return Err(StateError::NotFound(format!("{checkpoint_id}/{gate_type}")));
        };

        Ok(rad_till_gate_result(&rad).await?)
    }
}
