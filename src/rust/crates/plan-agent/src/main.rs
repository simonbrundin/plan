//! Agent-API:t.
//!
//! Rust-ersättning för `src/agent/cmd/agent-api`. Routern bygger samma
//! vägar som Go-agenten - paritetstestet i `routes/table.rs` håller de två
//! i synk.

use std::process::ExitCode;
use std::sync::Arc;
use std::time::Duration;

use axum::http::{header, HeaderValue, Method, StatusCode};
use axum::response::{IntoResponse, Response};
use axum::Router;
use tokio::net::TcpListener;
use tower_http::trace::TraceLayer;

use plan_agent::planclient::PlanClient;
use plan_agent::routes::table;
use plan_agent::state::AppState;

/// Agentens databas om `DATABASE_URL` inte är satt - samma som i Go.
const DEFAULT_DATABASE_URL: &str =
    "postgres://postgres:postgres@postgres:5432/software_agent?sslmode=disable";

#[tokio::main]
async fn main() -> ExitCode {
    tracing_subscriber::fmt()
        .with_env_filter(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "info,plan_agent=debug".into()),
        )
        .init();

    let port: u16 = std::env::var("PORT")
        .ok()
        .and_then(|value| value.parse().ok())
        .unwrap_or(8080);

    // `agent-api healthcheck` används av Dockerfile:s HEALTHCHECK.
    if std::env::args().nth(1).as_deref() == Some("healthcheck") {
        return match healthcheck(port) {
            0 => ExitCode::SUCCESS,
            _ => ExitCode::FAILURE,
        };
    }

    // Plan-kopplingen är valfri. Utan PLAN_API_KEY startas bara API:t.
    let plan_client = match PlanClient::new() {
        Ok(client) => {
            tracing::info!("Plan API-klient skapad mot {}", client.base_url());
            Some(client)
        }
        Err(err) => {
            tracing::warn!("Ingen Plan-klient ({err}) - agent-API:t startar utan Plan-koppling");
            None
        }
    };

    let api_key = std::env::var("AGENT_API_KEY")
        .ok()
        .filter(|key| !key.is_empty());
    if api_key.is_none() {
        tracing::warn!("AGENT_API_KEY är inte satt - agent-API:t tar emot anrop utan nyckel");
    }

    // Agentens egen databas. Utan DATABASE_URL används Go-agentens
    // standard-DSN, och kan den inte anslutas startar inte agenten -
    // precis som Go-agenten gör.
    let database_url = std::env::var("DATABASE_URL")
        .ok()
        .filter(|url| !url.is_empty())
        .unwrap_or_else(|| DEFAULT_DATABASE_URL.to_string());

    let db = match connect_with_retry(&database_url, db_connect_timeout()).await {
        Ok(db) => {
            tracing::info!("Agentens databas ansluten");
            db
        }
        Err(err) => {
            tracing::error!(
                "Kunde inte ansluta agentens databas: {err}. \
                 Startar om - podden körs om och försöker igen."
            );
            return ExitCode::FAILURE;
        }
    };

    // Store:n med förslag som väntar på ett beslut. Den finns alltid -
    // den kan användas utan Plan-koppling.
    let approval_store = match plan_agent::approvals::Store::new() {
        Ok(store) => Some(Arc::new(store)),
        Err(err) => {
            tracing::error!("Kunde inte skapa godkännandestoren: {err}");
            None
        }
    };

    // Orchestratorn är redo.
    //
    // Go:s `InitOrchestrator` faller tillbaka på minneslagring om
    // statestoret saknas och returnerar därför aldrig ett fel - bara
    // en fel i `orchestrator.New` når health. Här markeras den redo
    // när agenten startat. Den riktiga orchestratorn (fas 4d) ersätter
    // det här och sätter tillståndet själv.
    plan_agent::orchestrator::set_ready();

    // Orchestratorn sparar checkpointplanerna i agentens databas, så
    // att en plan överlever en omstart.
    let orchestrator = plan_agent::orchestrator::SharedOrchestrator::new(
        plan_agent::orchestrator::Orchestrator::new(plan_agent::state_store::StateStore::Database(
            plan_agent::state_store::CheckpointStore::new(db.pool().clone()),
        )),
    );

    let state = AppState {
        plan_client: plan_client.clone(),
        api_key,
        db: Some(db),
        approval_store: approval_store.clone(),
        orchestrator: Some(orchestrator),
    };

    // Vad som startas styrs av AGENT_WORKER_ENABLED och
    // AGENT_SCHEDULER_ENABLED. I klustret är båda avstängda - där går
    // arbetet via Kubernetes Jobs i stället.
    //
    // AGENT_GOAL_MODE är deprecated. Den används bara om ingen av
    // flaggorna är satta, så att en gammal deployment inte tyst slutar
    // att starta något.
    let mut worker_enabled = std::env::var("AGENT_WORKER_ENABLED").ok();
    let mut scheduler_enabled = std::env::var("AGENT_SCHEDULER_ENABLED").ok();

    if worker_enabled.as_deref() == Some("") && scheduler_enabled.as_deref() == Some("") {
        match std::env::var("AGENT_GOAL_MODE").ok().as_deref() {
            Some("worker") => worker_enabled = Some("true".to_string()),
            Some("") | Some("approval") => scheduler_enabled = Some("true".to_string()),
            _ => {
                tracing::warn!("Unknown AGENT_GOAL_MODE - starting approval scheduler");
                scheduler_enabled = Some("true".to_string());
            }
        }
    }

    if worker_enabled.as_deref() == Some("true") {
        starta_workern(plan_client.clone());
    }

    if scheduler_enabled.as_deref() == Some("true") {
        starta_schemaläggaren(plan_client.clone(), approval_store.clone());
    }

    // Jobbcontrollern startar Kubernetes Jobs för mål som behöver en
    // agentkörning. Den kan stängas av när jobb inte ska köras.
    if std::env::var("AGENT_JOB_CONTROLLER_ENABLED").as_deref() == Ok("true") {
        starta_jobbcontrollern(plan_client.clone());
    }

    let app = Router::new()
        .merge(table::build())
        .layer(axum::middleware::from_fn(cors))
        .layer(axum::middleware::from_fn(require_api_key))
        .layer(TraceLayer::new_for_http())
        .with_state(state);

    let address = format!("0.0.0.0:{port}");
    let listener = match TcpListener::bind(&address).await {
        Ok(listener) => listener,
        Err(err) => {
            tracing::error!("Kunde inte lyssna på {address}: {err}");
            return ExitCode::FAILURE;
        }
    };

    tracing::info!("Agent-API:t lyssnar på {address}");
    tracing::info!("Scalar UI: http://localhost:{port}/docs");

    if let Err(err) = axum::serve(listener, app)
        .with_graceful_shutdown(shutdown_signal())
        .await
    {
        tracing::error!("Servern stoppades med fel: {err}");
        return ExitCode::FAILURE;
    }

    ExitCode::SUCCESS
}

/// Hur länge agenten väntar på databasen innan den ger upp.
///
/// Samma princip som API:t: en ny CNPG-kluster behöver tid för initdb,
/// och agenten ska vänta i stället för att starta om i en crashloop.
fn db_connect_timeout() -> Duration {
    Duration::from_secs(
        std::env::var("DB_CONNECT_TIMEOUT_SECONDS")
            .ok()
            .and_then(|value| value.parse().ok())
            .unwrap_or(120),
    )
}

/// Väntar på agentens databas och kör migreringen.
///
/// Agentens databas är ett eget CNPG-kluster. Direkt efter att den
/// skapats kör den initdb, och då svarar den inte. Utan väntan skulle
/// agenten starta om i en crashloop tills databasen är klar - varje
/// försök med en egen backoff.
///
/// `AgentDb::connect` skapar tabellerna om de saknas (`CREATE TABLE IF
/// NOT EXISTS`), så agentens schema är klart när den här funktionen
/// returnerar.
async fn connect_with_retry(
    database_url: &str,
    timeout: Duration,
) -> Result<plan_agent::db::AgentDb, sqlx::Error> {
    let start = std::time::Instant::now();
    let mut attempt: u32 = 1;

    loop {
        match plan_agent::db::AgentDb::connect(Some(database_url)).await {
            Ok(Some(db)) => {
                tracing::info!("Agentens databas ansluten (försök {attempt})");
                return Ok(db);
            }
            Ok(None) => {
                tracing::error!("DATABASE_URL saknas - agent-API:t kan inte starta");
                return Err(sqlx::Error::Configuration("DATABASE_URL saknas".into()));
            }
            Err(err) => {
                if start.elapsed() >= timeout {
                    return Err(err);
                }

                tracing::warn!(
                    "Kunde inte ansluta agentens databas (försök {attempt}): {err} - \
                     väntar 3s och försöker igen"
                );
                tokio::time::sleep(Duration::from_secs(3)).await;
                attempt += 1;
            }
        }
    }
}

/// Kräver agentens egen nyckel på alla anrop.
///
/// Nyckeln sätts i `AGENT_API_KEY`. Utan den öppna nyckeln är tjänsten
/// öppen för alla som når den - och då kan någon besluta om ett
/// godkännande åt människor. Därför loggas det tydligt när tjänsten
/// startas utan nyckel.
///
/// Hälsokontrollen är undantagen, den säger ingenting om vad tjänsten får.
async fn require_api_key(
    request: axum::extract::Request,
    next: axum::middleware::Next,
) -> Response {
    let expected = std::env::var("AGENT_API_KEY").unwrap_or_default();
    if expected.is_empty() {
        return next.run(request).await;
    }

    let path = request.uri().path();
    if is_public_path(path) {
        return next.run(request).await;
    }

    let provided = request
        .headers()
        .get(header::AUTHORIZATION)
        .and_then(|value| value.to_str().ok())
        .map(str::trim)
        .and_then(|value| value.strip_prefix("Bearer "))
        .map(str::trim);

    match provided {
        Some(key) if key == expected => next.run(request).await,
        _ => (
            StatusCode::UNAUTHORIZED,
            axum::Json(serde_json::json!({
                "error": "Unauthorized",
                "message": "En giltig API-nyckel krävs",
            })),
        )
            .into_response(),
    }
}

/// Anger vilka anrop som är öppna utan nyckel.
///
/// Det är hälsokontrollen och dokumentationen - de behöver fungera utan
/// nyckel så att drift kan se att tjänsten lever och så att API:et går
/// att läsa. Go registrerar dem utanför nyckelskyddet, så här gör vi
/// samma sak.
fn is_public_path(path: &str) -> bool {
    matches!(
        path,
        "/api/v1/health" | "/api/v1/status" | "/health" | "/openapi.json"
    ) || path == "/docs"
        || path.starts_with("/docs/")
}

/// CORS - samma uppsättning som Go-agenten (alla ursprung tillåtna).
async fn cors(request: axum::extract::Request, next: axum::middleware::Next) -> Response {
    if request.method() == Method::OPTIONS {
        let mut response = StatusCode::NO_CONTENT.into_response();
        let headers = response.headers_mut();
        headers.insert(
            header::ACCESS_CONTROL_ALLOW_ORIGIN,
            HeaderValue::from_static("*"),
        );
        headers.insert(
            header::ACCESS_CONTROL_ALLOW_HEADERS,
            HeaderValue::from_static("Origin, Content-Type, Accept, Authorization"),
        );
        return response;
    }

    let mut response = next.run(request).await;
    let headers = response.headers_mut();
    headers.insert(
        header::ACCESS_CONTROL_ALLOW_ORIGIN,
        HeaderValue::from_static("*"),
    );
    response
}

/// Startar workern som tar sig an mål med agentflaggan.
///
/// Den här workern är skild från schemaläggaren: schemaläggaren
/// definierar mål och lämnar förslaget till en människa, workern tar sig
/// an de mål som är markerade med agentflaggan.
///
/// Go: `startAgentWorker`.
fn starta_workern(plan: Option<PlanClient>) {
    let Some(plan) = plan else {
        tracing::info!("Plan client is nil, agent worker not started");

        return;
    };

    let (leverantör, modell) = match plan_agent::provider::Factory::skapa_från_miljö() {
        Ok(leverans) => leverans,
        Err(err) => {
            tracing::error!("Failed to create LLM provider: {err}");
            tracing::info!("Agent worker not started (set MINIMAX_API_KEY or OPENAI_API_KEY)");

            return;
        }
    };

    tracing::info!(
        "Agent worker using LLM provider: {} ({})",
        modell.provider,
        modell.name
    );

    let definierare = Arc::new(plan_agent::definer::Definer::ny(leverantör, modell.clone()));
    let worker = plan_agent::worker::Worker::ny(
        plan,
        definierare,
        plan_agent::worker::Config {
            pollintervall: plan_agent::worker::pollintervall_ur_miljö(),
            leverantör: modell.provider.clone(),
            modell: modell.name.clone(),
            agent_id: agent_id_ur_miljö(),
            ..Default::default()
        },
    );

    tokio::spawn(async move {
        worker.run().await;
    });
}

/// Startar schemaläggaren som definierar mål och ber om godkännande.
///
/// Go: `startGoalScheduler`.
fn starta_schemaläggaren(
    plan: Option<PlanClient>,
    approval_store: Option<Arc<plan_agent::approvals::Store>>,
) {
    let Some(plan) = plan else {
        tracing::info!("Plan client is nil, scheduler not started");

        return;
    };

    let Some(approval_store) = approval_store else {
        tracing::info!("Approval store is nil, scheduler not started");

        return;
    };

    tracing::info!("Starting goal scheduler...");

    let (leverantör, modell) = match plan_agent::provider::Factory::skapa_från_miljö() {
        Ok(leverans) => leverans,
        Err(err) => {
            tracing::error!("Failed to create LLM provider: {err}");
            tracing::info!("Goal scheduler not started (set MINIMAX_API_KEY or OPENAI_API_KEY)");

            return;
        }
    };

    tracing::info!(
        "LLM provider created: {} ({})",
        modell.provider,
        modell.name
    );

    let definierare = Arc::new(plan_agent::definer::Definer::ny(leverantör, modell));

    let scheduler = plan_agent::scheduler::Scheduler::ny(
        plan,
        approval_store,
        definierare,
        plan_agent::scheduler::Config {
            initiator: "scheduler:agent-api".to_string(),
            ..Default::default()
        },
    );

    tokio::spawn(async move {
        scheduler.run().await;
    });
}

/// Vilken agent som kör.
///
/// Namnet används för att filtrera kostnader per agent.
fn agent_id_ur_miljö() -> String {
    if let Some(raw) = std::env::var("AGENT_ID").ok().filter(|v| !v.is_empty()) {
        return raw;
    }

    if let Some(värdnamn) = värdnamn() {
        if !värdnamn.is_empty() {
            return format!("agent:{värdnamn}");
        }
    }

    "agent:agent-api".to_string()
}

/// Värdnamnet på maskinen.
fn värdnamn() -> Option<String> {
    std::fs::read_to_string("/etc/hostname")
        .ok()
        .map(|namn| namn.trim().to_string())
        .filter(|namn| !namn.is_empty())
}

/// Startar controllern som skapar Kubernetes Jobs.
///
/// Controllern är kopplad mellan Plan och klustret: Plan äger målen,
/// klustret utför arbetet. Den behöver rättigheter att skapa jobb - inga
/// andra.
///
/// Go: `startJobController`.
fn starta_jobbcontrollern(plan: Option<PlanClient>) {
    let Some(plan) = plan else {
        tracing::info!("Plan client is nil, job controller not started");

        return;
    };

    let klient = match plan_agent::k8s::Client::in_cluster() {
        Ok(klient) => klient,
        Err(err) => {
            tracing::error!("Failed to read in-cluster config: {err}");
            tracing::info!("Job controller not started");

            return;
        }
    };

    let runner = plan_agent::jobrunner::Runner::ny(klient, jobbkonfiguration_ur_miljö());

    let controller = plan_agent::jobcontroller::Controller::ny(
        plan,
        runner,
        plan_agent::jobcontroller::Config {
            pollintervall: jobb_pollintervall_ur_miljö(),
            namespace: std::env::var("AGENT_JOB_NAMESPACE").unwrap_or_default(),
            max_concurrent_jobs: jobb_max_parallellt_ur_miljö(),
        },
    );

    tokio::spawn(async move {
        controller.run().await;
    });
}

/// Bygger jobbkonfigurationen från miljön.
///
/// Go: `jobConfigFromEnv`.
fn jobbkonfiguration_ur_miljö() -> plan_agent::jobrunner::Config {
    let mut config = plan_agent::jobrunner::Config::default();

    if let Some(värde) = miljö_värde("AGENT_JOB_NAMESPACE") {
        config.namespace = värde;
    }

    if let Some(värde) = miljö_värde("AGENT_JOB_IMAGE") {
        config.image = värde;
    }

    if let Some(värde) = miljö_värde("AGENT_JOB_SERVICE_ACCOUNT") {
        config.service_account = värde;
    }

    if let Some(värde) = miljö_värde("AGENT_JOB_REPO_URL") {
        config.repo_url = värde;
    }

    if let Some(värde) = miljö_värde("AGENT_JOB_REPO_BRANCH") {
        config.repo_branch = värde;
    }

    if let Some(värde) = miljö_värde("AGENT_JOB_PLAN_API_URL") {
        config.plan_api_url = värde;
    }

    if let Some(värde) = miljö_värde("AGENT_JOB_API_KEY_SECRET") {
        config.api_key_secret_name = värde;
    }

    if let Some(värde) = miljö_värde("AGENT_JOB_LLM_SECRET") {
        config.llm_secret_name = värde;
    }

    if let Some(värde) = miljö_värde("AGENT_JOB_GIT_SECRET") {
        config.git_secret_name = värde;
    }

    if let Some(värde) = miljö_värde("AGENT_JOB_CPU_LIMIT") {
        config.cpu_limit = värde;
    }

    if let Some(värde) = miljö_värde("AGENT_JOB_MEMORY_LIMIT") {
        config.memory_limit = värde;
    }

    if let Some(värde) = miljö_värde("AGENT_JOB_DEADLINE_SECONDS") {
        if let Ok(seconds) = värde.parse::<i64>() {
            config.active_deadline_seconds = seconds;
        }
    }

    config
}

/// Läser den globala jobbkapaciteten.
///
/// Go: `jobMaxConcurrentJobsFromEnv`.
fn jobb_max_parallellt_ur_miljö() -> usize {
    let rå = std::env::var("AGENT_JOB_MAX_CONCURRENT").unwrap_or_default();

    if rå.is_empty() {
        return 4;
    }

    match rå.parse::<usize>() {
        Ok(värde) if värde >= 1 => värde,
        _ => {
            tracing::warn!("Invalid AGENT_JOB_MAX_CONCURRENT {rå:?}, using 4");

            4
        }
    }
}

/// Läser hur ofta controllern läser Plan och klustret.
///
/// Go: `jobPollIntervalFromEnv`.
fn jobb_pollintervall_ur_miljö() -> Duration {
    let rå = std::env::var("AGENT_JOB_POLL_INTERVAL").unwrap_or_default();

    if rå.is_empty() {
        return Duration::from_secs(30);
    }

    match parse_go_duration(&rå) {
        Some(interval) => interval,
        None => {
            tracing::warn!("Invalid AGENT_JOB_POLL_INTERVAL {rå:?}, using 30s");

            Duration::from_secs(30)
        }
    }
}

/// Läser en miljövariabel som inte är tom.
///
/// Go: `os.Getenv` med `value != ""`.
fn miljö_värde(namn: &str) -> Option<String> {
    std::env::var(namn).ok().filter(|värde| !värde.is_empty())
}

/// Tolkar en Go-duration ("30s", "2m", "1h30m").
///
/// Go använder `time.ParseDuration` - formatet ska tolkas likadant.
fn parse_go_duration(text: &str) -> Option<Duration> {
    // Go godkänner "0" utan enhet.
    if text == "0" {
        return Some(Duration::ZERO);
    }

    let mut total = Duration::ZERO;
    let mut tal = String::new();
    let mut enhet = String::new();

    for tecken in text.chars() {
        if tecken.is_ascii_digit() || tecken == '.' {
            if !enhet.is_empty() {
                // Enheten är slut - siffran efter den hör till nästa del.
                total += längd_för(&tal, &enhet)?;
                tal.clear();
                enhet.clear();
            }

            tal.push(tecken);

            continue;
        }

        if tecken == ' ' {
            continue;
        }

        enhet.push(tecken);
    }

    if !tal.is_empty() {
        total += längd_för(&tal, &enhet)?;
    } else if !enhet.is_empty() {
        // En enhet utan tal är ett fel.
        return None;
    }

    Some(total)
}

/// Gör om ett tal med en enhet till en längd.
///
/// Go: enheterna i `time.ParseDuration`.
fn längd_för(tal: &str, enhet: &str) -> Option<Duration> {
    let värde: f64 = tal.parse().ok()?;

    let längd = match enhet {
        "ns" => Duration::from_nanos(värde as u64),
        "us" | "µs" => Duration::from_micros(värde as u64),
        "ms" => Duration::from_millis(värde as u64),
        "s" => Duration::from_secs_f64(värde),
        "m" => Duration::from_secs_f64(värde * 60.0),
        "h" => Duration::from_secs_f64(värde * 3_600.0),
        _ => return None,
    };

    Some(längd)
}

/// TCP-probe mot den egna porten. Returnerar exit-kod 0/1.
fn healthcheck(port: u16) -> i32 {
    let adress = format!("127.0.0.1:{port}");
    match std::net::TcpStream::connect_timeout(
        &adress.parse().expect("ogiltig adress"),
        Duration::from_secs(2),
    ) {
        Ok(_) => 0,
        Err(err) => {
            eprintln!("healthcheck misslyckades ({adress}): {err}");
            1
        }
    }
}

/// Väntar på SIGINT/SIGTERM och stänger sedan servern rent.
async fn shutdown_signal() {
    let ctrl_c = async {
        let _ = tokio::signal::ctrl_c().await;
    };

    #[cfg(unix)]
    let terminate = async {
        match tokio::signal::unix::signal(tokio::signal::unix::SignalKind::terminate()) {
            Ok(mut signal) => {
                signal.recv().await;
            }
            Err(err) => tracing::warn!("Kunde inte lyssna på SIGTERM: {err}"),
        }
    };

    #[cfg(not(unix))]
    let terminate = std::future::pending::<()>();

    tokio::select! {
        _ = ctrl_c => tracing::info!("Fick Ctrl-C, stänger av"),
        _ = terminate => tracing::info!("Fick SIGTERM, stänger av"),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn go_durationer_tolkas() {
        assert_eq!(parse_go_duration("30s"), Some(Duration::from_secs(30)));
        assert_eq!(parse_go_duration("2m"), Some(Duration::from_secs(120)));
        assert_eq!(parse_go_duration("1h30m"), Some(Duration::from_secs(5400)));
        assert_eq!(parse_go_duration("500ms"), Some(Duration::from_millis(500)));
        assert_eq!(parse_go_duration("0"), Some(Duration::ZERO));

        assert_eq!(parse_go_duration("30"), None);
        assert_eq!(parse_go_duration("abc"), None);
        assert_eq!(parse_go_duration("30x"), None);
    }
}
