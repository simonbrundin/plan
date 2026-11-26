//! OpenAPI-specen.
//!
//! Specen är *code-first*: den genereras ur `#[utoipa::path]`-annotationerna
//! på handlerna, så dokumentationen inte kan avvika från implementationen.
//! Go-tjänsten använde swaggo med en genererad swagger.json - det är
//! samma mål, men utan fil att glömma regenerera.

use utoipa::openapi::server::Server;
use utoipa::OpenApi;

use crate::routes::{
    agent_approvals, agent_runs, agent_usage, api_keys, auth, calendar_events, calendars, goals,
    handlers, places, relations, statuses, tags,
};

/// Grund för OpenAPI-specen.
#[derive(OpenApi)]
#[openapi(
    info(
        title = "Goals Management API",
        version = "1.0",
        description = "API för att hantera mål med förälder-barn-relationer och beroenden. \
                       Specen (Swagger 3.1) genereras ur koden med utoipa.",
        license(name = "MIT")
    ),
    tags(
        (name = "health", description = "Hälsokontroll"),
        (name = "auth", description = "Autentisering och sessioner"),
        (name = "goals", description = "Mål"),
        (name = "calendars", description = "Kalendrar och kalenderhändelser"),
        (name = "places", description = "Platser"),
        (name = "tags", description = "Taggar"),
        (name = "agent", description = "Agentkörningar, godkännanden och budgets"),
        (name = "api-keys", description = "API-nycklar"),
        (name = "statuses", description = "Statusar och statushistorik"),
        (name = "relations", description = "Förälder-barn-relationer"),
        (name = "dependencies", description = "Beroenden mellan mål")
    ),
    paths(
        handlers::health,
        auth::login,
        auth::callback,
        auth::dev_login,
        auth::session,
        auth::logout,
        statuses::get_statuses,
        statuses::get_goal_status,
        statuses::update_goal_status,
        statuses::get_status_history,
        goals::get_goals,
        goals::get_goal,
        goals::get_goal_graph,
        goals::get_prioritized_goals,
        goals::complete_goal_occurrence,
        goals::undo_goal_occurrence,
        goals::create_goal,
        goals::update_goal,
        goals::delete_goal,
        tags::list,
        tags::get_goal_tags,
        tags::set_goal_tags,
        tags::delete_goal_tag,
        relations::create_relation,
        relations::update_relation,
        relations::update_relation_weights,
        relations::delete_relation,
        relations::reorder_relations,
        relations::reorder_goals,
        relations::reorder_priority_goals,
        relations::create_dependency,
        relations::delete_dependency,
        places::list,
        places::create,
        places::update,
        places::delete,
        places::nearby,
        places::get_goal_places,
        places::set_goal_places,
        places::get_event_places,
        places::set_event_places,
        places::get_current_place,
        places::set_current_place,
        calendars::list,
        calendars::create,
        calendars::update,
        calendars::delete,
        calendar_events::list,
        calendar_events::get,
        calendar_events::create,
        calendar_events::update,
        calendar_events::delete,
        calendar_events::create_occurrence_override,
        calendar_events::split_series,
        agent_runs::lifecycle,
        agent_runs::list,
        agent_runs::create,
        agent_runs::update,
        agent_runs::cancel,
        agent_runs::list_events,
        agent_runs::create_event,
        agent_usage::get_usage,
        agent_usage::record_usage,
        agent_usage::list_usage,
        agent_usage::list_budgets,
        agent_usage::create_budget,
        agent_usage::budget_status,
        agent_usage::update_budget,
        agent_usage::delete_budget,
        agent_approvals::list_for_run,
        agent_approvals::create,
        agent_approvals::list_pending,
        agent_approvals::decide,
        api_keys::list,
        api_keys::create,
        api_keys::delete,
    )
)]
pub struct ApiDoc;

/// Bygger OpenAPI-specen.
///
/// `API_HOST` sätts i `servers` när den är satt - samma semantik som
/// Go-tjänsten (`API_HOST` -> `host` i swagger-specen).
pub fn openapi(api_host: Option<&str>) -> utoipa::openapi::OpenApi {
    let mut spec = <ApiDoc as utoipa::OpenApi>::openapi();

    if let Some(host) = api_host.map(str::trim).filter(|host| !host.is_empty()) {
        let url = if host.starts_with("http://") || host.starts_with("https://") {
            host.to_string()
        } else {
            format!("http://{host}")
        };
        spec.servers = Some(vec![Server::new(url)]);
    }

    spec
}
