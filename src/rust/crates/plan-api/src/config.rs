//! Konfiguration från miljövariabler.
//!
//! Variabelnamnen är desamma som i Go-tjänsten, så deploymenten (Flux,
//! Kubernetes) behöver inte ändras.

use std::net::SocketAddr;

#[derive(Debug, Clone)]
// Fält som ännu inte används (sessioner, agent-nyckel) tas i bruk i de
// kommande faserna av migreringen.
#[allow(dead_code)]
pub struct Config {
    /// Porten API:t lyssnar på. Standard 8080, som i Go.
    pub port: u16,
    /// Adressen API:t binder till. Tom sträng betyder "alla interface".
    pub host: String,
    /// Anslutningssträng till PostgreSQL.
    pub database_url: Option<String>,
    /// Samma sessionshemlighet som Nuxt använder (nuxt-auth-utils).
    pub session_secret: String,
    /// Host som skrivs in i OpenAPI-specen.
    pub api_host: Option<String>,
    /// Sökväg till swagger.json - Scalar läser specen därifrån.
    pub docs_path: Option<String>,
    /// Gör att API-nyckeln i AGENT_API_KEY får agent-scope.
    pub agent_api_key: Option<String>,
    /// Utvecklingsnyckel som bypassar autentisering.
    pub api_key_bypass: Option<String>,
    /// Hur länge API:t väntar på databasen innan den ger upp och låter
    /// podden starta om. Default 120s.
    pub db_connect_timeout_secs: u64,
}

impl Config {
    /// Läser konfigurationen från processens miljö.
    pub fn from_env() -> Self {
        let session_secret = std::env::var("NUXT_SESSION_PASSWORD")
            .or_else(|_| std::env::var("SESSION_SECRET"))
            .unwrap_or_else(|_| "dev-secret-change-in-production".to_string());

        Self {
            port: std::env::var("PORT")
                .ok()
                .and_then(|value| value.parse().ok())
                .unwrap_or(8080),
            host: std::env::var("HOST").unwrap_or_default(),
            database_url: std::env::var("DATABASE_URL").ok().filter(|v| !v.is_empty()),
            session_secret,
            api_host: std::env::var("API_HOST").ok().filter(|v| !v.is_empty()),
            docs_path: std::env::var("DOCS_PATH").ok().filter(|v| !v.is_empty()),
            agent_api_key: std::env::var("AGENT_API_KEY")
                .ok()
                .filter(|v| !v.is_empty()),
            api_key_bypass: std::env::var("API_KEY_BYPASS")
                .ok()
                .filter(|v| !v.is_empty()),
            db_connect_timeout_secs: std::env::var("DB_CONNECT_TIMEOUT_SECONDS")
                .ok()
                .and_then(|value| value.parse().ok())
                .unwrap_or(120),
        }
    }

    /// Adressen servern lyssnar på.
    pub fn listen_addr(&self) -> SocketAddr {
        let host = if self.host.is_empty() {
            "0.0.0.0"
        } else {
            self.host.as_str()
        };
        format!("{host}:{}", self.port)
            .parse()
            .unwrap_or_else(|_| SocketAddr::from(([0, 0, 0, 0], self.port)))
    }
}
