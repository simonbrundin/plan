//! Delat tillstånd för alla handlers.

use sqlx::postgres::PgPool;

use crate::config::Config;

#[derive(Clone)]
pub struct AppState {
    /// Databaspoolen. `None` betyder mock-läge - API:t startar men
    /// endpoints som behöver data svarar med ett fel.
    #[allow(dead_code)]
    pub pool: Option<PgPool>,
    /// Konfigurationen API:t startade med.
    pub config: Config,
}
