//! Meddelanden i agentens samtal.
//!
//! Go-agentens `internal/core/messages/types.go`. Meddelandena bäddar
//! innehåll i block - text, bilder, verktygsanrop och verktygsresultat.

use serde::{Deserialize, Serialize};

/// Rollen i ett meddelande.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum Role {
    /// Användaren.
    #[serde(rename = "user")]
    User,
    /// Agenten.
    #[serde(rename = "assistant")]
    Assistant,
    /// Resultatet av ett verktygsanrop.
    #[serde(rename = "toolResult")]
    ToolResult,
    /// Systemmeddelandet.
    #[serde(rename = "system")]
    System,
}

impl Role {
    /// Rollen som text - så som Go skriver den.
    pub fn as_str(&self) -> &'static str {
        match self {
            Role::User => "user",
            Role::Assistant => "assistant",
            Role::ToolResult => "toolResult",
            Role::System => "system",
        }
    }
}

/// Ett innehållsblock.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "type", rename_all = "camelCase")]
pub enum ContentBlock {
    /// Text.
    Text {
        /// Texten.
        text: String,
    },
    /// En bild.
    Image {
        /// Bildens adress.
        #[serde(skip_serializing_if = "String::is_empty", default)]
        url: String,
        /// Bilden som base64.
        #[serde(skip_serializing_if = "String::is_empty", default)]
        base64: String,
        /// Alternativtext.
        #[serde(skip_serializing_if = "String::is_empty", default)]
        alt: String,
    },
    /// Ett verktygsanrop.
    ToolCall {
        /// Anropets id.
        id: String,
        /// Verktygets namn.
        name: String,
        /// Argumenten.
        arguments: serde_json::Value,
    },
    /// Resultatet av ett verktygsanrop.
    ToolResult {
        /// Id:t på anropet.
        #[serde(rename = "toolCallId")]
        tool_call_id: String,
        /// Verktygets namn.
        #[serde(rename = "toolName")]
        tool_name: String,
        /// Innehållet i resultatet.
        content: Vec<ContentBlock>,
        /// Om resultatet var ett fel.
        #[serde(rename = "isError", skip_serializing_if = "is_false", default)]
        is_error: bool,
    },
}

/// Om värdet är falskt - så att fältet kan hoppas över.
fn is_false(v: &bool) -> bool {
    !*v
}

/// Ett meddelande i samtalet.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "role", rename_all = "camelCase")]
pub enum Message {
    /// Användarens meddelande.
    User {
        /// Innehållet.
        content: Vec<ContentBlock>,
        /// Tidpunkten i millisekunder.
        timestamp: i64,
    },
    /// Agentens meddelande.
    Assistant {
        /// Innehållet.
        content: Vec<ContentBlock>,
        /// Varför svaret slutade.
        #[serde(
            rename = "stopReason",
            skip_serializing_if = "String::is_empty",
            default
        )]
        stop_reason: String,
        /// Tidpunkten i millisekunder.
        timestamp: i64,
        /// Tokenanvändningen.
        #[serde(skip_serializing_if = "Option::is_none", default)]
        usage: Option<Usage>,
    },
    /// Resultatet av ett verktygsanrop.
    ToolResult {
        /// Id:t på anropet.
        #[serde(rename = "toolCallId")]
        tool_call_id: String,
        /// Verktygets namn.
        #[serde(rename = "toolName")]
        tool_name: String,
        /// Innehållet i resultatet.
        content: Vec<ContentBlock>,
        /// Om resultatet var ett fel.
        #[serde(rename = "isError", skip_serializing_if = "is_false", default)]
        is_error: bool,
        /// Extra detaljer.
        #[serde(skip_serializing_if = "Option::is_none", default)]
        details: Option<serde_json::Value>,
        /// Tokenanvändningen.
        #[serde(skip_serializing_if = "Option::is_none", default)]
        usage: Option<Usage>,
        /// Tidpunkten i millisekunder.
        timestamp: i64,
    },
    /// Systemmeddelandet.
    System {
        /// Innehållet.
        content: Vec<ContentBlock>,
        /// Tidpunkten i millisekunder.
        timestamp: i64,
    },
}

impl Message {
    /// Rollen som text.
    pub fn role(&self) -> Role {
        match self {
            Message::User { .. } => Role::User,
            Message::Assistant { .. } => Role::Assistant,
            Message::ToolResult { .. } => Role::ToolResult,
            Message::System { .. } => Role::System,
        }
    }

    /// Innehållet i meddelandet.
    pub fn content(&self) -> &[ContentBlock] {
        match self {
            Message::User { content, .. }
            | Message::Assistant { content, .. }
            | Message::ToolResult { content, .. }
            | Message::System { content, .. } => content,
        }
    }

    /// Tidpunkten i millisekunder.
    pub fn timestamp(&self) -> i64 {
        match self {
            Message::User { timestamp, .. }
            | Message::Assistant { timestamp, .. }
            | Message::ToolResult { timestamp, .. }
            | Message::System { timestamp, .. } => *timestamp,
        }
    }

    /// Tokenanvändningen, om meddelandet bär en sådan.
    pub fn usage(&self) -> Option<&Usage> {
        match self {
            Message::Assistant { usage, .. } | Message::ToolResult { usage, .. } => usage.as_ref(),
            Message::User { .. } | Message::System { .. } => None,
        }
    }

    /// Skapar ett användarmeddelande.
    pub fn user(text: impl Into<String>) -> Self {
        Message::User {
            content: vec![ContentBlock::Text { text: text.into() }],
            timestamp: nu(),
        }
    }

    /// Skapar ett systemmeddelande.
    pub fn system(text: impl Into<String>) -> Self {
        Message::System {
            content: vec![ContentBlock::Text { text: text.into() }],
            timestamp: nu(),
        }
    }
}

/// Nuvarande tid i millisekunder.
pub fn nu() -> i64 {
    chrono::Utc::now().timestamp_millis()
}

/// Tokenanvändningen för ett svar.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Usage {
    /// Tokens som gick in.
    #[serde(skip_serializing_if = "is_zero", default)]
    pub input_tokens: i64,
    /// Tokens som kom ut.
    #[serde(skip_serializing_if = "is_zero", default)]
    pub output_tokens: i64,
    /// Tokens totalt.
    #[serde(skip_serializing_if = "is_zero", default)]
    pub total_tokens: i64,
    /// Tokens som lästs från modellens cache.
    ///
    /// De räknas separat eftersom de har ett eget pris.
    #[serde(skip_serializing_if = "is_zero", default)]
    pub cache_tokens: i64,
}

/// Om värdet är noll - så att fältet kan hoppas över.
fn is_zero(v: &i64) -> bool {
    *v == 0
}

/// Gör om innehållsblocken till text.
pub fn text_innehåll(blocker: &[ContentBlock]) -> String {
    let mut text = String::new();

    for block in blocker {
        if let ContentBlock::Text { text: stycke } = block {
            text.push_str(stycke);
        }
    }

    text
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn anvandarmeddelandet_ser_text() {
        let meddelande = Message::user("hej");

        assert_eq!(meddelande.role(), Role::User);
        assert_eq!(text_innehåll(meddelande.content()), "hej");
    }

    #[test]
    fn tomma_usage_falt_hoppas_over() {
        let anvandning = Usage::default();
        let json = serde_json::to_string(&anvandning).unwrap();

        assert_eq!(json, "{}");
    }

    #[test]
    fn roller_serialiseras_som_go() {
        assert_eq!(
            serde_json::to_string(&Role::ToolResult).unwrap(),
            "\"toolResult\""
        );
        assert_eq!(serde_json::to_string(&Role::User).unwrap(), "\"user\"");
    }
}
