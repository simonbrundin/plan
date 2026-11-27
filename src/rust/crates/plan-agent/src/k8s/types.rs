//! Typer för Kubernetes-manifest.
//!
//! Go använder `k8s.io/api` - här skrivs de typerna som manifestet
//! faktiskt innehåller. Manifestet ska vara likvärdigt med Go:s, så
//! fältnamnen följer API:ets camelCase.

use std::collections::BTreeMap;

use serde::{Deserialize, Serialize};

/// Metadata för ett objekt.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct ObjectMeta {
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub name: Option<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub namespace: Option<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub labels: Option<BTreeMap<String, String>>,
}

/// Ett Kubernetes Job.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Job {
    #[serde(
        rename = "apiVersion",
        default,
        skip_serializing_if = "Option::is_none"
    )]
    pub api_version: Option<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub kind: Option<String>,
    #[serde(default)]
    pub metadata: ObjectMeta,
    #[serde(default)]
    pub spec: JobSpec,
    #[serde(default)]
    pub status: JobStatus,
}

/// Ett Kubernetes Pod.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Pod {
    #[serde(
        rename = "apiVersion",
        default,
        skip_serializing_if = "Option::is_none"
    )]
    pub api_version: Option<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub kind: Option<String>,
    #[serde(default)]
    pub metadata: ObjectMeta,
    #[serde(default)]
    pub spec: PodSpec,
}

/// Specen för ett Job.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct JobSpec {
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub backoff_limit: Option<i32>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub active_deadline_seconds: Option<i64>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub ttl_seconds_after_finished: Option<i32>,
    #[serde(default)]
    pub template: PodTemplateSpec,
}

/// Statusen för ett Job.
///
/// Fältena är valbara - Go skriver ut dem med `omitempty`, så ett jobb
/// utan status serialiseras som `{}`.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct JobStatus {
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub succeeded: Option<i32>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub failed: Option<i32>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub active: Option<i32>,
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub conditions: Vec<JobCondition>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub start_time: Option<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub completion_time: Option<String>,
}

impl JobStatus {
    /// Antalet lyckade försök.
    pub fn succeeded(&self) -> i32 {
        self.succeeded.unwrap_or(0)
    }

    /// Antalet misslyckade försök.
    pub fn failed(&self) -> i32 {
        self.failed.unwrap_or(0)
    }

    /// Antalet aktiva podar.
    pub fn active(&self) -> i32 {
        self.active.unwrap_or(0)
    }
}

/// Ett villkor på ett Job.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct JobCondition {
    #[serde(rename = "type")]
    pub condition_type: String,
    #[serde(default)]
    pub status: String,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub reason: Option<String>,
}

/// Mallen för poden i ett Job.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct PodTemplateSpec {
    #[serde(default)]
    pub metadata: ObjectMeta,
    #[serde(default)]
    pub spec: PodSpec,
}

/// Specen för en pod.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct PodSpec {
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub restart_policy: Option<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub service_account_name: Option<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub automount_service_account_token: Option<bool>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub security_context: Option<PodSecurityContext>,
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub init_containers: Vec<Container>,
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub containers: Vec<Container>,
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub volumes: Vec<Volume>,
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub image_pull_secrets: Vec<LocalObjectReference>,
}

/// En referens till ett lokalt objekt.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct LocalObjectReference {
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub name: Option<String>,
}

/// Säkerhetskontexten för poden.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct PodSecurityContext {
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub run_as_non_root: Option<bool>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub run_as_user: Option<i64>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub run_as_group: Option<i64>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub fs_group: Option<i64>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub seccomp_profile: Option<SeccompProfile>,
}

/// En seccomp-profil.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct SeccompProfile {
    #[serde(rename = "type")]
    pub profile_type: String,
}

/// Säkerhetskontexten för containern.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct ContainerSecurityContext {
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub allow_privilege_escalation: Option<bool>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub read_only_root_filesystem: Option<bool>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub run_as_non_root: Option<bool>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub run_as_user: Option<i64>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub capabilities: Option<Capabilities>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub seccomp_profile: Option<SeccompProfile>,
}

/// Kapaciteter som containern får behålla eller tappa.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct Capabilities {
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub drop: Vec<String>,
}

/// En container i poden.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Container {
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub name: Option<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub image: Option<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub image_pull_policy: Option<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub command: Option<Vec<String>>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub args: Option<Vec<String>>,
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub env: Vec<EnvVar>,
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub volume_mounts: Vec<VolumeMount>,
    #[serde(default)]
    pub resources: Resources,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub security_context: Option<ContainerSecurityContext>,
}

/// Resursgränserna för en container.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct Resources {
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub requests: Option<BTreeMap<String, String>>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub limits: Option<BTreeMap<String, String>>,
}

/// En miljövariabel.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct EnvVar {
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub name: Option<String>,
    /// Värdet skrivs inte ut när det är tomt - Go gör samma sak.
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub value: String,
    #[serde(default, rename = "valueFrom", skip_serializing_if = "Option::is_none")]
    pub value_from: Option<EnvVarSource>,
}

/// Var värdet till en miljövariabel kommer ifrån.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct EnvVarSource {
    #[serde(
        default,
        rename = "secretKeyRef",
        skip_serializing_if = "Option::is_none"
    )]
    pub secret_key_ref: Option<SecretKeySelector>,
}

/// En nyckel i en hemlighet.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct SecretKeySelector {
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub name: Option<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub key: Option<String>,
}

/// En volym i poden.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct Volume {
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub name: Option<String>,
    #[serde(default, rename = "emptyDir", skip_serializing_if = "Option::is_none")]
    pub empty_dir: Option<EmptyDirVolumeSource>,
}

/// En tom volym - den fylls när poden startar.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct EmptyDirVolumeSource {}

/// En montering av en volym.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct VolumeMount {
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub name: Option<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub mount_path: Option<String>,
}
