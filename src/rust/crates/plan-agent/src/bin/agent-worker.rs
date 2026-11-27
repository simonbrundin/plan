//! Agentworkern - pollar Plan efter mål som är markerade för agentarbete.
//!
//! Go: `cmd/agent-worker`. Workern kan köras för ett enda mål:
//!
//! ```text
//! AGENT_GOAL_ID=449 agent-worker
//! ```

use std::sync::Arc;

use plan_agent::definer::{Definer, Input};
use plan_agent::planclient::PlanClient;
use plan_agent::profile;
use plan_agent::provider::Factory;
use plan_agent::runusage::Recorder;
use plan_agent::worker::{Config, GoalDefiner, Worker};
use tracing::{info, warn};

/// Definieraren som workern använder.
struct DefinerAdapter {
    /// Den egentlige definieraren.
    definer: Definer,
}

impl GoalDefiner for DefinerAdapter {
    fn definiera<'a>(
        &'a self,
        input: &'a Input,
    ) -> std::pin::Pin<
        Box<
            dyn std::future::Future<Output = Result<plan_agent::definer::Definition, String>>
                + Send
                + 'a,
        >,
    > {
        Box::pin(self.definer.definiera(input))
    }
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt()
        .with_env_filter(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| tracing_subscriber::EnvFilter::new("info")),
        )
        .init();

    let plan = match PlanClient::new() {
        Ok(plan) => plan,
        Err(fel) => {
            eprintln!("Failed to create Plan client: {fel}");
            std::process::exit(1);
        }
    };

    let (leverantör, modell) = match Factory::skapa_från_miljö() {
        Ok(leverans) => leverans,
        Err(fel) => {
            eprintln!("Failed to create LLM provider: {fel}");
            std::process::exit(1);
        }
    };

    info!("Using LLM provider: {} ({})", modell.provider, modell.name);

    let definierare = Arc::new(DefinerAdapter {
        definer: Definer::ny(leverantör, modell.clone()),
    });

    // Recordern samlar tokenanvändningen för körningen. Plan räknar
    // sedan fram kostnaden ur sin prislista.
    let _recorder = Recorder::ny(&modell.provider, &modell.name);

    let worker = Worker::ny(
        plan,
        definierare,
        Config {
            pollintervall: plan_agent::worker::pollintervall_ur_miljö(),
            enda_mål_id: mål_id_ur_miljö(),
            initiator: initiator_ur_miljö(),
            profil: profil_ur_miljö(),
            kodrunner: None,
            stop_when_finished: stop_when_finished_ur_miljö(),
            leverantör: modell.provider.clone(),
            modell: modell.name.clone(),
            agent_id: agent_id_ur_miljö(),
        },
    );

    worker.run().await;
}

/// Vilken agent som kör.
///
/// Namnet används för att filtrera kostnader per agent.
fn agent_id_ur_miljö() -> String {
    if let Some(rå) = std::env::var("AGENT_ID").ok().filter(|v| !v.is_empty()) {
        return rå;
    }

    if let Some(värdnamn) = värdnamn() {
        if !värdnamn.is_empty() {
            return format!("agent:{värdnamn}");
        }
    }

    "agent:agent-worker".to_string()
}

/// Vem eller vad som startade körningen.
fn initiator_ur_miljö() -> String {
    if let Some(rå) = std::env::var("AGENT_INITIATOR")
        .ok()
        .filter(|v| !v.is_empty())
    {
        return rå;
    }

    if let Some(värdnamn) = värdnamn() {
        if !värdnamn.is_empty() {
            return format!("worker:{värdnamn}");
        }
    }

    "worker:agent-worker".to_string()
}

/// Läser `AGENT_PROFILE` - coding eller life.
fn profil_ur_miljö() -> String {
    match std::env::var("AGENT_PROFILE").ok().as_deref() {
        Some(profile::LIFE) => profile::LIFE.to_string(),
        Some(profile::CODING) => profile::CODING.to_string(),
        None | Some("") => profile::CODING.to_string(),
        Some(annat) => {
            warn!("Invalid AGENT_PROFILE {annat:?}, using {}", profile::CODING);

            profile::CODING.to_string()
        }
    }
}

/// Läser `AGENT_STOP_WHEN_FINISHED`.
fn stop_when_finished_ur_miljö() -> bool {
    let rå = std::env::var("AGENT_STOP_WHEN_FINISHED")
        .unwrap_or_default()
        .to_lowercase();

    rå == "true" || rå == "1" || rå == "yes"
}

/// Läser `AGENT_GOAL_ID` - om den sätts körs workern bara för det målet.
fn mål_id_ur_miljö() -> i64 {
    let rå = std::env::var("AGENT_GOAL_ID").unwrap_or_default();

    if rå.is_empty() {
        return 0;
    }

    match rå.parse::<i64>() {
        Ok(id) => id,
        Err(_) => {
            warn!("Invalid AGENT_GOAL_ID {rå:?}, running for all agent goals");

            0
        }
    }
}

/// Värdnamnet på maskinen.
fn värdnamn() -> Option<String> {
    std::fs::read_to_string("/etc/hostname")
        .ok()
        .map(|namn| namn.trim().to_string())
        .filter(|namn| !namn.is_empty())
}
