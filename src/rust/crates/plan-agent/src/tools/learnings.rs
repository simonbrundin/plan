//! Verktygen `add_learning` och `get_learnings`.
//!
//! Go-agentens `internal/core/orchestrator/learnings_tools.go`. Verktygen
//! registreras inte i agentens vanliga verktygslista - de används när
//! agenten kör med en orchestrator.

use std::sync::Arc;

use serde_json::json;

use super::{text_resultat, ToolCall, ToolDefinition};
use crate::messages::ContentBlock;
use crate::orchestrator::Orchestrator;
use crate::provider::ToolResult;

/// Verktyget som sparar en lärdom.
pub fn add_learning_verktyg(orchestrator: Arc<Orchestrator>) -> ToolDefinition {
    ToolDefinition {
        name: "add_learning".to_string(),
        description: "Record a learning for future reference. Use when you discover something important, make a mistake, or find a better approach.".to_string(),
        input_schema: json!({
            "type": "object",
            "properties": {
                "category": {
                    "type": "string",
                    "description": "Category: test, bug, pattern, optimization, security, refactor, or general",
                    "enum": ["test", "bug", "pattern", "optimization", "security", "refactor", "general"]
                },
                "title": {
                    "type": "string",
                    "description": "Brief title for this learning"
                },
                "description": {
                    "type": "string",
                    "description": "What happened or what you discovered"
                },
                "lesson": {
                    "type": "string",
                    "description": "The key takeaway or best practice to follow"
                }
            },
            "required": ["category", "title", "lesson"]
        }),
        execute: Arc::new(move |anrop: ToolCall| {
            let orchestrator = Arc::clone(&orchestrator);
            Box::pin(async move { Ok(verkställ_add_learning(&orchestrator, anrop)) })
        }),
    }
}

/// Verktyget som läser lärdomarna.
pub fn get_learnings_verktyg(orchestrator: Arc<Orchestrator>) -> ToolDefinition {
    ToolDefinition {
        name: "get_learnings".to_string(),
        description: "Retrieve all recorded learnings. Use this to learn from past mistakes and follow best practices.".to_string(),
        // Go lämnar schemat som nil - leverantören fyller i ett tomt objekt.
        input_schema: json!({}),
        execute: Arc::new(move |anrop: ToolCall| {
            let orchestrator = Arc::clone(&orchestrator);
            Box::pin(async move { Ok(verkställ_get_learnings(&orchestrator, anrop)) })
        }),
    }
}

/// Sparar en lärdom.
fn verkställ_add_learning(orchestrator: &Orchestrator, anrop: ToolCall) -> ToolResult {
    let kategori = anrop.str_arg("category", "general");
    let titel = anrop.str_arg("title", "");
    let beskrivning = anrop.str_arg("description", "");
    let lärdom = anrop.str_arg("lesson", "");

    if titel.is_empty() {
        return fel_text("title is required");
    }

    if lärdom.is_empty() {
        return fel_text("lesson is required");
    }

    match orchestrator.add_learning(&kategori, &titel, &beskrivning, &lärdom) {
        Ok(()) => text_resultat(format!("Learning recorded: {titel}")),
        Err(err) => fel_text(&format!("Failed to save learning: {err}")),
    }
}

/// Läser lärdomarna.
fn verkställ_get_learnings(orchestrator: &Orchestrator, _anrop: ToolCall) -> ToolResult {
    let kontext = orchestrator.generate_learnings_context(&[]);

    if kontext.is_empty() {
        return text_resultat("No learnings recorded yet.");
    }

    text_resultat(kontext)
}

/// Ett felresultat med samma text som Go - utan prefixen "Error:".
fn fel_text(text: &str) -> ToolResult {
    ToolResult {
        content: vec![ContentBlock::Text {
            text: text.to_string(),
        }],
        is_error: true,
        ..Default::default()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::learnings::LearningsManager;
    use crate::state_store::{InMemoryStateStore, StateStore};

    /// Bygger en orchestrator med en egen lärdomsfil.
    fn orchestrator(temp: &tempfile::TempDir) -> Arc<Orchestrator> {
        let lärdomar = LearningsManager::ny(temp.path().join("learnings.md"));
        let lager = StateStore::Memory(InMemoryStateStore::new());
        Arc::new(Orchestrator::with_learnings(lager, lärdomar))
    }

    /// Skapar ett verktygsanrop med givna argument.
    fn anrop(args: serde_json::Value) -> ToolCall {
        ToolCall {
            id: "anrop-1".to_string(),
            args,
            signal: crate::tools::cancel::CancellationManager::new().register("anrop-1"),
            on_update: None,
        }
    }

    /// Plockar ut texten ur ett resultat.
    fn texten(resultat: &ToolResult) -> String {
        match resultat.content.first() {
            Some(ContentBlock::Text { text }) => text.clone(),
            _ => String::new(),
        }
    }

    #[tokio::test]
    async fn lärdom_sparas_och_bekräftas() {
        let temp = tempfile::tempdir().unwrap();
        let orchestrator = orchestrator(&temp);
        let verktyg = add_learning_verktyg(Arc::clone(&orchestrator));

        let resultat = (verktyg.execute)(anrop(json!({
            "category": "bug",
            "title": "En titel",
            "description": "En beskrivning",
            "lesson": "En lärdom"
        })))
        .await
        .unwrap();

        assert!(!resultat.is_error);
        assert_eq!(texten(&resultat), "Learning recorded: En titel");
    }

    #[tokio::test]
    async fn titel_och_lärdom_krävs() {
        let temp = tempfile::tempdir().unwrap();
        let orchestrator = orchestrator(&temp);
        let verktyg = add_learning_verktyg(Arc::clone(&orchestrator));

        let utan_titel =
            (verktyg.execute)(anrop(json!({"category": "bug", "lesson": "En lärdom"})))
                .await
                .unwrap();

        assert!(utan_titel.is_error);
        assert_eq!(texten(&utan_titel), "title is required");

        let utan_lärdom = (verktyg.execute)(anrop(json!({"category": "bug", "title": "En titel"})))
            .await
            .unwrap();

        assert!(utan_lärdom.is_error);
        assert_eq!(texten(&utan_lärdom), "lesson is required");
    }

    #[tokio::test]
    async fn tom_fil_ge_kontext_lika_go() {
        let temp = tempfile::tempdir().unwrap();
        let orchestrator = orchestrator(&temp);
        let verktyg = get_learnings_verktyg(orchestrator);

        let resultat = (verktyg.execute)(anrop(json!({}))).await.unwrap();

        // Go returnerar aldrig tomt: `GetLearnings` ger alltid en
        // standardtext, och den hamnar i kontexten.
        assert!(!resultat.is_error);
        assert_eq!(
            texten(&resultat),
            "\n\n## Relevant Learnings\n\n# Learnings\n\nNo learnings recorded yet.\n"
        );
    }
}
