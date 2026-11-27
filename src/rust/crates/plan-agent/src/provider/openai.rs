//! OpenAI-leverantören.
//!
//! Go: `internal/core/provider/openai.go`. Anropar OpenAI-liknande API:er
//! som svarar med NDJSON.

use std::time::Duration;

use reqwest::Client;
use serde_json::Value;
use tokio::sync::mpsc;

use super::{
    konvertera_meddelanden, verktyg_till_openai_format, LlmProvider, Model, PartialAssistant,
    ProviderError, StreamEvent, StreamingOptions, ToolDefinition,
};
use crate::messages::{ContentBlock, Message, Usage};

/// Leverantören för OpenAI-liknande API:er.
pub struct OpenAiProvider {
    /// HTTP-klienten.
    klient: Client,
    /// Basadressen till API:et.
    endpoint: String,
}

/// Hur många extra stycken som läses efter att modellen angett att den
/// är klar.
const MAX_CHUNKS_AFTER_STOP: u32 = 5;

impl OpenAiProvider {
    /// Skapar en leverantör mot en adress.
    pub fn new(endpoint: impl Into<String>) -> Self {
        Self {
            klient: Client::builder()
                .timeout(Duration::from_secs(120))
                .build()
                .unwrap_or_default(),
            endpoint: endpoint.into(),
        }
    }

    /// Bygger förfrågan.
    fn bygg_förfrågan(
        &self,
        modell: &Model,
        meddelanden: &[Message],
        verktyg: &[ToolDefinition],
        opts: &StreamingOptions,
    ) -> (Vec<u8>, String) {
        let mut kropp = serde_json::Map::new();
        kropp.insert("model".to_string(), Value::String(modell.name.clone()));
        kropp.insert("stream".to_string(), Value::Bool(true));
        kropp.insert(
            "messages".to_string(),
            Value::Array(konvertera_meddelanden(meddelanden)),
        );

        // OpenAI-liknande leverantörer skickar användningen i ett eget
        // stycke när detta begärs.
        if opts.usage {
            kropp.insert(
                "stream_options".to_string(),
                serde_json::json!({ "include_usage": true }),
            );
        }

        if let Some(max_tokens) = opts.max_tokens {
            kropp.insert("max_tokens".to_string(), Value::from(max_tokens));
        }

        if let Some(temperatur) = opts.temperature {
            kropp.insert("temperature".to_string(), Value::from(temperatur));
        }

        if !verktyg.is_empty() {
            kropp.insert(
                "tools".to_string(),
                Value::Array(verktyg_till_openai_format(verktyg)),
            );
        }

        let mut endpoint = format!("{}/chat/completions", self.endpoint);
        if !modell.endpoint.is_empty() {
            endpoint = modell.endpoint.clone();
        }

        (
            serde_json::to_vec(&Value::Object(kropp)).unwrap_or_default(),
            endpoint,
        )
    }

    /// Läser ett stycke och returnerar händelsen och stoppreasonen.
    fn tolka_stycke(
        data: &str,
        partial: &mut PartialAssistant,
    ) -> (Option<StreamEvent>, Option<String>) {
        let Ok(tolkat) = serde_json::from_str::<Value>(data) else {
            return (None, None);
        };

        // Användningen sparas på meddelandet, så att den kan rapporteras
        // vidare. Utan den går det inte att räkna ut vad körningen kostade.
        if let Some(användning) = tolkat.get("usage").and_then(|u| u.as_object()) {
            let cache_tokens = användning
                .get("prompt_tokens_details")
                .and_then(|d| d.get("cached_tokens"))
                .and_then(|v| v.as_i64())
                .unwrap_or(0);

            let ny = Usage {
                input_tokens: användning
                    .get("prompt_tokens")
                    .and_then(|v| v.as_i64())
                    .unwrap_or(0),
                output_tokens: användning
                    .get("completion_tokens")
                    .and_then(|v| v.as_i64())
                    .unwrap_or(0),
                total_tokens: användning
                    .get("total_tokens")
                    .and_then(|v| v.as_i64())
                    .unwrap_or(0),
                cache_tokens,
            };

            // En tom användning ska inte skriva över en riktig.
            if super::har_användning(Some(&ny)) || partial.usage.is_none() {
                partial.usage = Some(ny);
            }
        }

        if let Some(meddelande) = tolkat
            .get("error")
            .and_then(|e| e.get("message"))
            .and_then(|v| v.as_str())
        {
            return (
                Some(StreamEvent::Error {
                    error: meddelande.to_string(),
                    message: partial.to_message(),
                }),
                None,
            );
        }

        let Some(val) = tolkat.get("choices").and_then(|c| c.as_array()) else {
            return (None, None);
        };

        for valg in val {
            let delta = valg.get("delta").and_then(|d| d.as_object());

            if let Some(innehåll) = delta
                .and_then(|d| d.get("content"))
                .and_then(|v| v.as_str())
            {
                partial.content.push(ContentBlock::Text {
                    text: innehåll.to_string(),
                });

                return (
                    Some(StreamEvent::TextDelta {
                        delta: innehåll.to_string(),
                        partial: partial.clone(),
                    }),
                    None,
                );
            }

            if let Some(anrop) = delta
                .and_then(|d| d.get("tool_calls"))
                .and_then(|v| v.as_array())
            {
                for anrop in anrop {
                    super::hantera_verktygsanrop(anrop, partial);
                }
            }

            if let Some(reason) = valg.get("finish_reason").and_then(|v| v.as_str()) {
                if !reason.is_empty() {
                    return (None, Some(reason.to_string()));
                }
            }
        }

        (None, None)
    }
}

impl LlmProvider for OpenAiProvider {
    fn name(&self) -> &str {
        "openai"
    }

    fn stream(
        &self,
        modell: Model,
        meddelanden: Vec<Message>,
        verktyg: Vec<ToolDefinition>,
        opts: StreamingOptions,
    ) -> std::pin::Pin<
        Box<
            dyn std::future::Future<Output = Result<mpsc::Receiver<StreamEvent>, ProviderError>>
                + Send
                + '_,
        >,
    > {
        Box::pin(async move {
            let (kropp, endpoint) = self.bygg_förfrågan(&modell, &meddelanden, &verktyg, &opts);

            let svar = self
                .klient
                .post(&endpoint)
                .header("Content-Type", "application/json")
                .header("Authorization", format!("Bearer {}", modell.api_key))
                .body(kropp)
                .send()
                .await
                .map_err(|fel| ProviderError::Request(format!("request failed: {fel}")))?;

            let status = svar.status();

            if status != reqwest::StatusCode::OK {
                let text = svar.text().await.unwrap_or_default();

                return Err(ProviderError::Api(format!("{status} - {text}")));
            }

            let (sändare, mottagare) = mpsc::channel(100);

            tokio::spawn(async move {
                hantera_strömmat_svar(svar, sändare).await;
            });

            Ok(mottagare)
        })
    }
}

/// Läser ett strömmat svar i NDJSON-format.
async fn hantera_strömmat_svar(svar: reqwest::Response, sändare: mpsc::Sender<StreamEvent>) {
    let mut partial = PartialAssistant::new();
    let mut stoppreason_seen = String::new();
    let mut extra_stycken = 0u32;

    let Ok(text) = svar.text().await else {
        return;
    };

    for rad in text.lines() {
        if rad.trim().is_empty() {
            continue;
        }

        let (händelse, stoppreason) = OpenAiProvider::tolka_stycke(rad, &mut partial);

        if let Some(händelse) = händelse {
            let _ = sändare.send(händelse).await;
        }

        let Some(stoppreason) = stoppreason else {
            continue;
        };

        partial.stop_reason = stoppreason.clone();

        // Användningen kommer ibland i ett eget stycke efter det sista
        // innehållet. Därför läses streamen vidare tills användningen
        // kommer - annars blir kostnaden noll.
        if super::har_användning(partial.usage.as_ref()) {
            let _ = sändare
                .send(StreamEvent::Done {
                    message: partial.to_message(),
                })
                .await;

            return;
        }

        stoppreason_seen = stoppreason;
        extra_stycken += 1;

        if extra_stycken > MAX_CHUNKS_AFTER_STOP {
            let _ = sändare
                .send(StreamEvent::Done {
                    message: partial.to_message(),
                })
                .await;

            return;
        }
    }

    if !partial.content.is_empty() {
        if stoppreason_seen.is_empty() {
            partial.stop_reason = "end_turn".to_string();
        } else {
            partial.stop_reason = stoppreason_seen;
        }

        let _ = sändare
            .send(StreamEvent::Done {
                message: partial.to_message(),
            })
            .await;
    }
}
