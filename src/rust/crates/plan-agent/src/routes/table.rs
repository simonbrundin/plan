//! Route-tabellen för agent-API:et.
//!
//! Tabellen är agentens kontrakt med omvärlden. Paritetstestet i slutet av
//! filen jämför tabellen mot route-registreringarna i Go-tjänsten
//! (`src/agent/internal/api/server.go`), så en endpoint som glöms bort
//! failar i CI.

use axum::routing::{delete, get, patch, post};
use axum::Router;

use crate::state::AppState;

/// Alla agent-vägar. Samma ordning som i Go-tjänsten.
pub const ROUTES: &[(&str, &str)] = &[
    // Hälsokontroll (utanför API-nyckeln)
    ("GET", "/health"),
    // Dokumentation
    ("GET", "/docs/{*any}"),
    ("GET", "/openapi.json"),
    // Status
    ("GET", "/api/v1/health"),
    ("GET", "/api/v1/status"),
    // Workflow
    ("POST", "/api/v1/workflow"),
    ("GET", "/api/v1/workflow/plans"),
    ("GET", "/api/v1/workflow/{id}"),
    ("POST", "/api/v1/workflow/{id}/approve"),
    ("POST", "/api/v1/workflow/{id}/reject"),
    ("GET", "/api/v1/workflow/current/checkpoint"),
    ("GET", "/api/v1/workflow/current/progress"),
    ("POST", "/api/v1/workflow/decision"),
    // Samtal
    ("GET", "/api/v1/conversations"),
    ("POST", "/api/v1/conversations"),
    ("GET", "/api/v1/conversations/{id}"),
    ("PATCH", "/api/v1/conversations/{id}"),
    // Meddelanden
    ("GET", "/api/v1/conversations/{id}/messages"),
    ("POST", "/api/v1/conversations/{id}/messages"),
    ("GET", "/api/v1/conversations/{id}/stream"),
    // API-nycklar
    ("GET", "/api/v1/api-keys"),
    ("POST", "/api/v1/api-keys"),
    ("DELETE", "/api/v1/api-keys/{provider}"),
    // Godkännanden - besluten tas i Plan
    ("GET", "/api/v1/approvals"),
    ("GET", "/api/v1/approvals/{goal_id}"),
    ("POST", "/api/v1/approvals/{goal_id}/decision"),
    // Dokumentation (README som markdown)
    ("GET", "/api/v1/readme"),
    // WebSocket för realtids-chatt
    ("GET", "/api/v1/ws/chat"),
    // Verktygskörning (avbryt)
    ("POST", "/api/v1/tools/cancel"),
    ("POST", "/api/v1/tools/cancel-all"),
    ("GET", "/api/v1/tools/status"),
];

/// Bygger en router med alla vägar i tabellen.
///
/// Handlers kopplas på i takt med att de portas från Go. Vägar som ännu
/// inte är portade svarar med 501 i stället för 404 - så att en förbise
/// endpoint är tydlig i loggen och i testerna.
pub fn build() -> Router<AppState> {
    let mut router = Router::new();

    for (method, path) in ROUTES {
        if *path == "/health" {
            router = router.route("/health", get(crate::handlers::health));
            continue;
        }

        router = router.route(
            path,
            match *method {
                "GET" => get(crate::handlers::not_implemented),
                "POST" => post(crate::handlers::not_implemented),
                "PUT" => axum::routing::put(crate::handlers::not_implemented),
                "PATCH" => patch(crate::handlers::not_implemented),
                "DELETE" => delete(crate::handlers::not_implemented),
                other => panic!("okänd HTTP-metod i route-tabellen: {other}"),
            },
        );
    }

    router
}

#[cfg(test)]
mod tests {
    use std::collections::BTreeSet;
    use std::fs;
    use std::path::PathBuf;

    use super::ROUTES;

    /// Sökvägen till Go-agentens server.go.
    fn go_server() -> PathBuf {
        PathBuf::from(env!("CARGO_MANIFEST_DIR")).join("../../../agent/internal/api/server.go")
    }

    /// Route-registreringarna i Go-agenten.
    fn go_routes() -> BTreeSet<(String, String)> {
        let kalla = fs::read_to_string(go_server())
            .expect("går att läsa Go-agentens server.go - annars fungerar paritetstestet inte");
        plan_parity::go_routes(&kalla)
    }

    /// Rust-tabellens routes.
    fn rust_routes() -> Vec<(String, String)> {
        ROUTES
            .iter()
            .map(|(metod, sokvag)| (metod.to_string(), sokvag.to_string()))
            .collect()
    }

    /// Varje route i Go-agenten ska finnas i Rust-tabellen.
    #[test]
    fn go_routes_finns_i_rust_tabellen() {
        let (saknas, _) = plan_parity::jämför(&go_routes(), &rust_routes());

        assert!(
            saknas.is_empty(),
            "följande routes finns i Go-agenten men inte i Rust-tabellen: {}",
            saknas.join(", ")
        );
    }

    /// Rust-tabellen ska inte ha fler routes än Go-agenten.
    #[test]
    fn rust_tabellen_har_inga_extra_routes() {
        let (_, extra) = plan_parity::jämför(&go_routes(), &rust_routes());

        assert!(
            extra.is_empty(),
            "följande routes finns i Rust-tabellen men inte i Go-agenten: {}",
            extra.join(", ")
        );
    }

    /// Go-agenten ska registrera routes - annars är paritetstestet tomt.
    #[test]
    fn go_routes_finns_att_jamfora_mot() {
        assert!(
            go_routes().len() > 20,
            "route-tabellen i Go-agenten verkar ofullständig - kontrollera parsningen"
        );
    }
}
