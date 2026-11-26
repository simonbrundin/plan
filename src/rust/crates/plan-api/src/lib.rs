//! Plan API - REST-API för Plan.
//!
//! Detta är Rust-ersättningen för Go-tjänsten i `src/api`. Gränssnittet
//! (Nuxt) och agenten pratar med API:t över HTTP, så kontraktet - vägar,
//! metoder, statuskoder och JSON-format - måste vara identiskt.

pub mod auth;
pub mod config;
pub mod contracts;
pub mod db;
pub mod docs;
pub mod error;
pub mod models;
pub mod openapi;
pub mod pricing;
pub mod repository;
pub mod routes;
