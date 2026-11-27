//! Bygger Kubernetes Jobs för agentkörningar.
//!
//! Go: `internal/jobrunner`. En agentkörning ska inte dela processutrymme
//! med API-servern. Ett Job ger isolering: egen pod, egna
//! resursgränser, egen timeout och en ServiceAccount utan rättigheter.
//! Jobbet kan avslutas utan att resten av systemet märker något.

use std::collections::BTreeMap;

use crate::k8s::types::{
    Capabilities, Container, ContainerSecurityContext, EmptyDirVolumeSource, EnvVar, EnvVarSource,
    Job, JobSpec, ObjectMeta, PodSecurityContext, PodTemplateSpec, Resources, SeccompProfile,
    SecretKeySelector, Volume, VolumeMount,
};
use crate::k8s::{Client, KubeError};
use crate::planclient::models::{RUN_STATUS_FAILED, RUN_STATUS_RUNNING, RUN_STATUS_SUCCEEDED};
use crate::profile::{self, Policy};

/// Hur jobb ska byggas.
#[derive(Debug, Clone)]
pub struct Config {
    /// Namespace jobbet skapas i.
    pub namespace: String,
    /// Image som kör jobbet.
    pub image: String,
    /// ServiceAccount jobbet använder.
    ///
    /// Den ska ha så få rättigheter som möjligt - jobbet behöver inte
    /// prata med Kubernetes-API:t alls.
    pub service_account: String,
    /// Repot som klonas innan agenten börjar.
    pub repo_url: String,
    /// Branchen som klonas.
    pub repo_branch: String,
    /// Adressen till Plan.
    pub plan_api_url: String,
    /// Hemligheten med Plan-nyckeln.
    pub api_key_secret_name: String,
    /// Nyckeln i hemligheten.
    pub api_key_secret_key: String,
    /// Hemligheten med modellnyckeln.
    pub llm_secret_name: String,
    /// Nyckeln i hemligheten.
    pub llm_secret_key: String,
    /// Hemligheten med git-autentiseringen.
    ///
    /// Tomt innebär att agenten bara kan läsa repot - den kan inte skicka
    /// brancher tillbaka.
    pub git_secret_name: String,
    /// Nyckeln i hemligheten.
    pub git_secret_key: String,
    /// CPU-förfrågan.
    pub cpu_request: String,
    /// Minnesförfrågan.
    pub memory_request: String,
    /// CPU-gränsen.
    pub cpu_limit: String,
    /// Minnesgränsen.
    pub memory_limit: String,
    /// Stoppar jobbet om det kör för länge.
    pub active_deadline_seconds: i64,
    /// Hur många gånger jobbet får starta om.
    pub backoff_limit: i32,
    /// Städar jobbet efter att det är klart.
    pub ttl_seconds_after_finished: i32,
    /// Hur imagen hämtas.
    pub image_pull_policy: String,
}

impl Default for Config {
    /// Ger en config med rimliga värden för en agentkörning.
    ///
    /// Image lämnas tom - då läses agentens egen image ur podens spec, så
    /// att jobben alltid kör den kod som är utbyggd.
    ///
    /// Go: `DefaultConfig`.
    fn default() -> Self {
        Self {
            namespace: "plan".to_string(),
            image: String::new(),
            service_account: "plan-agent-worker".to_string(),
            repo_url: String::new(),
            repo_branch: String::new(),
            plan_api_url: String::new(),
            api_key_secret_name: String::new(),
            api_key_secret_key: String::new(),
            llm_secret_name: String::new(),
            llm_secret_key: String::new(),
            git_secret_name: String::new(),
            git_secret_key: String::new(),
            cpu_request: "100m".to_string(),
            memory_request: "256Mi".to_string(),
            cpu_limit: "1000m".to_string(),
            memory_limit: "2Gi".to_string(),
            active_deadline_seconds: 3600,
            backoff_limit: 0,
            ttl_seconds_after_finished: 3600,
            image_pull_policy: "IfNotPresent".to_string(),
        }
    }
}

/// Vad ett jobb ska göra.
#[derive(Debug, Clone)]
pub struct Spec {
    /// Målet som ska utföras.
    pub goal_id: i64,
    /// Körningen som jobbet hör till.
    pub run_id: i64,
    /// Jobbets namn. Det måste vara unikt i namespace.
    pub name: String,
    /// Agentprofilen - coding eller life.
    ///
    /// Profilen bestämmer om jobbet får klona repot och vilka
    /// resursgränser som gäller.
    pub profile: String,
}

/// Skapar och läser Kubernetes Jobs.
#[derive(Debug, Clone)]
pub struct Runner {
    /// Klienten mot klustret.
    client: Client,
    /// Jobbkonfigurationen.
    config: Config,
}

impl Runner {
    /// Skapar en runner mot den givna klienten.
    ///
    /// Go: `jobrunner.New`.
    pub fn ny(client: Client, config: Config) -> Self {
        Self { client, config }
    }

    /// Skapar jobbet för en körning.
    ///
    /// Go: `Runner.Create`.
    pub async fn create(&self, spec: &Spec) -> Result<Job, KubeError> {
        let image = self.resolve_image().await?;
        let job = self.build(spec, &image)?;

        let skapad: Job = self
            .client
            .post(
                &format!("/apis/batch/v1/namespaces/{}/jobs", self.config.namespace),
                &job,
            )
            .await
            .map_err(|err| {
                KubeError::Annat(format!("failed to create job {}: {err}", spec.name))
            })?;

        tracing::info!(
            "Created agent job {} for goal #{} (run #{}) with image {}",
            skapad.metadata.name.clone().unwrap_or_default(),
            spec.goal_id,
            spec.run_id,
            image
        );

        Ok(skapad)
    }

    /// Bestämmer vilken image jobbet ska köra.
    ///
    /// En angiven image har företräde. Annars läses agentens egen image ur
    /// podens spec - jobbet ska alltid köra exakt den kod som är utbyggd,
    /// aldrig en äldre tagg.
    ///
    /// Go: `Runner.resolveImage`.
    async fn resolve_image(&self) -> Result<String, KubeError> {
        if !self.config.image.is_empty() {
            return Ok(self.config.image.clone());
        }

        let pod_name = std::env::var("POD_NAME").unwrap_or_default();

        if pod_name.is_empty() {
            return Err(KubeError::Annat(
                "no image configured and POD_NAME is not set".to_string(),
            ));
        }

        let pod: crate::k8s::types::Pod = self
            .client
            .get(&format!(
                "/api/v1/namespaces/{}/pods/{}",
                self.config.namespace, pod_name
            ))
            .await
            .map_err(|err| KubeError::Annat(format!("failed to read own pod {pod_name}: {err}")))?;

        if pod.spec.containers.is_empty() {
            return Err(KubeError::Annat(format!(
                "pod {pod_name} has no containers"
            )));
        }

        Ok(pod.spec.containers[0].image.clone().unwrap_or_default())
    }

    /// Hämtar ett jobb.
    ///
    /// Go: `Runner.Get`.
    pub async fn get(&self, name: &str) -> Result<Job, KubeError> {
        self.client
            .get(&format!(
                "/apis/batch/v1/namespaces/{}/jobs/{}",
                self.config.namespace, name
            ))
            .await
    }

    /// Svarar på om ett mål redan har ett pågående jobb.
    ///
    /// Ett mål ska ha högst ett jobb i taget - annars kan två
    /// agentkörningar skriva i samma repo och förstöra varandras ändringar.
    ///
    /// Go: `Runner.HasActiveJobForGoal`.
    pub async fn has_active_job_for_goal(
        &self,
        goal_id: i64,
        uteslut_körning_id: i64,
    ) -> Result<bool, KubeError> {
        let jobb = self
            .list_jobs(Some(&format!("plan.dev/goal-id={goal_id}")))
            .await?;
        let målets_id = goal_id.to_string();
        let uteslut = uteslut_körning_id.to_string();

        for job in &jobb {
            let etiketter = job.metadata.labels.clone().unwrap_or_default();

            if etiketter.get("plan.dev/goal-id").map(String::as_str) != Some(målets_id.as_str()) {
                continue;
            }

            if etiketter.get("plan.dev/agent-run-id").map(String::as_str) == Some(uteslut.as_str())
            {
                continue;
            }

            if job.status.succeeded() > 0 || job.status.failed() > 0 {
                continue;
            }

            return Ok(true);
        }

        Ok(false)
    }

    /// Returnerar antalet jobb som fortfarande kör.
    ///
    /// Avslutade jobb räknas inte, även om TTL-städningen ännu inte
    /// hunnit ske.
    ///
    /// Go: `Runner.ActiveJobCount`.
    pub async fn active_job_count(&self) -> Result<usize, KubeError> {
        let jobb = self.list_jobs(Some("app=agent-worker")).await?;
        let aktiva = jobb
            .iter()
            .filter(|job| job.status.succeeded() == 0 && job.status.failed() == 0)
            .count();

        Ok(aktiva)
    }

    /// Listar jobb i namespace.
    async fn list_jobs(&self, etikettfilter: Option<&str>) -> Result<Vec<Job>, KubeError> {
        #[derive(serde::Deserialize)]
        struct JobLista {
            #[serde(default)]
            items: Vec<Job>,
        }

        let lista: JobLista = self
            .client
            .list(
                &format!("/apis/batch/v1/namespaces/{}/jobs", self.config.namespace),
                etikettfilter,
            )
            .await
            .map_err(|err| KubeError::Annat(format!("failed to list active agent jobs: {err}")))?;

        Ok(lista.items)
    }

    /// Tar bort ett jobb och dess pods.
    ///
    /// Go: `Runner.Delete`.
    pub async fn delete(&self, name: &str) -> Result<(), KubeError> {
        self.client
            .delete(&format!(
                "/apis/batch/v1/namespaces/{}/jobs/{}",
                self.config.namespace, name
            ))
            .await
            .map_err(|err| KubeError::Annat(format!("failed to delete job {name}: {err}")))
    }

    /// Bygger jobbets manifest.
    ///
    /// Jobbet har tre säkerhetsgränser:
    ///
    /// - ingen ServiceAccount-token monteras, jobbet kan inte röra API-servern
    /// - resursgränser sätter en tak på CPU och minne
    /// - en deadline stoppar jobbet om det fastnar
    ///
    /// Repot klonas av en init-container, så agenten arbetar på en kopia
    /// som bara finns under jobbets livstid.
    ///
    /// Go: `Runner.Build`.
    pub fn build(&self, spec: &Spec, image: &str) -> Result<Job, KubeError> {
        if spec.name.is_empty() {
            return Err(KubeError::Annat("job name is required".to_string()));
        }

        if image.is_empty() {
            return Err(KubeError::Annat("job image is required".to_string()));
        }

        if spec.goal_id == 0 {
            return Err(KubeError::Annat("goal id is required".to_string()));
        }

        let policy = profile::policy_for(&spec.profile).map_err(KubeError::Annat)?;

        let cpu_request = normalisera_mängd(&värde_eller(&self.config.cpu_request, "100m"))
            .map_err(|err| KubeError::Annat(format!("failed to parse cpu request: {err}")))?;
        let memory_request = normalisera_mängd(&värde_eller(&self.config.memory_request, "256Mi"))
            .map_err(|err| KubeError::Annat(format!("failed to parse memory request: {err}")))?;

        let cpu_limit = normalisera_mängd(&värde_eller(
            &policy.cpu_limit,
            &värde_eller(&self.config.cpu_limit, "1000m"),
        ))
        .map_err(|err| KubeError::Annat(format!("failed to parse cpu limit: {err}")))?;
        let memory_limit = normalisera_mängd(&värde_eller(
            &policy.memory_limit,
            &värde_eller(&self.config.memory_limit, "2Gi"),
        ))
        .map_err(|err| KubeError::Annat(format!("failed to parse memory limit: {err}")))?;

        let mut etiketter = BTreeMap::new();
        etiketter.insert("app".to_string(), "agent-worker".to_string());
        etiketter.insert("plan.dev/goal-id".to_string(), spec.goal_id.to_string());
        etiketter.insert("plan.dev/agent-run-id".to_string(), spec.run_id.to_string());

        let resurser = Resources {
            requests: Some(resurskarta(&cpu_request, &memory_request)),
            limits: Some(resurskarta(&cpu_limit, &memory_limit)),
        };

        let volym = Volume {
            name: Some("workspace".to_string()),
            empty_dir: Some(EmptyDirVolumeSource {}),
        };
        let tmp = Volume {
            name: Some("tmp".to_string()),
            empty_dir: Some(EmptyDirVolumeSource {}),
        };

        let monteringar = vec![
            VolumeMount {
                name: Some("workspace".to_string()),
                mount_path: Some("/workspace".to_string()),
            },
            VolumeMount {
                name: Some("tmp".to_string()),
                mount_path: Some("/tmp".to_string()),
            },
        ];

        let env = self.env(spec, &policy)?;

        let job = Job {
            api_version: None,
            kind: None,
            metadata: ObjectMeta {
                name: Some(spec.name.clone()),
                namespace: Some(self.config.namespace.clone()),
                labels: Some(etiketter.clone()),
            },
            spec: JobSpec {
                backoff_limit: Some(self.config.backoff_limit),
                active_deadline_seconds: Some(self.config.active_deadline_seconds),
                ttl_seconds_after_finished: Some(self.config.ttl_seconds_after_finished),
                template: PodTemplateSpec {
                    metadata: ObjectMeta {
                        name: None,
                        namespace: None,
                        labels: Some(etiketter),
                    },
                    spec: crate::k8s::types::PodSpec {
                        restart_policy: Some("Never".to_string()),
                        service_account_name: Some(self.config.service_account.clone()),
                        // Jobbet ska inte kunna prata med Kubernetes-API:t.
                        automount_service_account_token: Some(false),
                        security_context: Some(PodSecurityContext {
                            run_as_non_root: Some(true),
                            run_as_user: Some(65532),
                            run_as_group: Some(65532),
                            fs_group: Some(65532),
                            seccomp_profile: Some(SeccompProfile {
                                profile_type: "RuntimeDefault".to_string(),
                            }),
                        }),
                        init_containers: vec![Container {
                            name: Some("clone".to_string()),
                            image: Some(image.to_string()),
                            image_pull_policy: Some(self.config.image_pull_policy.clone()),
                            command: Some(vec![
                                "sh".to_string(),
                                "-c".to_string(),
                                clone_script(&self.repo_url(&policy), &self.config.repo_branch),
                            ]),
                            args: None,
                            env: Vec::new(),
                            volume_mounts: monteringar.clone(),
                            resources: resurser.clone(),
                            security_context: Some(container_security_context()),
                        }],
                        containers: vec![Container {
                            name: Some("agent-worker".to_string()),
                            image: Some(image.to_string()),
                            image_pull_policy: Some(self.config.image_pull_policy.clone()),
                            command: Some(vec!["/app/agent-worker".to_string()]),
                            args: None,
                            env,
                            volume_mounts: monteringar,
                            resources: resurser,
                            security_context: Some(container_security_context()),
                        }],
                        volumes: vec![volym, tmp],
                        image_pull_secrets: Vec::new(),
                    },
                },
            },
            status: Default::default(),
        };

        Ok(job)
    }

    /// Bygger miljövariablerna för jobbet.
    ///
    /// Hemligheter kommer från Kubernetes Secrets - de skrivs aldrig i
    /// manifestet och syns inte i loggar.
    ///
    /// Go: `Runner.env`.
    fn env(&self, spec: &Spec, policy: &Policy) -> Result<Vec<EnvVar>, KubeError> {
        // En profil som inte får röra kod får inte ens en sökväg till repot.
        let repo_path = if policy.needs_repo {
            "/workspace/repo"
        } else {
            ""
        };

        let mut env = vec![
            EnvVar {
                name: Some("AGENT_GOAL_ID".to_string()),
                value: spec.goal_id.to_string(),
                value_from: None,
            },
            EnvVar {
                name: Some("AGENT_REPO_PATH".to_string()),
                value: repo_path.to_string(),
                value_from: None,
            },
            EnvVar {
                name: Some("AGENT_PROFILE".to_string()),
                value: policy.name.clone(),
                value_from: None,
            },
            EnvVar {
                name: Some("AGENT_INITIATOR".to_string()),
                value: format!("job:{}", spec.name),
                value_from: None,
            },
            EnvVar {
                name: Some("AGENT_POLL_INTERVAL".to_string()),
                value: "15s".to_string(),
                value_from: None,
            },
            EnvVar {
                name: Some("AGENT_STOP_WHEN_FINISHED".to_string()),
                value: "true".to_string(),
                value_from: None,
            },
            EnvVar {
                name: Some("HOME".to_string()),
                value: "/tmp".to_string(),
                value_from: None,
            },
        ];

        if !self.config.plan_api_url.is_empty() {
            env.push(EnvVar {
                name: Some("PLAN_API_URL".to_string()),
                value: self.config.plan_api_url.clone(),
                value_from: None,
            });
        }

        if !self.config.api_key_secret_name.is_empty() {
            env.push(EnvVar {
                name: Some("PLAN_API_KEY".to_string()),
                value: String::new(),
                value_from: Some(EnvVarSource {
                    secret_key_ref: Some(SecretKeySelector {
                        name: Some(self.config.api_key_secret_name.clone()),
                        key: Some(if self.config.api_key_secret_key.trim().is_empty() {
                            "PLAN_API_KEY".to_string()
                        } else {
                            self.config.api_key_secret_key.clone()
                        }),
                    }),
                }),
            });
        }

        if !self.config.git_secret_name.is_empty() && policy.may_change_code {
            env.push(EnvVar {
                name: Some("GIT_AUTH_HEADER".to_string()),
                value: String::new(),
                value_from: Some(EnvVarSource {
                    secret_key_ref: Some(SecretKeySelector {
                        name: Some(self.config.git_secret_name.clone()),
                        key: Some(if self.config.git_secret_key.trim().is_empty() {
                            "GIT_AUTH_HEADER".to_string()
                        } else {
                            self.config.git_secret_key.clone()
                        }),
                    }),
                }),
            });
        }

        if !self.config.llm_secret_name.is_empty() {
            env.push(EnvVar {
                name: Some("MINIMAX_API_KEY".to_string()),
                value: String::new(),
                value_from: Some(EnvVarSource {
                    secret_key_ref: Some(SecretKeySelector {
                        name: Some(self.config.llm_secret_name.clone()),
                        key: Some(if self.config.llm_secret_key.trim().is_empty() {
                            "MINIMAX_API_KEY".to_string()
                        } else {
                            self.config.llm_secret_key.clone()
                        }),
                    }),
                }),
            });
        }

        Ok(env)
    }

    /// Ger repot som profilen får klona.
    ///
    /// En profil som inte får röra kod får inte ens se repot.
    ///
    /// Go: `Runner.repoURL`.
    fn repo_url(&self, policy: &Policy) -> String {
        if !policy.needs_repo {
            return String::new();
        }

        self.config.repo_url.clone()
    }
}

/// Ger ett värde, eller en ersättning om värdet är tomt.
///
/// Go: `valueOr`.
fn värde_eller(värde: &str, ersättning: &str) -> String {
    if värde.trim().is_empty() {
        return ersättning.to_string();
    }

    värde.trim().to_string()
}

/// Bygger resurskartan för en container.
fn resurskarta(cpu: &str, minne: &str) -> BTreeMap<String, String> {
    let mut resurser = BTreeMap::new();

    resurser.insert("cpu".to_string(), cpu.to_string());
    resurser.insert("memory".to_string(), minne.to_string());

    resurser
}

/// Gör om en resursmängd till sitt kanoniska format.
///
/// Kubernetes läser "1000m" och "1" som samma sak, men Go skriver ut det
/// kanoniska formatet - och det hamnar i manifestet som jämförs mellan
/// Go och Rust.
///
/// Go: `resource.ParseQuantity`.
fn normalisera_mängd(text: &str) -> Result<String, String> {
    let text = text.trim();

    if text.is_empty() {
        return Err("quantities must match the regular expression".to_string());
    }

    // Binära suffix (Ki, Mi, Gi, Ti) behålls som de är.
    for suffix in ["Ki", "Mi", "Gi", "Ti"] {
        if let Some(siffror) = text.strip_suffix(suffix) {
            if siffror.parse::<f64>().is_ok() {
                return Ok(text.to_string());
            }
        }
    }

    let (siffror, suffix) =
        match text.find(|tecken: char| !tecken.is_ascii_digit() && tecken != '.') {
            Some(index) => (&text[..index], &text[index..]),
            None => (text, ""),
        };

    let värde: f64 = siffror
        .parse()
        .map_err(|_| "quantities must match the regular expression".to_string())?;

    let basvärde = match suffix {
        "" => värde,
        "m" => värde / 1_000.0,
        "k" | "K" => värde * 1_000.0,
        "M" => värde * 1_000_000.0,
        "G" => värde * 1_000_000_000.0,
        "T" => värde * 1_000_000_000_000.0,
        _ => return Err("quantities must match the regular expression".to_string()),
    };

    // Hela tal skrivs utan suffix - så blir "1000m" till "1".
    if basvärde.fract() == 0.0 {
        if basvärde % 1_000_000_000.0 == 0.0 && basvärde.abs() >= 1_000_000_000.0 {
            return Ok(format!("{}G", basvärde / 1_000_000_000.0));
        }

        if basvärde % 1_000_000.0 == 0.0 && basvärde.abs() >= 1_000_000.0 {
            return Ok(format!("{}M", basvärde / 1_000_000.0));
        }

        if basvärde % 1_000.0 == 0.0 && basvärde.abs() >= 1_000.0 {
            return Ok(format!("{}k", basvärde / 1_000.0));
        }

        return Ok(format!("{}", basvärde as i64));
    }

    // Bråktal skrivs i milli - så blir "0.5" till "500m".
    let milli = basvärde * 1_000.0;

    if milli.fract() == 0.0 {
        return Ok(format!("{}m", milli as i64));
    }

    Ok(text.to_string())
}

/// Ger containern minimala rättigheter.
///
/// Go: `containerSecurityContext`.
fn container_security_context() -> ContainerSecurityContext {
    ContainerSecurityContext {
        allow_privilege_escalation: Some(false),
        read_only_root_filesystem: Some(true),
        run_as_non_root: Some(true),
        run_as_user: Some(65532),
        capabilities: Some(Capabilities {
            drop: vec!["ALL".to_string()],
        }),
        seccomp_profile: Some(SeccompProfile {
            profile_type: "RuntimeDefault".to_string(),
        }),
    }
}

/// Bygger skriptet som klonar repot.
///
/// Go: `cloneScript`.
fn clone_script(url: &str, branch: &str) -> String {
    if url.is_empty() {
        return "mkdir -p /workspace/repo".to_string();
    }

    let mut script = String::from("set -eu\n");
    script.push_str("git clone --depth 1");

    if !branch.is_empty() {
        script.push_str(" --branch ");
        script.push_str(branch);
    }

    script.push(' ');
    script.push_str(url);
    script.push_str(" /workspace/repo\n");

    script
}

/// Ger ett jobbnamn som är unikt för en körning.
///
/// Go: `JobName`.
pub fn job_name(run_id: i64) -> String {
    format!("agent-worker-run-{run_id}")
}

/// Ger ett jobbnamn för ett mål.
///
/// Namnet används för att hitta samma jobb igen - controllern ska inte
/// starta två jobb för samma mål.
///
/// Go: `JobNameForGoal`.
pub fn job_name_for_goal(goal_id: i64) -> String {
    format!("agent-worker-goal-{goal_id}")
}

/// Översätter ett jobbs tillstånd till en agentkörningsstatus.
///
/// Översättningen följer agentkörningarnas egna statusar, så att en
/// körning som startats i ett Job ser samma ut i Plan som en körning
/// som startats av schemaläggaren.
///
/// Go: `JobStatus`.
pub fn job_status(job: Option<&Job>) -> String {
    let Some(job) = job else {
        return "queued".to_string();
    };

    if job.status.succeeded() > 0 {
        return RUN_STATUS_SUCCEEDED.to_string();
    }

    if job.status.failed() > 0 {
        return RUN_STATUS_FAILED.to_string();
    }

    if job.status.active() > 0 {
        return RUN_STATUS_RUNNING.to_string();
    }

    "queued".to_string()
}

#[cfg(test)]
mod tests {
    use super::*;

    fn spec() -> Spec {
        Spec {
            goal_id: 42,
            run_id: 7,
            name: job_name(7),
            profile: profile::CODING.to_string(),
        }
    }

    #[test]
    fn manifestet_innehaller_säkerhetsgränserna() {
        let runner = Runner::ny(
            Client::ny("https://127.0.0.1:6443", None).expect("klient"),
            Config::default(),
        );

        let job = runner.build(&spec(), "plan/agent:test").expect("manifest");

        assert_eq!(job.api_version, None);
        assert_eq!(job.metadata.name.as_deref(), Some("agent-worker-run-7"));

        let pod = &job.spec.template.spec;

        assert_eq!(pod.restart_policy.as_deref(), Some("Never"));
        assert_eq!(pod.automount_service_account_token, Some(false));
        assert_eq!(
            pod.service_account_name.as_deref(),
            Some("plan-agent-worker")
        );

        let säkerhet = pod.security_context.as_ref().expect("säkerhetskontext");

        assert_eq!(säkerhet.run_as_non_root, Some(true));
        assert_eq!(säkerhet.run_as_user, Some(65532));
        assert_eq!(säkerhet.fs_group, Some(65532));

        // Init-containern klonar repot, agenten kör i work space.
        assert_eq!(pod.init_containers.len(), 1);
        assert_eq!(pod.init_containers[0].name.as_deref(), Some("clone"));
        assert_eq!(pod.containers.len(), 1);
        assert_eq!(pod.containers[0].name.as_deref(), Some("agent-worker"));

        // Hemligheter skrivs inte i manifestet.
        assert!(pod.containers[0]
            .env
            .iter()
            .all(|variabel| variabel.value_from.is_none()));
    }

    #[test]
    fn livprofilen_får_inte_röra_kod() {
        let runner = Runner::ny(
            Client::ny("https://127.0.0.1:6443", None).expect("klient"),
            Config::default(),
        );

        let mut spec = spec();

        spec.profile = profile::LIFE.to_string();

        let job = runner.build(&spec, "plan/agent:test").expect("manifest");
        let env = &job.spec.template.spec.containers[0].env;

        let repo_path = env
            .iter()
            .find(|variabel| variabel.name.as_deref() == Some("AGENT_REPO_PATH"))
            .map(|variabel| variabel.value.clone())
            .unwrap_or_default();

        assert_eq!(repo_path, "");

        // Life-profilen har lägre gränser än coding.
        let gränser = job.spec.template.spec.containers[0]
            .resources
            .limits
            .clone()
            .expect("gränser");

        assert_eq!(gränser.get("cpu").map(String::as_str), Some("500m"));
        assert_eq!(gränser.get("memory").map(String::as_str), Some("1Gi"));
    }

    #[test]
    fn klonskriptet_följer_branchen() {
        assert_eq!(
            clone_script("https://github.com/simon/plan.git", "main"),
            "set -eu\ngit clone --depth 1 --branch main https://github.com/simon/plan.git /workspace/repo\n"
        );
        assert_eq!(clone_script("", ""), "mkdir -p /workspace/repo");
    }

    #[test]
    fn statusen_översätts() {
        let mut job = Job {
            api_version: None,
            kind: None,
            metadata: ObjectMeta::default(),
            spec: JobSpec::default(),
            status: Default::default(),
        };

        assert_eq!(job_status(None), "queued");
        assert_eq!(job_status(Some(&job)), "queued");

        job.status.active = Some(1);

        assert_eq!(job_status(Some(&job)), "running");

        job.status.succeeded = Some(1);

        assert_eq!(job_status(Some(&job)), "succeeded");

        job.status.succeeded = Some(0);
        job.status.failed = Some(1);

        assert_eq!(job_status(Some(&job)), "failed");
    }
}
