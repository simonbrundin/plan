//! Agenttjänsten för Plan.
//!
//! Rust-ersättning för Go-tjänsten i `src/agent`. Modulerna är delade
//! mellan `agent-api` (HTTP-tjänsten) och `agent-worker` (workern som
//! utför körningarna).

pub mod approvals;
pub mod checkpoint;
pub mod coder;
pub mod contract;
pub mod core;
pub mod db;
pub mod definer;
pub mod error;
pub mod executor;
pub mod go_num;
pub mod go_time;
pub mod goaldefinition;
pub mod handlers;
pub mod jobcontroller;
pub mod jobrunner;
pub mod k8s;
pub mod learnings;
pub mod messages;
pub mod orchestrator;
pub mod planclient;
pub mod profile;
pub mod proposer;
pub mod provider;
pub mod routes;
pub mod runusage;
pub mod scheduler;
pub mod state;
pub mod state_store;
pub mod tools;
pub mod worker;
pub mod workspace;
