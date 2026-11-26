//! Agenttjänsten för Plan.
//!
//! Rust-ersättning för Go-tjänsten i `src/agent`. Modulerna är delade
//! mellan `agent-api` (HTTP-tjänsten) och `agent-worker` (workern som
//! utför körningarna).

pub mod approvals;
pub mod contract;
pub mod db;
pub mod error;
pub mod handlers;
pub mod orchestrator;
pub mod planclient;
pub mod routes;
pub mod state;
pub mod tools;
