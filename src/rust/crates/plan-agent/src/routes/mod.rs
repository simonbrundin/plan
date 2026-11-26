//! Router för agent-API:et.

pub mod table;

use axum::Router;

use crate::state::AppState;

/// Bygger agent-API:ts router med tillstånd.
pub fn router(state: AppState) -> Router {
    Router::new().merge(table::build()).with_state(state)
}
