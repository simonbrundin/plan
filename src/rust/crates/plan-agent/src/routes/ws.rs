//! Agentens WebSocket-chatt.
//!
//! Go: `internal/api/handlers/websocket.go` - `HandleChat`.
//!
//! Observera att Go även har en agentvariant i `agent_ws.go`, men den
//! registreras inte - det är den här som används.

use axum::extract::ws::{Message as WsMeddelande, WebSocket, WebSocketUpgrade};
use axum::extract::State;
use axum::response::Response;
use serde::{Deserialize, Serialize};

use crate::provider::anthropic::{AnthropicClient, AnthropicStycke};
use crate::state::AppState;

/// Förfrågan som klienten skickar.
///
/// Go: `WsChatRequest`.
#[derive(Debug, Default, Deserialize)]
pub struct WsChatRequest {
    /// Typen av meddelande - "chat", "ping" eller något annat.
    #[serde(rename = "type")]
    pub typ: String,
    /// Modellens namn.
    #[serde(rename = "model", default)]
    pub modell: String,
    /// Meddelandena i chatten.
    #[serde(rename = "messages", default)]
    pub meddelanden: Vec<WsMeddelandeInnehåll>,
    /// Temperaturen.
    #[serde(rename = "temperature", default)]
    pub temperatur: f64,
    /// Samtalets id - används om klienten skickar med ett.
    #[serde(rename = "convId", default)]
    pub conv_id: String,
}

/// Ett meddelande i chatten.
///
/// Go: `llm.Message`.
#[derive(Debug, Default, Deserialize)]
pub struct WsMeddelandeInnehåll {
    /// Meddelandets roll - "user", "assistant" eller "system".
    pub role: String,
    /// Meddelandets innehåll.
    #[serde(default)]
    pub content: String,
}

/// Ett stycke som skickas till klienten.
///
/// Go: `llm.StreamChunk`. Fältens ordning är densamma som i Go.
#[derive(Debug, Default, Serialize)]
pub struct WsStrömmStycke {
    /// Styckets typ.
    #[serde(rename = "type")]
    pub typ: String,
    /// Innehållet.
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub content: String,
    /// Antalet tokens.
    #[serde(default, skip_serializing_if = "is_zero_i32")]
    pub tokens: i32,
    /// Latensen i millisekunder.
    #[serde(default, skip_serializing_if = "is_zero_i64")]
    pub latency: i64,
    /// Anledningen till att svaret slutade.
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub reason: String,
    /// Felmeddelandet.
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub error: String,
    /// Resonemanget.
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub thinking: String,
}

/// Ett meddelande i Go:s format - en karta med sorterade nycklar.
///
/// Go skickar "connected" och "done" som `map[string]string`, och felen
/// som `map[string]interface{}`. Gin's JSON skriver kartans nycklar i
/// sorterad ordning - därför ligger fälten i samma ordning här.
#[derive(Debug, Serialize)]
pub struct WsKarta {
    /// Samtalets id.
    #[serde(rename = "convId", skip_serializing_if = "Option::is_none")]
    pub conv_id: Option<String>,
    /// Felmeddelandet.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub error: Option<String>,
    /// Styckets typ.
    #[serde(rename = "type")]
    pub typ: String,
}

/// Om värdet är noll.
fn is_zero_i32(värde: &i32) -> bool {
    *värde == 0
}

/// Om värdet är noll.
fn is_zero_i64(värde: &i64) -> bool {
    *värde == 0
}

/// Hanterar WebSocket-chatten.
///
/// Go: `WebSocketHandler.HandleChat`.
pub async fn chat(State(_state): State<AppState>, uppgradering: WebSocketUpgrade) -> Response {
    uppgradering.on_upgrade(hantera)
}

/// Hanterar en uppkopplad WebSocket.
async fn hantera(mut socket: WebSocket) {
    // Samtalets id skapas vid uppkopplingen - som i Go.
    let mut conv_id = uuid::Uuid::new_v4().to_string();

    while let Some(meddelande) = socket.recv().await {
        let Ok(meddelande) = meddelande else {
            break;
        };

        let text = match meddelande {
            WsMeddelande::Text(text) => text.to_string(),
            WsMeddelande::Ping(_) | WsMeddelande::Pong(_) => continue,
            WsMeddelande::Binary(data) => match String::from_utf8(data.to_vec()) {
                Ok(text) => text,
                Err(_) => {
                    let _ = skicka_karta(
                        &mut socket,
                        &WsKarta {
                            conv_id: None,
                            error: Some("invalid JSON: invalid UTF-8 data".to_string()),
                            typ: "error".to_string(),
                        },
                    )
                    .await;

                    continue;
                }
            },
            WsMeddelande::Close(_) => break,
        };

        let req = match crate::routes::json_fel::tolka::<WsChatRequest>(
            text.as_bytes(),
            "handlers.WsChatRequest",
        ) {
            Ok(req) => req,
            Err(fel) => {
                let _ = skicka_karta(
                    &mut socket,
                    &WsKarta {
                        conv_id: None,
                        error: Some(format!("invalid JSON: {fel}")),
                        typ: "error".to_string(),
                    },
                )
                .await;

                continue;
            }
        };

        match req.typ.as_str() {
            "chat" => {
                // Klientens convId vinner över det genererade.
                if !req.conv_id.is_empty() {
                    conv_id = req.conv_id.clone();
                }

                hantera_chat(&mut socket, &req, &conv_id).await;
            }
            "ping" => {
                let _ = skicka_karta(
                    &mut socket,
                    &WsKarta {
                        conv_id: None,
                        error: None,
                        typ: "pong".to_string(),
                    },
                )
                .await;
            }
            annat => {
                let _ = skicka_karta(
                    &mut socket,
                    &WsKarta {
                        conv_id: None,
                        error: Some(format!("unknown message type: {annat}")),
                        typ: "error".to_string(),
                    },
                )
                .await;
            }
        }
    }
}

/// Kör chatten och skriver svaret till klienten.
///
/// Go: `handleChat`.
async fn hantera_chat(socket: &mut WebSocket, req: &WsChatRequest, conv_id: &str) {
    // Klienten får veta vilket samtal svaret hör till.
    let _ = skicka_karta(
        socket,
        &WsKarta {
            conv_id: Some(conv_id.to_string()),
            error: None,
            typ: "connected".to_string(),
        },
    )
    .await;

    // Go skickar bara det senaste användarmeddelandet.
    let användarmeddelande = req
        .meddelanden
        .iter()
        .rev()
        .find(|meddelande| meddelande.role == "user")
        .map(|meddelande| meddelande.content.clone())
        .unwrap_or_default();

    if användarmeddelande.is_empty() {
        let _ = skicka_stycke(
            socket,
            &WsStrömmStycke {
                typ: "error".to_string(),
                error: "No user message found".to_string(),
                ..Default::default()
            },
        )
        .await;
    } else {
        strömma_svar(socket, req, &användarmeddelande).await;
    }

    // Slutsignalen skickas alltid.
    let _ = skicka_karta(
        socket,
        &WsKarta {
            conv_id: Some(conv_id.to_string()),
            error: None,
            typ: "done".to_string(),
        },
    )
    .await;
}

/// Strömmar modellens svar till klienten.
async fn strömma_svar(socket: &mut WebSocket, req: &WsChatRequest, meddelande: &str) {
    // Go mäter latensen från början av anropet - även felmeddelandet
    // bär den tiden.
    let start = std::time::Instant::now();

    let klient = AnthropicClient::från_miljö();

    let (sändare, mut mottagare) = tokio::sync::mpsc::channel::<AnthropicStycke>(32);

    let modell = req.modell.clone();
    let temperatur = req.temperatur;
    let meddelande = meddelande.to_string();

    let strömning = tokio::spawn(async move {
        klient
            .strömma(&modell, &meddelande, temperatur, sändare)
            .await
    });

    while let Some(stycke) = mottagare.recv().await {
        let stycke = match stycke {
            AnthropicStycke::Text(text) => WsStrömmStycke {
                typ: "content".to_string(),
                content: text,
                ..Default::default()
            },
            AnthropicStycke::Tänking(text) => WsStrömmStycke {
                typ: "thinking".to_string(),
                thinking: text,
                ..Default::default()
            },
            AnthropicStycke::Klar { reason, latency_ms } => WsStrömmStycke {
                typ: "done".to_string(),
                reason,
                latency: latency_ms,
                ..Default::default()
            },
        };

        if skicka_stycke(socket, &stycke).await.is_err() {
            break;
        }
    }

    match strömning.await {
        Ok(Ok(())) => {}
        Ok(Err(fel)) => {
            let _ = skicka_stycke(
                socket,
                &WsStrömmStycke {
                    typ: "error".to_string(),
                    error: fel,
                    latency: start.elapsed().as_millis() as i64,
                    ..Default::default()
                },
            )
            .await;
        }
        Err(fel) => {
            let _ = skicka_stycke(
                socket,
                &WsStrömmStycke {
                    typ: "error".to_string(),
                    error: fel.to_string(),
                    latency: start.elapsed().as_millis() as i64,
                    ..Default::default()
                },
            )
            .await;
        }
    }
}

/// Skickar ett stycke till klienten.
async fn skicka_stycke(socket: &mut WebSocket, stycke: &WsStrömmStycke) -> Result<(), ()> {
    let text = serde_json::to_string(stycke).map_err(|_| ())?;

    socket
        .send(WsMeddelande::Text(text.into()))
        .await
        .map_err(|_| ())
}

/// Skickar en karta till klienten.
async fn skicka_karta(socket: &mut WebSocket, karta: &WsKarta) -> Result<(), ()> {
    let text = serde_json::to_string(karta).map_err(|_| ())?;

    socket
        .send(WsMeddelande::Text(text.into()))
        .await
        .map_err(|_| ())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn ogiltig_json_ger_go_meddelandet() {
        let fel = crate::routes::json_fel::tolka::<WsChatRequest>(b"{", "handlers.WsChatRequest")
            .unwrap_err();

        assert_eq!(fel, "unexpected end of JSON input");
    }

    #[test]
    fn kartan_sorteras_som_go() {
        let karta = WsKarta {
            conv_id: Some("abc".to_string()),
            error: None,
            typ: "connected".to_string(),
        };

        assert_eq!(
            serde_json::to_string(&karta).unwrap(),
            r#"{"convId":"abc","type":"connected"}"#
        );
    }

    #[test]
    fn felet_sorteras_som_go() {
        let karta = WsKarta {
            conv_id: None,
            error: Some("unknown message type: x".to_string()),
            typ: "error".to_string(),
        };

        assert_eq!(
            serde_json::to_string(&karta).unwrap(),
            r#"{"error":"unknown message type: x","type":"error"}"#
        );
    }
}
