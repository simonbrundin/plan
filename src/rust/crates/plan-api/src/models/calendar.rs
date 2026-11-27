//! Kalendrar.
//!
//! Port av `src/api/internal/models/calendar.go`.

use serde::{Deserialize, Serialize};
use utoipa::ToSchema;

/// Färgen en kalender får om ingen annan anges.
pub const DEFAULT_CALENDAR_COLOR: &str = "#6366f1";

/// En användares egen kalender.
///
/// Användaren kan skapa hur många kalendrar som helst - till exempel en
/// jobb- och en fritidskalender - och varje kalender har en egen färg.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, ToSchema)]
pub struct Calendar {
    /// Kalenderns id.
    pub id: i64,
    /// Ägarens id.
    pub user_id: i64,
    /// Kalenderns namn.
    pub name: String,
    /// Kalenderns färg som hexkod.
    pub color: String,
    /// Ikonen som visas - tom sträng om ingen ikon satts.
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub icon: String,
    /// Om kalendern är användarens standardkalender.
    ///
    /// Nya händelser hamnar där om ingen kalender väljs uttryckligen.
    pub is_default: bool,
    /// Kalenderns placering i listan.
    pub sort_order: i32,
    /// När kalendern skapades.
    #[serde(serialize_with = "crate::models::go_time::serialize")]
    pub created_at: chrono::DateTime<chrono::Utc>,
    /// När kalendern senast ändrades.
    #[serde(serialize_with = "crate::models::go_time::serialize")]
    pub updated_at: chrono::DateTime<chrono::Utc>,
}

/// Request-body för att skapa en kalender.
#[derive(Debug, Clone, Deserialize, ToSchema)]
pub struct CreateCalendarRequest {
    /// Kalenderns namn - krävs.
    pub name: String,
    /// Färgen som hexkod.
    #[serde(default)]
    pub color: Option<String>,
    /// Ikonen som visas.
    #[serde(default)]
    pub icon: Option<String>,
    /// Placeringen i listan.
    #[serde(default)]
    pub sort_order: Option<i32>,
    /// Sätts när kalendern ska vara standardkalender.
    #[serde(default)]
    pub is_default: Option<bool>,
}

/// Request-body för att uppdatera en kalender.
///
/// Namnet sätts när det skickas - ett tomt värde är ogiltigt.
#[derive(Debug, Clone, Deserialize, ToSchema)]
pub struct UpdateCalendarRequest {
    /// Nytt namn.
    #[serde(default)]
    pub name: Option<String>,
    /// Färgen.
    ///
    /// Har samma tre-tillståndssemantik som på händelser: inget betyder
    /// "behåll nuvarande värde", en tom sträng "töm" och ett värde "sätt".
    #[serde(default)]
    pub color: Option<String>,
    /// Ikonen - samma tre-tillståndssemantik som färgen.
    #[serde(default)]
    pub icon: Option<String>,
    /// Ny placering i listan.
    #[serde(default)]
    pub sort_order: Option<i32>,
    /// Standardkalendern.
    ///
    /// Inget betyder "behåll nuvarande värde", sant "gör denna till
    /// standardkalender" och falskt "avmarkera den som standardkalender".
    #[serde(default)]
    pub is_default: Option<bool>,
}
