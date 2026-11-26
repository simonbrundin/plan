//! Delat tillstånd för agent-API:et.

use crate::planclient::PlanClient;

/// Agent-API:ets tillstånd.
#[derive(Clone)]
pub struct AppState {
    /// Klienten mot Plan API. `None` betyder att agenten inte är
    /// konfigurerad för Plan - då startas bara API:t.
    pub plan_client: Option<PlanClient>,
    /// Agentens egen nyckel - klienten som anropar agent-API:t måste
    /// skicka den.
    pub api_key: Option<String>,
}
