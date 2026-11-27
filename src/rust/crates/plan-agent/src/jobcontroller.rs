//! Håller ihop Plan och Kubernetes.
//!
//! Go: `internal/jobcontroller`. Plan äger målen och körningarna. Klustert
//! äger jobben. Controllern gör de två världarna mötas: en körning som
//! står i kö - en uttrycklig begäran om att få köras - får ett eget
//! Kubernetes Job, och jobbets utfall förs tillbaka till Plan.

use std::time::Duration;

use tracing::{info, warn};

use crate::jobrunner::{self, Runner, Spec};
use crate::k8s::types::Job;
use crate::k8s::KubeError;
use crate::planclient::models::{
    AgentRun, EVENT_RUN_FAILED, RUN_STATUS_CANCELLED, RUN_STATUS_FAILED, RUN_STATUS_QUEUED,
    RUN_STATUS_RUNNING, RUN_STATUS_SUCCEEDED,
};
use crate::planclient::{CreateAgentRunEventRequest, PlanClient, UpdateAgentRunRequest};

/// Styr hur controllern arbetar.
#[derive(Debug, Clone)]
pub struct Config {
    /// Hur ofta Plan och klustret läses av.
    pub pollintervall: Duration,
    /// Namespace jobben skapas i.
    pub namespace: String,
    /// Begränsar hur många agentjobb som får köra samtidigt.
    ///
    /// Noll använder standardvärdet fyra.
    pub max_concurrent_jobs: usize,
}

/// Håller Plan och Kubernetes i synk.
#[derive(Debug, Clone)]
pub struct Controller {
    /// Klienten mot Plan.
    plan: PlanClient,
    /// Runnern som skapar jobb.
    runner: Runner,
    /// Konfigurationen.
    config: Config,
}

impl Controller {
    /// Skapar en controller.
    ///
    /// Go: `jobcontroller.New`.
    pub fn ny(plan: PlanClient, runner: Runner, config: Config) -> Self {
        let pollintervall = if config.pollintervall.is_zero() {
            Duration::from_secs(60)
        } else {
            config.pollintervall
        };

        Self {
            plan,
            runner,
            config: Config {
                pollintervall,
                namespace: config.namespace,
                max_concurrent_jobs: max_concurrent_jobs(config.max_concurrent_jobs),
            },
        }
    }

    /// Kör controllern tills den avbryts.
    ///
    /// Go: `Controller.Run`.
    pub async fn run(&self) {
        info!(
            "Job controller starting, polling every {:?}",
            self.config.pollintervall
        );

        if let Err(fel) = self.reconcile().await {
            warn!("Job controller reconcile failed: {fel}");
        }

        loop {
            tokio::time::sleep(self.config.pollintervall).await;

            if let Err(fel) = self.reconcile().await {
                warn!("Job controller reconcile failed: {fel}");
            }
        }
    }

    /// Gör en genomgång av körningarna och jobben.
    ///
    /// En körning som står i kö får ett Job. En körning som redan har ett
    /// Job lämnas ifred - bara ett utfall rapporteras tillbaka.
    ///
    /// Go: `Controller.Reconcile`.
    pub async fn reconcile(&self) -> Result<(), String> {
        let runs = self
            .plan
            .agent_runs(None)
            .await
            .map_err(|err| format!("failed to fetch agent runs: {err}"))?;

        for run in &runs {
            if let Err(fel) = self.reconcile_run(run).await {
                warn!("Failed to reconcile run {}: {fel}", run.id);
            }
        }

        Ok(())
    }

    /// Ser till att en köad körning får ett Job.
    ///
    /// Go: `Controller.reconcileRun`.
    async fn reconcile_run(&self, run: &AgentRun) -> Result<(), String> {
        let job_name = jobrunner::job_name(run.id);

        // En avbruten körning ska inte lämna resurser efter sig.
        if run.status == RUN_STATUS_CANCELLED {
            return self
                .clean_up_cancelled(run, &job_name)
                .await
                .map_err(|err| err.to_string());
        }

        let job = match self.runner.get(&job_name).await {
            Ok(job) => job,
            Err(KubeError::NotFound(_)) => {
                // Jobbet finns inte. Bara en köad körning får skapa ett nytt
                // jobb.
                if run.status == RUN_STATUS_QUEUED {
                    return self.skapa_jobb(run, &job_name).await;
                }

                // En körning som redan är running utan motsvarande Job kan
                // inte återupptas säkert: workspace och process är borta.
                // Markera den som failed i stället för att lämna den fast
                // för alltid eller skapa ett nytt jobb med okänd historik.
                // waiting_approval lämnas orörd eftersom jobbet kan vara
                // färdigt medan människan beslutar.
                if run.status == RUN_STATUS_RUNNING {
                    return self.report_missing_job(run, &job_name).await;
                }

                return Ok(());
            }
            Err(err) => {
                return Err(format!("failed to get job {job_name}: {err}"));
            }
        };

        // Jobbet finns - för över ett misslyckande till körningen i Plan.
        self.report_failure(run, &job).await
    }

    /// Skapar ett jobb för en köad körning.
    async fn skapa_jobb(&self, run: &AgentRun, job_name: &str) -> Result<(), String> {
        let aktiva = self
            .runner
            .active_job_count()
            .await
            .map_err(|err| err.to_string())?;

        if aktiva >= self.config.max_concurrent_jobs {
            info!(
                "Job capacity reached ({}) - run {} stays queued",
                self.config.max_concurrent_jobs, run.id
            );

            return Ok(());
        }

        // Ett mål får högst ett pågående jobb - två körningar mot samma
        // repo skulle skriva över varandra.
        let aktivt = self
            .runner
            .has_active_job_for_goal(run.goal_id, run.id)
            .await
            .map_err(|err| err.to_string())?;

        if aktivt {
            info!(
                "Goal #{} already has a running job - run {} stays queued",
                run.goal_id, run.id
            );

            return Ok(());
        }

        self.runner
            .create(&Spec {
                goal_id: run.goal_id,
                run_id: run.id,
                name: job_name.to_string(),
                profile: run.agent_profile.clone(),
            })
            .await
            .map(|_| ())
            .map_err(|err| err.to_string())
    }

    /// Tar bort jobbet och podden för en avbruten körning.
    ///
    /// Avbrytning ska lämna klustret som det var - annars fortsätter en
    /// agent att köra kod mot repot efter att en människa sagt stopp.
    ///
    /// Go: `Controller.cleanUpCancelled`.
    async fn clean_up_cancelled(&self, run: &AgentRun, job_name: &str) -> Result<(), KubeError> {
        match self.runner.get(job_name).await {
            Ok(_) => {}
            // Jobbet finns redan inte - inget att städa.
            Err(KubeError::NotFound(_)) => return Ok(()),
            Err(err) => return Err(err),
        }

        self.runner.delete(job_name).await.map_err(|err| {
            KubeError::Annat(format!("failed to delete cancelled job {job_name}: {err}"))
        })?;

        info!("Deleted job {job_name} - run {} was cancelled", run.id);

        Ok(())
    }

    /// Markerar en running-körning som misslyckad när dess Job är borta.
    ///
    /// Det är säkrare än att återskapa jobbet, eftersom den gamla
    /// workspacen och processhistoriken inte längre går att identifiera.
    ///
    /// Go: `Controller.reportMissingJob`.
    async fn report_missing_job(&self, run: &AgentRun, job_name: &str) -> Result<(), String> {
        let status = RUN_STATUS_FAILED;
        let summary = format!("Jobbet {job_name} försvann medan körningen pågick");
        let result =
            "AgentRun markerades som failed eftersom dess Kubernetes Job inte längre finns";
        let finished_at = chrono::Utc::now();

        self.plan
            .update_agent_run(
                run.id,
                UpdateAgentRunRequest {
                    status: Some(status.to_string()),
                    summary: Some(summary),
                    result: Some(result.to_string()),
                    finished_at: Some(finished_at),
                    ..Default::default()
                },
            )
            .await
            .map_err(|err| format!("failed to mark orphaned run {} as failed: {err}", run.id))?;

        if let Err(err) = self
            .plan
            .create_agent_run_event(
                run.id,
                CreateAgentRunEventRequest {
                    r#type: EVENT_RUN_FAILED.to_string(),
                    phase: None,
                    payload: Some(serde_json::json!({
                        "reason": "job_missing",
                        "job": job_name,
                        "goal_id": run.goal_id,
                        "message": "running AgentRun saknar sitt Kubernetes Job",
                    })),
                },
            )
            .await
        {
            warn!(
                "Failed to record missing-job event for run {}: {err}",
                run.id
            );
        }

        Ok(())
    }

    /// Markerar körningen som misslyckad om jobbet föll eller överskred sin
    /// aktiva deadline.
    ///
    /// Go: `Controller.reportFailure`.
    async fn report_failure(&self, run: &AgentRun, job: &Job) -> Result<(), String> {
        if !job_failed(job) {
            return Ok(());
        }

        if run.status == RUN_STATUS_FAILED || run.status == RUN_STATUS_SUCCEEDED {
            return Ok(());
        }

        let job_name = job.metadata.name.clone().unwrap_or_default();
        let mut summary = format!("Jobbet {job_name} misslyckades");
        let mut result = format!("Jobbet {job_name} misslyckades utan att ge ett resultat");

        if job_deadline_exceeded(job) {
            summary = format!("Jobbet {job_name} överskred sin tidsgräns");
            result =
                format!("Jobbet {job_name} stoppades av Kubernetes efter att deadline passerat");
        }

        let finished_at = chrono::Utc::now();

        self.plan
            .update_agent_run(
                run.id,
                UpdateAgentRunRequest {
                    status: Some(RUN_STATUS_FAILED.to_string()),
                    summary: Some(summary),
                    result: Some(result),
                    finished_at: Some(finished_at),
                    ..Default::default()
                },
            )
            .await
            .map_err(|err| format!("failed to mark run {} as failed: {err}", run.id))?;

        if let Err(err) = self
            .plan
            .create_agent_run_event(
                run.id,
                CreateAgentRunEventRequest {
                    r#type: EVENT_RUN_FAILED.to_string(),
                    phase: None,
                    payload: Some(serde_json::json!({
                        "error": "the kubernetes job failed",
                        "job": job_name,
                        "goal_id": run.goal_id,
                    })),
                },
            )
            .await
        {
            warn!("Failed to record run_failed for run {}: {err}", run.id);
        }

        info!(
            "Job {} failed - run {} for goal #{} marked as failed",
            job_name, run.id, run.goal_id
        );

        Ok(())
    }
}

/// Begränsar antalet samtidiga jobb.
///
/// Go: `maxConcurrentJobs`.
fn max_concurrent_jobs(value: usize) -> usize {
    if value == 0 {
        return 4;
    }

    value
}

/// Har jobbet fallit?
///
/// Go: `jobFailed`.
fn job_failed(job: &Job) -> bool {
    job.status.failed() > 0 || job_deadline_exceeded(job)
}

/// Har jobbet överskridit sin deadline?
///
/// Go: `jobDeadlineExceeded`.
fn job_deadline_exceeded(job: &Job) -> bool {
    job.status
        .conditions
        .iter()
        .any(|villkor| villkor.condition_type == "DeadlineExceeded" && villkor.status == "True")
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn kapaciteten_falls_tillbaka_till_fyra() {
        assert_eq!(max_concurrent_jobs(0), 4);
        assert_eq!(max_concurrent_jobs(2), 2);
    }

    #[test]
    fn misslyckade_jobb_känns_igen() {
        let mut job = Job {
            api_version: None,
            kind: None,
            metadata: Default::default(),
            spec: Default::default(),
            status: Default::default(),
        };

        assert!(!job_failed(&job));

        job.status.failed = Some(1);

        assert!(job_failed(&job));

        job.status.failed = Some(0);
        job.status.conditions.push(crate::k8s::types::JobCondition {
            condition_type: "DeadlineExceeded".to_string(),
            status: "True".to_string(),
            reason: None,
        });

        assert!(job_deadline_exceeded(&job));
        assert!(job_failed(&job));
    }
}
