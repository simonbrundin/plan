//! Leverantörer av språkmodeller.
//!
//! Go-agentens `internal/core/provider`. Leverantörerna strömmar svar
//! från språkmodeller - MiniMax och OpenAI-liknande API:er.

pub mod anthropic;
pub mod minimax;
pub mod openai;

use std::future::Future;
use std::pin::Pin;
use std::sync::Arc;

use serde::{Deserialize, Serialize};
use tokio::sync::mpsc;

use crate::messages::{ContentBlock, Message, Usage};

/// Fel från en leverantör.
#[derive(Debug, thiserror::Error)]
pub enum ProviderError {
    /// Anropet misslyckades.
    #[error("{0}")]
    Request(String),
    /// Leverantören svarade med ett fel.
    #[error("API error: {0}")]
    Api(String),
    /// Ingen leverantör är konfigurerad.
    #[error("no LLM provider configured: set MINIMAX_API_KEY or OPENAI_API_KEY")]
    NotConfigured,
    /// Alla leverantörer misslyckades.
    #[error("all providers failed: {0}")]
    AllFailed(String),
}

/// En modell - leverantör, namn och nyckel.
#[derive(Debug, Clone, Default)]
pub struct Model {
    /// Leverantörens namn, som "minimax" eller "openai".
    pub provider: String,
    /// Modellens namn.
    pub name: String,
    /// API-nyckeln.
    pub api_key: String,
    /// En egen adress, om den ska användas i stället för standard.
    pub endpoint: String,
    /// Modell-specifika parametrar.
    pub parameters: std::collections::HashMap<String, serde_json::Value>,
}

/// Hur mycket leverantören ska tänka efter.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum ThinkingLevel {
    /// Inget resonemang.
    Off,
    /// Minimalt resonemang.
    Minimal,
    /// Lite resonemang.
    Low,
    /// Måttligt resonemang.
    Medium,
    /// Mycket resonemang.
    High,
    /// Extremt resonemang.
    XHigh,
    /// Maximalt resonemang.
    Max,
}

/// Alternativ för strömmad generering.
#[derive(Debug, Clone, Default)]
pub struct StreamingOptions {
    /// Temperaturen.
    pub temperature: Option<f64>,
    /// Maximalt antal tokens.
    pub max_tokens: Option<i64>,
    /// Hur mycket leverantören ska tänka efter.
    pub thinking: Option<ThinkingLevel>,
    /// Systemprompten.
    pub system_prompt: String,
    /// Ber om att tokenanvändningen ska följa med svaret.
    ///
    /// Vissa leverantörer ger inte användningen i ett strömmat svar - då
    /// används det vanliga svaret i stället. Det kostar lite mer
    /// fördröjning men ger riktiga siffror att räkna kostnad med.
    pub usage: bool,
}

/// Ett verktyg som modellen får använda.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ToolDefinition {
    /// Verktygets namn.
    pub name: String,
    /// Beskrivningen av verktyget.
    pub description: String,
    /// Schemat för argumenten.
    #[serde(rename = "inputSchema")]
    pub input_schema: serde_json::Value,
}

/// Resultatet av ett verktygsanrop.
#[derive(Debug, Clone, Default)]
pub struct ToolResult {
    /// Innehållet i resultatet.
    pub content: Vec<ContentBlock>,
    /// Extra detaljer.
    pub details: Option<serde_json::Value>,
    /// Om resultatet var ett fel.
    pub is_error: bool,
    /// Tokenanvändningen.
    pub usage: Option<Usage>,
    /// Om agenten ska avsluta efter verktyget.
    pub terminate: bool,
}

/// Ett meddelande som byggs upp under en ström.
#[derive(Debug, Clone, Default)]
pub struct PartialAssistant {
    /// Innehållet hittills.
    pub content: Vec<ContentBlock>,
    /// Varför svaret slutade.
    pub stop_reason: String,
    /// Tokenanvändningen.
    pub usage: Option<Usage>,
    /// Tidpunkten i millisekunder.
    pub timestamp: i64,
}

impl PartialAssistant {
    /// Skapar en tom delvis assistent.
    pub fn new() -> Self {
        Self {
            content: Vec::new(),
            stop_reason: String::new(),
            usage: None,
            timestamp: crate::messages::nu(),
        }
    }

    /// Gör om till ett meddelande.
    pub fn to_message(&self) -> Message {
        Message::Assistant {
            content: self.content.clone(),
            stop_reason: self.stop_reason.clone(),
            timestamp: self.timestamp,
            usage: self.usage.clone(),
        }
    }
}

/// En händelse i strömmen.
#[derive(Debug, Clone)]
pub enum StreamEvent {
    /// Strömmen har börjat.
    Start {
        /// Det delvisa meddelandet.
        partial: PartialAssistant,
    },
    /// Ny text.
    TextDelta {
        /// Texten som kom.
        delta: String,
        /// Det delvisa meddelandet.
        partial: PartialAssistant,
    },
    /// Ett verktygsanrop kom.
    ToolCallDelta {
        /// Anropets id.
        id: String,
        /// Verktygets namn.
        name: String,
        /// Argumenten hittills.
        arguments: String,
        /// Det delvisa meddelandet.
        partial: PartialAssistant,
    },
    /// Svaret är klart.
    Done {
        /// Det färdiga meddelandet.
        message: Message,
    },
    /// Något gick fel.
    Error {
        /// Felmeddelandet.
        error: String,
        /// Det delvisa meddelandet.
        message: Message,
    },
}

/// En leverantör av språkmodeller.
pub trait LlmProvider: Send + Sync {
    /// Leverantörens namn.
    fn name(&self) -> &str;

    /// Startar en ström.
    fn stream(
        &self,
        modell: Model,
        meddelanden: Vec<Message>,
        verktyg: Vec<ToolDefinition>,
        opts: StreamingOptions,
    ) -> Pin<Box<dyn Future<Output = Result<mpsc::Receiver<StreamEvent>, ProviderError>> + Send + '_>>;
}

/// En leverantör som delas mellan flera.
pub type SharedProvider = Arc<dyn LlmProvider>;

// --- Hjälpfunktioner ---

/// Gör om innehållsblocken till text.
pub fn konvertera_text_innehåll(blocker: &[ContentBlock]) -> String {
    crate::messages::text_innehåll(blocker)
}

/// Gör om ett assistentmeddelande till OpenAI-format.
///
/// Returnerar texten och verktygsanropen.
pub fn konvertera_assistent_innehåll(
    blocker: &[ContentBlock],
) -> (Option<String>, Vec<serde_json::Value>) {
    let mut text: Option<String> = None;
    let mut anrop = Vec::new();

    for block in blocker {
        match block {
            ContentBlock::Text { text: stycke } => {
                text = Some(match text {
                    Some(existing) => existing + stycke,
                    None => stycke.clone(),
                });
            }
            ContentBlock::ToolCall {
                id,
                name,
                arguments,
            } => anrop.push(serde_json::json!({
                "type": "function",
                "id": id,
                "function": {
                    "name": name,
                    "arguments": formatera_argument(arguments),
                },
            })),
            _ => {}
        }
    }

    (text, anrop)
}

/// Formaterar argumenten som JSON-text.
pub fn formatera_argument(argument: &serde_json::Value) -> String {
    match argument {
        serde_json::Value::Null => "{}".to_string(),
        value => value.to_string(),
    }
}

/// Tolkar JSON-argumenten från ett verktygsanrop.
///
/// Trasig JSON ger `_raw` med den råa texten - så att inget går
/// förlorat.
pub fn tolka_json_argument(argument: &str) -> serde_json::Value {
    if argument.is_empty() || argument == "{}" {
        return serde_json::json!({});
    }

    match serde_json::from_str::<serde_json::Value>(argument) {
        Ok(value) if value.is_object() => value,
        _ => serde_json::json!({ "_raw": argument }),
    }
}

/// Plockar ut ett verktygsanrop ur ett stycke.
///
/// Vid strömmning lagras argumenten rått i `_raw_args`.
pub fn extrahera_verktygsanrop(anrop: &serde_json::Value) -> Option<crate::messages::ContentBlock> {
    let index = anrop.get("index").and_then(|v| v.as_i64()).unwrap_or(0);
    let mut id = anrop
        .get("id")
        .and_then(|v| v.as_str())
        .unwrap_or_default()
        .to_string();
    let name = anrop
        .get("function")
        .and_then(|f| f.get("name"))
        .and_then(|v| v.as_str())
        .unwrap_or_default()
        .to_string();
    let argument = anrop
        .get("function")
        .and_then(|f| f.get("arguments"))
        .and_then(|v| v.as_str())
        .unwrap_or_default()
        .to_string();

    if name.is_empty() {
        return None;
    }

    if id.is_empty() {
        id = format!("call_{index}_{name}");
    }

    Some(ContentBlock::ToolCall {
        id,
        name,
        arguments: serde_json::json!({ "_raw_args": argument }),
    })
}

/// Lägger till argumenten i ett pågående verktygsanrop.
pub fn samla_verktygsargument(anrop: &mut crate::messages::ContentBlock, nya_argument: &str) {
    if let ContentBlock::ToolCall { arguments, .. } = anrop {
        let befintliga = arguments
            .get("_raw_args")
            .and_then(|v| v.as_str())
            .unwrap_or_default()
            .to_string();

        arguments["_raw_args"] = serde_json::Value::String(befintliga + nya_argument);
    }
}

/// Tolkar de samlade argumenten.
pub fn tolka_verktygsargument(anrop: &mut crate::messages::ContentBlock) -> serde_json::Value {
    if let ContentBlock::ToolCall { arguments, .. } = anrop {
        let Some(rått) = arguments.get("_raw_args").and_then(|v| v.as_str()) else {
            return arguments.clone();
        };

        let tolkat = tolka_json_argument(rått);

        if tolkat.get("_raw").is_some() {
            // JSON:en var inte komplett - det som finns lämnas som det är.
            return tolkat;
        }

        *arguments = tolkat.clone();

        return tolkat;
    }

    serde_json::json!({})
}

/// Gör om meddelandena till OpenAI-format.
pub fn konvertera_meddelanden(meddelanden: &[Message]) -> Vec<serde_json::Value> {
    let mut resultat = Vec::new();

    for meddelande in meddelanden {
        if let Some(värde) = meddelande_till_json(meddelande) {
            resultat.push(värde);
        }
    }

    resultat
}

/// Gör om ett meddelande till OpenAI-format.
fn meddelande_till_json(meddelande: &Message) -> Option<serde_json::Value> {
    match meddelande {
        Message::User { content, .. } => {
            let text = konvertera_text_innehåll(content);
            if text.is_empty() {
                return None;
            }

            Some(serde_json::json!({ "role": "user", "content": text }))
        }
        Message::Assistant {
            content,
            stop_reason: _,
            timestamp: _,
            usage: _,
        } => {
            let (text, anrop) = konvertera_assistent_innehåll(content);

            let mut meddelande = serde_json::Map::new();
            meddelande.insert(
                "role".to_string(),
                serde_json::Value::String("assistant".to_string()),
            );

            if let Some(text) = text {
                meddelande.insert("content".to_string(), serde_json::Value::String(text));
            }

            if !anrop.is_empty() {
                meddelande.insert("tool_calls".to_string(), serde_json::Value::Array(anrop));
            }

            Some(serde_json::Value::Object(meddelande))
        }
        Message::ToolResult {
            tool_call_id,
            content,
            ..
        } => {
            let text = content
                .iter()
                .find_map(|block| match block {
                    ContentBlock::Text { text } => Some(text.clone()),
                    _ => None,
                })
                .unwrap_or_default();

            Some(serde_json::json!({
                "role": "tool",
                "tool_call_id": tool_call_id,
                "content": text,
            }))
        }
        Message::System { content, .. } => {
            let text = konvertera_text_innehåll(content);
            if text.is_empty() {
                return None;
            }

            Some(serde_json::json!({ "role": "system", "content": text }))
        }
    }
}

/// Gör om verktygsdefinitionerna till OpenAI-format.
pub fn verktyg_till_openai_format(verktyg: &[ToolDefinition]) -> Vec<serde_json::Value> {
    verktyg
        .iter()
        .map(|verktyg| {
            // Ett verktyg utan schema måste ändå ha ett objekt-schema -
            // MiniMax avvisar "parameters": null.
            let schema = if verktyg.input_schema.is_null() {
                serde_json::json!({ "type": "object", "properties": {} })
            } else {
                verktyg.input_schema.clone()
            };

            serde_json::json!({
                "type": "function",
                "function": {
                    "name": verktyg.name,
                    "description": verktyg.description,
                    "parameters": schema,
                },
            })
        })
        .collect()
}

/// Hanterar ett stycke med ett verktygsanrop.
///
/// Finns anropet redan läggs argumenten ihop - annars läggs det till.
pub fn hantera_verktygsanrop(
    anrop: &serde_json::Value,
    partial: &mut PartialAssistant,
) -> Option<crate::messages::ContentBlock> {
    let block = extrahera_verktygsanrop(anrop)?;

    if let ContentBlock::ToolCall { id, .. } = &block {
        let id = id.clone();

        for befintligt in partial.content.iter_mut() {
            if let ContentBlock::ToolCall {
                id: befintligt_id, ..
            } = befintligt
            {
                if befintligt_id == &id {
                    let argument = match &block {
                        ContentBlock::ToolCall { arguments, .. } => arguments
                            .get("_raw_args")
                            .and_then(|v| v.as_str())
                            .unwrap_or_default()
                            .to_string(),
                        _ => String::new(),
                    };

                    samla_verktygsargument(befintligt, &argument);

                    return Some(befintligt.clone());
                }
            }
        }
    }

    partial.content.push(block.clone());

    Some(block)
}

/// Om användningen innehåller några tokens.
pub fn har_användning(användning: Option<&Usage>) -> bool {
    användning.is_some_and(|användning| {
        användning.total_tokens > 0 || användning.input_tokens > 0 || användning.output_tokens > 0
    })
}

/// Gör om leverantörens stoppreason till agentens.
pub fn normalisera_stoppreason(reason: &str) -> String {
    match reason {
        "tool_calls" | "tool_use" => "tool_use".to_string(),
        "length" => "max_tokens".to_string(),
        "stop" => "end_turn".to_string(),
        reason => reason.to_string(),
    }
}

// --- Upprepning med exponentiell bakoff ---

/// Inställningar för upprepning.
#[derive(Debug, Clone, Copy)]
pub struct RetryConfig {
    /// Maximalt antal försök.
    pub max_attempts: u32,
    /// Den första väntetiden.
    pub base_delay: std::time::Duration,
    /// Den längsta väntetiden.
    pub max_delay: std::time::Duration,
    /// Multiplikatorn mellan försöken.
    pub multiplier: f64,
}

impl Default for RetryConfig {
    fn default() -> Self {
        Self {
            max_attempts: 3,
            base_delay: std::time::Duration::from_secs(1),
            max_delay: std::time::Duration::from_secs(30),
            multiplier: 2.0,
        }
    }
}

/// Om ett fel går att försöka igen.
pub fn är_återförsökbart(fel: &str) -> bool {
    let fel = fel.to_lowercase();

    // Taktbegränsning.
    if fel.contains("rate limit")
        || fel.contains("429")
        || fel.contains("too many requests")
        || fel.contains("rate_limit_exceeded")
    {
        return true;
    }

    // Nätverksfel.
    if fel.contains("connection reset")
        || fel.contains("connection refused")
        || fel.contains("timeout")
        || fel.contains("temporary failure")
        || fel.contains("i/o timeout")
        || fel.contains("network")
    {
        return true;
    }

    // Serverfel.
    fel.contains("500")
        || fel.contains("502")
        || fel.contains("503")
        || fel.contains("504")
        || fel.contains("internal server error")
}

/// Kör ett anrop igen om felet går att försöka igen.
pub async fn försök_igen<F, Fut, T>(config: RetryConfig, anrop: F) -> Result<T, ProviderError>
where
    F: Fn() -> Fut,
    Fut: std::future::Future<Output = Result<T, ProviderError>>,
{
    let mut sista_felet: Option<ProviderError> = None;

    for försök in 0..config.max_attempts {
        match anrop().await {
            Ok(värde) => return Ok(värde),
            Err(fel) => {
                if !är_återförsökbart(&fel.to_string()) {
                    return Err(fel);
                }

                sista_felet = Some(fel);
            }
        }

        // Ingen väntan efter det sista försöket.
        if försök + 1 >= config.max_attempts {
            break;
        }

        let mut fördröjning =
            config.base_delay.as_secs_f64() * config.multiplier.powi(försök as i32);
        if fördröjning > config.max_delay.as_secs_f64() {
            fördröjning = config.max_delay.as_secs_f64();
        }

        // Slumpmässig avvikelse på ±25 %.
        let avvikelse = fördröjning * 0.25 * (rand::random::<f64>() * 2.0 - 1.0);
        let väntetid = std::time::Duration::from_secs_f64(fördröjning + avvikelse);

        tracing::warn!(
            "Försök {}/{} igen efter {:?}: {:?}",
            försök + 1,
            config.max_attempts,
            väntetid,
            sista_felet
        );

        tokio::time::sleep(väntetid).await;
    }

    Err(ProviderError::Request(format!(
        "after {} attempts: {}",
        config.max_attempts,
        sista_felet
            .map(|fel| fel.to_string())
            .unwrap_or_else(|| "okänt fel".to_string())
    )))
}

/// Chef för flera leverantörer, med reserv mellan dem.
#[derive(Default)]
pub struct ProviderManager {
    /// Leverantörerna i prioritetsordning.
    leverantörer: Vec<(String, SharedProvider)>,
}

impl ProviderManager {
    /// Skapar en chef utan leverantörer.
    pub fn new() -> Self {
        Self::default()
    }

    /// Lägger till en leverantör.
    pub fn add(&mut self, namn: impl Into<String>, leverantör: SharedProvider) {
        self.leverantörer.push((namn.into(), leverantör));
    }

    /// Leverantörernas namn.
    pub fn namn(&self) -> Vec<String> {
        self.leverantörer
            .iter()
            .map(|(namn, _)| namn.clone())
            .collect()
    }

    /// Startar en ström - och provar nästa leverantör om denne misslyckas.
    pub async fn stream(
        &self,
        modell: &Model,
        meddelanden: &[Message],
        verktyg: &[ToolDefinition],
        opts: &StreamingOptions,
    ) -> Result<(String, mpsc::Receiver<StreamEvent>), ProviderError> {
        let modell = modell.clone();
        let meddelanden = meddelanden.to_vec();
        let verktyg = verktyg.to_vec();
        let opts = opts.clone();

        let mut sista_felet: Option<ProviderError> = None;

        for (namn, leverantör) in &self.leverantörer {
            match leverantör
                .stream(
                    modell.clone(),
                    meddelanden.clone(),
                    verktyg.clone(),
                    opts.clone(),
                )
                .await
            {
                Ok(händelser) => {
                    tracing::info!("Använder leverantören {namn}");

                    return Ok((namn.clone(), händelser));
                }
                Err(fel) => {
                    tracing::warn!("Leverantören {namn} misslyckades: {fel} - provar nästa");
                    sista_felet = Some(fel);
                }
            }
        }

        match sista_felet {
            Some(fel) => Err(ProviderError::AllFailed(fel.to_string())),
            None => Err(ProviderError::NotConfigured),
        }
    }
}

/// Skapar leverantörer utifrån miljövariablerna.
///
/// Go: `internal/core/provider/factory.go`.
pub struct Factory;

impl Factory {
    /// Skapar en leverantör och en modell utifrån miljövariablerna.
    pub fn skapa_från_miljö() -> Result<(SharedProvider, Model), ProviderError> {
        if let Some(nyckel) = env_nyckel("MINIMAX_API_KEY") {
            let namn = env_nyckel("MINIMAX_MODEL").unwrap_or_else(|| "MiniMax-M2.7".to_string());
            let adress = env_nyckel("MINIMAX_ENDPOINT")
                .unwrap_or_else(|| "https://api.minimax.io/v1".to_string());

            let leverantör = minimax::MinimaxProvider::new(adress);

            return Ok((
                Arc::new(leverantör),
                Model {
                    provider: "minimax".to_string(),
                    name: namn,
                    api_key: nyckel,
                    endpoint: String::new(),
                    parameters: Default::default(),
                },
            ));
        }

        if let Some(nyckel) = env_nyckel("OPENAI_API_KEY") {
            let namn = env_nyckel("OPENAI_MODEL").unwrap_or_else(|| "gpt-4o".to_string());
            let adress = env_nyckel("OPENAI_BASE_URL")
                .unwrap_or_else(|| "https://api.openai.com/v1".to_string());

            let leverantör = openai::OpenAiProvider::new(adress);

            return Ok((
                Arc::new(leverantör),
                Model {
                    provider: "openai".to_string(),
                    name: namn,
                    api_key: nyckel,
                    endpoint: String::new(),
                    parameters: Default::default(),
                },
            ));
        }

        Err(ProviderError::NotConfigured)
    }

    /// Skapar en chef med alla leverantörer som är konfigurerade.
    pub fn skapa_chef() -> Result<ProviderManager, ProviderError> {
        let mut chef = ProviderManager::new();

        if let Some(nyckel) = env_nyckel("MINIMAX_API_KEY") {
            let adress = env_nyckel("MINIMAX_ENDPOINT")
                .unwrap_or_else(|| "https://api.minimax.io/v1".to_string());

            let _ = nyckel;
            chef.add("minimax", Arc::new(minimax::MinimaxProvider::new(adress)));
        }

        if env_nyckel("OPENAI_API_KEY").is_some() {
            let adress = env_nyckel("OPENAI_BASE_URL")
                .unwrap_or_else(|| "https://api.openai.com/v1".to_string());

            chef.add("openai", Arc::new(openai::OpenAiProvider::new(adress)));
        }

        if chef.leverantörer.is_empty() {
            return Err(ProviderError::NotConfigured);
        }

        Ok(chef)
    }
}

/// Läser en miljövariabel - tomma värden räknas som saknade.
fn env_nyckel(namn: &str) -> Option<String> {
    std::env::var(namn).ok().filter(|värde| !värde.is_empty())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn trasiga_argument_hamtas_undan() {
        let värde = tolka_json_argument("{\"a\": 1");
        assert_eq!(värde.get("_raw").unwrap(), "{\"a\": 1");
    }

    #[test]
    fn tomma_argument_blir_tomt_objekt() {
        assert_eq!(tolka_json_argument(""), serde_json::json!({}));
        assert_eq!(tolka_json_argument("{}"), serde_json::json!({}));
    }

    #[test]
    fn verktygsanrop_samlar_argument() {
        let mut anrop = extrahera_verktygsanrop(&serde_json::json!({
            "index": 0,
            "id": "anrop-1",
            "function": { "name": "läs", "arguments": "{\"fil\":" },
        }))
        .unwrap();

        samla_verktygsargument(&mut anrop, "\"a.txt\"}");

        let argument = tolka_verktygsargument(&mut anrop);
        assert_eq!(argument.get("fil").unwrap(), "a.txt");
    }

    #[test]
    fn taktbegransning_ar_aterforsokbar() {
        assert!(är_återförsökbart("rate limit exceeded"));
        assert!(är_återförsökbart("503 Service Unavailable"));
        assert!(!är_återförsökbart("invalid api key"));
    }

    #[test]
    fn stoppreason_normaliseras() {
        assert_eq!(normalisera_stoppreason("tool_calls"), "tool_use");
        assert_eq!(normalisera_stoppreason("length"), "max_tokens");
        assert_eq!(normalisera_stoppreason("stop"), "end_turn");
    }
}
