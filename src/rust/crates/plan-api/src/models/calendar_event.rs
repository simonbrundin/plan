//! Kalenderhändelser.
//!
//! Port av `src/api/internal/models/calendar_event.go`.

use serde::{Deserialize, Serialize};
use utoipa::ToSchema;

use crate::models::place::Place;

/// En händelse i kalendern.
///
/// Återkommande serier lagras som en händelse med en `recurrence_rule`.
/// När serien expanderas skapas en kopia per förekomst, med ett eget
/// `occurrence_id`.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, ToSchema)]
pub struct CalendarEvent {
    /// Händelsens id.
    pub id: i64,
    /// Id för en enskild förekomst i en återkommande serie.
    ///
    /// Sätts när serien expanderas - då har förekomsten både ett
    /// `occurrence_id` och ett `original_start_at`.
    #[serde(skip_serializing_if = "String::is_empty")]
    pub occurrence_id: String,
    /// Ägarens id.
    pub user_id: i64,
    /// Målet händelsen hör till, om något.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub goal_id: Option<i64>,
    /// Kalendern händelsen ligger i.
    ///
    /// Inget betyder att händelsen inte tillhör någon kalender.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub calendar_id: Option<i64>,
    /// Händelsens titel.
    pub title: String,
    /// Beskrivningen.
    pub description: String,
    /// Ikonen - en emoji eller ett Iconify-namn.
    #[serde(skip_serializing_if = "String::is_empty")]
    pub icon: String,
    /// När händelsen börjar.
    pub start_at: chrono::DateTime<chrono::Utc>,
    /// När händelsen slutar.
    pub end_at: chrono::DateTime<chrono::Utc>,
    /// Tidszonen händelsen ligger i.
    pub time_zone: String,
    /// Om händelsen pågår hela dagen.
    pub all_day: bool,
    /// Status - "confirmed", "tentative" eller "cancelled".
    pub status: String,
    /// Upprepningsregeln, om händelsen är återkommande.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub recurrence_rule: Option<String>,
    /// Förälderserien, om händelsen är en enskild förekomst.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub parent_event_id: Option<i64>,
    /// Förekomstens ursprungliga start, om händelsen är en override.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub original_start_at: Option<chrono::DateTime<chrono::Utc>>,
    /// När händelsen skapades.
    pub created_at: chrono::DateTime<chrono::Utc>,
    /// När händelsen senast ändrades.
    pub updated_at: chrono::DateTime<chrono::Utc>,
    /// Platser där händelsen sker.
    ///
    /// Skickas alltid ut. `None` blir "null" - det händer för de svar där
    /// platser inte fylls i, som skapade förekomster och delade serier.
    pub places: Option<Vec<Place>>,
}

/// Request-body för att skapa en händelse.
#[derive(Debug, Clone, Deserialize, ToSchema)]
pub struct CreateCalendarEventRequest {
    /// Händelsens titel - krävs.
    ///
    /// `None` betyder att titeln saknas, vilket är ett fel - likaså en tom
    /// sträng.
    pub title: Option<String>,
    /// Beskrivningen.
    #[serde(default)]
    pub description: Option<String>,
    /// Ikonen.
    #[serde(default)]
    pub icon: Option<String>,
    /// Starttid - krävs.
    pub start_at: Option<String>,
    /// Sluttid - krävs.
    pub end_at: Option<String>,
    /// Tidszonen.
    #[serde(default)]
    pub time_zone: Option<String>,
    /// Om händelsen pågår hela dagen.
    #[serde(default)]
    pub all_day: Option<bool>,
    /// Status.
    #[serde(default)]
    pub status: Option<String>,
    /// Upprepningsregeln.
    #[serde(default)]
    pub recurrence_rule: Option<String>,
    /// Målet händelsen hör till.
    #[serde(default)]
    pub goal_id: Option<i64>,
    /// Kalendern händelsen ska ligga i.
    ///
    /// Inget lämnar händelsen i standardkalendern, och noll betyder
    /// uttryckligen "ingen kalender".
    #[serde(default)]
    pub calendar_id: Option<i64>,
}

/// Request-body för att uppdatera en händelse.
///
/// Fält som skickas ändras, fält som utelämnas lämnas oförändrade. Ikonen
/// har tre tillstånd: inget betyder "behåll nuvarande ikon", en tom sträng
/// "töm" och ett värde "sätt".
#[derive(Debug, Clone, Deserialize, ToSchema)]
pub struct UpdateCalendarEventRequest {
    /// Ny titel.
    #[serde(default)]
    pub title: Option<String>,
    /// Ny beskrivning.
    #[serde(default)]
    pub description: Option<String>,
    /// Ny ikon - tre tillstånd.
    #[serde(default)]
    pub icon: Option<String>,
    /// Ny starttid.
    #[serde(default)]
    pub start_at: Option<String>,
    /// Ny sluttid.
    #[serde(default)]
    pub end_at: Option<String>,
    /// Ny tidszon.
    #[serde(default)]
    pub time_zone: Option<String>,
    /// Om händelsen pågår hela dagen.
    #[serde(default)]
    pub all_day: Option<bool>,
    /// Ny status.
    #[serde(default)]
    pub status: Option<String>,
    /// Ny upprepningsregel.
    #[serde(default)]
    pub recurrence_rule: Option<String>,
    /// Nytt mål.
    #[serde(default)]
    pub goal_id: Option<i64>,
    /// Ny kalender.
    ///
    /// Inget behåller nuvarande kalender och noll tar bort den.
    #[serde(default)]
    pub calendar_id: Option<i64>,
}

/// Request-body för att skapa en override av en förekomst.
///
/// Fält som inte skickas ärvs från förälderserien.
#[derive(Debug, Clone, Deserialize, ToSchema)]
pub struct CreateCalendarOccurrenceRequest {
    /// Förekomstens ursprungliga start - krävs.
    pub original_start_at: Option<String>,
    /// Ny starttid - krävs.
    pub start_at: Option<String>,
    /// Ny sluttid - krävs.
    pub end_at: Option<String>,
    /// Ny titel.
    #[serde(default)]
    pub title: Option<String>,
    /// Ny beskrivning.
    #[serde(default)]
    pub description: Option<String>,
    /// Ny ikon - tre tillstånd.
    #[serde(default)]
    pub icon: Option<String>,
    /// Ny tidszon.
    #[serde(default)]
    pub time_zone: Option<String>,
    /// Om förekomsten pågår hela dagen.
    #[serde(default)]
    pub all_day: Option<bool>,
    /// Ny status.
    #[serde(default)]
    pub status: Option<String>,
    /// Nytt mål.
    #[serde(default)]
    pub goal_id: Option<i64>,
    /// Ny kalender.
    #[serde(default)]
    pub calendar_id: Option<i64>,
}

/// Request-body för att dela en återkommande serie.
///
/// Den gamla serien avslutas strax före `original_start_at` och en ny serie
/// fortsätter därifrån.
#[derive(Debug, Clone, Deserialize, ToSchema)]
pub struct SplitCalendarEventRequest {
    /// Förekomsten som ska bli början på den nya serien - krävs.
    pub original_start_at: Option<String>,
    /// Den nya seriens starttid - krävs.
    pub start_at: Option<String>,
    /// Den nya seriens sluttid - krävs.
    pub end_at: Option<String>,
    /// Den nya seriens upprepningsregel - krävs.
    pub recurrence_rule: Option<String>,
    /// Ny titel.
    #[serde(default)]
    pub title: Option<String>,
    /// Ny beskrivning.
    #[serde(default)]
    pub description: Option<String>,
    /// Ny ikon - tre tillstånd.
    #[serde(default)]
    pub icon: Option<String>,
    /// Ny tidszon.
    #[serde(default)]
    pub time_zone: Option<String>,
    /// Om serien pågår hela dagen.
    #[serde(default)]
    pub all_day: Option<bool>,
    /// Nytt mål.
    #[serde(default)]
    pub goal_id: Option<i64>,
    /// Ny kalender.
    #[serde(default)]
    pub calendar_id: Option<i64>,
}
