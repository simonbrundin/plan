//! Router för agent-API:et.

pub mod api_keys;
pub mod approvals;
pub mod chat;
pub mod conversations;
pub mod docs;
pub mod health;
pub mod json_fel;
pub mod readme;
pub mod table;
pub mod tools;
pub mod workflow;

use axum::Router;

use crate::state::AppState;

/// Bygger agent-API:ts router med tillstånd.
pub fn router(state: AppState) -> Router {
    Router::new().merge(table::build()).with_state(state)
}
