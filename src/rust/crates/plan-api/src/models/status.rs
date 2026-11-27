//! Statusar.
//!
//! Port av `models.Status`, `models.StatusUpdate` och
//! `models.UpdateStatusRequest` i `src/api/internal/models/goal.go`.

use serde::{Deserialize, Serialize};
use utoipa::ToSchema;

/// En status som ett mål kan ha.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, ToSchema)]
pub struct Status {
    /// Statusens id.
    pub id: i64,
    /// Statusens namn, t.ex. "define".
    pub name: String,
    /// Etiketten som visas, t.ex. "Defining".
    pub label: String,
    /// Skapad-tid - skickas inte ut om den är tom, som i Go.
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub created: String,
}

/// En statusändring i historiken.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, ToSchema)]
pub struct StatusUpdate {
    /// Radens id.
    pub id: i64,
    /// Målets id.
    pub goal_id: i64,
    /// Statusen målet hade innan - saknas om målet inte hade någon.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub from_status_id: Option<i64>,
    /// Statusen målet har nu.
    pub to_status_id: i64,
    /// När ändringen skedde.
    #[serde(serialize_with = "crate::models::go_time::serialize")]
    pub changed_at: chrono::DateTime<chrono::Utc>,
    /// Namnet på den gamla statusen - tom sträng om den saknas.
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub from_status_name: String,
    /// Namnet på den nya statusen.
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub to_status_name: String,
}

/// Request-body för att byta status.
#[derive(Debug, Clone, Deserialize, ToSchema)]
pub struct UpdateStatusRequest {
    /// Den nya statusens id.
    pub status_id: i64,
}
