//! Orchestratorns tillstånd och planhantering.
//!
//! Orchestratorn koordinerar checkpoint-gate-flödet. Här ligger
//! tillståndet - initieringsfelet sparas så att det syns i både health
//! och status, och planerna hanteras här.
//!
//! Själva körningen (gates, subagenter, viewers) portas i fas 4d -
//! tills dess är det här planerna och deras godkännanden som finns.

use std::sync::{Arc, RwLock};

use crate::checkpoint::CheckpointPlan;
use crate::state_store::{StateError, StateStore};

/// Initieringsfelet, om orchestratorn inte startade.
static INIT_ERROR: RwLock<Option<String>> = RwLock::new(None);

/// Om orchestratorn är redo.
static REDO: RwLock<bool> = RwLock::new(false);

/// Sparar initieringsfelet.
pub fn set_init_error(err: impl Into<String>) {
    if let Ok(mut fel) = INIT_ERROR.write() {
        *fel = Some(err.into());
    }
}

/// Initieringsfelet, om det finns.
pub fn init_error() -> Option<String> {
    INIT_ERROR.read().ok().and_then(|fel| fel.clone())
}

/// Markera orchestratorn som redo.
pub fn set_ready() {
    if let Ok(mut redo) = REDO.write() {
        *redo = true;
    }
}

/// Om orchestratorn är redo.
pub fn is_ready() -> bool {
    REDO.read().map(|redo| *redo).unwrap_or(false)
}

/// Agentens orchestrator.
///
/// Orchestratorn håller den aktuella planen och sparar planerna i
/// tillståndslagret.
pub struct Orchestrator {
    /// Lagret där planerna sparas.
    lager: StateStore,
    /// Planen som körs just nu.
    aktuell_plan: RwLock<Option<CheckpointPlan>>,
}

impl Orchestrator {
    /// Skapar orchestratorn med ett tillståndslager.
    pub fn new(lager: StateStore) -> Self {
        Self {
            lager,
            aktuell_plan: RwLock::new(None),
        }
    }

    /// Skapar en plan för en funktion.
    ///
    /// Go skapar planen utan checkpoints när ingen planner är satt - det
    /// är det fallet i produktion, där agenten inte har någon
    /// checkpointplanner konfigurerad.
    pub async fn create_plan(
        &self,
        feature: &str,
        description: &str,
    ) -> Result<CheckpointPlan, StateError> {
        let plan = CheckpointPlan::new(uuid::Uuid::new_v4().to_string(), feature, description);

        // Ett fel här ska inte stoppa arbetet - planen finns i minnet
        // även om den inte gick att spara.
        if let Err(err) = self.lager.save_plan(&plan).await {
            tracing::warn!("Kunde inte spara planen: {err}");
        }

        if let Ok(mut aktuell) = self.aktuell_plan.write() {
            *aktuell = Some(plan.clone());
        }

        Ok(plan)
    }

    /// Läser en plan.
    pub async fn plan(&self, id: &str) -> Result<CheckpointPlan, StateError> {
        self.lager.get_plan(id).await
    }

    /// Läser alla planer.
    pub async fn all_plans(&self) -> Result<Vec<CheckpointPlan>, StateError> {
        self.lager.get_all_plans().await
    }

    /// Godkänner en plan.
    pub async fn approve_plan(&self, id: &str, approved_by: &str) -> Result<(), StateError> {
        let mut plan = self
            .lager
            .get_plan(id)
            .await
            .map_err(|err| StateError::Message(format!("plan not found: {err}")))?;
        plan.approve(approved_by);

        self.lager.save_plan(&plan).await?;

        Ok(())
    }

    /// Avvisar en plan.
    pub async fn reject_plan(&self, id: &str) -> Result<(), StateError> {
        let mut plan = self
            .lager
            .get_plan(id)
            .await
            .map_err(|err| StateError::Message(format!("plan not found: {err}")))?;
        plan.reject();

        self.lager.save_plan(&plan).await?;

        Ok(())
    }

    /// Planen som körs just nu.
    pub fn current_plan(&self) -> Option<CheckpointPlan> {
        self.aktuell_plan.read().ok().and_then(|plan| plan.clone())
    }

    /// Checkpointen som körs just nu.
    pub async fn current_checkpoint(&self) -> Option<crate::checkpoint::Checkpoint> {
        self.current_plan().and_then(|plan| plan.current().cloned())
    }

    /// Framsteget i den aktuella planen.
    ///
    /// Go räknar varje gate i varje checkpoint - alltså antalet gates
    /// som godkänts, inte antalet checkpoints.
    pub async fn progress(&self) -> (usize, usize) {
        let Some(plan) = self.current_plan() else {
            return (0, 0);
        };

        let mut passed = 0;
        let mut total = 0;

        for checkpoint in plan.checkpoints() {
            for gate in &checkpoint.gates {
                total += 1;

                let godkänd = checkpoint
                    .gate_results
                    .as_ref()
                    .and_then(|resultat| resultat.get(gate))
                    .map(|result| result.status == crate::checkpoint::gate_status::PASSED)
                    .unwrap_or(false);

                if godkänd {
                    passed += 1;
                }
            }
        }

        (passed, total)
    }
}

/// Delar orchestratorn mellan flera.
pub type SharedOrchestrator = Arc<Orchestrator>;
