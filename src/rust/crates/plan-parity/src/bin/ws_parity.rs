//! Jämför agentens WebSocket-chatt mellan Go och Rust.
//!
//! Verktyget ansluter till båda tjänsterna, skickar samma meddelanden
//! och jämför svarets sekvens. Skillnader som slumpade
//! conversations-id:n och latenstider normaliseras bort - de är olika
//! på varje körning.
//!
//! Användning:
//!   ws_parity [go-url] [rust-url] [nyckel]
//!
//! Standard: http://127.0.0.1:8198 och http://127.0.0.1:8199.

use std::time::Duration;

use futures_util::{SinkExt, StreamExt};
use serde_json::Value;
use tokio_tungstenite::connect_async;
use tokio_tungstenite::tungstenite::client::IntoClientRequest;
use tokio_tungstenite::tungstenite::http::HeaderValue;

/// Meddelanden som skickas till båda tjänsterna.
const MEDDELANDEN: &[(&str, &str)] = &[
    ("ping", r#"{"type":"ping"}"#),
    ("okänd typ", r#"{"type":"unknown"}"#),
    ("trasig json", "{"),
    ("chatt utan meddelande", r#"{"type":"chat"}"#),
];

/// Hur länge väntar verktyget på nästa ram.
const RAMTID: Duration = Duration::from_millis(600);

/// Bygger om en http-adress till en ws-adress.
fn ws_adress(url: &str) -> String {
    if let Some(rest) = url.strip_prefix("http://") {
        return format!("ws://{rest}");
    }

    if let Some(rest) = url.strip_prefix("https://") {
        return format!("wss://{rest}");
    }

    url.to_string()
}

/// Normaliserar en ram så att två körningar kan jämföras.
///
/// Samtalets id är slumpat och latensen beror på maskinens fart - båda
/// tas bort. Det som är kvar är meddelandets ordning och innehåll.
fn normalisera(mut ram: Value) -> Value {
    if let Some(objekt) = ram.as_object_mut() {
        objekt.remove("convId");
        objekt.remove("latency");
    }

    ram
}

/// Läser en sekvens av ramar från en uppkoppling.
async fn läs_ramar<S>(läsare: &mut S) -> Vec<Value>
where
    S: StreamExt<
            Item = Result<
                tokio_tungstenite::tungstenite::Message,
                tokio_tungstenite::tungstenite::Error,
            >,
        > + Unpin,
{
    let mut ramar = Vec::new();

    while let Ok(Some(ram)) = tokio::time::timeout(RAMTID, läsare.next()).await {
        let Ok(ram) = ram else {
            break;
        };

        let text = match ram {
            tokio_tungstenite::tungstenite::Message::Text(text) => text.to_string(),
            tokio_tungstenite::tungstenite::Message::Close(_) => break,
            _ => continue,
        };

        let Ok(värde) = serde_json::from_str::<Value>(&text) else {
            ramar.push(Value::String(text));
            continue;
        };

        ramar.push(normalisera(värde));
    }

    ramar
}

/// Kör ett meddelande mot en tjänst och returnerar svarens sekvens.
async fn kör_meddelande(
    basadress: &str,
    nyckel: &str,
    meddelande: &str,
) -> Result<Vec<Value>, String> {
    let adress = format!("{}/api/v1/ws/chat", ws_adress(basadress));

    let mut begäran = adress
        .as_str()
        .into_client_request()
        .map_err(|fel| format!("ogiltig adress: {fel}"))?;

    begäran.headers_mut().insert(
        "Authorization",
        HeaderValue::from_str(&format!("Bearer {nyckel}"))
            .map_err(|fel| format!("ogiltig nyckel: {fel}"))?,
    );

    let (mut anslutning, _svar) = connect_async(begäran)
        .await
        .map_err(|fel| format!("anslutningen misslyckades: {fel}"))?;

    anslutning
        .send(tokio_tungstenite::tungstenite::Message::Text(
            meddelande.to_string().into(),
        ))
        .await
        .map_err(|fel| format!("kunde inte skicka: {fel}"))?;

    let ramar = läs_ramar(&mut anslutning).await;

    Ok(ramar)
}

/// Kör hela jämförelsen.
async fn jämför(go_url: &str, rust_url: &str, nyckel: &str) -> bool {
    let mut avvikelser = 0;

    for (namn, meddelande) in MEDDELANDEN {
        let go_svar = match kör_meddelande(go_url, nyckel, meddelande).await {
            Ok(svar) => svar,
            Err(fel) => {
                println!("FEL (go, {namn}): {fel}");
                avvikelser += 1;
                continue;
            }
        };

        let rust_svar = match kör_meddelande(rust_url, nyckel, meddelande).await {
            Ok(svar) => svar,
            Err(fel) => {
                println!("FEL (rust, {namn}): {fel}");
                avvikelser += 1;
                continue;
            }
        };

        if go_svar == rust_svar {
            println!("OK: {namn} ({} ramar)", go_svar.len());

            if std::env::var("PARITY_VERBOSE").is_ok() {
                println!("  {}", serde_json::to_string(&go_svar).unwrap_or_default());
            }

            continue;
        }

        avvikelser += 1;

        println!("AVVIKELSE: {namn}");
        println!(
            "  go:   {}",
            serde_json::to_string(&go_svar).unwrap_or_default()
        );
        println!(
            "  rust: {}",
            serde_json::to_string(&rust_svar).unwrap_or_default()
        );
    }

    avvikelser == 0
}

#[tokio::main]
async fn main() {
    let go_url = std::env::args()
        .nth(1)
        .unwrap_or_else(|| "http://127.0.0.1:8198".to_string());
    let rust_url = std::env::args()
        .nth(2)
        .unwrap_or_else(|| "http://127.0.0.1:8199".to_string());
    let nyckel = std::env::args()
        .nth(3)
        .unwrap_or_else(|| "testkey".to_string());

    if jämför(&go_url, &rust_url, &nyckel).await {
        println!();
        println!("WebSocket-chatten ger samma svar i båda tjänsterna.");
        return;
    }

    println!();
    println!("Avvikelser hittade - se ovan.");
    std::process::exit(1);
}
