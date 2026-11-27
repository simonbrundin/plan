//! Route-tabellen.
//!
//! Tabellen är Plan API:s kontrakt med gränssnittet och agenten. Varje rad
//! är en (metod, sökväg) som serveras. Paritetstestet i slutet av filen
//! jämför tabellen mot route-registreringarna i Go-tjänsten
//! (`src/api/cmd/api/main.go`), så en endpoint som glöms bort failar i CI.

use axum::routing::{delete, get, patch, post, put, MethodRouter};
use axum::Router;

use crate::routes::state::AppState;

/// Alla API-vägar. Samma ordning som i Go-tjänsten.
pub const ROUTES: &[(&str, &str)] = &[
    // Hälsokontroll - utan autentisering
    ("GET", "/health"),
    // Auth
    ("GET", "/api/v1/auth/login"),
    ("GET", "/api/v1/auth/callback"),
    ("POST", "/api/v1/auth/dev-login"),
    ("GET", "/api/v1/auth/session"),
    ("POST", "/api/v1/auth/logout"),
    // Kalendrar
    ("GET", "/api/v1/calendars"),
    ("POST", "/api/v1/calendars"),
    ("PATCH", "/api/v1/calendars/{id}"),
    ("DELETE", "/api/v1/calendars/{id}"),
    // Kalenderhändelser
    ("GET", "/api/v1/calendar-events"),
    ("POST", "/api/v1/calendar-events"),
    ("GET", "/api/v1/calendar-events/{id}"),
    ("PATCH", "/api/v1/calendar-events/{id}"),
    ("POST", "/api/v1/calendar-events/{id}/occurrences"),
    ("POST", "/api/v1/calendar-events/{id}/split"),
    ("DELETE", "/api/v1/calendar-events/{id}"),
    // Mål
    ("GET", "/api/v1/goals"),
    ("GET", "/api/v1/goals/prioritized"),
    ("GET", "/api/v1/goals/graph"),
    ("POST", "/api/v1/goals"),
    ("GET", "/api/v1/goals/{id}"),
    ("PATCH", "/api/v1/goals/{id}"),
    ("DELETE", "/api/v1/goals/{id}"),
    ("POST", "/api/v1/goals/{id}/completions"),
    ("DELETE", "/api/v1/goals/{id}/completions"),
    ("PATCH", "/api/v1/goals/reorder"),
    ("PATCH", "/api/v1/goals/priority/reorder"),
    // Status
    ("GET", "/api/v1/statuses"),
    ("GET", "/api/v1/goals/{id}/status"),
    ("PATCH", "/api/v1/goals/{id}/status"),
    ("GET", "/api/v1/goals/{id}/status/history"),
    // Fasmaskin och verifiering
    ("POST", "/api/v1/goals/{id}/phase"),
    ("POST", "/api/v1/goals/{id}/complete"),
    ("GET", "/api/v1/goals/{id}/verification"),
    ("PUT", "/api/v1/goals/{id}/verification"),
    ("POST", "/api/v1/goals/{id}/evidence"),
    ("POST", "/api/v1/goals/{id}/verification/run"),
    ("POST", "/api/v1/verification-results/{id}/approve"),
    // Relationer
    ("POST", "/api/v1/goals/relations"),
    ("PATCH", "/api/v1/goals/relations"),
    ("PATCH", "/api/v1/goals/relations/weights"),
    ("DELETE", "/api/v1/goals/relations"),
    ("PATCH", "/api/v1/goals/relations/reorder"),
    // Beroenden
    ("POST", "/api/v1/goals/dependencies"),
    ("DELETE", "/api/v1/goals/dependencies"),
    // Platser
    ("GET", "/api/v1/places"),
    ("POST", "/api/v1/places"),
    ("GET", "/api/v1/places/nearby"),
    ("PATCH", "/api/v1/places/{id}"),
    ("DELETE", "/api/v1/places/{id}"),
    ("GET", "/api/v1/goals/{id}/places"),
    ("PUT", "/api/v1/goals/{id}/places"),
    ("GET", "/api/v1/calendar-events/{id}/places"),
    ("PUT", "/api/v1/calendar-events/{id}/places"),
    ("GET", "/api/v1/me/current-place"),
    ("PUT", "/api/v1/me/current-place"),
    // Taggar
    ("GET", "/api/v1/tags"),
    ("GET", "/api/v1/goals/{id}/tags"),
    ("PUT", "/api/v1/goals/{id}/tags"),
    ("DELETE", "/api/v1/goals/{id}/tags/{name}"),
    // Agentkörningar
    ("GET", "/api/v1/agent-run-lifecycle"),
    ("GET", "/api/v1/agent-runs"),
    ("POST", "/api/v1/agent-runs"),
    ("PATCH", "/api/v1/agent-runs/{id}"),
    ("POST", "/api/v1/agent-runs/{id}/cancel"),
    ("GET", "/api/v1/agent-runs/{id}/usage"),
    ("POST", "/api/v1/agent-runs/{id}/usage"),
    ("GET", "/api/v1/agent-usage"),
    ("GET", "/api/v1/agent-budgets"),
    ("POST", "/api/v1/agent-budgets"),
    ("GET", "/api/v1/agent-budgets/{id}/status"),
    ("PATCH", "/api/v1/agent-budgets/{id}"),
    ("DELETE", "/api/v1/agent-budgets/{id}"),
    ("GET", "/api/v1/agent-runs/{id}/events"),
    ("POST", "/api/v1/agent-runs/{id}/events"),
    // Godkännanden
    ("GET", "/api/v1/agent-runs/{id}/approvals"),
    ("POST", "/api/v1/agent-runs/{id}/approvals"),
    ("GET", "/api/v1/agent-approvals"),
    ("POST", "/api/v1/agent-approvals/{id}/decision"),
    // API-nycklar
    ("GET", "/api/v1/api-keys"),
    ("POST", "/api/v1/api-keys"),
    ("DELETE", "/api/v1/api-keys"),
];

/// Bygger en router med alla vägar i tabellen.
///
/// Handlers kopplas på i takt med att de portas från Go. Vägar som ännu
/// inte är portade svarar med 501 i stället för 404 - så att en förbise
/// endpoint är tydlig i loggen och i testerna.
pub fn build() -> Router<AppState> {
    let mut router = Router::new();

    for (method, path) in ROUTES {
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
        ("GET", "/health") => get(crate::routes::handlers::health),

        // Autentisering
        ("GET", "/api/v1/auth/login") => get(crate::routes::auth::login),
        ("GET", "/api/v1/auth/callback") => get(crate::routes::auth::callback),
        ("POST", "/api/v1/auth/dev-login") => post(crate::routes::auth::dev_login),
        ("GET", "/api/v1/auth/session") => get(crate::routes::auth::session),
        ("POST", "/api/v1/auth/logout") => post(crate::routes::auth::logout),

        ("GET", "/api/v1/goals") => get(crate::routes::goals::get_goals),
        ("POST", "/api/v1/goals") => post(crate::routes::goals::create_goal),
        ("GET", "/api/v1/goals/{id}") => get(crate::routes::goals::get_goal),
        ("PATCH", "/api/v1/goals/{id}") => patch(crate::routes::goals::update_goal),
        ("DELETE", "/api/v1/goals/{id}") => delete(crate::routes::goals::delete_goal),
        ("GET", "/api/v1/goals/graph") => get(crate::routes::goals::get_goal_graph),
        ("GET", "/api/v1/goals/prioritized") => get(crate::routes::goals::get_prioritized_goals),
        ("POST", "/api/v1/goals/{id}/completions") => {
            post(crate::routes::goals::complete_goal_occurrence)
        }
        ("DELETE", "/api/v1/goals/{id}/completions") => {
            delete(crate::routes::goals::undo_goal_occurrence)
        }

        ("POST", "/api/v1/goals/relations") => post(crate::routes::relations::create_relation),
        ("PATCH", "/api/v1/goals/relations") => patch(crate::routes::relations::update_relation),
        ("DELETE", "/api/v1/goals/relations") => delete(crate::routes::relations::delete_relation),
        ("PATCH", "/api/v1/goals/relations/weights") => {
            patch(crate::routes::relations::update_relation_weights)
        }
        ("PATCH", "/api/v1/goals/relations/reorder") => {
            patch(crate::routes::relations::reorder_relations)
        }
        ("PATCH", "/api/v1/goals/reorder") => patch(crate::routes::relations::reorder_goals),
        ("PATCH", "/api/v1/goals/priority/reorder") => {
            patch(crate::routes::relations::reorder_priority_goals)
        }
        ("POST", "/api/v1/goals/dependencies") => post(crate::routes::relations::create_dependency),
        ("DELETE", "/api/v1/goals/dependencies") => {
            delete(crate::routes::relations::delete_dependency)
        }

        ("GET", "/api/v1/calendar-events") => get(crate::routes::calendar_events::list),
        ("POST", "/api/v1/calendar-events") => post(crate::routes::calendar_events::create),
        ("GET", "/api/v1/calendar-events/{id}") => get(crate::routes::calendar_events::get),
        ("PATCH", "/api/v1/calendar-events/{id}") => patch(crate::routes::calendar_events::update),
        ("DELETE", "/api/v1/calendar-events/{id}") => {
            delete(crate::routes::calendar_events::delete)
        }
        ("POST", "/api/v1/calendar-events/{id}/occurrences") => {
            post(crate::routes::calendar_events::create_occurrence_override)
        }
        ("POST", "/api/v1/calendar-events/{id}/split") => {
            post(crate::routes::calendar_events::split_series)
        }

        ("GET", "/api/v1/calendars") => get(crate::routes::calendars::list),
        ("POST", "/api/v1/calendars") => post(crate::routes::calendars::create),
        ("PATCH", "/api/v1/calendars/{id}") => patch(crate::routes::calendars::update),
        ("DELETE", "/api/v1/calendars/{id}") => delete(crate::routes::calendars::delete),

        ("GET", "/api/v1/places") => get(crate::routes::places::list),
        ("POST", "/api/v1/places") => post(crate::routes::places::create),
        ("GET", "/api/v1/places/nearby") => get(crate::routes::places::nearby),
        ("PATCH", "/api/v1/places/{id}") => patch(crate::routes::places::update),
        ("DELETE", "/api/v1/places/{id}") => delete(crate::routes::places::delete),
        ("GET", "/api/v1/goals/{id}/places") => get(crate::routes::places::get_goal_places),
        ("PUT", "/api/v1/goals/{id}/places") => put(crate::routes::places::set_goal_places),
        ("GET", "/api/v1/calendar-events/{id}/places") => {
            get(crate::routes::places::get_event_places)
        }
        ("PUT", "/api/v1/calendar-events/{id}/places") => {
            put(crate::routes::places::set_event_places)
        }
        ("GET", "/api/v1/me/current-place") => get(crate::routes::places::get_current_place),
        ("PUT", "/api/v1/me/current-place") => put(crate::routes::places::set_current_place),

        ("GET", "/api/v1/tags") => get(crate::routes::tags::list),
        ("GET", "/api/v1/goals/{id}/tags") => get(crate::routes::tags::get_goal_tags),
        ("PUT", "/api/v1/goals/{id}/tags") => put(crate::routes::tags::set_goal_tags),
        ("DELETE", "/api/v1/goals/{id}/tags/{name}") => {
            delete(crate::routes::tags::delete_goal_tag)
        }

        ("GET", "/api/v1/statuses") => get(crate::routes::statuses::get_statuses),
        ("GET", "/api/v1/goals/{id}/status") => get(crate::routes::statuses::get_goal_status),
        ("PATCH", "/api/v1/goals/{id}/status") => {
            patch(crate::routes::statuses::update_goal_status)
        }
        ("GET", "/api/v1/goals/{id}/status/history") => {
            get(crate::routes::statuses::get_status_history)
        }

        // Fasmaskin och verifiering - inte portad till Rust ännu.
        // Go-tjänsten är referensimplementationen; se PR:n för fasmaskinen.
        ("POST", "/api/v1/goals/{id}/phase") => post(crate::routes::handlers::not_implemented),
        ("POST", "/api/v1/goals/{id}/complete") => post(crate::routes::handlers::not_implemented),
        ("GET", "/api/v1/goals/{id}/verification") => get(crate::routes::handlers::not_implemented),
        ("PUT", "/api/v1/goals/{id}/verification") => put(crate::routes::handlers::not_implemented),
        ("POST", "/api/v1/goals/{id}/evidence") => post(crate::routes::handlers::not_implemented),
        ("POST", "/api/v1/goals/{id}/verification/run") => {
            post(crate::routes::handlers::not_implemented)
        }
        ("POST", "/api/v1/verification-results/{id}/approve") => {
            post(crate::routes::handlers::not_implemented)
        }

        // Agentkörningar
        ("GET", "/api/v1/agent-run-lifecycle") => get(crate::routes::agent_runs::lifecycle),
        ("GET", "/api/v1/agent-runs") => get(crate::routes::agent_runs::list),
        ("POST", "/api/v1/agent-runs") => post(crate::routes::agent_runs::create),
        ("PATCH", "/api/v1/agent-runs/{id}") => patch(crate::routes::agent_runs::update),
        ("POST", "/api/v1/agent-runs/{id}/cancel") => post(crate::routes::agent_runs::cancel),
        ("GET", "/api/v1/agent-runs/{id}/events") => get(crate::routes::agent_runs::list_events),
        ("POST", "/api/v1/agent-runs/{id}/events") => post(crate::routes::agent_runs::create_event),
        ("GET", "/api/v1/agent-runs/{id}/usage") => get(crate::routes::agent_usage::get_usage),
        ("POST", "/api/v1/agent-runs/{id}/usage") => post(crate::routes::agent_usage::record_usage),
        ("GET", "/api/v1/agent-usage") => get(crate::routes::agent_usage::list_usage),
        ("GET", "/api/v1/agent-budgets") => get(crate::routes::agent_usage::list_budgets),
        ("POST", "/api/v1/agent-budgets") => post(crate::routes::agent_usage::create_budget),
        ("GET", "/api/v1/agent-budgets/{id}/status") => {
            get(crate::routes::agent_usage::budget_status)
        }
        ("PATCH", "/api/v1/agent-budgets/{id}") => patch(crate::routes::agent_usage::update_budget),
        ("DELETE", "/api/v1/agent-budgets/{id}") => {
            delete(crate::routes::agent_usage::delete_budget)
        }

        // Godkännanden
        ("GET", "/api/v1/agent-runs/{id}/approvals") => {
            get(crate::routes::agent_approvals::list_for_run)
        }
        ("POST", "/api/v1/agent-runs/{id}/approvals") => {
            post(crate::routes::agent_approvals::create)
        }
        ("GET", "/api/v1/agent-approvals") => get(crate::routes::agent_approvals::list_pending),
        ("POST", "/api/v1/agent-approvals/{id}/decision") => {
            post(crate::routes::agent_approvals::decide)
        }

        // API-nycklar
        ("GET", "/api/v1/api-keys") => get(crate::routes::api_keys::list),
        ("POST", "/api/v1/api-keys") => post(crate::routes::api_keys::create),
        ("DELETE", "/api/v1/api-keys") => delete(crate::routes::api_keys::delete),

        ("GET", _) => get(crate::routes::handlers::not_implemented),
        ("POST", _) => post(crate::routes::handlers::not_implemented),
        ("PUT", _) => put(crate::routes::handlers::not_implemented),
        ("PATCH", _) => patch(crate::routes::handlers::not_implemented),
        ("DELETE", _) => delete(crate::routes::handlers::not_implemented),
        other => panic!("okänd HTTP-metod i route-tabellen: {other:?}"),
    }
}

#[cfg(test)]
mod tests {
    use std::collections::BTreeSet;
    use std::fs;
    use std::path::PathBuf;

    use super::ROUTES;

    /// Sökvägen till Go-tjänstens main.go.
    fn go_main() -> PathBuf {
        PathBuf::from(env!("CARGO_MANIFEST_DIR")).join("../../../api/cmd/api/main.go")
    }

    /// Route-registreringarna i Go-tjänsten.
    fn go_routes() -> BTreeSet<(String, String)> {
        let kalla = fs::read_to_string(go_main())
            .expect("går att läsa Go-tjänstens main.go - annars fungerar paritetstestet inte");
        plan_parity::go_routes(&kalla)
    }

    /// Rust-tabellens routes.
    fn rust_routes() -> Vec<(String, String)> {
        ROUTES
            .iter()
            .map(|(metod, sokvag)| (metod.to_string(), sokvag.to_string()))
            .collect()
    }

    /// Varje route i Go-tjänsten ska finnas i Rust-tabellen.
    ///
    /// Dokumentationsvägen (/docs) hoppas över - den serveras av Scalar i
    /// båda implementationerna.
    #[test]
    fn go_routes_finns_i_rust_tabellen() {
        let (saknas, _) = plan_parity::jämför(&go_routes(), &rust_routes());

        assert!(
            saknas.is_empty(),
            "följande routes finns i Go-tjänsten men inte i Rust-tabellen: {}",
            saknas.join(", ")
        );
    }

    /// Varje route i tabellen ska ha en handler.
    ///
    /// En route utan handler svarar med 501 - det fångar en endpoint som
    /// glömts bort när den portades.
    #[test]
    fn varje_route_i_tabellen_har_en_handler() {
        let kalla = include_str!("table.rs");
        let funktion = kalla
            .split("fn route_for(method: &str, path: &str) -> MethodRouter<AppState> {")
            .nth(1)
            .expect("route_for finns i filen");
        let funktion = funktion
            .split("\n    }\n")
            .next()
            .expect("route_for har en avslutande parentes");

        // Varje arm börjar med (metod, "sökväg") => ...
        let mut handlagda = BTreeSet::new();
        for rad in funktion.lines() {
            let rad = rad.trim();
            let Some(rest) = rad.strip_prefix("(\"") else {
                continue;
            };
            let Some((metod, rest)) = rest.split_once("\", \"") else {
                continue;
            };
            let Some(sokvag) = rest.split_once("\"") else {
                continue;
            };
            handlagda.insert((metod.to_string(), sokvag.0.to_string()));
        }

        let tabellen: BTreeSet<(String, String)> = rust_routes().into_iter().collect();
        let utan_handler: Vec<String> = tabellen
            .difference(&handlagda)
            .map(|(metod, sokvag)| format!("{metod} {sokvag}"))
            .collect();

        assert!(
            utan_handler.is_empty(),
            "följande routes saknar handler: {}",
            utan_handler.join(", ")
        );
    }

    /// Rust-tabellen ska inte ha fler routes än Go-tjänsten.
    #[test]
    fn rust_tabellen_har_inga_extra_routes() {
        let (_, extra) = plan_parity::jämför(&go_routes(), &rust_routes());

        assert!(
            extra.is_empty(),
            "följande routes finns i Rust-tabellen men inte i Go-tjänsten: {}",
            extra.join(", ")
        );
    }

    /// Go-tjänsten ska registrera routes - annars är paritetstestet tomt.
    #[test]
    fn go_routes_finns_att_jamfora_mot() {
        assert!(
            go_routes().len() > 50,
            "route-tabellen i Go-tjänsten verkar ofullständig - kontrollera parsningen"
        );
    }
}
