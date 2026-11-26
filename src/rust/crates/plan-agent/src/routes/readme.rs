//! Agentens README.
//!
//! Go-agentens `internal/api/handlers/docs.go` - samma sökvägar, samma
//! innehåll och samma innehållstyp.

use axum::response::{IntoResponse, Response};

use crate::error::AgentError;

/// Returnerar README.md som text.
pub async fn readme() -> Result<Response, AgentError> {
    let Some(innehåll) = hitta_projektfil("README.md") else {
        return Ok((axum::http::StatusCode::NOT_FOUND, "README.md not found").into_response());
    };

    Ok((
        axum::http::StatusCode::OK,
        [(
            axum::http::header::CONTENT_TYPE,
            "text/plain; charset=utf-8",
        )],
        innehåll,
    )
        .into_response())
}

/// Läser en fil från repots rot.
///
/// Sökvägen löses i samma ordning som i Go:
///  1. `AGENT_DOCS_DIR` - explicit katalog med filerna
///  2. `PROJECT_ROOT` - repots rot
///  3. arbetskatalogen
///  4. repots rot, inbyggd vid kompilering
fn hitta_projektfil(filnamn: &str) -> Option<String> {
    for katalog in projektkataloger() {
        let sökväg = std::path::Path::new(&katalog).join(filnamn);
        if let Ok(innehåll) = std::fs::read_to_string(&sökväg) {
            return Some(innehåll);
        }
    }

    None
}

/// Katalogerna som letas i.
fn projektkataloger() -> Vec<String> {
    let mut kataloger = Vec::new();

    if let Ok(katalog) = std::env::var("AGENT_DOCS_DIR") {
        if !katalog.is_empty() {
            kataloger.push(katalog);
        }
    }

    if let Ok(katalog) = std::env::var("PROJECT_ROOT") {
        if !katalog.is_empty() {
            kataloger.push(katalog);
        }
    }

    if let Ok(katalog) = std::env::current_dir() {
        kataloger.push(katalog.to_string_lossy().to_string());
    }

    // Repots rot - inbyggd när crate:en kompilerades.
    kataloger.push(repotsrot().to_string());

    kataloger
}

/// Repots rot, räknad baklänges från crate:en.
fn repotsrot() -> &'static str {
    // crates/plan-agent -> crates -> rust -> repots rot
    concat!(env!("CARGO_MANIFEST_DIR"), "/../../..")
}
