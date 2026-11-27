//! MiniMax-leverantören.
//!
//! Go: `internal/core/provider/minimax.go`. Anropar MiniMax:s
//! OpenAI-kompatibla API.

use std::time::Duration;

use reqwest::Client;
use serde_json::Value;
use tokio::sync::mpsc;

use super::{
    hantera_verktygsanrop, har_användning, konvertera_meddelanden, normalisera_stoppreason,
    tolka_json_argument, verktyg_till_openai_format, LlmProvider, Model, PartialAssistant,
    ProviderError, StreamEvent, StreamingOptions, ToolDefinition,
};
use crate::messages::{ContentBlock, Message, Usage};

/// Leverantören för MiniMax.
pub struct MinimaxProvider {
    /// HTTP-klienten.
    klient: Client,
    /// Basadressen till API:et.
    endpoint: String,
}

/// Hur många extra stycken som läses efter att modellen angett att den
/// är klar. Det ska räcka för att användningen ska hinna komma.
const MAX_CHUNKS_AFTER_STOP: u32 = 5;

impl MinimaxProvider {
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
        // Strömmad respons när det inte finns verktyg. När användningen
        // efterfrågas används i stället ett vanligt svar - ett strömmat
        // svar från MiniMax saknar användningen.
        let ström = verktyg.is_empty() && !opts.usage;

        let mut kropp = serde_json::Map::new();
        kropp.insert("model".to_string(), Value::String(modell.name.clone()));
        kropp.insert("stream".to_string(), Value::Bool(ström));
        kropp.insert(
            "messages".to_string(),
            Value::Array(konvertera_meddelanden(meddelanden)),
        );

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

    /// Skickar förfrågan.
    async fn skicka_förfrågan(
        &self,
        endpoint: &str,
        nyckel: &str,
        kropp: Vec<u8>,
    ) -> Result<reqwest::Response, ProviderError> {
        let svar = self
            .klient
            .post(endpoint)
            .header("Content-Type", "application/json")
            .header("Authorization", format!("Bearer {nyckel}"))
            .body(kropp)
            .send()
            .await
            .map_err(|fel| ProviderError::Request(format!("request failed: {fel}")))?;

        let status = svar.status();

        if status != reqwest::StatusCode::OK {
            let text = svar.text().await.unwrap_or_default();

            return Err(ProviderError::Api(format!("{status} - {text}")));
        }

        Ok(svar)
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
                cache_tokens: användning
                    .get("cached_tokens")
                    .and_then(|v| v.as_i64())
                    .unwrap_or(0),
            };

            // En tom användning ska inte skriva över en riktig.
            if har_användning(Some(&ny)) || partial.usage.is_none() {
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

        let mut sista_händelse: Option<StreamEvent> = None;

        let Some(val) = tolkat.get("choices").and_then(|c| c.as_array()) else {
            return (sista_händelse, None);
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

                sista_händelse = Some(StreamEvent::TextDelta {
                    delta: innehåll.to_string(),
                    partial: partial.clone(),
                });

                continue;
            }

            if let Some(anrop) = delta
                .and_then(|d| d.get("tool_calls"))
                .and_then(|v| v.as_array())
            {
                for anrop in anrop {
                    if let Some(block) = hantera_verktygsanrop(anrop, partial) {
                        let (id, namn, argument) = match &block {
                            ContentBlock::ToolCall {
                                id,
                                name,
                                arguments,
                            } => (
                                id.clone(),
                                name.clone(),
                                arguments
                                    .get("_raw_args")
                                    .and_then(|v| v.as_str())
                                    .unwrap_or_default()
                                    .to_string(),
                            ),
                            _ => (String::new(), String::new(), String::new()),
                        };

                        sista_händelse = Some(StreamEvent::ToolCallDelta {
                            id,
                            name: namn,
                            arguments: argument,
                            partial: partial.clone(),
                        });
                    }
                }
            }

            if let Some(reason) = valg.get("finish_reason").and_then(|v| v.as_str()) {
                if !reason.is_empty() {
                    return (sista_händelse, Some(reason.to_string()));
                }
            }
        }

        (sista_händelse, None)
    }
}

impl LlmProvider for MinimaxProvider {
    fn name(&self) -> &str {
        "minimax"
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
                .skicka_förfrågan(&endpoint, &modell.api_key, kropp)
                .await?;

            let (sändare, mottagare) = mpsc::channel(100);

            let användning = opts.usage;
            let har_verktyg = !verktyg.is_empty();

            tokio::spawn(async move {
                if har_verktyg || användning {
                    hantera_icke_strömmat_svar(svar, sändare).await;
                } else {
                    hantera_strömmat_svar(svar, sändare).await;
                }
            });

            Ok(mottagare)
        })
    }
}

/// Läser ett strömmat svar i SSE-format.
async fn hantera_strömmat_svar(svar: reqwest::Response, sändare: mpsc::Sender<StreamEvent>) {
    let mut partial = PartialAssistant::new();
    let mut stoppreason_seen = String::new();
    let mut extra_stycken = 0u32;

    // Strömmen läses rad för rad - varje rad är ett "data: {...}"-stycke.
    let mut buffer = String::new();
    let sista_händelse: Option<StreamEvent> = None;
    let mut svar = svar;

    loop {
        let stycke = match svar.chunk().await {
            Ok(stycke) => stycke,
            Err(_) => break,
        };

        let Some(stycke) = stycke else { break };

        buffer.push_str(&String::from_utf8_lossy(&stycke));

        while let Some(position) = buffer.find('\n') {
            let rad = buffer[..position].to_string();
            buffer = buffer[position + 1..].to_string();

            let Some(data) = rad.strip_prefix("data: ") else {
                continue;
            };

            if data == "[DONE]" {
                if !partial.content.is_empty() {
                    partial.stop_reason = "end_turn".to_string();

                    let _ = sändare
                        .send(StreamEvent::Done {
                            message: partial.to_message(),
                        })
                        .await;
                }

                return;
            }

            let (händelse, stoppreason) = MinimaxProvider::tolka_stycke(data, &mut partial);

            if let Some(händelse) = händelse {
                // Händelsen skickas direkt - annars skulle bara den
                // sista delen av svaret nå konsumenten.
                let är_fel = matches!(händelse, StreamEvent::Error { .. });

                let _ = sändare.send(händelse).await;

                if är_fel {
                    return;
                }
            }

            let Some(stoppreason) = stoppreason else {
                continue;
            };

            partial.stop_reason = normalisera_stoppreason(&stoppreason);

            // Användningen kommer ibland i ett eget stycke efter det
            // sista innehållet. Därför läses streamen vidare tills
            // användningen kommer - annars blir kostnaden noll.
            if har_användning(partial.usage.as_ref()) {
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
    }

    if let Some(händelse) = sista_händelse {
        let _ = sändare.send(händelse).await;
    }

    if !partial.content.is_empty() {
        if stoppreason_seen.is_empty() {
            partial.stop_reason = "end_turn".to_string();
        } else {
            partial.stop_reason = normalisera_stoppreason(&stoppreason_seen);
        }

        let _ = sändare
            .send(StreamEvent::Done {
                message: partial.to_message(),
            })
            .await;
    }
}

/// Läser ett vanligt svar - det som används när verktyg eller
/// användning efterfrågas.
async fn hantera_icke_strömmat_svar(svar: reqwest::Response, sändare: mpsc::Sender<StreamEvent>) {
    let data = match svar.text().await {
        Ok(data) => data,
        Err(fel) => {
            let _ = sändare
                .send(StreamEvent::Error {
                    error: fel.to_string(),
                    message: Message::Assistant {
                        content: Vec::new(),
                        stop_reason: String::new(),
                        timestamp: crate::messages::nu(),
                        usage: None,
                    },
                })
                .await;

            return;
        }
    };

    let Ok(tolkat) = serde_json::from_str::<Value>(&data) else {
        return;
    };

    if let Some(meddelande) = tolkat
        .get("error")
        .and_then(|e| e.get("message"))
        .and_then(|v| v.as_str())
    {
        let _ = sändare
            .send(StreamEvent::Error {
                error: meddelande.to_string(),
                message: Message::Assistant {
                    content: Vec::new(),
                    stop_reason: String::new(),
                    timestamp: crate::messages::nu(),
                    usage: None,
                },
            })
            .await;

        return;
    }

    let Some(val) = tolkat
        .get("choices")
        .and_then(|c| c.as_array())
        .and_then(|c| c.first())
    else {
        return;
    };

    let mut partial = PartialAssistant::new();

    partial.stop_reason = normalisera_stoppreason(valg_finish(val).unwrap_or_default());

    // Användningen sparas på meddelandet, så att den kan rapporteras
    // vidare. Utan den går det inte att räkna ut vad körningen kostade.
    if let Some(användning) = tolkat.get("usage").and_then(|u| u.as_object()) {
        partial.usage = Some(Usage {
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
            cache_tokens: användning
                .get("cached_tokens")
                .and_then(|v| v.as_i64())
                .unwrap_or(0),
        });
    }

    if let Some(innehåll) = val
        .get("message")
        .and_then(|m| m.get("content"))
        .and_then(|v| v.as_str())
    {
        if !innehåll.is_empty() {
            partial.content.push(ContentBlock::Text {
                text: innehåll.to_string(),
            });

            let _ = sändare
                .send(StreamEvent::TextDelta {
                    delta: innehåll.to_string(),
                    partial: partial.clone(),
                })
                .await;
        }
    }

    if let Some(anrop) = val
        .get("message")
        .and_then(|m| m.get("tool_calls"))
        .and_then(|v| v.as_array())
    {
        for anrop in anrop {
            if let Some(block) = verktygsanrop_från_icke_strömmat(anrop) {
                partial.content.push(block);
            }
        }
    }

    let _ = sändare
        .send(StreamEvent::Done {
            message: partial.to_message(),
        })
        .await;
}

/// Läser `finish_reason` ur ett val.
fn valg_finish(val: &Value) -> Option<&str> {
    val.get("finish_reason").and_then(|v| v.as_str())
}

/// Gör om ett verktygsanrop från ett vanligt svar.
fn verktygsanrop_från_icke_strömmat(anrop: &Value) -> Option<ContentBlock> {
    let index = anrop.get("index").and_then(|v| v.as_i64()).unwrap_or(0);
    let mut id = anrop
        .get("id")
        .and_then(|v| v.as_str())
        .unwrap_or_default()
        .to_string();

    let funktion = anrop.get("function")?;

    let namn = funktion
        .get("name")
        .and_then(|v| v.as_str())
        .unwrap_or_default();

    if namn.is_empty() {
        return None;
    }

    if id.is_empty() {
        id = format!("call_{index}_{namn}");
    }

    let argument = funktion
        .get("arguments")
        .and_then(|v| v.as_str())
        .unwrap_or_default();

    Some(ContentBlock::ToolCall {
        id,
        name: namn.to_string(),
        arguments: tolka_json_argument(argument),
    })
}
