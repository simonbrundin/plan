//! Route-tabellen för agent-API:et.
//!
//! Tabellen är agentens kontrakt med omvärlden. Paritetstestet i slutet av
//! filen jämför tabellen mot route-registreringarna i Go-tjänsten
//! (`src/agent/internal/api/server.go`), så en endpoint som glöms bort
//! failar i CI.

use axum::routing::{delete, get, patch, post, MethodRouter};
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

        // Gin skickar "/docs" vidare till "/docs/" med sin trailing-slash-
        // redirect. Axum gör inte det, så det händer här.
        if *path == "/docs/{*any}" {
            router = router
                .route("/docs", get(crate::routes::docs::docs_redirect))
                .route("/docs/", get(crate::routes::docs::docs_root));
        }

        router = router.route(path, route_for(method, path));
    }

    router
}

/// Handler för en väg.
///
/// Portade vägar pekar på sin handler - resten svarar med 501, så att en
/// förbise endpoint är tydlig i loggen och i testerna.
fn route_for(method: &str, path: &str) -> MethodRouter<AppState> {
    match (method, path) {
        ("GET", "/api/v1/conversations") => get(crate::routes::conversations::list),
        ("POST", "/api/v1/conversations") => post(crate::routes::conversations::create),
        ("GET", "/api/v1/conversations/{id}") => get(crate::routes::conversations::get),
        ("PATCH", "/api/v1/conversations/{id}") => patch(crate::routes::conversations::update),
        ("GET", "/api/v1/conversations/{id}/messages") => {
            get(crate::routes::conversations::messages)
        }
        ("POST", "/api/v1/conversations/{id}/messages") => {
            post(crate::routes::conversations::create_message)
        }
        ("GET", "/api/v1/conversations/{id}/stream") => get(crate::routes::conversations::stream),

        ("GET", "/api/v1/health") => get(crate::routes::health::health),
        ("GET", "/api/v1/status") => get(crate::routes::health::status),
        ("GET", "/api/v1/readme") => get(crate::routes::readme::readme),

        ("POST", "/api/v1/workflow") => post(crate::routes::workflow::start),
        ("GET", "/api/v1/workflow/plans") => get(crate::routes::workflow::plans),
        ("GET", "/api/v1/workflow/{id}") => get(crate::routes::workflow::get),
        ("POST", "/api/v1/workflow/{id}/approve") => post(crate::routes::workflow::approve),
        ("POST", "/api/v1/workflow/{id}/reject") => post(crate::routes::workflow::reject),
        ("GET", "/api/v1/workflow/current/checkpoint") => {
            get(crate::routes::workflow::current_checkpoint)
        }
        ("GET", "/api/v1/workflow/current/progress") => {
            get(crate::routes::workflow::current_progress)
        }
        ("POST", "/api/v1/workflow/decision") => post(crate::routes::workflow::decision),

        ("GET", "/docs/{*any}") => get(crate::routes::docs::docs),
        ("GET", "/openapi.json") => get(crate::routes::docs::openapi_json),

        ("POST", "/api/v1/tools/cancel") => post(crate::routes::tools::cancel),
        ("POST", "/api/v1/tools/cancel-all") => post(crate::routes::tools::cancel_all),
        ("GET", "/api/v1/tools/status") => get(crate::routes::tools::status),

        ("GET", "/api/v1/approvals") => get(crate::routes::approvals::list),
        ("GET", "/api/v1/approvals/{goal_id}") => get(crate::routes::approvals::get),
        ("POST", "/api/v1/approvals/{goal_id}/decision") => post(crate::routes::approvals::decide),

        ("GET", "/api/v1/ws/chat") => get(crate::routes::ws::chat),

        ("GET", "/api/v1/api-keys") => get(crate::routes::api_keys::list),
        ("POST", "/api/v1/api-keys") => post(crate::routes::api_keys::set),
        ("DELETE", "/api/v1/api-keys/{provider}") => delete(crate::routes::api_keys::delete),

        ("GET", _) => get(crate::handlers::not_implemented),
        ("POST", _) => post(crate::handlers::not_implemented),
        ("PUT", _) => axum::routing::put(crate::handlers::not_implemented),
        ("PATCH", _) => patch(crate::handlers::not_implemented),
        ("DELETE", _) => delete(crate::handlers::not_implemented),
        other => panic!("okänd HTTP-metod i route-tabellen: {other:?}"),
    }
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
