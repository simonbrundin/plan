//! Anthropic-kompatibla leverantören.
//!
//! Go: `internal/api/llm`. Anropar Anthropics messages-API - eller
//! MiniMax:s Anthropic-kompatibla adress.
//!
//! Skillnaden mot leverantörerna i [`super`] är att historiken inte
//! skickas med - bara det senaste användarmeddelandet - och att
//! resonemanget kommer i egna stycken.

use std::time::Duration;

use reqwest::Client;
use serde_json::Value;
use tokio::sync::mpsc;

/// Klienten mot ett Anthropic-kompatibelt API.
pub struct AnthropicClient {
    /// HTTP-klienten.
    klient: Client,
    /// Basadressen, som "https://api.minimax.io/anthropic".
    basadress: String,
    /// API-nyckeln.
    nyckel: String,
}

/// Ett stycke som kommer från strömmen.
#[derive(Debug, Clone)]
pub enum AnthropicStycke {
    /// Text som kommit.
    Text(String),
    /// Resonemang som kommit.
    Tänking(String),
    /// Strömmen är klar.
    Klar {
        /// Anledningen till att svaret slutade.
        reason: String,
        /// Tiden det tog, i millisekunder.
        latency_ms: i64,
    },
}

impl AnthropicClient {
    /// Skapar en klient mot en adress.
    pub fn new(basadress: impl Into<String>, nyckel: impl Into<String>) -> Self {
        Self {
            klient: Client::builder()
                .timeout(Duration::from_secs(120))
                .build()
                .unwrap_or_default(),
            basadress: basadress.into(),
            nyckel: nyckel.into(),
        }
    }

    /// Skapar en klient utifrån miljön.
    ///
    /// Go: `llm.DefaultClient`. MiniMax används om dess nyckel finns,
    /// annars Anthropic.
    pub fn från_miljö() -> Self {
        if let Ok(nyckel) = std::env::var("MINIMAX_API_KEY") {
            if !nyckel.is_empty() {
                return Self::new("https://api.minimax.io/anthropic", nyckel);
            }
        }

        if let Ok(nyckel) = std::env::var("ANTHROPIC_API_KEY") {
            if !nyckel.is_empty() {
                return Self::new("https://api.anthropic.com", nyckel);
            }
        }

        Self::new("https://api.minimax.io/anthropic", String::new())
    }

    /// Leverantörens namn.
    pub fn namn(&self) -> &'static str {
        if self.basadress.contains("minimax") {
            "minimax"
        } else {
            "anthropic"
        }
    }

    /// Strömmar ett svar för ett meddelande.
    ///
    /// `meddelande` är det senaste användarmeddelandet - historiken
    /// skickas inte med, precis som i Go.
    pub async fn strömma(
        &self,
        modell: &str,
        meddelande: &str,
        temperatur: f64,
        sändare: mpsc::Sender<AnthropicStycke>,
    ) -> Result<(), String> {
        let modell = if modell.is_empty() {
            "MiniMax-M2.7".to_string()
        } else {
            modell.to_string()
        };

        let mut kropp = serde_json::Map::new();
        kropp.insert("model".to_string(), Value::String(modell));
        kropp.insert("max_tokens".to_string(), Value::Number(8192.into()));
        kropp.insert(
            "messages".to_string(),
            serde_json::json!([{ "role": "user", "content": meddelande }]),
        );
        kropp.insert("stream".to_string(), Value::Bool(true));

        if temperatur > 0.0 {
            kropp.insert(
                "temperature".to_string(),
                serde_json::Number::from_f64(temperatur)
                    .map(Value::Number)
                    .unwrap_or(Value::Null),
            );
        }

        let svar = self
            .klient
            .post(format!("{}/v1/messages", self.basadress))
            .header("x-api-key", &self.nyckel)
            .header("anthropic-version", "2023-06-01")
            .header("content-type", "application/json")
            .body(serde_json::to_vec(&Value::Object(kropp)).unwrap_or_default())
            .send()
            .await
            .map_err(|fel| fel.to_string())?;

        if !svar.status().is_success() {
            let status = svar.status();
            let text = svar.text().await.unwrap_or_default();

            return Err(format!("{status} - {text}"));
        }

        let start = std::time::Instant::now();

        let mut buffer = String::new();
        let mut svar = svar;

        loop {
            let stycke = match svar.chunk().await {
                Ok(stycke) => stycke,
                Err(fel) => return Err(fel.to_string()),
            };

            let Some(stycke) = stycke else { break };

            buffer.push_str(&String::from_utf8_lossy(&stycke));

            while let Some(position) = buffer.find('\n') {
                let rad = buffer[..position].to_string();
                buffer = buffer[position + 1..].to_string();

                let Some(data) = rad.strip_prefix("data: ") else {
                    continue;
                };

                let Ok(tolkad) = serde_json::from_str::<Value>(data) else {
                    continue;
                };

                match tolkad.get("type").and_then(|t| t.as_str()) {
                    Some("content_block_delta") => {
                        let delta = tolkad.get("delta");

                        match delta.and_then(|d| d.get("type")).and_then(|t| t.as_str()) {
                            Some("text_delta") => {
                                if let Some(text) =
                                    delta.and_then(|d| d.get("text")).and_then(|t| t.as_str())
                                {
                                    if sändare
                                        .send(AnthropicStycke::Text(text.to_string()))
                                        .await
                                        .is_err()
                                    {
                                        return Ok(());
                                    }
                                }
                            }
                            Some("thinking_delta") => {
                                if let Some(text) =
                                    delta.and_then(|d| d.get("text")).and_then(|t| t.as_str())
                                {
                                    if sändare
                                        .send(AnthropicStycke::Tänking(text.to_string()))
                                        .await
                                        .is_err()
                                    {
                                        return Ok(());
                                    }
                                }
                            }
                            _ => {}
                        }
                    }
                    Some("message_delta") => {
                        let reason = tolkad
                            .get("delta")
                            .and_then(|d| d.get("stop_reason"))
                            .and_then(|r| r.as_str())
                            .unwrap_or_default();

                        if !reason.is_empty() {
                            let _ = sändare
                                .send(AnthropicStycke::Klar {
                                    reason: reason.to_string(),
                                    latency_ms: start.elapsed().as_millis() as i64,
                                })
                                .await;

                            return Ok(());
                        }
                    }
                    _ => {}
                }
            }
        }

        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn minmax_väljs_för_minimax_nyckeln() {
        let klient = AnthropicClient::new("https://api.minimax.io/anthropic", "nyckel");

        assert_eq!(klient.namn(), "minimax");
    }

    #[test]
    fn anthropic_väljs_för_anthropic_adressen() {
        let klient = AnthropicClient::new("https://api.anthropic.com", "nyckel");

        assert_eq!(klient.namn(), "anthropic");
    }
}
