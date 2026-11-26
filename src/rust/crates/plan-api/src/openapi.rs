//! OpenAPI-specen.
//!
//! Specen är *code-first*: den genereras ur `#[utoipa::path]`-annotationerna
//! på handlerna, så dokumentationen inte kan avvika från implementationen.
//! Go-tjänsten använde swaggo med en genererad swagger.json - det är
//! samma mål, men utan fil att glömma regenerera.

use utoipa::openapi::server::Server;
use utoipa::OpenApi;

use crate::routes::{goals, handlers, statuses, tags};

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
        (name = "statuses", description = "Statusar och statushistorik")
    ),
    paths(
        handlers::health,
        statuses::get_statuses,
        statuses::get_goal_status,
        statuses::update_goal_status,
        statuses::get_status_history,
        goals::get_goals,
        goals::get_goal,
        goals::get_goal_graph,
        goals::get_prioritized_goals,
        goals::create_goal,
        goals::update_goal,
        goals::delete_goal,
        tags::list,
        tags::get_goal_tags,
        tags::set_goal_tags,
        tags::delete_goal_tag,
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
