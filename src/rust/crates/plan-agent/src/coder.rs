//! Driver en kodändring från förslag till verifierat resultat.
//!
//! Go: `internal/coder`. Flödet är: agenten föreslår en ändring, en människa
//! godkänner den, ändringen utförs i en isolerad workspace och resultatet
//! verifieras med projektets egna tester. Först när människan sagt ja
//! ändras något.

use std::pin::Pin;
use std::sync::Arc;

use chrono::Utc;
use serde_json::json;
use tracing::{info, warn};

use crate::executor::{Change, Executor, Request, Utfall};
use crate::planclient::models::{
    AgentApproval, AgentRun, Goal, APPROVAL_KIND_CODE_CHANGE, EVENT_APPROVAL_REQUESTED,
    EVENT_ARTIFACT, EVENT_PHASE_CHANGED, EVENT_RUN_COMPLETED, EVENT_RUN_FAILED, EVENT_VERIFICATION,
};
use crate::planclient::{
    CreateAgentApprovalRequest, CreateAgentRunEventRequest, PlanClient, UpdateAgentRunRequest,
};
use crate::profile;
use crate::proposer::Proposal;
use crate::runusage::{Recorder, Snapshot};
use crate::workspace::Workspace;

/// En framtid som äger sina referenser.
type Framtid<'a, T> = Pin<Box<dyn std::future::Future<Output = T> + Send + 'a>>;

/// Föreslår kodändringar.
///
/// Go: `coder.Proposer`.
pub trait ProposalProvider: Send + Sync {
    /// Läser repot och föreslår en ändring.
    fn föreslå<'a>(
        &'a self,
        ws: &'a Workspace,
        mål_titel: &'a str,
        mål_beskrivning: &'a str,
    ) -> Framtid<'a, Result<Proposal, String>>;

    /// Skriver om testkommandot så att det går igenom på innehållet.
    fn reparera_test<'a>(
        &'a self,
        ws: &'a Workspace,
        mål_titel: &'a str,
        mål_beskrivning: &'a str,
        förslag: &'a Proposal,
        fel: &'a str,
    ) -> Framtid<'a, Result<String, String>>;

    /// Rättar ett förslag vars egen verifiering inte gick igenom.
    fn revidera<'a>(
        &'a self,
        ws: &'a Workspace,
        mål_titel: &'a str,
        mål_beskrivning: &'a str,
        förslag: &'a Proposal,
        fel: &'a str,
    ) -> Framtid<'a, Result<Proposal, String>>;
}

impl ProposalProvider for crate::proposer::Proposer {
    fn föreslå<'a>(
        &'a self,
        ws: &'a Workspace,
        mål_titel: &'a str,
        mål_beskrivning: &'a str,
    ) -> Framtid<'a, Result<Proposal, String>> {
        Box::pin(async move {
            crate::proposer::Proposer::föreslå(self, ws, mål_titel, mål_beskrivning).await
        })
    }

    fn reparera_test<'a>(
        &'a self,
        ws: &'a Workspace,
        mål_titel: &'a str,
        mål_beskrivning: &'a str,
        förslag: &'a Proposal,
        fel: &'a str,
    ) -> Framtid<'a, Result<String, String>> {
        Box::pin(async move {
            crate::proposer::Proposer::reparera_test(
                self,
                ws,
                mål_titel,
                mål_beskrivning,
                förslag,
                fel,
            )
            .await
        })
    }

    fn revidera<'a>(
        &'a self,
        ws: &'a Workspace,
        mål_titel: &'a str,
        mål_beskrivning: &'a str,
        förslag: &'a Proposal,
        fel: &'a str,
    ) -> Framtid<'a, Result<Proposal, String>> {
        Box::pin(async move {
            crate::proposer::Proposer::revidera(self, ws, mål_titel, mål_beskrivning, förslag, fel)
                .await
        })
    }
}

/// Driver en kodändring genom hela flödet.
pub struct Coder {
    /// Klienten mot Plan.
    plan: PlanClient,
    /// Föreslår kodändringar.
    proposer: Arc<dyn ProposalProvider>,
    /// Samlar tokenanvändningen för körningen.
    användning: Option<Arc<Recorder>>,
    /// Repot som kodändringen görs i.
    repo_path: String,
    /// Begränsar hur stor del av diffen som sparas i historiken.
    max_diff_lines: usize,
}

/// Konfigurerar codern.
#[derive(Clone, Default)]
pub struct Config {
    /// Repot som ska ändras.
    pub repo_path: String,
    /// Begränsar diffen som sparas i historiken.
    pub max_diff_lines: usize,
    /// Samlar tokenanvändningen för körningen.
    pub användning: Option<Arc<Recorder>>,
}

impl Coder {
    /// Skapar en coder.
    ///
    /// Go: `coder.New`.
    pub fn ny(plan: PlanClient, proposer: Arc<dyn ProposalProvider>, config: Config) -> Self {
        let max_diff_lines = if config.max_diff_lines == 0 {
            200
        } else {
            config.max_diff_lines
        };

        Self {
            plan,
            proposer,
            användning: config.användning,
            repo_path: config.repo_path,
            max_diff_lines,
        }
    }

    /// Den samlade tokenanvändningen för körningen.
    ///
    /// Go: `Coder.Usage`.
    pub fn användning(&self) -> Option<Snapshot> {
        self.användning.as_ref().map(|recorder| recorder.snapshot())
    }

    /// Tar hand om ett mål som är redo för kodändring.
    ///
    /// Körningen måste vara godkänd av en människa innan något ändras. Är
    /// den inte godkänd avbryts körningen utan att något skrivs.
    ///
    /// Go: `Coder.Process`.
    pub async fn process(&self, run: &AgentRun, goal: &Goal) -> Result<(), String> {
        info!("Coder processing goal #{}: {}", goal.id, goal.title);

        // Profilen avgör om körningen får röra kod. Det är en teknisk
        // gräns - inte något modellen kan övertala.
        let policy = profile::policy_for(&run.agent_profile)
            .map_err(|fel| format!("failed to resolve agent profile: {fel}"))?;

        if !policy.may_change_code {
            info!(
                "Profile {} may not change code - skipping goal #{}",
                policy.name, goal.id
            );

            return Ok(());
        }

        let godkännanden = self
            .plan
            .agent_approvals(run.id)
            .await
            .map_err(|fel| format!("failed to check approvals: {fel}"))?;

        if let Some(godkänt) = hitta_godkänt(&godkännanden) {
            return self.execute(run, goal, &godkänt).await;
        }

        // Ett förslag som redan väntar på beslut ska inte ersättas av ett
        // nytt.
        if hitta_väntande(&godkännanden).is_some() {
            info!(
                "Goal #{} already has a proposal waiting for a decision",
                goal.id
            );

            return Ok(());
        }

        // Ett avslaget förslag ska inte köras om - det är människan som
        // avgör om agenten får försöka igen.
        if hitta_avslaget(&godkännanden).is_some() {
            info!(
                "Goal #{} has a rejected proposal - waiting for a human decision",
                goal.id
            );

            return Ok(());
        }

        self.propose(run, goal).await
    }

    /// Tar fram ett förslag och lämnar det till en människa.
    ///
    /// Förslaget verifieras först med sitt eget testkommando. Det ger två
    /// saker: modellen får chans att rätta ett trasigt förslag, och
    /// människan ser om förslaget faktiskt håller när det lämnas över.
    ///
    /// Go: `Coder.propose`.
    async fn propose(&self, run: &AgentRun, goal: &Goal) -> Result<(), String> {
        let ws = Workspace::ny(&self.repo_path)
            .map_err(|fel| format!("failed to create workspace: {fel}"))?;

        let mut förslag = match self
            .proposer
            .föreslå(&ws, &goal.title, &goal.description)
            .await
        {
            Ok(förslag) => förslag,
            Err(fel) => {
                self.mislyckas(run.id, &fel).await;

                return Err(format!("failed to propose change: {fel}"));
            }
        };

        let mut verifiering = self.verifiera_förslag(&förslag).await;

        if !verifiering.passed {
            info!(
                "Proposal for goal #{} failed its own verification - asking the model to fix the test",
                goal.id
            );

            // Först provas ett nytt testkommando. Innehållet behålls - det
            // är oftast testet som är för strikt.
            match self
                .proposer
                .reparera_test(
                    &ws,
                    &goal.title,
                    &goal.description,
                    &förslag,
                    &verifiering.output,
                )
                .await
            {
                Ok(kommando) if !kommando.is_empty() => {
                    förslag.test_command = kommando;
                    verifiering = self.verifiera_förslag(&förslag).await;
                }
                Ok(_) => {}
                Err(fel) => {
                    warn!(
                        "Failed to get a new test command for goal #{}: {fel}",
                        goal.id
                    );
                }
            }
        }

        if !verifiering.passed {
            info!(
                "Proposal for goal #{} still fails - asking the model to revise the whole proposal",
                goal.id
            );

            match self
                .proposer
                .revidera(
                    &ws,
                    &goal.title,
                    &goal.description,
                    &förslag,
                    &verifiering.output,
                )
                .await
            {
                Ok(reviderat) => {
                    förslag = reviderat;
                    verifiering = self.verifiera_förslag(&förslag).await;
                }
                Err(fel) => warn!("Failed to revise proposal for goal #{}: {fel}", goal.id),
            }
        }

        // Förslaget sparas i godkännandet, så att exakt det som människan
        // godkänner är det som appliceras senare.
        let mut payload = förslag_payload(&förslag);
        payload["pre_verification"] = json!({
            "passed": verifiering.passed,
            "output": korta(&verifiering.output, 20),
        });

        self.plan
            .create_agent_approval(
                run.id,
                CreateAgentApprovalRequest {
                    title: förslag.title.clone(),
                    description: förslagsbeskrivning(&förslag, &verifiering),
                    consequence: beskriv_konsekvens(&förslag),
                    verification_plan: förslag.test_command.clone(),
                    payload: Some(payload),
                },
            )
            .await
            .map_err(|fel| format!("failed to create approval: {fel}"))?;

        self.registrera(
            run.id,
            EVENT_APPROVAL_REQUESTED,
            Some(json!({
                "title": förslag.title,
                "changed_files": ändrade_filer(&förslag),
                "verification_plan": förslag.test_command,
                "pre_verified": verifiering.passed,
            })),
        )
        .await;

        // Körningen väntar på en människa - statusen säger det tydligt.
        let status = crate::worker::STATUS_WAITING_APPROVAL.to_string();
        let summary = format!("Förslag inväntar mänskligt godkännande: {}", förslag.title);

        self.plan
            .update_agent_run(
                run.id,
                UpdateAgentRunRequest {
                    status: Some(status),
                    summary: Some(summary),
                    ..Default::default()
                },
            )
            .await
            .map_err(|fel| format!("failed to mark run as blocked: {fel}"))?;

        info!(
            "Goal #{} has a proposal waiting for approval: {}",
            goal.id, förslag.title
        );

        Ok(())
    }

    /// Utför den godkända ändringen och verifierar den.
    ///
    /// Ändringen kommer från godkännandet - inte från en ny modellkörning.
    /// Därför är det garanterat att det som appliceras är det en människa
    /// godkände.
    ///
    /// Go: `Coder.execute`.
    async fn execute(
        &self,
        run: &AgentRun,
        goal: &Goal,
        godkänt: &AgentApproval,
    ) -> Result<(), String> {
        let (ändringar, testkommando) = match ändringar_från_godkännande(godkänt) {
            Ok(värden) => värden,
            Err(fel) => {
                self.mislyckas(run.id, &fel).await;

                return Err(format!("failed to read approved change: {fel}"));
            }
        };

        // Körningen går från väntan till utförande.
        self.starta_utförande(run, goal).await?;

        let mut exekverare = match Executor::ny(&self.repo_path) {
            Ok(exekverare) => exekverare,
            Err(fel) => {
                self.mislyckas(run.id, &fel).await;

                return Err(format!("failed to create executor: {fel}"));
            }
        };

        if let Some(recorder) = &self.användning {
            exekverare = exekverare.med_användning(recorder.clone());
        }

        let resultat = match exekverare
            .execute(&Request {
                changes: ändringar,
                test_command: testkommando.clone(),
            })
            .await
        {
            Ok(resultat) => resultat,
            Err(fel) => {
                self.mislyckas(run.id, &fel).await;

                return Err(format!("failed to execute change: {fel}"));
            }
        };

        self.registrera(
            run.id,
            EVENT_ARTIFACT,
            Some(json!({
                "kind": "diff",
                "changed_files": resultat.changed_files,
                "diff": resultat.diff_förhandsvisning(self.max_diff_lines),
            })),
        )
        .await;

        self.registrera(
            run.id,
            EVENT_VERIFICATION,
            Some(json!({
                "tests_passed": resultat.tests_passed,
                "test_command": testkommando,
                "test_output": korta(&resultat.test_output, self.max_diff_lines),
                "test_error": resultat.test_error,
            })),
        )
        .await;

        if !resultat.tests_passed {
            self.mislyckas(run.id, "testerna gick inte igenom").await;

            return Err(format!("verification failed for goal #{}", goal.id));
        }

        // Den godkända ändringen landar på en egen branch - huvudgrenen
        // rörs aldrig, en människa väljer om den ska mergeas.
        //
        // Branchen bär körningens nummer, så varje försök får sin egen
        // branch att granska. Utan det skulle ett nytt försök krocka med
        // den förra branchen i repot.
        let branch = format!("agent/goal-{}-run-{}", goal.id, run.id);
        let mut pr_url = String::new();
        let mut branch_publicerad = false;

        match exekverare.committa(&branch, &commitmeddelande(goal, &resultat)) {
            Ok(()) => {
                branch_publicerad = true;

                self.registrera(
                    run.id,
                    EVENT_ARTIFACT,
                    Some(json!({ "kind": "branch", "branch": branch })),
                )
                .await;

                let pr_titel = format!("Agent: {}", goal.title);
                let pr_text = pull_request_text(run, goal, &resultat, &testkommando);

                match exekverare.skapa_pull_request(&pr_titel, &pr_text, "main") {
                    Ok(url) => {
                        pr_url = url;

                        self.registrera(
                            run.id,
                            EVENT_ARTIFACT,
                            Some(json!({
                                "kind": "pull_request",
                                "status": "created",
                                "branch": branch,
                                "url": pr_url,
                            })),
                        )
                        .await;
                    }
                    Err(fel) => {
                        // En publicerad branch är fortfarande ett användbart
                        // resultat. PR-leveransen rapporteras separat så att
                        // ett tillfälligt gh- eller GitHub-fel inte döljer
                        // att ändringen är verifierad.
                        warn!(
                            "Goal #{} branch published but PR creation failed: {fel}",
                            goal.id
                        );

                        self.registrera(
                            run.id,
                            EVENT_ARTIFACT,
                            Some(json!({
                                "kind": "pull_request",
                                "status": "failed",
                                "branch": branch,
                                "error": fel,
                            })),
                        )
                        .await;
                    }
                }
            }
            Err(fel) => {
                if !fel.contains(crate::workspace::ErrNotPublished.to_string().as_str()) {
                    self.mislyckas(run.id, &fel).await;

                    return Err(format!("failed to publish the change: {fel}"));
                }

                // Ändringen är gjord och verifierad men kom inte ut till
                // repot. Det är inte ett fel i ändringen - det är en sak
                // som behöver nämnas tydligt i historiken.
                info!(
                    "Goal #{} was verified but could not be published: {fel}",
                    goal.id
                );

                self.registrera(
                    run.id,
                    EVENT_ARTIFACT,
                    Some(json!({
                        "kind": "branch",
                        "branch": branch,
                        "publish": "failed",
                        "error": fel,
                    })),
                )
                .await;
            }
        }

        let status = crate::worker::STATUS_SUCCEEDED.to_string();
        let phase = crate::worker::PHASE_VERIFY.to_string();

        let mut result_text = if branch_publicerad {
            format!("Ändringen är gjord på branchen {branch} och testerna gick igenom")
        } else {
            "Ändringen är gjord och testerna gick igenom, men branchen kunde inte publiceras"
                .to_string()
        };

        if !pr_url.is_empty() {
            result_text.push_str(&format!(" Pull Request: {pr_url}"));
        }

        let mut uppdatering = UpdateAgentRunRequest {
            status: Some(status),
            phase: Some(phase.clone()),
            summary: Some(result_text.clone()),
            result: Some(result_text),
            finished_at: Some(Utc::now()),
            ..Default::default()
        };

        if !pr_url.is_empty() {
            uppdatering.artifact_url = Some(pr_url);
        }

        self.plan
            .update_agent_run(run.id, uppdatering)
            .await
            .map_err(|fel| format!("failed to update agent run: {fel}"))?;

        // Målet går till verifiering - en människa avgör om det är klart.
        self.plan
            .update_goal_phase(goal.id, &phase)
            .await
            .map_err(|fel| format!("failed to move goal to verification: {fel}"))?;

        self.registrera(
            run.id,
            EVENT_RUN_COMPLETED,
            Some(json!({
                "changed_files": resultat.changed_files,
                "branch": branch,
            })),
        )
        .await;

        self.rapportera_användning(run.id).await;

        info!("Goal #{} executed and verified", goal.id);

        Ok(())
    }

    /// Markerar att körningen börjar utföra den godkända ändringen.
    ///
    /// Go: `Coder.startExecuting`.
    async fn starta_utförande(&self, run: &AgentRun, goal: &Goal) -> Result<(), String> {
        let phase = crate::worker::PHASE_EXECUTE.to_string();
        let status = crate::worker::STATUS_RUNNING.to_string();
        let summary = "Agenten utför den godkända ändringen".to_string();

        self.plan
            .update_agent_run(
                run.id,
                UpdateAgentRunRequest {
                    phase: Some(phase.clone()),
                    status: Some(status.clone()),
                    summary: Some(summary),
                    ..Default::default()
                },
            )
            .await
            .map_err(|fel| format!("failed to start execution: {fel}"))?;

        self.plan
            .update_goal_phase(goal.id, &phase)
            .await
            .map_err(|fel| format!("failed to move goal to execution: {fel}"))?;

        self.registrera(
            run.id,
            EVENT_PHASE_CHANGED,
            Some(json!({ "phase": phase, "status": status })),
        )
        .await;

        Ok(())
    }

    /// Applicerar förslaget i en isolerad kopia och kör dess eget
    /// testkommando.
    ///
    /// Ingenting rapporteras till Plan här - verifieringen är bara till för
    /// att fånga trasiga förslag innan en människa läser dem.
    ///
    /// Go: `Coder.verifyProposal`.
    async fn verifiera_förslag(&self, förslag: &Proposal) -> Verifiering {
        if förslag.test_command.is_empty() {
            return Verifiering {
                passed: true,
                output: String::new(),
            };
        }

        let Ok(mut exekverare) = Executor::ny(&self.repo_path) else {
            return Verifiering {
                passed: false,
                output: "failed to create executor".to_string(),
            };
        };

        if let Some(recorder) = &self.användning {
            exekverare = exekverare.med_användning(recorder.clone());
        }

        let resultat = match exekverare
            .execute(&Request {
                changes: förslag.changes.clone(),
                test_command: förslag.test_command.clone(),
            })
            .await
        {
            Ok(resultat) => resultat,
            Err(fel) => {
                return Verifiering {
                    passed: false,
                    output: fel,
                }
            }
        };

        let mut output = resultat.test_output.clone();

        if !resultat.tests_passed && !resultat.test_error.is_empty() {
            output = format!("{}\n{}", resultat.test_error, output);
        }

        Verifiering {
            passed: resultat.tests_passed,
            output,
        }
    }

    /// Registrerar en händelse i körningens historik.
    ///
    /// Go: `Coder.record`.
    async fn registrera(&self, körning_id: i64, typ: &str, payload: Option<serde_json::Value>) {
        if let Err(fel) = self
            .plan
            .create_agent_run_event(
                körning_id,
                CreateAgentRunEventRequest {
                    r#type: typ.to_string(),
                    phase: None,
                    payload,
                },
            )
            .await
        {
            warn!("Failed to record {typ} event for agent run {körning_id}: {fel}");
        }
    }

    /// Markerar körningen som misslyckad.
    ///
    /// Go: `Coder.fail`.
    async fn mislyckas(&self, körning_id: i64, orsak: &str) {
        let status = crate::worker::STATUS_FAILED.to_string();
        let summary = format!("Kodändringen misslyckades: {orsak}");
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

        self.registrera(
            körning_id,
            EVENT_RUN_FAILED,
            Some(json!({ "error": orsak })),
        )
        .await;
    }

    /// Rapporterar tokenanvändningen för körningen till Plan.
    ///
    /// Plan räknar fram kostnaden. Ett fel här får inte stoppa körningen -
    /// kostnaden kan rapporteras i efterhand.
    ///
    /// Go: `Coder.reportUsage`.
    async fn rapportera_användning(&self, körning_id: i64) {
        let Some(recorder) = &self.användning else {
            return;
        };

        let snapshot = recorder.snapshot();

        if let Err(fel) = self
            .plan
            .record_agent_run_usage(
                körning_id,
                crate::planclient::AgentRunUsageRequest {
                    agent_id: None,
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
        }
    }
}

impl crate::worker::CodeRunner for Coder {
    fn process<'a>(
        &'a self,
        run: &'a AgentRun,
        goal: &'a Goal,
    ) -> Pin<Box<dyn std::future::Future<Output = Result<(), String>> + Send + 'a>> {
        Box::pin(async move { Coder::process(self, run, goal).await })
    }

    fn användning(&self) -> Option<Snapshot> {
        Coder::användning(self)
    }
}

/// Utfallet av en förslagsverifiering.
#[derive(Debug, Clone, Default)]
pub struct Verifiering {
    /// Säger om testkommandot gick igenom.
    pub passed: bool,
    /// Utdatan från testkörningen.
    pub output: String,
}

/// Ger det godkända förslaget, om det finns.
///
/// Bara förslag av den givna typen räknas - ett godkänt förslag om
/// målets utseende ska inte köras som en kodändring.
///
/// Go: `findApproved`.
pub fn hitta_godkänt(godkännanden: &[AgentApproval]) -> Option<AgentApproval> {
    godkännanden
        .iter()
        .find(|godkännande| {
            godkännande.status == "approved"
                && godkännande.approval_kind() == APPROVAL_KIND_CODE_CHANGE
        })
        .cloned()
}

/// Ger det förslag som fortfarande väntar på beslut.
///
/// Go: `findPending`.
pub fn hitta_väntande(godkännanden: &[AgentApproval]) -> Option<AgentApproval> {
    godkännanden
        .iter()
        .find(|godkännande| {
            godkännande.status == "pending"
                && godkännande.approval_kind() == APPROVAL_KIND_CODE_CHANGE
        })
        .cloned()
}

/// Ger det förslag som en människa har avslagit.
///
/// Go: `findRejected`.
pub fn hitta_avslaget(godkännanden: &[AgentApproval]) -> Option<AgentApproval> {
    godkännanden
        .iter()
        .find(|godkännande| {
            godkännande.status == "rejected"
                && godkännande.approval_kind() == APPROVAL_KIND_CODE_CHANGE
        })
        .cloned()
}

/// Gör om ett förslag till något som kan sparas i Plan.
///
/// Go: `proposalPayload`.
pub fn förslag_payload(förslag: &Proposal) -> serde_json::Value {
    let ändringar: Vec<serde_json::Value> = förslag
        .changes
        .iter()
        .map(|ändring| {
            json!({
                "path": ändring.path,
                "content": ändring.content,
                "reason": ändring.reason,
            })
        })
        .collect();

    json!({
        "kind": APPROVAL_KIND_CODE_CHANGE,
        "changes": ändringar,
        "test_command": förslag.test_command,
    })
}

/// Läser de godkända ändringarna ur ett godkännande.
///
/// Go: `changesFromApproval`.
pub fn ändringar_från_godkännande(
    godkännande: &AgentApproval,
) -> Result<(Vec<Change>, String), String> {
    let Some(payload) = godkännande.payload.as_ref() else {
        return Err("godkännandet saknar det föreslagna innehållet".to_string());
    };

    let Some(rå) = payload.get("changes").and_then(|v| v.as_array()) else {
        return Err("godkännandet innehåller inga ändringar".to_string());
    };

    if rå.is_empty() {
        return Err("godkännandet innehåller inga ändringar".to_string());
    }

    let mut ändringar: Vec<Change> = Vec::with_capacity(rå.len());

    for post in rå {
        let Some(entry) = post.as_object() else {
            return Err("en ändring i godkännandet har fel format".to_string());
        };

        let sökväg = entry
            .get("path")
            .and_then(|v| v.as_str())
            .unwrap_or_default();

        if sökväg.is_empty() {
            return Err("en ändring i godkännandet saknar sökväg".to_string());
        }

        let innehåll = entry
            .get("content")
            .and_then(|v| v.as_str())
            .unwrap_or_default();

        let orsak = entry
            .get("reason")
            .and_then(|v| v.as_str())
            .unwrap_or_default();

        ändringar.push(Change {
            path: sökväg.to_string(),
            content: innehåll.to_string(),
            reason: orsak.to_string(),
        });
    }

    let mut testkommando = payload
        .get("test_command")
        .and_then(|v| v.as_str())
        .unwrap_or_default()
        .to_string();

    if testkommando.is_empty() {
        testkommando = godkännande.verification_plan.clone();
    }

    Ok((ändringar, testkommando))
}

/// Beskriver förslaget och hur det verifierades.
///
/// Go: `proposalDescription`.
pub fn förslagsbeskrivning(förslag: &Proposal, verifiering: &Verifiering) -> String {
    let mut beskrivning = förslag.description.clone();

    if beskrivning.is_empty() {
        beskrivning = förslag.title.clone();
    }

    if verifiering.passed {
        return format!("{beskrivning}\n\nFörslaget har verifierats med sitt eget testkommando.");
    }

    format!(
        "{beskrivning}\n\nVarning: förslagets eget testkommando gick inte igenom:\n{}",
        korta(&verifiering.output, 20)
    )
}

/// Beskriver vad förslaget innebär.
///
/// Go: `describeConsequence`.
pub fn beskriv_konsekvens(förslag: &Proposal) -> String {
    let filer = ändrade_filer(förslag);

    if filer.is_empty() {
        return "Inga filer ändras.".to_string();
    }

    format!("{} fil(er) ändras: {:?}", filer.len(), filer)
}

/// Listar filerna i ett förslag.
///
/// Go: `changedFiles`.
pub fn ändrade_filer(förslag: &Proposal) -> Vec<String> {
    förslag
        .changes
        .iter()
        .map(|ändring| ändring.path.clone())
        .collect()
}

/// Beskriver ändringen i committen.
///
/// Go: `commitMessage`.
pub fn commitmeddelande(goal: &Goal, resultat: &Utfall) -> String {
    format!(
        "{}\n\nÄndringen är gjord av agenten för mål #{}.\nÄndrade filer: {:?}",
        goal.title, goal.id, resultat.changed_files
    )
}

/// Bygger brödtexten till Pull Requesten.
///
/// Go: `pullRequestBody`.
pub fn pull_request_text(
    run: &AgentRun,
    goal: &Goal,
    resultat: &Utfall,
    testkommando: &str,
) -> String {
    format!(
        "## AgentRun #{}\n\nMål: #{} – {}\n\n### Sammanfattning\nÄndringen skapades efter mänskligt godkännande av förslaget och verifierades i den isolerade workspacen.\n\n### Ändrade filer\n{}\n\n### Verifiering\n{}\n\nAgentRun: {}\nBranch: {}\n",
        run.id,
        goal.id,
        goal.title,
        resultat.changed_files.join("\n"),
        testkommando,
        run.id,
        resultat.branch
    )
}

/// Kortar av en lång text till ett antal rader.
///
/// Go: `truncate`.
pub fn korta(text: &str, max_rader: usize) -> String {
    if max_rader == 0 {
        return text.to_string();
    }

    let rader: Vec<&str> = text.split('\n').collect();

    if rader.len() <= max_rader {
        return text.to_string();
    }

    format!("{}\n...(förkortat)", rader[..max_rader].join("\n"))
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::planclient::models::AgentApproval;

    fn godkännande(payload: serde_json::Value) -> AgentApproval {
        AgentApproval {
            id: 1,
            agent_run_id: 1,
            goal_id: 1,
            title: "Ändring".to_string(),
            description: String::new(),
            status: "approved".to_string(),
            feedback: String::new(),
            consequence: String::new(),
            verification_plan: "go test ./...".to_string(),
            payload: Some(payload),
            decided_at: None,
        }
    }

    #[test]
    fn ändringar_läses_ur_godkännandet() {
        let godkänt = godkännande(json!({
            "kind": APPROVAL_KIND_CODE_CHANGE,
            "changes": [{"path": "a.go", "content": "x", "reason": "bugg"}],
            "test_command": "go test ./...",
        }));

        let (ändringar, testkommando) = ändringar_från_godkännande(&godkänt).unwrap();

        assert_eq!(ändringar.len(), 1);
        assert_eq!(ändringar[0].path, "a.go");
        assert_eq!(ändringar[0].content, "x");
        assert_eq!(testkommando, "go test ./...");
    }

    #[test]
    fn testkommandot_falls_back_till_verifieringsplanen() {
        let mut godkänt = godkännande(json!({
            "kind": APPROVAL_KIND_CODE_CHANGE,
            "changes": [{"path": "a.go", "content": "x"}],
        }));

        godkänt.verification_plan = "cargo test".to_string();

        let (_, testkommando) = ändringar_från_godkännande(&godkänt).unwrap();

        assert_eq!(testkommando, "cargo test");
    }

    #[test]
    fn tomma_ändringar_avvisas() {
        let godkänt = godkännande(json!({ "changes": [] }));

        assert_eq!(
            ändringar_från_godkännande(&godkänt).unwrap_err(),
            "godkännandet innehåller inga ändringar"
        );
    }

    #[test]
    fn saknad_payload_avvisas() {
        let mut godkänt = godkännande(json!({}));
        godkänt.payload = None;

        assert_eq!(
            ändringar_från_godkännande(&godkänt).unwrap_err(),
            "godkännandet saknar det föreslagna innehållet"
        );
    }

    #[test]
    fn endast_kodförslag_räknas() {
        let mut godkänt = godkännande(json!({}));
        godkänt.status = "pending".to_string();

        assert!(hitta_väntande(&[godkänt.clone()]).is_some());
        assert!(hitta_godkänt(&[godkänt.clone()]).is_none());

        godkänt.status = "approved".to_string();

        assert!(hitta_godkänt(&[godkänt.clone()]).is_some());

        godkänt.status = "rejected".to_string();

        assert!(hitta_avslaget(&[godkänt]).is_some());
    }

    #[test]
    fn text_förkortas() {
        assert_eq!(korta("a\nb\nc", 2), "a\nb\n...(förkortat)");
        assert_eq!(korta("a\nb", 5), "a\nb");
    }
}
