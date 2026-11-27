//! En liten klient mot Kubernetes-API:t.
//!
//! Go använder `client-go`. Här görs anropen direkt mot API:et med
//! `reqwest` - agenten behöver bara läsa och skriva jobb och poder, och
//! det är mindre beroenden att dra in.
//!
//! Konfigurationen läses som i Go: först in-cluster-konfigurationen, sedan
//! kubeconfig-filen.

use std::path::PathBuf;

use reqwest::header::{HeaderValue, AUTHORIZATION};
use serde::de::DeserializeOwned;
use serde::Deserialize;

pub mod types;

/// Var in-cluster-hemligheterna ligger.
const SERVICE_ACCOUNT_DIR: &str = "/var/run/secrets/kubernetes.io/serviceaccount";

/// Ett fel från Kubernetes.
#[derive(Debug, thiserror::Error)]
pub enum KubeError {
    /// Objektet finns inte.
    #[error("not found: {0}")]
    NotFound(String),
    /// API:t svarade med ett fel.
    #[error("{0}: {1}")]
    Status(u16, String),
    /// Något gick fel innan anropet kom fram.
    #[error("{0}")]
    Annat(String),
}

/// Klienten mot Kubernetes-API:t.
#[derive(Debug, Clone)]
pub struct Client {
    /// API-serverns adress.
    base_url: String,
    /// Bearer-token, om det finns.
    token: Option<String>,
    /// HTTP-klienten.
    http: reqwest::Client,
}

impl Client {
    /// Skapar en klient mot en given adress.
    pub fn ny(base_url: impl Into<String>, token: Option<String>) -> Result<Self, KubeError> {
        let http = reqwest::Client::builder()
            .build()
            .map_err(|err| KubeError::Annat(format!("kunde inte bygga HTTP-klienten: {err}")))?;

        Ok(Self {
            base_url: base_url.into().trim_end_matches('/').to_string(),
            token,
            http,
        })
    }

    /// Läser in-cluster-konfigurationen.
    ///
    /// Go: `rest.InClusterConfig`.
    pub fn in_cluster() -> Result<Self, KubeError> {
        let host = std::env::var("KUBERNETES_SERVICE_HOST").unwrap_or_default();
        let port = std::env::var("KUBERNETES_SERVICE_PORT").unwrap_or_default();

        if host.is_empty() || port.is_empty() {
            return Err(KubeError::Annat(
                "not running in a cluster (KUBERNETES_SERVICE_HOST is not set)".to_string(),
            ));
        }

        let token_path = PathBuf::from(SERVICE_ACCOUNT_DIR).join("token");
        let token = std::fs::read_to_string(&token_path)
            .map_err(|err| {
                KubeError::Annat(format!(
                    "failed to read service account token {}: {err}",
                    token_path.display()
                ))
            })?
            .trim()
            .to_string();

        let mut builder = reqwest::Client::builder();

        // Klustrets eget CA läggs in - annars går inte https mot API:t.
        let ca_path = PathBuf::from(SERVICE_ACCOUNT_DIR).join("ca.crt");

        if let Ok(ca) = std::fs::read(&ca_path) {
            if let Ok(cert) = reqwest::Certificate::from_pem(&ca) {
                builder = builder.add_root_certificate(cert);
            }
        }

        let http = builder
            .build()
            .map_err(|err| KubeError::Annat(format!("kunde inte bygga HTTP-klienten: {err}")))?;

        Ok(Self {
            base_url: format!("https://{host}:{port}"),
            token: Some(token),
            http,
        })
    }

    /// Läser konfigurationen från en kubeconfig-fil.
    ///
    /// Go: `clientcmd.BuildConfigFromFlags`.
    pub fn from_kubeconfig(sökväg: &std::path::Path) -> Result<Self, KubeError> {
        let innehåll = std::fs::read_to_string(sökväg).map_err(|err| {
            KubeError::Annat(format!(
                "failed to read kubeconfig {}: {err}",
                sökväg.display()
            ))
        })?;

        let config: Kubeconfig = serde_yaml::from_str(&innehåll).map_err(|err| {
            KubeError::Annat(format!(
                "failed to parse kubeconfig {}: {err}",
                sökväg.display()
            ))
        })?;

        let context = config
            .current_context
            .as_ref()
            .and_then(|namn| {
                config
                    .contexts
                    .iter()
                    .find(|ctx| ctx.name.as_deref() == Some(namn.as_str()))
            })
            .ok_or_else(|| {
                KubeError::Annat(format!(
                    "kubeconfig {} har ingen aktuell context",
                    sökväg.display()
                ))
            })?;

        let cluster = config
            .clusters
            .iter()
            .find(|kluster| kluster.name.as_deref() == context.context.cluster.as_deref())
            .ok_or_else(|| {
                KubeError::Annat("kubeconfigens context pekar på en okänd kluster".to_string())
            })?;

        let user = config
            .users
            .iter()
            .find(|användare| användare.name.as_deref() == context.context.user.as_deref());

        let mut builder = reqwest::Client::builder();

        if let Some(ca_data) = &cluster.cluster.certificate_authority_data {
            if let Ok(cert) = reqwest::Certificate::from_pem(ca_data.as_bytes()) {
                builder = builder.add_root_certificate(cert);
            }
        } else if let Some(ca_path) = &cluster.cluster.certificate_authority {
            if let Ok(ca) = std::fs::read(ca_path) {
                if let Ok(cert) = reqwest::Certificate::from_pem(&ca) {
                    builder = builder.add_root_certificate(cert);
                }
            }
        }

        let http = builder
            .build()
            .map_err(|err| KubeError::Annat(format!("kunde inte bygga HTTP-klienten: {err}")))?;

        let token = user.and_then(|användare| {
            användare.user.token.clone().or_else(|| {
                användare.user.token_file.as_ref().and_then(|sökväg| {
                    std::fs::read_to_string(sökväg)
                        .ok()
                        .map(|token| token.trim().to_string())
                })
            })
        });

        Ok(Self {
            base_url: cluster.cluster.server.trim_end_matches('/').to_string(),
            token,
            http,
        })
    }

    /// Skapar en klient från miljön.
    ///
    /// In-cluster först, sedan kubeconfig - precis som Go.
    pub fn from_env() -> Result<Self, KubeError> {
        match Self::in_cluster() {
            Ok(client) => Ok(client),
            Err(in_cluster_fel) => {
                let kubeconfig = match std::env::var("KUBECONFIG") {
                    Ok(sökväg) if !sökväg.is_empty() => PathBuf::from(sökväg),
                    _ => {
                        let hem = std::env::var("HOME").unwrap_or_else(|_| "/root".to_string());

                        PathBuf::from(hem).join(".kube").join("config")
                    }
                };

                Self::from_kubeconfig(&kubeconfig).map_err(|kubeconfig_fel| {
                    KubeError::Annat(format!(
                        "failed to create Kubernetes client: {in_cluster_fel}; {kubeconfig_fel}"
                    ))
                })
            }
        }
    }

    /// Läser ett objekt.
    pub async fn get<T: DeserializeOwned>(&self, sökväg: &str) -> Result<T, KubeError> {
        let svar = self.skicka(self.http.get(self.url(sökväg))).await?;

        svar.json::<T>().await.map_err(|err| {
            KubeError::Annat(format!("failed to decode response from {sökväg}: {err}"))
        })
    }

    /// Skapar ett objekt.
    pub async fn post<B: serde::Serialize, T: DeserializeOwned>(
        &self,
        sökväg: &str,
        kropp: &B,
    ) -> Result<T, KubeError> {
        let svar = self
            .skicka(self.http.post(self.url(sökväg)).json(kropp))
            .await?;

        svar.json::<T>().await.map_err(|err| {
            KubeError::Annat(format!("failed to decode response from {sökväg}: {err}"))
        })
    }

    /// Tar bort ett objekt.
    pub async fn delete(&self, sökväg: &str) -> Result<(), KubeError> {
        self.skicka(self.http.delete(self.url(sökväg))).await?;

        Ok(())
    }

    /// Listar objekt med en etikettfilter.
    pub async fn list<T: DeserializeOwned>(
        &self,
        sökväg: &str,
        etikettfilter: Option<&str>,
    ) -> Result<T, KubeError> {
        let mut req = self.http.get(self.url(sökväg));

        if let Some(filter) = etikettfilter {
            req = req.query(&[("labelSelector", filter)]);
        }

        let svar = self.skicka(req).await?;

        svar.json::<T>().await.map_err(|err| {
            KubeError::Annat(format!("failed to decode response from {sökväg}: {err}"))
        })
    }

    /// Bygger en fullständig adress.
    fn url(&self, sökväg: &str) -> String {
        format!("{}{}", self.base_url, sökväg)
    }

    /// Skickar anropet och hanterar API:ets svar.
    async fn skicka(&self, anrop: reqwest::RequestBuilder) -> Result<reqwest::Response, KubeError> {
        let mut anrop = anrop;

        if let Some(token) = &self.token {
            let värde = HeaderValue::from_str(&format!("Bearer {token}"))
                .map_err(|err| KubeError::Annat(format!("ogiltig token: {err}")))?;

            anrop = anrop.header(AUTHORIZATION, värde);
        }

        let svar = anrop
            .send()
            .await
            .map_err(|err| KubeError::Annat(format!("request failed: {err}")))?;

        let status = svar.status();

        if status.is_success() {
            return Ok(svar);
        }

        let kod = status.as_u16();
        let text = svar.text().await.unwrap_or_default();
        let meddelande = meddelande_ur_svar(&text);

        if kod == 404 {
            return Err(KubeError::NotFound(meddelande));
        }

        Err(KubeError::Status(kod, meddelande))
    }
}

/// Plockar ut meddelandet ur ett Status-svar.
fn meddelande_ur_svar(text: &str) -> String {
    #[derive(Deserialize)]
    struct Status {
        #[serde(default)]
        message: String,
    }

    match serde_json::from_str::<Status>(text) {
        Ok(status) if !status.message.is_empty() => status.message,
        _ => text.to_string(),
    }
}

/// En kubeconfig-fil.
#[derive(Debug, Default, Deserialize)]
struct Kubeconfig {
    #[serde(default, rename = "current-context")]
    current_context: Option<String>,
    #[serde(default)]
    clusters: Vec<NamedCluster>,
    #[serde(default)]
    contexts: Vec<NamedContext>,
    #[serde(default)]
    users: Vec<NamedUser>,
}

/// En kluster i kubeconfig.
#[derive(Debug, Default, Deserialize)]
struct NamedCluster {
    #[serde(default)]
    name: Option<String>,
    #[serde(default)]
    cluster: Cluster,
}

/// Klusteruppgifterna.
#[derive(Debug, Default, Deserialize)]
struct Cluster {
    #[serde(default)]
    server: String,
    #[serde(default, rename = "certificate-authority-data")]
    certificate_authority_data: Option<String>,
    #[serde(default, rename = "certificate-authority")]
    certificate_authority: Option<String>,
}

/// En context i kubeconfig.
#[derive(Debug, Default, Deserialize)]
struct NamedContext {
    #[serde(default)]
    name: Option<String>,
    #[serde(default)]
    context: Context,
}

/// Contextens uppgifter.
#[derive(Debug, Default, Deserialize)]
struct Context {
    #[serde(default)]
    cluster: Option<String>,
    #[serde(default)]
    user: Option<String>,
}

/// En användare i kubeconfig.
#[derive(Debug, Default, Deserialize)]
struct NamedUser {
    #[serde(default)]
    name: Option<String>,
    #[serde(default)]
    user: User,
}

/// Användarens uppgifter.
#[derive(Debug, Default, Deserialize)]
struct User {
    #[serde(default)]
    token: Option<String>,
    #[serde(default, rename = "tokenFile")]
    token_file: Option<String>,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn statusmeddelandet_plockas_ur_svaret() {
        assert_eq!(
            meddelande_ur_svar(r#"{"message":"jobs.batch \"agent\" not found"}"#),
            "jobs.batch \"agent\" not found"
        );
        assert_eq!(meddelande_ur_svar("ren text"), "ren text");
    }

    #[test]
    fn kubeconfig_läses() {
        let yaml = r#"
current-context: test
clusters:
- name: c1
  cluster:
    server: https://127.0.0.1:6443
contexts:
- name: test
  context:
    cluster: c1
    user: u1
users:
- name: u1
  user:
    token: abc
"#;

        let config: Kubeconfig = serde_yaml::from_str(yaml).expect("kubeconfig");

        assert_eq!(config.current_context.as_deref(), Some("test"));
        assert_eq!(config.clusters.len(), 1);
        assert_eq!(config.clusters[0].cluster.server, "https://127.0.0.1:6443");
        assert_eq!(config.users[0].user.token.as_deref(), Some("abc"));
    }
}
