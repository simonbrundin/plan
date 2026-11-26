//! Plan API - REST-API för Plan.
//!
//! Detta är Rust-ersättningen för Go-tjänsten i `src/api`. Gränssnittet
//! (Nuxt) och agenten pratar med API:t över HTTP, så kontraktet - vägar,
//! metoder, statuskoder och JSON-format - måste vara identiskt.

use std::process::ExitCode;
use std::time::Duration;

use tokio::net::TcpListener;

use plan_api::config::Config;
use plan_api::db;
use plan_api::routes;

#[tokio::main]
async fn main() -> ExitCode {
    tracing_subscriber::fmt()
        .with_env_filter(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "info,plan_api=debug".into()),
        )
        .init();

    let config = Config::from_env();

    // `plan-api healthcheck` används av Dockerfile:s HEALTHCHECK - Alpine
    // saknar curl/wget, så samma binär gör TCP-proben.
    if std::env::args().nth(1).as_deref() == Some("healthcheck") {
        return match healthcheck(config.port) {
            0 => ExitCode::SUCCESS,
            _ => ExitCode::FAILURE,
        };
    }

    // Databasen är valfri i Go-tjänsten ("mock mode"). Samma beteende här:
    // utan DATABASE_URL startar API:t, men endpoints som behöver data
    // svarar med ett tydligt fel i stället för att krascha.
    let pool = match db::connect(&config).await {
        Ok(pool) => {
            tracing::info!("Ansluten till databasen");
            if let Err(err) = db::ensure_seed_data(&pool).await {
                tracing::warn!("Kunde inte seeda data: {err}");
            } else {
                tracing::info!("Seed-data verifierad");
            }
            Some(pool)
        }
        Err(err) => {
            tracing::warn!("Kunde inte ansluta till databasen: {err}");
            tracing::warn!("Startar i mock-läge utan databasåtkomst");
            None
        }
    };

    let app = routes::router(pool, &config);

    let addr = config.listen_addr();
    let listener = match TcpListener::bind(addr).await {
        Ok(listener) => listener,
        Err(err) => {
            tracing::error!("Kunde inte lyssna på {addr}: {err}");
            return ExitCode::FAILURE;
        }
    };

    tracing::info!("Startar servern på {addr}");
    tracing::info!("Scalar UI finns på http://localhost:{}/docs/", config.port);

    if let Err(err) = axum::serve(listener, app)
        .with_graceful_shutdown(shutdown_signal())
        .await
    {
        tracing::error!("Servern stoppades med fel: {err}");
        return ExitCode::FAILURE;
    }

    tracing::info!("Servern avslutad");
    ExitCode::SUCCESS
}

/// TCP-probe mot den egna API-porten. Returnerar exit-kod 0/1.
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
            Ok(mut sig) => {
                sig.recv().await;
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
