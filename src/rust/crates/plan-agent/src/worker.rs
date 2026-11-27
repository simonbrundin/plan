//! Workern som tar sig an mål som är markerade för agentarbete.
//!
//! Go: `internal/agentworker`. Arbetsflödet är avsiktligt litet:
//! agenten specificerar målet, skriver den egna definitionen tillbaka i
//! Plan och sätter målets status till "plan".

use std::collections::HashMap;
use std::sync::Arc;
use std::time::Duration;

use chrono::Utc;
use serde_json::json;
use tracing::{info, warn};

use crate::definer::{Definition, Input};
use crate::goaldefinition::{self, Context as Förslagskontext, Proposal};
use crate::planclient::models::{AgentRun, Goal, Place};
use crate::planclient::{
    CreateAgentApprovalRequest, CreateAgentRunRequest, PlanClient, UpdateAgentRunRequest,
};
use crate::profile::{self, Policy};
use crate::runusage::Snapshot;

/// Faserna speglar `agent_runs`-tabellen i Plan.
pub const PHASE_DEFINE: &str = "define";
/// Fasen där målet planeras.
pub const PHASE_PLAN: &str = "plan";
/// Fasen där ändringarna utförs.
pub const PHASE_EXECUTE: &str = "execute";
/// Fasen där ändringarna verifieras.
pub const PHASE_VERIFY: &str = "verify";
/// Fasen där allt är klart.
pub const PHASE_COMPLETED: &str = "completed";

/// Körningen pågår.
pub const STATUS_RUNNING: &str = "running";
/// Körningen lyckades.
pub const STATUS_SUCCEEDED: &str = "succeeded";
/// Körningen misslyckades.
pub const STATUS_FAILED: &str = "failed";
/// Körningen väntar på en människa.
pub const STATUS_WAITING_APPROVAL: &str = "waiting_approval";
/// Körningen väntar på att få ett jobb.
pub const STATUS_QUEUED: &str = "queued";
/// Körningen avbröts av en människa.
pub const STATUS_CANCELLED: &str = "cancelled";

/// Körningen startades av agenten själv.
pub const TRIGGER_AGENT: &str = "agent";

/// Målet är färdigt och workern kan avslutas.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct ErrFinished;

impl std::fmt::Display for ErrFinished {
    fn fmt(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(formatter, "goal finished")
    }
}

impl std::error::Error for ErrFinished {}

/// Agenten som definierar mål.
pub trait GoalDefiner: Send + Sync {
    /// Tar fram ett förslag på hur målet ska se ut.
    fn definiera<'a>(
        &'a self,
        input: &'a Input,
    ) -> std::pin::Pin<Box<dyn std::future::Future<Output = Result<Definition, String>> + Send + 'a>>;
}

/// Kodrunner som driver kodändringar för ett mål.
pub trait CodeRunner: Send + Sync {
    /// Utför arbetet för en körning.
    fn process<'a>(
        &'a self,
        run: &'a AgentRun,
        goal: &'a Goal,
    ) -> std::pin::Pin<Box<dyn std::future::Future<Output = Result<(), String>> + Send + 'a>>;

    /// Användningen för körningen, om den finns.
    fn användning(&self) -> Option<Snapshot> {
        None
    }
}

/// Workern pollar Plan efter agentmål och definierar dem.
pub struct Worker {
    /// Klienten mot Plan.
    plan: PlanClient,
    /// Agenten som definierar mål.
    definierare: Arc<dyn GoalDefiner>,
    /// Kodrunnern, om profilen får röra kod.
    kodrunner: Option<Arc<dyn CodeRunner>>,
    /// Hur ofta Plan läses.
    pollintervall: Duration,
    /// Leverantören som används.
    leverantör: String,
    /// Modellen som används.
    modell: String,
    /// Vilken agent som kör.
    agent_id: String,
    /// Begränsar workern till ett mål. Noll betyder alla.
    enda_mål_id: i64,
    /// Vem eller vad som startade körningen.
    initiator: String,
    /// Profilen som kör.
    profil: String,
    /// Gör att workern slutar när målet är färdigt.
    stop_when_finished: bool,
}

/// Konfigurerar workern.
#[derive(Clone, Default)]
pub struct Config {
    /// Pollintervallet. Noll ger en minut.
    pub pollintervall: Duration,
    /// Begränsar workern till ett enda mål.
    pub enda_mål_id: i64,
    /// Vem eller vad som startade körningen.
    pub initiator: String,
    /// Vilken sorts agent som kör.
    pub profil: String,
    /// Kodrunnern. `None` betyder att workern bara definierar mål.
    pub kodrunner: Option<Arc<dyn CodeRunner>>,
    /// Gör att workern slutar när målet är färdigt.
    pub stop_when_finished: bool,
    /// Leverantören och modellen - Plan behöver dem för att prissätta
    /// körningen.
    pub leverantör: String,
    /// Modellen som används.
    pub modell: String,
    /// Vilken agent som kör.
    pub agent_id: String,
}

impl Worker {
    /// Skapar en worker.
    ///
    /// Go: `agentworker.New`.
    pub fn ny(plan: PlanClient, definierare: Arc<dyn GoalDefiner>, config: Config) -> Self {
        let pollintervall = if config.pollintervall.is_zero() {
            Duration::from_secs(60)
        } else {
            config.pollintervall
        };

        // Profilen bestämmer vad workern får göra - den är en teknisk
        // gräns, inte en instruktion till modellen.
        let policy = match profile::policy_for(&config.profil) {
            Ok(policy) => policy,
            Err(fel) => {
                warn!(
                    "Invalid profile {:?}, using {}",
                    config.profil,
                    profile::CODING
                );
                let _ = fel;
                Policy::coding()
            }
        };

        // En profil som inte får röra kod får inte köra codern.
        let kodrunner = if policy.may_change_code {
            config.kodrunner
        } else {
            None
        };

        Self {
            plan,
            definierare,
            kodrunner,
            pollintervall,
            leverantör: config.leverantör,
            modell: config.modell,
            agent_id: config.agent_id,
            enda_mål_id: config.enda_mål_id,
            initiator: config.initiator,
            profil: policy.name,
            stop_when_finished: config.stop_when_finished,
        }
    }

    /// Kör workern tills kontexten avbryts.
    ///
    /// Go: `Worker.Run`.
    pub async fn run(&self) {
        info!(
            "Agent worker starting, polling every {:?} for goals with agent=true",
            self.pollintervall
        );

        if let Err(fel) = self.run_once().await {
            if fel == "finished" {
                info!("Agent worker finished the goal and stops");
                return;
            }

            warn!("Agent worker run failed: {fel}");
        }

        loop {
            tokio::time::sleep(self.pollintervall).await;

            match self.run_once().await {
                Ok(()) => continue,
                Err(fel) if fel == "finished" => {
                    info!("Agent worker finished the goal and stops");
                    return;
                }
                Err(fel) => warn!("Agent worker run failed: {fel}"),
            }
        }
    }

    /// Behandlar agentmålen en gång.
    ///
    /// Go: `Worker.RunOnce`.
    pub async fn run_once(&self) -> Result<(), String> {
        let mål = self
            .plan
            .agent_goals()
            .await
            .map_err(|fel| format!("failed to fetch agent goals: {fel}"))?;

        let define_status_id = self
            .plan
            .status_id_by_name(PHASE_DEFINE)
            .await
            .map_err(|fel| format!("failed to resolve define phase: {fel}"))?;

        let körningar = self
            .plan
            .agent_runs(None)
            .await
            .map_err(|fel| format!("failed to fetch agent runs: {fel}"))?;

        // Körningar per mål - den senaste används.
        let mut senaste: HashMap<i64, AgentRun> = HashMap::new();
        let mut antal: HashMap<i64, usize> = HashMap::new();

        for körning in &körningar {
            let entry = senaste.entry(körning.goal_id).or_insert(körning.clone());

            if körning.id > entry.id {
                *entry = körning.clone();
            }

            *antal.entry(körning.goal_id).or_insert(0) += 1;
        }

        // En köad körning är en uttrycklig begäran om att få köras.
        let köade = hitta_köade_körning(&körningar);

        // Mål som redan behandlats.
        let mut hanterade: Vec<i64> = Vec::new();

        let plan_status_id = self
            .plan
            .status_id_by_name(PHASE_PLAN)
            .await
            .map_err(|fel| format!("failed to resolve plan phase: {fel}"))?;

        for mål in &mål {
            if !mål.agent {
                continue;
            }

            if self.enda_mål_id != 0 && mål.id != self.enda_mål_id {
                continue;
            }

            let (har_körning, körning) = match senaste.get(&mål.id) {
                Some(körning) => (true, körning.clone()),
                None => (false, AgentRun::empty()),
            };

            // En avbruten körning ska inte återupptas - en människa har
            // sagt stopp.
            if har_körning && körning.status == STATUS_CANCELLED {
                info!(
                    "Run {} for goal #{} was cancelled - the worker stops",
                    körning.id, mål.id
                );

                return Err("finished".to_string());
            }

            // En köad körning tas över och arbetet utförs i den.
            if let Some(köad) = köade.get(&mål.id) {
                hanterade.push(mål.id);

                if let Err(fel) = self.kör_köad(mål, köad.clone(), define_status_id).await {
                    warn!("Failed to run queued run {}: {fel}", köad.id);
                }

                continue;
            }

            // Ett mål som väntar på ett mänskligt beslut ska återupptas
            // så snart beslutet är fattat.
            if har_körning && körning.status == STATUS_WAITING_APPROVAL {
                if mål.status_id == define_status_id {
                    let återupptagen = match self.återuppta_definition(&körning, mål).await {
                        Ok(återupptagen) => återupptagen,
                        Err(fel) => {
                            warn!("Failed to resume the definition of goal #{}: {fel}", mål.id);
                            false
                        }
                    };

                    if återupptagen {
                        hanterade.push(mål.id);
                        continue;
                    }
                }

                if let Some(kodrunner) = &self.kodrunner {
                    if let Err(fel) = kodrunner.process(&körning, mål).await {
                        warn!("Failed to resume goal #{}: {fel}", mål.id);
                    }
                }

                continue;
            }

            // En pågående körning lämnas ifred.
            if har_körning && körning.status == STATUS_RUNNING {
                continue;
            }

            // Ett mål i definitionfasen specificeras först.
            if mål.status_id == define_status_id {
                if let Err(fel) = self.processa_mål(mål, har_körning, körning.clone()).await {
                    warn!("Failed to process goal #{}: {fel}", mål.id);
                }

                continue;
            }

            // Ett definierat mål lämnas över till codern. Det sker bara
            // en gång per mål.
            if self.kodrunner.is_some()
                && mål.status_id == plan_status_id
                && antal.get(&mål.id).copied().unwrap_or(0) < 2
            {
                if let Err(fel) = self.starta_körring(mål, false, AgentRun::empty()).await {
                    warn!("Failed to start code run for goal #{}: {fel}", mål.id);
                }

                continue;
            }

            // Ett färdigt mål avslutar workern.
            if self.stop_when_finished && self.är_färdigt(mål, &körning, har_körning).await {
                return Err("finished".to_string());
            }
        }

        // En köad körning är en uttrycklig begäran och ska inte glömmas
        // bara för att målet inte är ett agentmål.
        for (mål_id, köad) in &köade {
            if hanterade.contains(mål_id) {
                continue;
            }

            if self.enda_mål_id != 0 && *mål_id != self.enda_mål_id {
                continue;
            }

            let mål = match self.plan.goal(*mål_id).await {
                Ok(mål) => mål,
                Err(fel) => {
                    warn!(
                        "Failed to fetch goal #{} for queued run {}: {fel}",
                        mål_id, köad.id
                    );
                    continue;
                }
            };

            if let Err(fel) = self.kör_köad(&mål, köad.clone(), define_status_id).await {
                warn!("Failed to run queued run {}: {fel}", köad.id);
            }
        }

        // När körningen är klar är jobbet klart - det ska inte stå och
        // vänta på en fas som aldrig kommer.
        if self.stop_when_finished && self.enda_mål_id != 0 {
            match self.målskörning_klar(self.enda_mål_id).await {
                Ok(true) => return Err("finished".to_string()),
                Ok(false) => {}
                Err(fel) => warn!(
                    "Failed to check whether goal #{} is finished: {fel}",
                    self.enda_mål_id
                ),
            }
        }

        Ok(())
    }

    /// Berättar om den senaste körningen för målet är klar.
    ///
    /// Go: `Worker.goalRunDone`.
    async fn målskörning_klar(&self, mål_id: i64) -> Result<bool, String> {
        let körningar = self
            .plan
            .agent_runs(None)
            .await
            .map_err(|fel| format!("failed to fetch agent runs: {fel}"))?;

        let senaste = körningar
            .iter()
            .filter(|körning| körning.goal_id == mål_id)
            .max_by_key(|körning| körning.id);

        match senaste {
            Some(körning) => Ok(körning.status == STATUS_SUCCEEDED),
            None => Ok(false),
        }
    }

    /// Tar över en köad körning och utför arbetet i den.
    ///
    /// Go: `Worker.runQueued`.
    async fn kör_köad(
        &self,
        mål: &Goal,
        köad: AgentRun,
        define_status_id: i32,
    ) -> Result<(), String> {
        self.adoptera_körning(&köad).await?;

        if mål.status_id == define_status_id {
            return self.processa_mål(mål, true, köad).await;
        }

        // En profil som inte får röra kod har ingen kodfas.
        if self.kodrunner.is_none() {
            return self.avsluta_utan_ändringar(mål, &köad).await;
        }

        self.starta_körring(mål, true, köad).await
    }

    /// Tar över en köad körning och markerar den som pågående.
    ///
    /// Go: `Worker.adoptRun`.
    async fn adoptera_körning(&self, körning: &AgentRun) -> Result<(), String> {
        let status = STATUS_RUNNING.to_string();
        let summary = "Körningen har startat".to_string();

        self.plan
            .update_agent_run(
                körning.id,
                UpdateAgentRunRequest {
                    status: Some(status),
                    summary: Some(summary),
                    ..Default::default()
                },
            )
            .await
            .map_err(|fel| format!("failed to adopt run {}: {fel}", körning.id))?;

        info!(
            "Adopted queued agent run {} for goal #{}",
            körning.id, körning.goal_id
        );

        Ok(())
    }

    /// Berättar om målet är färdigt och ingen körning pågår.
    ///
    /// Go: `Worker.isFinished`.
    async fn är_färdigt(&self, mål: &Goal, körning: &AgentRun, har_körning: bool) -> bool {
        if !har_körning {
            return false;
        }

        // En misslyckad körning är inte färdig - jobbet ska inte
        // avslutas innan en människa eller en ny körning tagit ställning.
        if körning.status != STATUS_SUCCEEDED {
            return false;
        }

        // En profil som får röra kod flyttar målet till verify när
        // ändringen är verifierad.
        if let Ok(verify_status_id) = self.plan.status_id_by_name(PHASE_VERIFY).await {
            if mål.status_id == verify_status_id {
                return true;
            }
        }

        // En profil som inte får röra kod lämnar målet i plan - men
        // körningen är klar.
        self.kodrunner.is_none()
    }

    /// Avslutar en körning som inte får göra kodändringar.
    ///
    /// Go: `Worker.closeRunWithoutChanges`.
    async fn avsluta_utan_ändringar(
        &self, mål: &Goal, körning: &AgentRun
    ) -> Result<(), String> {
        let status = STATUS_SUCCEEDED.to_string();
        let summary = format!(
            "Profilen {} får inte ändra kod - körningen avslutades utan ändringar",
            self.profil
        );

        self.plan
            .update_agent_run(
                körning.id,
                UpdateAgentRunRequest {
                    status: Some(status),
                    summary: Some(summary.clone()),
                    result: Some(summary),
                    finished_at: Some(Utc::now()),
                    ..Default::default()
                },
            )
            .await
            .map_err(|fel| format!("failed to close agent run: {fel}"))?;

        self.registrera_event(
            körning.id,
            crate::planclient::CreateAgentRunEventRequest {
                r#type: crate::planclient::models::EVENT_RUN_STARTED.to_string(),
                phase: Some(körning.phase.clone()),
                payload: Some(json!({
                    "goal_id": mål.id,
                    "trigger": körning.trigger,
                    "initiator": self.initiator,
                    "agent_profile": self.profil,
                })),
            },
        )
        .await;

        self.registrera_event(
            körning.id,
            crate::planclient::CreateAgentRunEventRequest {
                r#type: crate::planclient::models::EVENT_RUN_COMPLETED.to_string(),
                phase: None,
                payload: Some(json!({
                    "reason": "profile does not allow code changes",
                    "profile": self.profil,
                })),
            },
        )
        .await;

        self.rapportera_användning(körning.id).await;

        Ok(())
    }

    /// Skapar en körning för ett definierat mål och lämnar över till
    /// codern.
    ///
    /// Go: `Worker.startCodeRun`.
    async fn starta_körring(
        &self,
        mål: &Goal,
        har_körning: bool,
        befintlig: AgentRun,
    ) -> Result<(), String> {
        let körning = if har_körning {
            befintlig
        } else {
            self.plan
                .create_agent_run(CreateAgentRunRequest {
                    goal_id: mål.id,
                    phase: Some(PHASE_PLAN.to_string()),
                    status: Some(STATUS_RUNNING.to_string()),
                    agent_id: Some(self.agent_id.clone()),
                    provider: Some(self.leverantör.clone()),
                    model: Some(self.modell.clone()),
                    trigger: Some(TRIGGER_AGENT.to_string()),
                    initiator: Some(self.initiator.clone()),
                    agent_profile: Some(self.profil.clone()),
                    summary: Some("Agenten tar fram ett kodförslag".to_string()),
                    ..Default::default()
                })
                .await
                .map_err(|fel| format!("failed to create agent run: {fel}"))?
        };

        self.registrera_event(
            körning.id,
            crate::planclient::CreateAgentRunEventRequest {
                r#type: crate::planclient::models::EVENT_RUN_STARTED.to_string(),
                phase: Some(PHASE_PLAN.to_string()),
                payload: Some(json!({
                    "goal_id": mål.id,
                    "trigger": TRIGGER_AGENT,
                    "initiator": self.initiator,
                    "agent_profile": self.profil,
                })),
            },
        )
        .await;

        let Some(kodrunner) = &self.kodrunner else {
            return Ok(());
        };

        kodrunner.process(&körning, mål).await
    }

    /// Definierar ett mål och lämnar resultatet till en människa.
    ///
    /// Go: `Worker.processGoal`.
    async fn processa_mål(
        &self,
        mål: &Goal,
        har_körning: bool,
        befintlig: AgentRun,
    ) -> Result<(), String> {
        info!("Agent worker processing goal #{}: {}", mål.id, mål.title);

        let körning = if har_körning {
            befintlig
        } else {
            self.plan
                .create_agent_run(CreateAgentRunRequest {
                    goal_id: mål.id,
                    phase: Some(PHASE_DEFINE.to_string()),
                    agent_id: Some(self.agent_id.clone()),
                    provider: Some(self.leverantör.clone()),
                    model: Some(self.modell.clone()),
                    status: Some(STATUS_RUNNING.to_string()),
                    trigger: Some(TRIGGER_AGENT.to_string()),
                    initiator: Some(self.initiator.clone()),
                    agent_profile: Some(self.profil.clone()),
                    summary: Some("Agenten specificerar målet".to_string()),
                    ..Default::default()
                })
                .await
                .map_err(|fel| format!("failed to create agent run: {fel}"))?
        };

        self.registrera_event(
            körning.id,
            crate::planclient::CreateAgentRunEventRequest {
                r#type: crate::planclient::models::EVENT_RUN_STARTED.to_string(),
                phase: Some(PHASE_DEFINE.to_string()),
                payload: Some(json!({
                    "goal_id": mål.id,
                    "trigger": TRIGGER_AGENT,
                    "initiator": self.initiator,
                    "agent_profile": self.profil,
                })),
            },
        )
        .await;

        // Platserna behövs för att agenten ska kunna föreslå platser på
        // målet.
        let platser = self.platser().await;

        let input = Input {
            title: mål.title.clone(),
            description: mål.description.clone(),
            feedback: String::new(),
            icon: mål.icon.clone(),
            now: mål.now,
            eternal: mål.eternal,
            agent: mål.agent,
            due: mål.due.clone().unwrap_or_default(),
            start_date: mål.start_date.clone().unwrap_or_default(),
            estimated_minutes: mål.estimated_minutes.unwrap_or(0),
            recurrence_rule: mål.recurrence_rule.clone().unwrap_or_default(),
            tags: mål.tags.clone(),
            places: goaldefinition::platsnamn(&mål.places),
        };

        let definition = match self.definierare.definiera(&input).await {
            Ok(definition) => definition,
            Err(fel) => {
                self.mislyckas_körning(körning.id, &fel).await;

                return Err(format!("failed to define goal: {fel}"));
            }
        };

        self.registrera_event(
            körning.id,
            crate::planclient::CreateAgentRunEventRequest {
                r#type: crate::planclient::models::EVENT_DEFINITION_CREATED.to_string(),
                phase: Some(PHASE_DEFINE.to_string()),
                payload: None,
            },
        )
        .await;

        let förslag = goaldefinition::bygg(&Förslagskontext {
            goal: mål.clone(),
            definition,
            places: platser,
            current_phase: self.fasnamn(mål.status_id).await,
            next_phase: PHASE_PLAN.to_string(),
        });

        // Utan data att föreslå finns inget att godkänna.
        if !goaldefinition::har_datafält(&förslag) {
            info!(
                "Goal #{} has nothing new to propose - moving it to planning",
                mål.id
            );

            return self.avsluta_definition(mål, &körning).await;
        }

        self.plan
            .create_agent_approval(
                körning.id,
                CreateAgentApprovalRequest {
                    title: Self::förslagstitel(&förslag, mål),
                    description: goaldefinition::sammanfattning(&förslag),
                    consequence: goaldefinition::konsekvens(&förslag),
                    verification_plan:
                        "Förslaget är validerat mot Plans regler för datum, ikon, uppskattad tid och upprepning."
                            .to_string(),
                    payload: Some(goaldefinition::payload(&förslag)),
                },
            )
            .await
            .map_err(|fel| format!("failed to create approval: {fel}"))?;

        self.registrera_event(
            körning.id,
            crate::planclient::CreateAgentRunEventRequest {
                r#type: crate::planclient::models::EVENT_APPROVAL_REQUESTED.to_string(),
                phase: Some(PHASE_DEFINE.to_string()),
                payload: Some(json!({
                    "title": Self::förslagstitel(&förslag, mål),
                    "kind": goaldefinition::KIND,
                    "changes": förslag.fields.len(),
                })),
            },
        )
        .await;

        // Körningen väntar på en människa - statusen säger det tydligt.
        let status = STATUS_WAITING_APPROVAL.to_string();
        let summary = "Definitionen inväntar mänskligt godkännande".to_string();

        self.plan
            .update_agent_run(
                körning.id,
                UpdateAgentRunRequest {
                    status: Some(status),
                    summary: Some(summary),
                    ..Default::default()
                },
            )
            .await
            .map_err(|fel| format!("failed to mark run as waiting for approval: {fel}"))?;

        info!("Goal #{} has a definition waiting for approval", mål.id);

        Ok(())
    }

    /// Hanterar ett förslag om målets utseende.
    ///
    /// Returnerar `true` om förslaget är agentens - då ska koderaren inte
    /// köras förrän definitionen är applicerad.
    ///
    /// Go: `Worker.resumeDefinition`.
    async fn återuppta_definition(
        &self, körning: &AgentRun, mål: &Goal
    ) -> Result<bool, String> {
        let godkännanden = self
            .plan
            .agent_approvals(körning.id)
            .await
            .map_err(|fel| format!("failed to check approvals: {fel}"))?;

        let Some(godkännande) = hitta_definitionsgodkännande(&godkännanden) else {
            return Ok(false);
        };

        match godkännande.status.as_str() {
            "approved" => {
                let Some(payload) = godkännande.payload.as_ref() else {
                    return Err(
                        "failed to read the approved definition: saknar payload".to_string()
                    );
                };

                let förslag = match goaldefinition::från_payload(payload) {
                    Ok(förslag) => förslag,
                    Err(fel) => {
                        self.mislyckas_körning(körning.id, &fel).await;

                        return Err(format!("failed to read the approved definition: {fel}"));
                    }
                };

                // Exakt de godkända värdena appliceras - inget genereras
                // om.
                if let Err(fel) = goaldefinition::applicera(&förslag, &self.plan, mål.id).await {
                    self.mislyckas_körning(körning.id, &fel).await;

                    return Err(format!("failed to apply the approved definition: {fel}"));
                }

                info!("Goal #{} got its approved definition", mål.id);

                // Med en coder igång fortsätter samma körning - men med
                // det uppdaterade målet.
                if let Some(kodrunner) = &self.kodrunner {
                    let uppdaterat = match self.plan.goal(mål.id).await {
                        Ok(mål) => mål,
                        Err(fel) => {
                            self.mislyckas_körning(körning.id, &fel.to_string()).await;

                            return Err(format!("failed to fetch the updated goal: {fel}"));
                        }
                    };

                    if let Err(fel) = kodrunner.process(körning, &uppdaterat).await {
                        return Err(format!("failed to hand over to the coder: {fel}"));
                    }

                    return Ok(true);
                }

                self.avsluta_definition(mål, körning).await?;

                Ok(true)
            }
            "rejected" => {
                // Ett avslaget förslag körs inte om - det är människan
                // som avgör om agenten får försöka igen.
                info!(
                    "Goal #{} has a rejected definition - waiting for a human decision",
                    mål.id
                );

                Ok(true)
            }
            // Förslaget väntar fortfarande på beslut.
            _ => Ok(true),
        }
    }

    /// Avslutar en körning där målet är definierat.
    ///
    /// Go: `Worker.finishDefinition`.
    async fn avsluta_definition(&self, mål: &Goal, körning: &AgentRun) -> Result<(), String> {
        if let Err(fel) = self.synka_fas(mål.id, körning.id, PHASE_PLAN).await {
            self.mislyckas_körning(körning.id, &fel).await;

            return Err(format!("failed to synchronize plan phase: {fel}"));
        }

        self.registrera_event(
            körning.id,
            crate::planclient::CreateAgentRunEventRequest {
                r#type: crate::planclient::models::EVENT_PHASE_CHANGED.to_string(),
                phase: Some(PHASE_PLAN.to_string()),
                payload: Some(json!({ "status": STATUS_SUCCEEDED })),
            },
        )
        .await;

        let status = STATUS_SUCCEEDED.to_string();
        let summary = mål.title.clone();
        let result = "Målet definierat och flyttat till plan".to_string();

        self.plan
            .update_agent_run(
                körning.id,
                UpdateAgentRunRequest {
                    status: Some(status),
                    summary: Some(summary),
                    result: Some(result),
                    finished_at: Some(Utc::now()),
                    ..Default::default()
                },
            )
            .await
            .map_err(|fel| format!("failed to update agent run: {fel}"))?;

        // En avslutad körning ska alltid sluta med run_completed.
        self.registrera_event(
            körning.id,
            crate::planclient::CreateAgentRunEventRequest {
                r#type: crate::planclient::models::EVENT_RUN_COMPLETED.to_string(),
                phase: Some(PHASE_PLAN.to_string()),
                payload: Some(json!({ "reason": "goal defined and moved to planning" })),
            },
        )
        .await;

        info!("Goal #{} defined and moved to planning", mål.id);

        Ok(())
    }

    /// Ger förslagets rubrik - den föreslagna titeln om det finns en ny.
    ///
    /// Go: `proposalTitle`.
    fn förslagstitel(förslag: &Proposal, mål: &Goal) -> String {
        if let Some(titel) = &förslag.values.title {
            if !titel.is_empty() {
                return titel.clone();
            }
        }

        mål.title.clone()
    }

    /// Hämtar användarens platser.
    ///
    /// Ett fel här ska inte stoppa agenten.
    ///
    /// Go: `Worker.places`.
    async fn platser(&self) -> Vec<Place> {
        match self.plan.places().await {
            Ok(platser) => platser,
            Err(fel) => {
                warn!("Failed to fetch places: {fel}");

                Vec::new()
            }
        }
    }

    /// Ger fasnamnet för ett status-id.
    ///
    /// Go: `Worker.phaseName`.
    async fn fasnamn(&self, status_id: i32) -> String {
        self.plan.phase_name(status_id).await.unwrap_or_default()
    }

    /// Registrerar en händelse i körningens historik.
    ///
    /// Go: `Worker.recordEvent`.
    async fn registrera_event(
        &self,
        körning_id: i64,
        händelse: crate::planclient::CreateAgentRunEventRequest,
    ) {
        if let Err(fel) = self.plan.create_agent_run_event(körning_id, händelse).await {
            warn!("Failed to record event for agent run {}: {fel}", körning_id);
        }
    }

    /// Synkar fasen mellan målet och körningen.
    ///
    /// Go: `Worker.syncPhase`.
    async fn synka_fas(&self, mål_id: i64, körning_id: i64, fas: &str) -> Result<(), String> {
        self.plan
            .update_goal_phase(mål_id, fas)
            .await
            .map_err(|fel| format!("failed to update goal phase: {fel}"))?;

        self.plan
            .update_agent_run(
                körning_id,
                UpdateAgentRunRequest {
                    phase: Some(fas.to_string()),
                    ..Default::default()
                },
            )
            .await
            .map_err(|fel| format!("failed to update agent run phase: {fel}"))?;

        Ok(())
    }

    /// Rapporterar tokenanvändningen för en körning.
    ///
    /// Plan räknar fram kostnaden ur sin prislista. Ett fel här ska inte
    /// stoppa arbetet.
    ///
    /// Go: `Worker.reportUsage`.
    async fn rapportera_användning(&self, körning_id: i64) {
        let Some(kodrunner) = &self.kodrunner else {
            return;
        };

        let Some(snapshot) = kodrunner.användning() else {
            return;
        };

        if let Err(fel) = self
            .plan
            .record_agent_run_usage(
                körning_id,
                crate::planclient::AgentRunUsageRequest {
                    agent_id: Some(self.agent_id.clone()),
                    provider: Some(snapshot.provider.clone()),
                    model: Some(snapshot.model.clone()),
                    model_version: Some(snapshot.model_version.clone()),
                    input_tokens: snapshot.input_tokens,
                    output_tokens: snapshot.output_tokens,
                    cache_read_tokens: snapshot.cache_read_tokens,
                    tool_calls: snapshot.tool_calls,
                    tool_names: Some(snapshot.tool_names.clone()),
                    started_at: Some(snapshot.started_at),
                    finished_at: Some(snapshot.finished_at),
                },
            )
            .await
        {
            warn!("Failed to report usage for run {körning_id}: {fel}");
            return;
        }

        info!(
            "Reported usage for run {}: {} input, {} output tokens",
            körning_id, snapshot.input_tokens, snapshot.output_tokens
        );
    }

    /// Markerar en körning som misslyckad.
    ///
    /// Go: `Worker.failRun`.
    async fn mislyckas_körning(&self, körning_id: i64, orsak: &str) {
        let status = STATUS_FAILED.to_string();
        let summary = format!("Agenten kunde inte specificera målet: {orsak}");
        let result = format!("Misslyckades: {orsak}");

        if let Err(fel) = self
            .plan
            .update_agent_run(
                körning_id,
                UpdateAgentRunRequest {
                    status: Some(status),
                    summary: Some(summary),
                    result: Some(result),
                    finished_at: Some(Utc::now()),
                    ..Default::default()
                },
            )
            .await
        {
            warn!("Failed to mark agent run {körning_id} as failed: {fel}");
        }

        self.registrera_event(
            körning_id,
            crate::planclient::CreateAgentRunEventRequest {
                r#type: crate::planclient::models::EVENT_RUN_FAILED.to_string(),
                phase: Some(PHASE_DEFINE.to_string()),
                payload: Some(json!({ "error": orsak })),
            },
        )
        .await;

        self.rapportera_användning(körning_id).await;
    }
}

/// De köade körningarna, per mål.
///
/// Go: `findQueuedRun`.
pub fn hitta_köade_körning(körningar: &[AgentRun]) -> HashMap<i64, AgentRun> {
    let mut köade: HashMap<i64, AgentRun> = HashMap::new();

    for körning in körningar {
        if körning.status != STATUS_QUEUED {
            continue;
        }

        let entry = köade
            .entry(körning.goal_id)
            .or_insert_with(|| körning.clone());

        if körning.id > entry.id {
            *entry = körning.clone();
        }
    }

    köade
}

/// Ger förslaget om målets utseende, om det finns.
///
/// Go: `findDefinitionApproval`.
pub fn hitta_definitionsgodkännande(
    godkännanden: &[crate::planclient::models::AgentApproval],
) -> Option<crate::planclient::models::AgentApproval> {
    godkännanden
        .iter()
        .find(|godkännande| godkännande.approval_kind() == goaldefinition::KIND)
        .cloned()
}

/// Pollintervallet ur miljön.
///
/// Go: `PollIntervalFromEnv`.
pub fn pollintervall_ur_miljö() -> Duration {
    let rå = std::env::var("AGENT_POLL_INTERVAL").unwrap_or_default();

    if rå.is_empty() {
        return Duration::from_secs(60);
    }

    match parse_duration(&rå) {
        Some(interval) => interval,
        None => {
            warn!("Invalid AGENT_POLL_INTERVAL {rå:?}, using one minute");

            Duration::from_secs(60)
        }
    }
}

/// Tolkar en Go-duration ("30s", "2m", "1h").
fn parse_duration(rå: &str) -> Option<Duration> {
    let trimmad = rå.trim();
    let (siffra, enhet) = trimmad.split_at(
        trimmad
            .find(|tecken: char| !tecken.is_ascii_digit())
            .unwrap_or(trimmad.len()),
    );

    let tal: u64 = siffra.parse().ok()?;
    let sekunder = match enhet.trim() {
        "" | "s" => tal,
        "m" => tal * 60,
        "h" => tal * 3600,
        "ms" => return Some(Duration::from_millis(tal)),
        _ => return None,
    };

    Some(Duration::from_secs(sekunder))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn durationer_tolkas() {
        assert_eq!(parse_duration("30s"), Some(Duration::from_secs(30)));
        assert_eq!(parse_duration("2m"), Some(Duration::from_secs(120)));
        assert_eq!(parse_duration("1h"), Some(Duration::from_secs(3600)));
        assert_eq!(parse_duration("500ms"), Some(Duration::from_millis(500)));
        assert_eq!(parse_duration("skräp"), None);
    }

    #[test]
    fn köade_körningar_finns_per_mål() {
        let körningar = vec![
            AgentRun {
                id: 1,
                goal_id: 10,
                status: STATUS_QUEUED.to_string(),
                ..AgentRun::empty()
            },
            AgentRun {
                id: 2,
                goal_id: 10,
                status: STATUS_QUEUED.to_string(),
                ..AgentRun::empty()
            },
            AgentRun {
                id: 3,
                goal_id: 11,
                status: STATUS_RUNNING.to_string(),
                ..AgentRun::empty()
            },
        ];

        let köade = hitta_köade_körning(&körningar);

        assert_eq!(köade.len(), 1);
        assert_eq!(köade.get(&10).unwrap().id, 2);
    }
}
