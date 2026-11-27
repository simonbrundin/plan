//! Agentens verktyg.
//!
//! Go-agentens `internal/core/tools`. Ramen definierar ett verktyg, dess
//! schema och hur argumenten valideras.

pub mod cancel;
pub mod fs;
pub mod glob;
pub mod learnings;
pub mod sanitize;
pub mod schemas;
pub mod validator;

use std::sync::Arc;

use serde_json::Value;

use crate::messages::ContentBlock;
use crate::provider::ToolResult;

/// En framtid som returnerar resultatet av ett verktygsanrop.
pub type ToolFuture = std::pin::Pin<
    Box<dyn std::future::Future<Output = Result<ToolResult, String>> + Send + 'static>,
>;

/// En verktygsfunktion.
pub type ToolFn = Arc<dyn Fn(ToolCall) -> ToolFuture + Send + Sync>;

/// Anropet som ett verktyg får.
pub struct ToolCall {
    /// Id:t på anropet.
    pub id: String,
    /// Argumenten.
    pub args: Value,
    /// Kanalen som signalerar avbrytning.
    pub signal: cancel::CancelChannel,
    /// Skickar delresultat medan verktyget kör.
    pub on_update: Option<tokio::sync::mpsc::Sender<ToolResult>>,
}

impl ToolCall {
    /// Läser ett strängargument.
    pub fn str_arg(&self, namn: &str, reserv: &str) -> String {
        self.args
            .get(namn)
            .and_then(|värde| värde.as_str())
            .unwrap_or(reserv)
            .to_string()
    }

    /// Läser ett talargument.
    pub fn float_arg(&self, namn: &str, reserv: f64) -> f64 {
        self.args
            .get(namn)
            .and_then(|värde| värde.as_f64())
            .unwrap_or(reserv)
    }

    /// Läser ett booleskt argument.
    pub fn bool_arg(&self, namn: &str, reserv: bool) -> bool {
        self.args
            .get(namn)
            .and_then(|värde| värde.as_bool())
            .unwrap_or(reserv)
    }

    /// Om verktyget har avbrutits.
    pub fn avbruten(&self) -> bool {
        *self.signal.borrow()
    }
}

/// Ett verktyg som agenten kan använda.
#[derive(Clone)]
pub struct ToolDefinition {
    /// Verktygets namn.
    pub name: String,
    /// Beskrivningen av verktyget.
    pub description: String,
    /// Schemat för argumenten.
    pub input_schema: Value,
    /// Funktionen som utför verktyget.
    pub execute: ToolFn,
}

impl ToolDefinition {
    /// Gör om definitionen till leverantörens format.
    pub fn till_leverantör(&self) -> crate::provider::ToolDefinition {
        crate::provider::ToolDefinition {
            name: self.name.clone(),
            description: self.description.clone(),
            input_schema: self.input_schema.clone(),
        }
    }
}

/// Resultatet av ett verktyg - ett fel.
pub fn fel_resultat(text: &str) -> ToolResult {
    ToolResult {
        content: vec![ContentBlock::Text {
            text: format!("Error: {text}"),
        }],
        is_error: true,
        ..Default::default()
    }
}

/// Resultatet av ett avbrutet verktyg.
pub fn avbruten_resultat(åtgärd: &str) -> ToolResult {
    ToolResult {
        content: vec![ContentBlock::Text {
            text: format!("{åtgärd} cancelled"),
        }],
        is_error: true,
        ..Default::default()
    }
}

/// Resultatet av ett verktyg - bara text.
pub fn text_resultat(text: impl Into<String>) -> ToolResult {
    ToolResult {
        content: vec![ContentBlock::Text { text: text.into() }],
        ..Default::default()
    }
}

/// De verktyg som agenten har tillgång till som standard.
///
/// Go: `internal/core/tools/tools.go` - `GetDefaultTools`.
pub fn standardverktyg() -> Vec<ToolDefinition> {
    vec![
        fs::bashverktyg(),
        fs::läsverktyg(),
        fs::skrivverktyg(),
        fs::redigeraverktyg(),
        fs::globverktyg(),
        fs::grepverktyg(),
        fs::läs_fleraverktyg(),
    ]
}

/// Systemprompten som används om ingen annan ges.
///
/// Go: `internal/core/tools/tools.go` - `DefaultSystemPrompt`.
pub fn standard_systemprompt() -> String {
    "You are a helpful AI coding assistant.\n\n\
You have access to the following tools:\n\
- bash: Execute bash commands\n\
- read: Read file contents\n\
- write: Write content to files\n\
- edit: Edit specific parts of files\n\
- glob: Find files matching patterns\n\
- grep: Search for text in files\n\
- read_multiple: Read multiple files at once\n\n\
When using tools:\n\
- Always validate arguments before execution\n\
- Use appropriate timeouts\n\
- Handle errors gracefully\n\
- Provide clear feedback\n\n\
For file operations, prefer edit over write when making small changes."
        .to_string()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn fel_resultat_borjar_med_error() {
        let resultat = fel_resultat("något gick fel");

        assert!(resultat.is_error);
        assert_eq!(
            crate::messages::text_innehåll(&resultat.content),
            "Error: något gick fel"
        );
    }

    #[test]
    fn avbruten_resultat_namnar_atgarden() {
        let resultat = avbruten_resultat("Command cancelled");

        assert!(resultat.is_error);
        assert_eq!(
            crate::messages::text_innehåll(&resultat.content),
            "Command cancelled cancelled"
        );
    }

    #[test]
    fn standardverktygen_ar_sju() {
        assert_eq!(standardverktyg().len(), 7);
    }
}
