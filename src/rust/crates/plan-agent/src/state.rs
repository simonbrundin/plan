//! Delat tillstånd för agent-API:et.

use crate::approvals::Store;
use crate::db::AgentDb;
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
    /// Agentens egen databas. `None` betyder att den inte är ansluten -
    /// då svarar samtalen och nycklarna med ett internt fel.
    pub db: Option<AgentDb>,
    /// Store:n med förslag som väntar på ett beslut.
    pub approval_store: Option<Store>,
}
