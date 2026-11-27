//! Schemaläggaren - pollar Plan efter agentmål och definierar dem.
//!
//! Go: `internal/scheduler`. Schemaläggaren är den enklare av de två
//! agenterna: den definierar mål och lämnar förslaget till en människa,
//! men rör inte kod. Kodfasen sköts av workern.

use std::sync::Arc;
use std::time::Duration;

use tracing::{info, warn};

use crate::approvals::{PendingApproval, Store, STATUS_PENDING};
use crate::definer::{Definer, Input};
use crate::planclient::models::{Goal, EVENT_APPROVAL_REQUESTED};
use crate::planclient::{
    CreateAgentRunEventRequest, CreateAgentRunRequest, PlanClient, UpdateAgentRunRequest,
};

/// Körningen startades av schemaläggningen.
pub const TRIGGER_SCHEDULE: &str = "schedule";

/// Pollar Plan och behandlar mål.
pub struct Scheduler {
    /// Klienten mot Plan.
    plan: PlanClient,
    /// Store för väntande förslag.
    approvals: Arc<Store>,
    /// Agenten som definierar mål.
    definer: Arc<Definer>,
    /// Hur ofta Plan läses.
    pollintervall: Duration,
    /// Adressen till gränssnittet - den syns i sammanfattningen.
    plan_url: String,
    /// Vem eller vad som startade körningen.
    initiator: String,
}

/// Konfigurerar schemaläggaren.
#[derive(Clone, Default)]
pub struct Config {
    /// Pollintervallet. Noll ger en minut.
    pub pollintervall: Duration,
    /// Adressen till gränssnittet.
    pub plan_url: String,
    /// Vem eller vad som startade körningen.
    pub initiator: String,
}

impl Scheduler {
    /// Skapar en schemaläggare.
    ///
    /// Go: `scheduler.New`.
    pub fn ny(
        plan: PlanClient,
        approvals: Arc<Store>,
        definer: Arc<Definer>,
        config: Config,
    ) -> Self {
        let pollintervall = if config.pollintervall.is_zero() {
            Duration::from_secs(60)
        } else {
            config.pollintervall
        };

        let plan_url = if config.plan_url.is_empty() {
            let ur_miljö = std::env::var("PLAN_WEB_URL").unwrap_or_default();

            if ur_miljö.is_empty() {
                "http://localhost:3000".to_string()
            } else {
                ur_miljö
            }
        } else {
            config.plan_url
        };

        Self {
            plan,
            approvals,
            definer,
            pollintervall,
            plan_url,
            initiator: config.initiator,
        }
    }

    /// Adressen till gränssnittet.
    ///
    /// Go: fältet `planURL` - det sätts men används inte ännu.
    pub fn plan_url(&self) -> &str {
        &self.plan_url
    }

    /// Startar schemaläggningen - den kör tills kontexten avbryts.
    ///
    /// Go: `Scheduler.Run`.
    pub async fn run(&self) {
        info!(
            "Scheduler starting, polling every {:?} for agent=true goals in define",
            self.pollintervall
        );

        if let Err(fel) = self.run_once().await {
            warn!("scheduler run failed: {fel}");
        }

        loop {
            tokio::time::sleep(self.pollintervall).await;

            if let Err(fel) = self.run_once().await {
                warn!("scheduler run failed: {fel}");
            }
        }
    }

    /// Behandlar målen en gång.
    ///
    /// Go: `Scheduler.RunOnce`.
    pub async fn run_once(&self) -> Result<(), String> {
        let goals = self
            .plan
            .agent_goals()
            .await
            .map_err(|fel| format!("failed to fetch agent goals: {fel}"))?;

        let define_status_id = self
            .plan
            .status_id_by_name("define")
            .await
            .map_err(|fel| format!("failed to resolve define phase: {fel}"))?;

        info!("Found {} agent goals", goals.len());

        for goal in &goals {
            if !is_agent_define_goal(goal, define_status_id) {
                continue;
            }

            self.processa_mål(goal).await;
        }

        Ok(())
    }

    /// Definierar ett mål och lämnar förslaget till en människa.
    ///
    /// Go: `Scheduler.processGoal`.
    async fn processa_mål(&self, goal: &Goal) {
        info!("Processing goal #{}: {}", goal.id, goal.title);

        // Ett befintligt förslag ger återkoppling och räknade försök.
        let existing = match self.approvals.get(goal.id) {
            Ok(existing) => existing,
            Err(fel) => {
                warn!(
                    "Failed to check existing approval for goal #{}: {fel}",
                    goal.id
                );

                return;
            }
        };

        let mut feedback = String::new();
        let mut attempts = 1;

        if let Some(existing) = &existing {
            feedback = existing.feedback.clone();
            attempts = existing.attempts + 1;
        }

        let definition = match self
            .definer
            .definiera(&Input {
                title: goal.title.clone(),
                description: goal.title.clone(),
                feedback,
                ..Default::default()
            })
            .await
        {
            Ok(definition) => definition,
            Err(fel) => {
                warn!("Failed to define goal #{}: {fel}", goal.id);

                self.publicera(
                    goal,
                    crate::worker::STATUS_FAILED,
                    &format!("Agenten kunde inte specificera målet: {fel}"),
                )
                .await;

                return;
            }
        };

        let approval = PendingApproval {
            goal_id: goal.id,
            title: definition.title.clone(),
            description: definition.description.clone(),
            created_at: chrono::Utc::now().format("%Y-%m-%dT%H:%M:%SZ").to_string(),
            attempts,
            feedback: String::new(),
            status: STATUS_PENDING.to_string(),
            conversation_id: String::new(),
        };

        if let Err(fel) = self.approvals.save(&approval) {
            warn!("Failed to save approval for goal #{}: {fel}", goal.id);

            return;
        }

        info!(
            "Saved pending approval for goal #{}: {}",
            goal.id, definition.title
        );

        // Förslaget syns i Plan - där tas beslutet. Slack används inte.
        self.publicera(
            goal,
            crate::worker::STATUS_WAITING_APPROVAL,
            &format!(
                "Förslag: {}\n\n{}",
                definition.title, definition.description
            ),
        )
        .await;

        self.registrera_godkännande(goal, &definition.title).await;
    }

    /// Skriver begäran om ett mänskligt beslut i körningens historik, så att
    /// det är tydligt att en människa - inte agenten - måste bestämma.
    ///
    /// Go: `Scheduler.recordApprovalRequest`.
    async fn registrera_godkännande(&self, goal: &Goal, titel: &str) {
        let runs = match self.plan.agent_runs(Some(goal.id)).await {
            Ok(runs) => runs,
            Err(fel) => {
                warn!("Failed to list agent runs for goal #{}: {fel}", goal.id);

                return;
            }
        };

        let Some(run) = runs.first() else {
            return;
        };

        if let Err(fel) = self
            .plan
            .create_agent_run_event(
                run.id,
                CreateAgentRunEventRequest {
                    r#type: EVENT_APPROVAL_REQUESTED.to_string(),
                    phase: Some("define".to_string()),
                    payload: Some(serde_json::json!({ "title": titel })),
                },
            )
            .await
        {
            warn!(
                "Failed to record approval request for goal #{}: {fel}",
                goal.id
            );
        }
    }

    /// Skapar eller uppdaterar agentkörningen i Plan, så att förslaget syns
    /// där målet bor.
    ///
    /// Go: `Scheduler.publish`.
    async fn publicera(&self, goal: &Goal, status: &str, summary: &str) {
        let runs = match self.plan.agent_runs(Some(goal.id)).await {
            Ok(runs) => runs,
            Err(fel) => {
                warn!("Failed to list agent runs for goal #{}: {fel}", goal.id);

                return;
            }
        };

        if let Some(run) = runs.first() {
            let phase = run.phase.clone();

            if let Err(fel) = self
                .plan
                .update_agent_run(
                    run.id,
                    UpdateAgentRunRequest {
                        phase: Some(phase),
                        status: Some(status.to_string()),
                        summary: Some(summary.to_string()),
                        ..Default::default()
                    },
                )
                .await
            {
                warn!("Failed to update agent run {}: {fel}", run.id);
            }

            return;
        }

        if let Err(fel) = self
            .plan
            .create_agent_run(CreateAgentRunRequest {
                goal_id: goal.id,
                phase: Some("define".to_string()),
                status: Some(status.to_string()),
                trigger: Some(TRIGGER_SCHEDULE.to_string()),
                initiator: Some(self.initiator.clone()),
                summary: Some(summary.to_string()),
                ..Default::default()
            })
            .await
        {
            warn!("Failed to create agent run for goal #{}: {fel}", goal.id);
        }
    }
}

/// Berättar om målet är ett agentmål i definitionfasen.
///
/// Go: `isAgentDefineGoal`.
pub fn is_agent_define_goal(goal: &Goal, define_status_id: i32) -> bool {
    goal.agent && goal.status_id == define_status_id
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn endast_definitionfasen_räknas() {
        let mut goal = Goal {
            id: 1,
            title: "Läsa rapporten".to_string(),
            description: String::new(),
            icon: String::new(),
            status_id: 1,
            created: chrono::Utc::now(),
            started: None,
            finished: None,
            inbox: 0,
            agent: true,
            now: false,
            eternal: false,
            due: None,
            start_date: None,
            estimated_minutes: None,
            recurrence_rule: None,
            recurrence_mode: None,
            tags: Vec::new(),
            places: Vec::new(),
        };

        assert!(is_agent_define_goal(&goal, 1));
        assert!(!is_agent_define_goal(&goal, 2));

        goal.agent = false;

        assert!(!is_agent_define_goal(&goal, 1));
    }
}
