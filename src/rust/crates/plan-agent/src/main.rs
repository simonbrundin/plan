//! Agent-API:t.
//!
//! Rust-ersättning för `src/agent/cmd/agent-api`. Routern bygger samma
//! vägar som Go-agenten - paritetstestet i `routes/table.rs` håller de två
//! i synk.

use std::process::ExitCode;
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

    let db = match plan_agent::db::AgentDb::connect(Some(&database_url)).await {
        Ok(Some(db)) => {
            tracing::info!("Agentens databas ansluten");
            db
        }
        Ok(None) => {
            tracing::error!("DATABASE_URL saknas - agent-API:t kan inte starta");
            return ExitCode::FAILURE;
        }
        Err(err) => {
            tracing::error!("Kunde inte ansluta agentens databas: {err}");
            return ExitCode::FAILURE;
        }
    };

    // Store:n med förslag som väntar på ett beslut. Den finns alltid -
    // den kan användas utan Plan-koppling.
    let approval_store = match plan_agent::approvals::Store::new() {
        Ok(store) => Some(store),
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

    let state = AppState {
        plan_client,
        api_key,
        db: Some(db),
        approval_store,
    };

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
/// Det är bara hälsokontrollen - den behöver fungera utan nyckel så att
/// drift kan se att tjänsten lever.
fn is_public_path(path: &str) -> bool {
    matches!(path, "/api/v1/health" | "/api/v1/status" | "/health")
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
