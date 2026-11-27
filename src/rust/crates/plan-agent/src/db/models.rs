//! Agentens modeller.
//!
//! JSON-fälten matchar Go-agentens `internal/api/db` - samma namn, samma
//! typ. `omitempty`-fälten utelämnas när de är tomma, precis som i Go.

use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use sqlx::FromRow;
use uuid::Uuid;

/// Ett samtal.
#[derive(Debug, Clone, Serialize, Deserialize, FromRow)]
pub struct Conversation {
    /// Samtalets id.
    pub id: Uuid,
    /// När samtalet skapades.
    pub created_at: DateTime<Utc>,
    /// När samtalet senast ändrades.
    pub updated_at: DateTime<Utc>,
    /// Samtalets titel.
    pub title: String,
    /// Modellen som används.
    pub model: String,
    /// Leverantören som används.
    pub provider: String,
    /// Målet i Plan som samtalet gäller.
    ///
    /// Utan det hittas samtalet inte tillbaka från målsidan.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub goal_id: Option<i64>,
}

/// Ett meddelande i ett samtal.
#[derive(Debug, Clone, Serialize, Deserialize, FromRow)]
pub struct Message {
    /// Meddelandets id.
    pub id: Uuid,
    /// Samtalet det tillhör.
    pub conversation_id: Uuid,
    /// Rollen - "user", "assistant", "tool_request" med flera.
    pub role: String,
    /// Meddelandets innehåll.
    pub content: String,
    /// Verktygets namn, om det är ett verktygsanrop.
    #[serde(skip_serializing_if = "String::is_empty")]
    pub tool_name: String,
    /// Verktygets argument - JSON som text, som i Go.
    #[serde(skip_serializing_if = "String::is_empty")]
    pub tool_args: String,
    /// Verktygets resultat.
    #[serde(skip_serializing_if = "String::is_empty")]
    pub tool_result: String,
    /// Om meddelandet är ett fel.
    #[serde(skip_serializing_if = "is_false")]
    pub is_error: bool,
    /// Modellens resonemang.
    #[serde(skip_serializing_if = "String::is_empty")]
    pub thinking: String,
    /// Antalet tokens som användes.
    pub tokens_used: i32,
    /// Latensen i millisekunder.
    pub latency_ms: i32,
    /// När meddelandet skapades.
    pub created_at: DateTime<Utc>,
}

/// Sant om värdet är falskt - då utelämnas fältet i JSON.
fn is_false(värde: &bool) -> bool {
    !*värde
}

/// En API-nyckel för en leverantör.
#[derive(Debug, Clone, Serialize, Deserialize, FromRow)]
pub struct ApiKey {
    /// Nyckelns id.
    pub id: i32,
    /// Leverantören nyckeln gäller.
    pub provider: String,
    /// Nyckeln.
    pub api_key: String,
    /// Modellen nyckeln gäller.
    pub model: String,
    /// När nyckeln skapades.
    pub created_at: DateTime<Utc>,
    /// När nyckeln senast ändrades.
    pub updated_at: DateTime<Utc>,
}

/// En API-nyckel utan själva nyckeln - så visas i listan.
#[derive(Debug, Clone, Serialize)]
pub struct MaskedApiKey {
    /// Nyckelns id.
    pub id: i32,
    /// Leverantören nyckeln gäller.
    pub provider: String,
    /// Om en nyckel finns sparad.
    pub has_key: bool,
    /// Modellen nyckeln gäller.
    pub model: String,
}
