//! Wikinreglerna i ett repos AGENTS.md.
//!
//! Go: `internal/wiki/init.go`.

use std::path::Path;

/// Markören som inleder wikinblocket.
pub const AGENTS_BLOCK_START: &str = "<!-- wiki:start -->";
/// Markören som avslutar wikinblocket.
pub const AGENTS_BLOCK_END: &str = "<!-- wiki:end -->";
/// Filen där wikinreglerna läggs in.
pub const AGENTS_FILE_NAME: &str = "AGENTS.md";

/// Texten som läggs in i ett repos AGENTS.md.
///
/// Den är kort - reglerna i sin helhet finns i den globala AGENTS.md.
pub const AGENTS_BLOCK: &str = concat!(
    "<!-- wiki:start -->\n",
    "## Kunskaps-wiki\n",
    "\n",
    "Det här repot delar en gemensam kunskaps-wiki med alla andra repor.\n",
    "Wikin ligger i `~/wiki` och är en persistent kunskapsgraf som\n",
    "agenterna underhåller - kunskapen härleds inte om vid varje fråga.\n",
    "\n",
    "**Fem regler:**\n",
    "\n",
    "1. **Läs wikin innan du frågar användaren.** Svaret finns ofta redan\n",
    "   dokumenterat - leta först.\n",
    "2. **Skriv en not i wikin inbox när något stort händer.** Incident, beslut,\n",
    "   ny feature eller en icke-obvious bugg.\n",
    "3. **Lägg in råkällor** du lär dig av med `wikictl source add`.\n",
    "   De är immutabla och ger spårbarhet.\n",
    "4. **Håll wikin aktuell.** Ett påstående som en ny källa motsäger ska\n",
    "   ändras, inte ligga kvar. `wikictl lint` hittar trasiga länkar,\n",
    "   ödelappar och föråldrad info.\n",
    "5. **Fila tillbaka värdefulla svar** som sidor, så att de inte försvinner\n",
    "   i chatthistoriken.\n",
    "\n",
    "```bash\n",
    "wikictl search \"bigram\"                  # sök i hela wikin\n",
    "wikictl search \"fel\" --project <repo>    # sök inom ett projekt\n",
    "wikictl note --title \"Rubrik\" --tags \"bug\" --body \"Vad som hände och varför.\"\n",
    "wikictl page --type concept --title \"Rubrik\" --body \"...\"   # skriv en sida\n",
    "wikictl link --from concepts/a --to concepts/b               # korslänka\n",
    "wikictl source add --file kalla.md --title \"Rubrik\"         # råkälla\n",
    "wikictl ingest                          # flytta inboxen in i grafen\n",
    "wikictl lint                            # hälsa\n",
    "```\n",
    "\n",
    "Anteckningar taggas automatiskt med `projekt:<repo>`.\n",
    "Schemat för wikin står i `~/wiki/AGENTS.md`.\n",
    "<!-- wiki:end -->"
);

/// Skriver wikinblocket in i ett repos AGENTS.md.
///
/// Om filen saknas skapas den, om blocket redan finns ersätts det.
/// Funktionen är idempotent: att köra den två gånger ger samma resultat.
///
/// Go: `wiki.EnsureAgentsBlock`.
pub fn ensure_agents_block(repo_dir: &Path) -> Result<bool, String> {
    let path = repo_dir.join(AGENTS_FILE_NAME);

    let existing = match std::fs::read_to_string(&path) {
        Ok(innehåll) => innehåll,
        Err(err) if err.kind() == std::io::ErrorKind::NotFound => String::new(),
        Err(err) => return Err(format!("kunde inte läsa {}: {err}", path.display())),
    };

    let updated = insert_block(&existing);

    if updated == existing {
        return Ok(false);
    }

    std::fs::write(&path, &updated)
        .map_err(|err| format!("kunde inte skriva {}: {err}", path.display()))?;

    Ok(true)
}

/// Sätter in wikinblocket i befintlig text.
fn insert_block(content: &str) -> String {
    let start = content.find(AGENTS_BLOCK_START);
    let end = content.find(AGENTS_BLOCK_END);

    if let (Some(start), Some(end)) = (start, end) {
        if end > start {
            return format!(
                "{}{}{}",
                &content[..start],
                AGENTS_BLOCK,
                &content[end + AGENTS_BLOCK_END.len()..]
            );
        }
    }

    if content.trim().is_empty() {
        return format!("# Agentinstruktioner\n\n{AGENTS_BLOCK}\n");
    }

    let mut content = content.to_string();

    if !content.ends_with('\n') {
        content.push('\n');
    }

    format!("{content}\n{AGENTS_BLOCK}\n")
}
