//! Verktygen som arbetar med filer och kommandon.
//!
//! Go: `internal/core/tools/{bash,read,write,glob,grep,read_multiple}.go`.

use std::path::Path;
use std::sync::{Arc, Mutex};
use std::time::Duration;

use regex::Regex;
use tokio::io::AsyncReadExt;
use tokio::process::Command;

use super::glob::glob;
use super::sanitize::{validera_kommando, validera_sökväg, ContentValidator};
use super::{avbruten_resultat, fel_resultat, text_resultat, ToolCall, ToolDefinition, ToolResult};
use crate::messages::ContentBlock;

/// Verktyget `bash`.
pub fn bashverktyg() -> ToolDefinition {
    ToolDefinition {
        name: "bash".to_string(),
        description: "Execute a bash command. Returns stdout and stderr.".to_string(),
        input_schema: super::schemas::bash_schema(),
        execute: Arc::new(verkställ_bash),
    }
}

/// Verktyget `read`.
pub fn läsverktyg() -> ToolDefinition {
    ToolDefinition {
        name: "read".to_string(),
        description: "Read the contents of a file. Supports text files.".to_string(),
        input_schema: super::schemas::fil_schema(),
        execute: Arc::new(verkställ_läs),
    }
}

/// Verktyget `write`.
pub fn skrivverktyg() -> ToolDefinition {
    ToolDefinition {
        name: "write".to_string(),
        description:
            "Write content to a file. Creates the file if it doesn't exist, overwrites if it does."
                .to_string(),
        input_schema: super::schemas::skriv_schema(),
        execute: Arc::new(verkställ_skriv),
    }
}

/// Verktyget `edit`.
pub fn redigeraverktyg() -> ToolDefinition {
    ToolDefinition {
        name: "edit".to_string(),
        description: "Edit specific sections of a file using exact text replacement.".to_string(),
        input_schema: super::schemas::redigera_schema(),
        execute: Arc::new(verkställ_redigera),
    }
}

/// Verktyget `glob`.
pub fn globverktyg() -> ToolDefinition {
    ToolDefinition {
        name: "glob".to_string(),
        description: "Find files matching a pattern using glob syntax.".to_string(),
        input_schema: super::schemas::glob_schema(),
        execute: Arc::new(verkställ_glob),
    }
}

/// Verktyget `grep`.
pub fn grepverktyg() -> ToolDefinition {
    ToolDefinition {
        name: "grep".to_string(),
        description: "Search for lines matching a pattern in a file.".to_string(),
        input_schema: super::schemas::grep_schema(),
        execute: Arc::new(verkställ_grep),
    }
}

/// Verktyget `read_multiple`.
pub fn läs_fleraverktyg() -> ToolDefinition {
    ToolDefinition {
        name: "read_multiple".to_string(),
        description: "Read multiple files in a single call.".to_string(),
        input_schema: super::schemas::läs_flera_schema(),
        execute: Arc::new(verkställ_läs_flera),
    }
}

/// Kör ett bash-kommando.
fn verkställ_bash(anrop: ToolCall) -> super::ToolFuture {
    Box::pin(async move {
        let Some(kommando) = anrop.args.get("command").and_then(|v| v.as_str()) else {
            return Ok(fel_resultat("command must be a string"));
        };

        let kommando = kommando.to_string();

        // Säkerhetskontroll av kommandot.
        let kontroll = validera_kommando(&kommando);
        if !kontroll.valid {
            return Ok(fel_resultat(&format!(
                "invalid command: {}",
                kontroll.message
            )));
        }

        let timeout = anrop.float_arg("timeout", 60.0);
        let arbetskatalog = anrop.str_arg("cwd", ".");
        let _strömma = anrop.bool_arg("stream", true);

        // En redan avbruten körning avbryts direkt.
        if anrop.avbruten() {
            return Ok(avbruten_resultat("Command cancelled"));
        }

        let mut kommando_process = Command::new("bash");
        kommando_process.arg("-c").arg(&kommando);

        if arbetskatalog != "." {
            kommando_process.current_dir(&arbetskatalog);
        }

        kommando_process.kill_on_drop(true);

        let mut child = match kommando_process
            .stdout(std::process::Stdio::piped())
            .stderr(std::process::Stdio::piped())
            .spawn()
        {
            Ok(child) => child,
            Err(_fel) => {
                // Reserv: försök med sh.
                let mut reserv = Command::new("sh");
                reserv.arg("-c").arg(&kommando);

                if arbetskatalog != "." {
                    reserv.current_dir(&arbetskatalog);
                }

                reserv.kill_on_drop(true);

                match reserv
                    .stdout(std::process::Stdio::piped())
                    .stderr(std::process::Stdio::piped())
                    .spawn()
                {
                    Ok(child) => child,
                    Err(fel) => {
                        return Ok(fel_resultat(&format!("failed to start command: {fel}")));
                    }
                }
            }
        };

        let utdata = Arc::new(Mutex::new(String::new()));
        let mut uppdateringar = Vec::new();

        if let Some(mut stdout) = child.stdout.take() {
            let utdata = Arc::clone(&utdata);
            let mut sender = anrop.on_update.clone();

            uppdateringar.push(tokio::spawn(async move {
                läs_ström(&mut stdout, Arc::clone(&utdata), sender.take(), false).await;
            }));
        }

        if let Some(mut stderr) = child.stderr.take() {
            let utdata = Arc::clone(&utdata);
            let mut sender = anrop.on_update.clone();

            uppdateringar.push(tokio::spawn(async move {
                läs_ström(&mut stderr, Arc::clone(&utdata), sender.take(), true).await;
            }));
        }

        // Processen väntas i en egen uppgift - så att den kan avbrytas
        // medan den väntar.
        let (status_sändare, status_mottagare) = tokio::sync::oneshot::channel();
        let (död_sändare, mut dödmottagare) = tokio::sync::mpsc::channel::<()>(1);

        tokio::spawn(async move {
            tokio::select! {
                status = child.wait() => {
                    let _ = status_sändare.send(status);
                }
                _ = dödmottagare.recv() => {
                    let _ = child.start_kill();
                    let _vänta = child.wait().await;
                    let _ = status_sändare.send(Err(std::io::Error::other("killed")));
                }
            }
        });

        let mut signal = anrop.signal.clone();

        let resultat = tokio::select! {
            _ = signal.changed() => {
                let _ = död_sändare.send(()).await;
                Err("Command cancelled".to_string())
            }
            _ = tokio::time::sleep(Duration::from_secs_f64(timeout.max(0.0))) => {
                let _ = död_sändare.send(()).await;
                Err("command timed out".to_string())
            }
            resultat = status_mottagare => match resultat {
                Ok(status) => status.map_err(|fel| fel.to_string()),
                Err(_) => Err("command failed".to_string()),
            },
        };

        // Vänta in strömmarna.
        for uppgift in uppdateringar {
            let _ = uppgift.await;
        }

        match resultat {
            Err(fel) if fel == "Command cancelled" => Ok(avbruten_resultat("Command cancelled")),
            Err(fel) if fel == "command timed out" => Ok(fel_resultat("command timed out")),
            Err(fel) => Ok(fel_resultat(&format!("failed to start command: {fel}"))),
            Ok(status) => {
                let utdata = utdata.lock().map(|data| data.clone()).unwrap_or_default();

                Ok(ToolResult {
                    content: vec![ContentBlock::Text { text: utdata }],
                    is_error: !status.success(),
                    details: Some(serde_json::json!({
                        "exit_code": status.code().unwrap_or(-1),
                    })),
                    ..Default::default()
                })
            }
        }
    })
}

/// Läser en ström och lägger innehållet i bufferten.
async fn läs_ström<R>(
    ström: &mut R,
    buffert: Arc<Mutex<String>>,
    uppdateringar: Option<tokio::sync::mpsc::Sender<ToolResult>>,
    är_fel: bool,
) where
    R: tokio::io::AsyncRead + Unpin,
{
    let mut byte = [0u8; 1024];
    let mut radbuffert = String::new();

    while let Ok(antal) = ström.read(&mut byte).await {
        if antal == 0 {
            break;
        }

        let text = String::from_utf8_lossy(&byte[..antal]).to_string();

        if let Ok(mut buffert) = buffert.lock() {
            buffert.push_str(&text);
        }

        radbuffert.push_str(&text);

        // Skicka en uppdatering för varje rad eller var ~1 KB.
        if radbuffert.len() >= 1024 || text.contains('\n') {
            if let Some(sender) = &uppdateringar {
                let _ = sender
                    .send(ToolResult {
                        content: vec![ContentBlock::Text {
                            text: std::mem::take(&mut radbuffert),
                        }],
                        is_error: är_fel,
                        ..Default::default()
                    })
                    .await;
            } else {
                radbuffert.clear();
            }
        }
    }

    // Skicka det som är kvar.
    if !radbuffert.is_empty() {
        if let Some(sender) = &uppdateringar {
            let _ = sender
                .send(ToolResult {
                    content: vec![ContentBlock::Text { text: radbuffert }],
                    is_error: är_fel,
                    ..Default::default()
                })
                .await;
        }
    }
}

/// Läser en fil.
fn verkställ_läs(anrop: ToolCall) -> super::ToolFuture {
    Box::pin(async move {
        let sökväg = anrop.str_arg("path", "");

        if sökväg.is_empty() {
            return Ok(fel_resultat("path is required"));
        }

        if anrop.avbruten() {
            return Ok(avbruten_resultat("Read"));
        }

        let kontroll = validera_sökväg(&sökväg);
        if !kontroll.valid {
            return Ok(fel_resultat(&format!("invalid path: {}", kontroll.message)));
        }

        let Ok(innehåll) = tokio::fs::read(&sökväg).await else {
            return Ok(fel_resultat(&format!("reading file: {}", feltext(&sökväg))));
        };

        let mut text = String::from_utf8_lossy(&innehåll).to_string();

        let offset = anrop.float_arg("offset", 0.0);
        let limit = anrop.float_arg("limit", 0.0);

        if offset > 0.0 {
            text = tilldela_offset(&text, offset as i64, limit as i64);
        } else if limit > 0.0 {
            text = tilldela_offset(&text, 1, limit as i64);
        }

        Ok(text_resultat(text))
    })
}

/// Plockar ut ett stycke ur texten.
fn tilldela_offset(innehåll: &str, offset: i64, limit: i64) -> String {
    let rader: Vec<&str> = innehåll.split('\n').collect();

    let start = (offset - 1).max(0) as usize;

    if start >= rader.len() {
        return String::new();
    }

    let mut delar = &rader[start..];

    if limit > 0 && limit < delar.len() as i64 {
        delar = &delar[..limit as usize];
    }

    delar.join("\n")
}

/// Skriver en fil.
fn verkställ_skriv(anrop: ToolCall) -> super::ToolFuture {
    Box::pin(async move {
        let sökväg = anrop.str_arg("path", "");
        let innehåll = anrop.str_arg("content", "");

        if anrop.avbruten() {
            return Ok(avbruten_resultat("Write"));
        }

        let kontroll = validera_sökväg(&sökväg);
        if !kontroll.valid {
            return Ok(fel_resultat(&format!("invalid path: {}", kontroll.message)));
        }

        let validator = ContentValidator::default();
        let innehållskontroll = validator.validera_innehåll(innehåll.as_bytes(), &sökväg);
        if !innehållskontroll.valid {
            return Ok(fel_resultat(&format!(
                "invalid content: {}",
                innehållskontroll.message
            )));
        }

        if let Some(katalog) = Path::new(&sökväg).parent() {
            if !katalog.as_os_str().is_empty() && katalog != Path::new(".") {
                if let Err(fel) = tokio::fs::create_dir_all(katalog).await {
                    return Ok(fel_resultat(&format!("creating directory: {fel}")));
                }
            }
        }

        if let Err(fel) = tokio::fs::write(&sökväg, &innehåll).await {
            return Ok(fel_resultat(&format!("writing file: {fel}")));
        }

        Ok(text_resultat(format!(
            "Successfully wrote {} bytes to {}",
            innehåll.len(),
            sökväg
        )))
    })
}

/// Redigerar en fil.
fn verkställ_redigera(anrop: ToolCall) -> super::ToolFuture {
    Box::pin(async move {
        let sökväg = anrop.str_arg("path", "");
        let gammal_text = anrop.str_arg("oldText", "");
        let ny_text = anrop.str_arg("newText", "");

        if anrop.avbruten() {
            return Ok(avbruten_resultat("Edit"));
        }

        let kontroll = validera_sökväg(&sökväg);
        if !kontroll.valid {
            return Ok(fel_resultat(&format!("invalid path: {}", kontroll.message)));
        }

        let Ok(innehåll) = tokio::fs::read(&sökväg).await else {
            return Ok(fel_resultat(&format!("reading file: {}", feltext(&sökväg))));
        };

        let text = String::from_utf8_lossy(&innehåll).to_string();

        if !text.contains(&gammal_text) {
            return Ok(fel_resultat(&format!(
                "oldText not found in file\nSearched for: {:?}",
                gammal_text
            )));
        }

        // Go byter ut den första förekomsten.
        let nytt_innehåll = text.replacen(&gammal_text, &ny_text, 1);

        if let Err(fel) = tokio::fs::write(&sökväg, nytt_innehåll).await {
            return Ok(fel_resultat(&format!("writing file: {fel}")));
        }

        Ok(text_resultat(format!("Successfully edited {sökväg}")))
    })
}

/// Hittar filer med ett glob-mönster.
fn verkställ_glob(anrop: ToolCall) -> super::ToolFuture {
    Box::pin(async move {
        let mönster = anrop.str_arg("pattern", "");
        let basmapp = anrop.str_arg("path", ".");

        let sökväg = Path::new(&basmapp).join(&mönster);

        let träffar = match glob(&sökväg.to_string_lossy()) {
            Ok(träffar) => träffar,
            Err(fel) => return Ok(fel_resultat(&format!("invalid glob pattern: {fel}"))),
        };

        let text: Vec<String> = träffar
            .iter()
            .map(|träff| träff.to_string_lossy().to_string())
            .collect();

        Ok(text_resultat(text.join("\n")))
    })
}

/// Söker i en fil.
fn verkställ_grep(anrop: ToolCall) -> super::ToolFuture {
    Box::pin(async move {
        let mönster = anrop.str_arg("pattern", "");
        let sökväg = anrop.str_arg("path", "");
        let ignora_storlek = anrop.bool_arg("ignoreCase", false);

        if anrop.avbruten() {
            return Ok(avbruten_resultat("Search"));
        }

        let regexmönster = if ignora_storlek {
            format!("(?i){mönster}")
        } else {
            mönster.clone()
        };

        let Ok(regex) = Regex::new(&regexmönster) else {
            return Ok(fel_resultat("invalid regex pattern"));
        };

        let Ok(innehåll) = tokio::fs::read_to_string(&sökväg).await else {
            return Ok(fel_resultat(&format!("opening file: {}", feltext(&sökväg))));
        };

        let mut träffar: Vec<String> = Vec::new();

        for (index, rad) in innehåll.lines().enumerate() {
            if regex.is_match(rad) {
                träffar.push(format!("{}:{}", index + 1, rad));
            }
        }

        if träffar.is_empty() {
            return Ok(text_resultat("No matches found"));
        }

        Ok(text_resultat(träffar.join("\n")))
    })
}

/// Läser flera filer.
fn verkställ_läs_flera(anrop: ToolCall) -> super::ToolFuture {
    Box::pin(async move {
        let Some(sökvägar) = anrop.args.get("paths").and_then(|v| v.as_array()) else {
            return Ok(fel_resultat("paths must be an array"));
        };

        if anrop.avbruten() {
            return Ok(avbruten_resultat("ReadMultiple"));
        }

        let mut resultat: Vec<String> = Vec::new();

        for post in sökvägar {
            let Some(sökväg) = post.as_str() else {
                resultat.push("Error: invalid path format".to_string());
                continue;
            };

            let kontroll = validera_sökväg(sökväg);
            if !kontroll.valid {
                resultat.push(format!(
                    "=== {sökväg} ===\nError: path outside allowed directories"
                ));
                continue;
            }

            let Ok(innehåll) = tokio::fs::read(sökväg).await else {
                resultat.push(format!("=== {sökväg} (error) ===\n{}", feltext(sökväg)));
                continue;
            };

            let mut text = String::from_utf8_lossy(&innehåll).to_string();

            if text.len() > 50000 {
                text = format!("{}\n... (truncated)", &text[..50000]);
            }

            resultat.push(format!("=== {sökväg} ===\n{text}"));
        }

        Ok(text_resultat(resultat.join("\n\n")))
    })
}

/// Feltexten för en fil - som Go:s `os.ReadFile`.
fn feltext(sökväg: &str) -> String {
    match std::fs::metadata(sökväg) {
        Ok(_) => format!("open {sökväg}: permission denied"),
        Err(fel) => format!(
            "open {sökväg}: {}",
            fel.to_string()
                .replace("os error 2, ", "no such file or directory: ")
        ),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn offset_och_limit_klipper_texten() {
        let text = "ett\ntvå\ntre\nfyra";

        assert_eq!(tilldela_offset(text, 2, 2), "två\ntre");
        assert_eq!(tilldela_offset(text, 1, 0), text);
        assert_eq!(tilldela_offset(text, 9, 0), "");
    }
}
