//! Agenttjänsten för Plan.
//!
//! Rust-ersättning för Go-tjänsten i `src/agent`. Modulerna är delade
//! mellan `agent-api` (HTTP-tjänsten) och `agent-worker` (workern som
//! utför körningarna).

pub mod approvals;
pub mod checkpoint;
pub mod contract;
pub mod core;
pub mod db;
pub mod error;
pub mod handlers;
pub mod messages;
pub mod orchestrator;
pub mod planclient;
pub mod provider;
pub mod routes;
pub mod state;
pub mod state_store;
pub mod tools;
