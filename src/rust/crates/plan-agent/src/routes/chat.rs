//! Den interaktiva chatten.
//!
//! Go-agentens `internal/api/handlers/agent.go`, `conversations.go` och
//! `run_recorder.go`. Chatten kör agenten mot språkmodellen och skriver
//! ut händelserna som NDJSON.

use std::sync::Arc;

use serde::{Deserialize, Serialize};
use serde_json::Value;

use crate::core::{AgentEvent, AgentLoopConfig, AgentRuntime, EventType};
use crate::db::models::{Conversation, Message};
use crate::messages::{ContentBlock, Message as AgentMessage};
use crate::planclient;
use crate::planclient::models::{
    CreateAgentRunEventRequest, EVENT_ERROR, EVENT_SUBGOAL_CREATED, EVENT_TOOL_CALL,
    EVENT_TOOL_RESULT,
};
use crate::provider::{Model, SharedProvider};
use crate::tools::ToolCall;
use crate::tools::ToolDefinition;

/// Sammanhanget om vilket mål chatten gäller.
#[derive(Debug, Clone, Default, Deserialize, Serialize)]
pub struct GoalContext {
    /// Målets id.
    pub id: i64,
    /// Målets titel.
    pub title: String,
    /// Målets beskrivning.
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub description: String,
}

/// Request för att skicka ett meddelande.
#[derive(Debug, Clone, Default, Deserialize)]
pub struct CreateMessageRequest {
    /// Meddelandets innehåll.
    pub content: String,
    /// Sammanhanget med målet, om chatten gäller ett mål.
    #[serde(default)]
    pub goal_context: Option<GoalContext>,
}

/// En leverantör som valts för chatten.
#[derive(Debug, Clone)]
pub struct SelectedProvider {
    /// Leverantörens namn.
    pub name: String,
    /// API-nyckeln.
    pub api_key: String,
    /// Modellens namn.
    pub model: String,
    /// Leverantörens adress.
    pub url: String,
}

/// En leverantör i prioritetsordning.
struct ProviderConfig {
    /// Namnet.
    name: &'static str,
    /// Miljövariabeln med API-nyckeln.
    key_env: &'static str,
    /// Miljövariabeln med modellnamnet.
    model_env: &'static str,
    /// Standardmodellen.
    default_model: &'static str,
    /// Miljövariabeln med adressen.
    url_env: &'static str,
    /// Standardadressen.
    default_url: &'static str,
}

/// Leverantörerna i prioritetsordning.
///
/// Go: `internal/api/handlers/agent.go` - `providerPriority`.
const LEVERANTÖRSPRIORITET: &[ProviderConfig] = &[
    ProviderConfig {
        name: "agentgateway",
        key_env: "AGENTGATEWAY_API_KEY",
        model_env: "AGENTGATEWAY_MODEL",
        default_model: "gpt-4o-mini",
        url_env: "AGENTGATEWAY_URL",
        default_url: "https://api.agentgateway.dev/v1",
    },
    ProviderConfig {
        name: "minimax",
        key_env: "MINIMAX_API_KEY",
        model_env: "MINIMAX_MODEL",
        default_model: "MiniMax-M2.7",
        url_env: "",
        default_url: "",
    },
    ProviderConfig {
        name: "openai",
        key_env: "OPENAI_API_KEY",
        model_env: "OPENAI_MODEL",
        default_model: "gpt-4o-mini",
        url_env: "",
        default_url: "",
    },
    ProviderConfig {
        name: "anthropic",
        key_env: "ANTHROPIC_API_KEY",
        model_env: "ANTHROPIC_MODEL",
        default_model: "claude-sonnet-4-7-25",
        url_env: "",
        default_url: "",
    },
];

/// Miljövariabeln med en leverantörs API-nyckel.
///
/// Go: `providerEnvKey`.
pub fn leverantörsnyckel_env(namn: &str) -> String {
    match namn {
        "agentgateway" => "AGENTGATEWAY_API_KEY".to_string(),
        "minimax" => "MINIMAX_API_KEY".to_string(),
        "openai" => "OPENAI_API_KEY".to_string(),
        "anthropic" => "ANTHROPIC_API_KEY".to_string(),
        annan => format!("{annan}_API_KEY"),
    }
}

/// Läser en miljövariabel.
fn env(namn: &str) -> String {
    std::env::var(namn).unwrap_or_default()
}

/// Väljer den bästa leverantören utifrån nycklarna och miljövariablerna.
///
/// Go: `selectProvider`.
pub fn välj_leverantör(
    nycklar: &std::collections::HashMap<String, String>,
    modeller: &std::collections::HashMap<String, String>,
) -> Option<SelectedProvider> {
    for config in LEVERANTÖRSPRIORITET {
        if let Some(nyckel) = nycklar.get(config.name) {
            if !nyckel.is_empty() {
                return Some(bygg_vald(config, nyckel, modeller.get(config.name)));
            }
        }

        let nyckel = env(config.key_env);

        if !nyckel.is_empty() {
            let modell = if config.model_env.is_empty() {
                String::new()
            } else {
                env(config.model_env)
            };

            return Some(bygg_vald(config, &nyckel, Some(&modell)));
        }
    }

    None
}

/// Bygger den valda leverantören.
///
/// Go: `buildSelected`.
fn bygg_vald(
    config: &ProviderConfig,
    nyckel: &str,
    modelltips: Option<&String>,
) -> SelectedProvider {
    let modell = match modelltips {
        Some(modell) if !modell.is_empty() => modell.clone(),
        _ => config.default_model.to_string(),
    };

    let url = if config.url_env.is_empty() {
        String::new()
    } else {
        let url = env(config.url_env);

        if url.is_empty() {
            config.default_url.to_string()
        } else {
            url
        }
    };

    SelectedProvider {
        name: config.name.to_string(),
        api_key: nyckel.to_string(),
        model: modell,
        url,
    }
}

/// Väljer leverantören för ett samtal.
///
/// Go: `resolveProvider`.
pub fn välj_leverantör_för_samtal(
    samtal: &Conversation,
    nycklar: &std::collections::HashMap<String, String>,
    modeller: &std::collections::HashMap<String, String>,
) -> Option<SelectedProvider> {
    // Samtalets leverantör företräds om den har en nyckel.
    if !samtal.provider.is_empty() && !samtal.model.is_empty() {
        let mut nyckel = nycklar.get(&samtal.provider).cloned().unwrap_or_default();

        if nyckel.is_empty() {
            nyckel = env(&leverantörsnyckel_env(&samtal.provider));
        }

        if !nyckel.is_empty() {
            let mut modell = samtal.model.clone();

            if let Some(annan) = modeller.get(&samtal.provider) {
                if !annan.is_empty() {
                    modell = annan.clone();
                }
            }

            return Some(SelectedProvider {
                name: samtal.provider.clone(),
                api_key: nyckel,
                model: modell,
                url: String::new(),
            });
        }
    }

    välj_leverantör(nycklar, modeller)
}

/// Skapar leverantören som agenten använder.
///
/// Go: `createProvider`.
pub fn skapa_leverantör(namn: &str, url: &str) -> SharedProvider {
    match namn {
        "agentgateway" => {
            let adress = if url.is_empty() {
                "https://api.agentgateway.dev/v1"
            } else {
                url
            };

            Arc::new(crate::provider::openai::OpenAiProvider::new(adress))
        }
        "minimax" => Arc::new(crate::provider::minimax::MinimaxProvider::new(
            "https://api.minimax.io/v1",
        )),
        _ => Arc::new(crate::provider::openai::OpenAiProvider::new(
            "https://api.openai.com/v1",
        )),
    }
}

/// Systemprompten som används i chatten.
///
/// Go: `defaultSystemPrompt`.
pub const STANDARD_SYSTEMPROMPT: &str = "You are a helpful AI assistant for planning work.\n\n\
Be concise and helpful. Respond in the same language as the user.\n\n\
Interactive chat is for understanding and planning a goal. You may create sub-goals\n\
only when the user explicitly asks you to add them; use the create_subgoal tool for\n\
that instead of merely describing them. Do not invent repository state, claim that\n\
files were changed, or pretend that tests were run. Ask focused questions when\n\
important information is missing. A separate isolated AgentRun is responsible for\n\
repository inspection and code changes after the plan is approved.";

/// Läser projektets instruktioner från AGENTS.md.
///
/// Go: `loadProjectSystemPrompt`.
pub fn läs_projektinstruktioner() -> String {
    let Ok(katalog) = std::env::current_dir() else {
        return String::new();
    };

    let mut nuvarande = katalog.as_path();

    loop {
        let sökväg = nuvarande.join("AGENTS.md");

        if let Ok(data) = std::fs::read(&sökväg) {
            if !data.is_empty() {
                return format!(
                    "\n\n--- Project Instructions (AGENTS.md) ---\n{}",
                    String::from_utf8_lossy(&data)
                );
            }
        }

        match nuvarande.parent() {
            Some(föregående) if föregående != nuvarande => nuvarande = föregående,
            _ => break,
        }
    }

    String::new()
}

/// Bygger systemprompten.
///
/// Go: `buildSystemPrompt`.
pub fn bygg_systemprompt() -> String {
    let mut prompt = STANDARD_SYSTEMPROMPT.to_string();

    let projektkontext = läs_projektinstruktioner();

    if !projektkontext.is_empty() {
        prompt.push_str(&projektkontext);
        prompt.push_str(
            "\n\nIMPORTANT: The project instructions above take precedence over \
the general instructions when they conflict.",
        );
    }

    prompt
}

/// Gör om databasens meddelanden till agentens meddelanden.
///
/// Go: `dbMessagesToAgentMessages`.
pub fn databasmeddelanden_till_agent(meddelanden: &[Message]) -> Vec<AgentMessage> {
    let mut resultat = Vec::with_capacity(meddelanden.len());

    for meddelande in meddelanden {
        // Verktygsmeddelanden skickas inte till modellen - de är
        // redan körda.
        match meddelande.role.as_str() {
            "tool_request" | "tool_result" | "thinking" => continue,
            "user" => resultat.push(AgentMessage::User {
                content: vec![ContentBlock::Text {
                    text: meddelande.content.clone(),
                }],
                timestamp: meddelande.created_at.timestamp_millis(),
            }),
            "assistant" => resultat.push(AgentMessage::Assistant {
                content: vec![ContentBlock::Text {
                    text: meddelande.content.clone(),
                }],
                stop_reason: String::new(),
                timestamp: meddelande.created_at.timestamp_millis(),
                usage: None,
            }),
            "system" => resultat.push(AgentMessage::System {
                content: vec![ContentBlock::Text {
                    text: meddelande.content.clone(),
                }],
                timestamp: meddelande.created_at.timestamp_millis(),
            }),
            _ => {}
        }
    }

    resultat
}

/// Formaterar målkontexten som skickas till modellen.
///
/// Go: `formatGoalContext`.
pub fn formatera_målkontext(mål: &GoalContext) -> String {
    let mut kontext = format!("GOAL #{}: {}", mål.id, mål.title);

    if !mål.description.is_empty() {
        kontext.push_str("\n\nGOAL DESCRIPTION:\n");
        kontext.push_str(&mål.description);
    }

    format!("We are discussing this Plan goal:\n\n{kontext}")
}

/// Plockar ut texten ur ett verktygsresultat.
///
/// Go: `extractTextResult`.
pub fn textsakning(resultat: &crate::provider::ToolResult) -> String {
    resultat
        .content
        .iter()
        .find_map(|block| match block {
            ContentBlock::Text { text } => Some(text.clone()),
            _ => None,
        })
        .unwrap_or_default()
}

/// Plockar ut resonemanget mellan `<think>` och `</think>`.
///
/// Go: `extractThinkBlocks`.
pub fn extrahera_resonemang(text: &str) -> String {
    const BÖRJAN: &[u8] = b"<think>";
    const SLUT: &[u8] = b"</think>";

    let mut resultat = String::new();
    let mut text = text.as_bytes().to_vec();

    loop {
        let mut start = 0;
        let mut slut = 0;

        let mut i = 0;

        while i + BÖRJAN.len() <= text.len() {
            if &text[i..i + BÖRJAN.len()] == BÖRJAN {
                start = i + BÖRJAN.len();
                break;
            }

            i += 1;
        }

        let mut i = start;

        while i + SLUT.len() <= text.len() {
            if &text[i..i + SLUT.len()] == SLUT {
                slut = i;
                break;
            }

            i += 1;
        }

        if start == 0 || slut == 0 || slut <= start {
            break;
        }

        if !resultat.is_empty() {
            resultat.push_str("\n\n");
        }

        resultat.push_str(&String::from_utf8_lossy(&text[start..slut]));
        text = text[slut + SLUT.len()..].to_vec();
    }

    resultat
}

/// Tar bort resonemanget ur texten.
///
/// Go: `stripThinkBlocks`.
pub fn ta_bort_resonemang(text: &str) -> String {
    const BÖRJAN: &[u8] = b"<think>";
    const SLUT: &[u8] = b"</think>";

    let mut buffert: Vec<u8> = Vec::new();
    let mut i_resonemang = false;
    let mut i = 0;

    let text = text.as_bytes();

    while i < text.len() {
        if i <= text.len().saturating_sub(BÖRJAN.len()) && &text[i..i + BÖRJAN.len()] == BÖRJAN {
            i_resonemang = true;
            i += BÖRJAN.len();
            continue;
        }

        if i <= text.len().saturating_sub(SLUT.len()) && &text[i..i + SLUT.len()] == SLUT {
            i_resonemang = false;
            i += SLUT.len();
            continue;
        }

        if !i_resonemang {
            buffert.push(text[i]);
        }

        i += 1;
    }

    String::from_utf8_lossy(&buffert).to_string()
}

/// Skriver händelser till en agentkörning i Plan.
///
/// Go: `runRecorder`.
#[derive(Clone)]
pub struct RunRecorder {
    /// Klienten mot Plan.
    client: planclient::PlanClient,
    /// Körningens id.
    run_id: i64,
}

impl RunRecorder {
    /// Hittar den senaste körningen för ett mål.
    ///
    /// Saknas körningen returneras ingen recorder - historiken är då
    /// tom, vilket är bättre än att skapa en körning bakom ryggen på
    /// chatten.
    pub async fn hitta(client: &planclient::PlanClient, goal_id: i64) -> Option<RunRecorder> {
        let körningar = match client.agent_runs(Some(goal_id)).await {
            Ok(körningar) => körningar,
            Err(fel) => {
                tracing::warn!("Kunde inte hitta agentkörning för mål #{goal_id}: {fel}");
                return None;
            }
        };

        let körning = körningar.first()?;

        Some(RunRecorder {
            client: client.clone(),
            run_id: körning.id,
        })
    }

    /// Körningens id.
    pub fn run_id(&self) -> i64 {
        self.run_id
    }

    /// Skriver en händelse. Fel ignoreras - historiken får inte stoppa
    /// chatten.
    pub async fn skriv(&self, typ: &str, fas: &str, nyttolast: Value) {
        let händelse = CreateAgentRunEventRequest {
            r#type: typ.to_string(),
            phase: if fas.is_empty() {
                None
            } else {
                Some(fas.to_string())
            },
            payload: Some(nyttolast),
        };

        if let Err(fel) = self
            .client
            .create_agent_run_event(self.run_id, händelse)
            .await
        {
            tracing::warn!(
                "Kunde inte skriva {typ}-händelse för körning {}: {fel}",
                self.run_id
            );
        }
    }
}

/// Verktyget som skapar undermål.
///
/// Go: `createSubgoalTool`.
pub fn skapa_undermålverktyg(
    client: planclient::PlanClient,
    förälder_id: i64,
    recorder: Option<RunRecorder>,
) -> ToolDefinition {
    let schema = serde_json::json!({
        "type": "object",
        "properties": {
            "title": {
                "type": "string",
                "description": "Short, actionable title for the sub-goal.",
            },
            "description": {
                "type": "string",
                "description": "Optional description and acceptance criteria.",
            },
        },
        "required": ["title"],
    });

    let execute = move |anrop: ToolCall| -> crate::tools::ToolFuture {
        let client = client.clone();
        let recorder = recorder.clone();

        Box::pin(async move {
            let titel = anrop.str_arg("title", "");
            let beskrivning = anrop.str_arg("description", "");

            if titel.is_empty() {
                return Ok(crate::tools::fel_resultat("title is required"));
            }

            match client
                .create_subgoal(förälder_id, &titel, &beskrivning)
                .await
            {
                Ok(mål) => {
                    if let Some(recorder) = &recorder {
                        recorder
                            .skriv(
                                EVENT_SUBGOAL_CREATED,
                                "",
                                serde_json::json!({
                                    "goal_id": mål.id,
                                    "parent_id": förälder_id,
                                    "title": mål.title,
                                }),
                            )
                            .await;
                    }

                    Ok(crate::tools::text_resultat(format!(
                        "Created sub-goal #{}: {}",
                        mål.id, mål.title
                    )))
                }
                Err(fel) => {
                    if let Some(recorder) = &recorder {
                        recorder
                            .skriv(
                                EVENT_ERROR,
                                "",
                                serde_json::json!({
                                    "tool": "create_subgoal",
                                    "error": fel.to_string(),
                                }),
                            )
                            .await;
                    }

                    Ok(crate::tools::fel_resultat(&fel.to_string()))
                }
            }
        })
    };

    ToolDefinition {
        name: "create_subgoal".to_string(),
        description: "Create one concrete sub-goal under the current Plan goal. Use only when the user explicitly asks to add or create sub-goals.".to_string(),
        input_schema: schema,
        execute: Arc::new(execute),
    }
}

/// Bygger agenten för ett samtal.
///
/// Go: `createAgentFromConversation`.
pub async fn bygg_agent(
    pool: &sqlx::PgPool,
    samtal: &Conversation,
    målkontext: Option<&GoalContext>,
    plan_client: Option<&planclient::PlanClient>,
    recorder: Option<RunRecorder>,
) -> Result<AgentRuntime, String> {
    let nycklar = crate::db::repository::get_api_keys(pool)
        .await
        .map_err(|fel| fel.to_string())?;

    let mut nyckelkarta = std::collections::HashMap::new();
    let mut modellkarta = std::collections::HashMap::new();

    for nyckel in nycklar {
        nyckelkarta.insert(nyckel.provider.clone(), nyckel.api_key.clone());
        modellkarta.insert(nyckel.provider.clone(), nyckel.model.clone());
    }

    let Some(vald) = välj_leverantör_för_samtal(samtal, &nyckelkarta, &modellkarta) else {
        return Err(
            "no API key configured: add one in settings or set MINIMAX_API_KEY / OPENAI_API_KEY environment variable"
                .to_string(),
        );
    };

    // Chatten har inga skal- eller filverktyg - den används för att
    // planera, inte för att ändra kod.
    let mut verktyg: Vec<ToolDefinition> = Vec::new();

    if let (Some(mål), Some(client)) = (målkontext, plan_client) {
        verktyg.push(skapa_undermålverktyg(client.clone(), mål.id, recorder));
    }

    let modell = Model {
        provider: vald.name.clone(),
        name: vald.model.clone(),
        api_key: vald.api_key.clone(),
        endpoint: vald.url.clone(),
        parameters: Default::default(),
    };

    let config = AgentLoopConfig {
        model: modell,
        tools: verktyg,
        system_prompt: bygg_systemprompt(),
        provider: Some(skapa_leverantör(&vald.name, &vald.url)),
        max_turns: max_turner(),
        ..Default::default()
    };

    Ok(AgentRuntime::ny(config))
}

/// Det maximala antalet turer i en chatt.
///
/// Go: `conversationMaxTurns`.
pub fn max_turner() -> usize {
    const STANDARD: usize = 12;

    match std::env::var("AGENT_MAX_TURNS") {
        Ok(värde) => värde
            .parse::<usize>()
            .ok()
            .filter(|v| *v >= 1)
            .unwrap_or(STANDARD),
        Err(_) => STANDARD,
    }
}

/// Samlar händelserna under en agentkörning.
///
/// Go: `runAgentStreaming` - händelserna skickas vidare till klienten
/// medan de kommer.
pub struct EventCollector {
    /// Händelserna som har samlats in.
    pub händelser: Vec<AgentEvent>,
}

impl EventCollector {
    /// Skapar en samlare.
    pub fn ny() -> Self {
        Self {
            händelser: Vec::new(),
        }
    }

    /// Lägger till en händelse.
    pub fn lägg_till(&mut self, händelse: &AgentEvent) {
        self.händelser.push(händelse.clone());
    }

    /// Om något verktyg anropades.
    pub fn har_verktyg(&self) -> bool {
        self.händelser
            .iter()
            .any(|händelse| händelse.typ == EventType::ToolStart)
    }
}

/// Ett verktygsanrop som pågår.
struct PendingTool {
    /// Anropets id.
    id: String,
    /// Verktygets namn.
    name: String,
    /// Argumenten som JSON.
    args: String,
    /// Resultatet som text.
    result: String,
    /// Om resultatet var ett fel.
    is_error: bool,
}

/// Kör agenten och skriver händelserna som NDJSON.
///
/// Go: `internal/api/handlers/conversations.go` - `runAgentStreaming`.
pub async fn kör_agenten(
    mut agent: AgentRuntime,
    historik: Vec<AgentMessage>,
    prompt: String,
    conv_id: uuid::Uuid,
    pool: sqlx::PgPool,
    recorder: Option<RunRecorder>,
    sändare: tokio::sync::mpsc::Sender<Result<axum::body::Bytes, std::io::Error>>,
) {
    let pending = std::sync::Arc::new(std::sync::Mutex::new(Vec::<PendingTool>::new()));
    let svartext = std::sync::Arc::new(std::sync::Mutex::new(String::new()));
    let tool_turn = std::sync::Arc::new(std::sync::Mutex::new(0u32));

    let pending_händelse = std::sync::Arc::clone(&pending);
    let svartext_händelse = std::sync::Arc::clone(&svartext);
    let tool_turn_händelse = std::sync::Arc::clone(&tool_turn);
    let pool_händelse = pool.clone();
    let recorder_händelse = recorder.clone();

    let sändare_händelse = sändare.clone();

    let handler: crate::core::EventHandler = std::sync::Arc::new(move |händelse: AgentEvent| {
        let sändare = sändare_händelse.clone();
        let pending = std::sync::Arc::clone(&pending_händelse);
        let svartext = std::sync::Arc::clone(&svartext_händelse);
        let tool_turn = std::sync::Arc::clone(&tool_turn_händelse);
        let pool = pool_händelse.clone();
        let recorder = recorder_händelse.clone();

        Box::pin(async move {
            match händelse.typ {
                EventType::MessageEnd => {
                    if let Some(meddelande) = händelse.message {
                        let mut text = locka(&svartext);

                        for block in meddelande.content() {
                            if let ContentBlock::Text { text: stycke } = block {
                                text.push_str(stycke);
                            }
                        }
                    }
                }
                EventType::ToolStart => {
                    let args_json = händelse
                        .args
                        .as_ref()
                        .map(|args| args.to_string())
                        .unwrap_or_else(|| "{}".to_string());

                    {
                        let mut pending = locka(&pending);

                        pending.push(PendingTool {
                            id: händelse.tool_call_id.clone(),
                            name: händelse.tool_name.clone(),
                            args: args_json.clone(),
                            result: String::new(),
                            is_error: false,
                        });
                    }

                    let tur = {
                        let mut tur = locka(&tool_turn);
                        *tur += 1;
                        *tur
                    };

                    if let Some(recorder) = &recorder {
                        recorder
                            .skriv(
                                EVENT_TOOL_CALL,
                                "",
                                serde_json::json!({
                                    "tool": händelse.tool_name,
                                    "args": händelse.args,
                                    "turn": tur,
                                }),
                            )
                            .await;
                    }

                    // Verktygsanropet sparas och skickas direkt.
                    if let Ok(meddelande) = crate::db::repository::create_tool_message(
                        &pool,
                        conv_id,
                        "tool_request",
                        &händelse.tool_name,
                        &args_json,
                        "",
                        false,
                    )
                    .await
                    {
                        skicka_chunk(
                            &sändare,
                            &MeddelandeStycke {
                                message: meddelande,
                                r#type: "tool_request".to_string(),
                            },
                        )
                        .await;
                    }
                }
                EventType::ToolUpdate => {
                    if let Some(resultat) = &händelse.result {
                        let output = textsakning(resultat);

                        if !output.is_empty() {
                            skicka_chunk(
                                &sändare,
                                &VerktygStycke {
                                    is_error: händelse.is_error,
                                    output,
                                    tool_call_id: händelse.tool_call_id.clone(),
                                    tool_name: händelse.tool_name.clone(),
                                    r#type: "tool_output".to_string(),
                                },
                            )
                            .await;
                        }
                    }
                }
                EventType::ToolEnd => {
                    let resultattext = händelse
                        .result
                        .as_ref()
                        .map(textsakning)
                        .unwrap_or_default();
                    let is_error = händelse.is_error;

                    {
                        let mut pending = locka(&pending);

                        for anrop in pending.iter_mut() {
                            if anrop.id == händelse.tool_call_id {
                                anrop.result = resultattext.clone();
                                anrop.is_error = is_error;
                                break;
                            }
                        }
                    }

                    if let Some(recorder) = &recorder {
                        recorder
                            .skriv(
                                EVENT_TOOL_RESULT,
                                "",
                                serde_json::json!({
                                    "tool": händelse.tool_name,
                                    "is_error": is_error,
                                    "result": resultattext,
                                }),
                            )
                            .await;
                    }

                    let (verktygsnamn, verktygsargument) = {
                        let pending = locka(&pending);

                        pending
                            .iter()
                            .find(|anrop| anrop.id == händelse.tool_call_id)
                            .map(|anrop| (anrop.name.clone(), anrop.args.clone()))
                            .unwrap_or_default()
                    };

                    if let Ok(meddelande) = crate::db::repository::create_tool_message(
                        &pool,
                        conv_id,
                        "tool_result",
                        &verktygsnamn,
                        &verktygsargument,
                        &resultattext,
                        is_error,
                    )
                    .await
                    {
                        skicka_chunk(
                            &sändare,
                            &MeddelandeStycke {
                                message: meddelande,
                                r#type: "tool_result".to_string(),
                            },
                        )
                        .await;
                    }
                }
                _ => {}
            }
        })
    });

    let resultat = agent
        .skicka_meddelande_med_historik(historik, &prompt, handler)
        .await;

    // Ett fel i agenten ska synas - annars ser chatten ut att ha tyst
    // misslyckats.
    if let Some(fel) = &resultat.error {
        tracing::warn!("Agentkörningen misslyckades för samtal {conv_id}: {fel}");

        if let Some(recorder) = &recorder {
            recorder
                .skriv(EVENT_ERROR, "", serde_json::json!({ "error": fel }))
                .await;
        }

        skicka_chunk(
            &sändare,
            &FelStycke {
                error: fel.clone(),
                r#type: "error".to_string(),
            },
        )
        .await;
    }

    let svartext = {
        let mut text = locka(&svartext);
        std::mem::take(&mut *text)
    };

    // Resonemanget sparas separat - det ska inte visas i svaret.
    let resonemang = extrahera_resonemang(&svartext);
    let svartext = ta_bort_resonemang(&svartext);

    if !resonemang.is_empty() {
        if let Ok(meddelande) =
            crate::db::repository::create_thinking_message(&pool, conv_id, &resonemang).await
        {
            skicka_chunk(
                &sändare,
                &MeddelandeStycke {
                    message: meddelande,
                    r#type: "thinking".to_string(),
                },
            )
            .await;
        }
    }

    if !svartext.is_empty() {
        if let Ok(meddelande) =
            crate::db::repository::create_message(&pool, conv_id, "assistant", &svartext, 0, 0)
                .await
        {
            skicka_chunk(
                &sändare,
                &MeddelandeStycke {
                    message: meddelande,
                    r#type: "assistant_message".to_string(),
                },
            )
            .await;
        }
    }

    skicka_chunk(
        &sändare,
        &KlartStycke {
            r#type: "done".to_string(),
        },
    )
    .await;
}

/// Låser en mutex och returnerar låset.
fn locka<T>(mutex: &std::sync::Mutex<T>) -> std::sync::MutexGuard<'_, T> {
    mutex
        .lock()
        .unwrap_or_else(|förgiftad| förgiftad.into_inner())
}

/// Skickar ett stycke som NDJSON.
///
/// Go bygger styckena med `gin.H` - en karta - och Gin's JSON skriver
/// nycklarna i sorterad ordning. Därför ligger fälten i samma
/// ordning här.
pub async fn skicka_chunk<T: serde::Serialize>(
    sändare: &tokio::sync::mpsc::Sender<Result<axum::body::Bytes, std::io::Error>>,
    data: &T,
) {
    let Ok(text) = serde_json::to_string(data) else {
        return;
    };

    let _ = sändare
        .send(Ok(axum::body::Bytes::from(format!("{text}\n"))))
        .await;
}

/// Ett stycke som innehåller ett meddelande.
///
/// Go: `gin.H{"type": ..., "message": ...}`.
#[derive(Debug, serde::Serialize)]
pub struct MeddelandeStycke {
    /// Meddelandet.
    pub message: crate::db::models::Message,
    /// Styckets typ.
    pub r#type: String,
}

/// Ett stycke med utdata från ett verktyg.
///
/// Go: `gin.H{"type": ..., "toolCallId": ..., ...}`.
#[derive(Debug, serde::Serialize)]
pub struct VerktygStycke {
    /// Om utdatan var ett fel.
    #[serde(rename = "isError")]
    pub is_error: bool,
    /// Utdatan.
    pub output: String,
    /// Verktygsanropets id.
    #[serde(rename = "toolCallId")]
    pub tool_call_id: String,
    /// Verktygets namn.
    #[serde(rename = "toolName")]
    pub tool_name: String,
    /// Styckets typ.
    pub r#type: String,
}

/// Ett stycke som innehåller ett fel.
///
/// Go: `gin.H{"type": "error", "error": ...}`.
#[derive(Debug, serde::Serialize)]
pub struct FelStycke {
    /// Felmeddelandet.
    pub error: String,
    /// Styckets typ.
    pub r#type: String,
}

/// Ett stycke som markerar att svaret är klart.
///
/// Go: `gin.H{"type": "done"}`.
#[derive(Debug, serde::Serialize)]
pub struct KlartStycke {
    /// Styckets typ.
    pub r#type: String,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn resonemang_extraheras_ur_texten() {
        let text = "före <think> resonemang </think> efter";

        // Go behåller det som står mellan taggarna.
        assert_eq!(extrahera_resonemang(text), " resonemang ");
        assert_eq!(ta_bort_resonemang(text), "före  efter");
    }

    #[test]
    fn flera_resonemangsblock_görs_till_ett() {
        let text = "<think>ett</think> text <think>två</think>";

        assert_eq!(extrahera_resonemang(text), "ett\n\ntvå");
        assert_eq!(ta_bort_resonemang(text), " text ");
    }

    #[test]
    fn malkontext_formateras_som_go() {
        let mål = GoalContext {
            id: 7,
            title: "Skriv dokumentation".to_string(),
            description: "Allt om API:et".to_string(),
        };

        assert_eq!(
            formatera_målkontext(&mål),
            "We are discussing this Plan goal:\n\nGOAL #7: Skriv dokumentation\n\nGOAL DESCRIPTION:\nAllt om API:et"
        );
    }

    #[test]
    fn miljovariabeln_for_leverantor_foljer_go() {
        assert_eq!(leverantörsnyckel_env("minimax"), "MINIMAX_API_KEY");
        assert_eq!(
            leverantörsnyckel_env("agentgateway"),
            "AGENTGATEWAY_API_KEY"
        );
        assert_eq!(leverantörsnyckel_env("egen"), "egen_API_KEY");
    }

    #[test]
    fn max_turner_faller_tillbaka_pa_standard() {
        // Utan AGENT_MAX_TURNS används standardvärdet.
        assert!(max_turner() >= 1);
    }
}
