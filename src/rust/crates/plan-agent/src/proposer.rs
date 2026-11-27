//! Låter agenten föreslå kodändringar.
//!
//! Go: `internal/proposer`. Agenten läser repot, föreslår en ändring och
//! lämnar in den för granskning. Den tillämpar inte själv ändringen - det
//! gör en människa efter att ha godkänt förslaget.

use std::path::{Path, PathBuf};
use std::sync::Arc;

use crate::executor::Change;
use crate::messages::Message;
use crate::provider::{LlmProvider, Model, StreamEvent};
use crate::runusage::Recorder;
use crate::workspace::Workspace;

/// Agentens förslag på en kodändring.
#[derive(Debug, Clone, Default)]
pub struct Proposal {
    /// Titeln beskriver ändringen kort.
    pub title: String,
    /// Beskrivningen förklarar varför ändringen behövs.
    pub description: String,
    /// De föreslagna ändringarna.
    pub changes: Vec<Change>,
    /// Kommandot som verifierar ändringen.
    pub test_command: String,
}

/// Hur många gånger modellen får svara innan förslaget lämnas över.
pub const MAX_FORSLAGSFÖRSÖK: usize = 2;

/// Tar fram förslag med hjälp av en språkmodell.
pub struct Proposer {
    /// Leverantören.
    leverantör: Arc<dyn LlmProvider>,
    /// Modellen.
    modell: Model,
    /// Recordern för tokenanvändning.
    användning: Option<std::sync::Arc<Recorder>>,
}

impl Proposer {
    /// Skapar en proposer.
    ///
    /// Go: `proposer.New`.
    pub fn ny(leverantör: Arc<dyn LlmProvider>, modell: Model) -> Self {
        Self {
            leverantör,
            modell,
            användning: None,
        }
    }

    /// Kopplar en recorder för tokenanvändning.
    ///
    /// Go: `Proposer.WithUsage`.
    pub fn med_användning(mut self, recorder: std::sync::Arc<Recorder>) -> Self {
        self.användning = Some(recorder);
        self
    }

    /// Läser repot och föreslår en ändring för ett mål.
    ///
    /// Om svaret inte går att tolka görs ett nytt försök - modellen kan
    /// svara fel format ibland.
    ///
    /// Go: `Proposer.Propose`.
    pub async fn föreslå(
        &self,
        ws: &Workspace,
        mål_titel: &str,
        mål_beskrivning: &str,
    ) -> Result<Proposal, String> {
        let filer = samla_filer(ws, 40);

        let mut sista_felet = String::new();

        for försök in 0..MAX_FORSLAGSFÖRSÖK {
            match self
                .föreslå_en_gång(ws, mål_titel, mål_beskrivning, &filer)
                .await
            {
                Ok(förslag) => return Ok(förslag),
                Err(fel) => {
                    tracing::warn!("Proposal attempt {} failed: {fel}", försök + 1);

                    sista_felet = fel;
                }
            }
        }

        Err(sista_felet)
    }

    /// Gör ett försök att få ett förslag.
    ///
    /// Go: `Proposer.proposeOnce`.
    async fn föreslå_en_gång(
        &self,
        ws: &Workspace,
        mål_titel: &str,
        mål_beskrivning: &str,
        filer: &[String],
    ) -> Result<Proposal, String> {
        let meddelanden = vec![
            Message::system(SYSTEMPROMPT),
            Message::user(bygg_userprompt(mål_titel, mål_beskrivning, filer)),
        ];

        let svar = self
            .streama(&meddelanden)
            .await
            .map_err(|fel| format!("failed to stream response: {fel}"))?;

        let mut förslag = tolka_förslag(&svar)?;

        relativisera(ws.root(), &mut förslag);

        Ok(förslag)
    }

    /// Rättar ett förslag vars egen verifiering inte gick igenom.
    ///
    /// Go: `Proposer.Revise`.
    pub async fn revidera(
        &self,
        ws: &Workspace,
        mål_titel: &str,
        mål_beskrivning: &str,
        förslag: &Proposal,
        fel: &str,
    ) -> Result<Proposal, String> {
        let meddelanden = vec![
            Message::system(REVISION_SYSTEMPROMPT),
            Message::user(bygg_revisionsprompt(
                mål_titel,
                mål_beskrivning,
                förslag,
                fel,
            )),
        ];

        let svar = self
            .streama(&meddelanden)
            .await
            .map_err(|fel| format!("failed to stream response: {fel}"))?;

        let mut reviderat = tolka_förslag(&svar)?;

        relativisera(ws.root(), &mut reviderat);

        Ok(reviderat)
    }

    /// Ber modellen skriva om testkommandot så att det går igenom på det
    /// föreslagna innehållet.
    ///
    /// Go: `Proposer.FixTest`.
    pub async fn reparera_test(
        &self,
        ws: &Workspace,
        mål_titel: &str,
        mål_beskrivning: &str,
        förslag: &Proposal,
        fel: &str,
    ) -> Result<String, String> {
        let meddelanden = vec![
            Message::system(FIXTEST_SYSTEMPROMPT),
            Message::user(bygg_fixtestprompt(mål_titel, mål_beskrivning, förslag, fel)),
        ];

        let svar = self
            .streama(&meddelanden)
            .await
            .map_err(|fel| format!("failed to stream response: {fel}"))?;

        let _ = ws;

        Ok(tolka_testkommando(&svar))
    }

    /// Streamar ett svar från modellen.
    async fn streama(&self, meddelanden: &[Message]) -> Result<String, String> {
        let mut händelser = self
            .leverantör
            .stream(
                self.modell.clone(),
                meddelanden.to_vec(),
                Vec::new(),
                crate::provider::StreamingOptions {
                    usage: true,
                    ..Default::default()
                },
            )
            .await
            .map_err(|fel| fel.to_string())?;

        let mut svar = String::new();

        while let Some(händelse) = händelser.recv().await {
            match händelse {
                StreamEvent::TextDelta { delta, .. } => svar.push_str(&delta),
                StreamEvent::Done { message } => {
                    if let Some(recorder) = &self.användning {
                        if let Some(användning) = message.usage() {
                            recorder.registrera(användning);
                        }
                    }

                    return Ok(svar);
                }
                StreamEvent::Error { error, .. } => return Err(error),
                _ => {}
            }
        }

        Err("stream ended without completion".to_string())
    }
}

/// Läser ett testkommando ur modellens svar.
///
/// Go: `parseTestCommand`.
pub fn tolka_testkommando(svar: &str) -> String {
    for rad in svar.split('\n') {
        let trimmed = rad.trim();

        if let Some(kommando) = trimmed.strip_prefix("TEST_COMMAND:") {
            return kommando.trim().to_string();
        }
    }

    String::new()
}

/// Bygger prompten med målet och filinnehållet.
///
/// Go: `buildUserPrompt`.
pub fn bygg_userprompt(mål_titel: &str, mål_beskrivning: &str, filer: &[String]) -> String {
    let mut text = String::new();

    text.push_str("## GOAL\n\n");
    text.push_str(mål_titel);
    text.push_str("\n\n");

    if !mål_beskrivning.is_empty() {
        text.push_str("## DESCRIPTION\n\n");
        text.push_str(mål_beskrivning);
        text.push_str("\n\n");
    }

    text.push_str("## REPOSITORY FILES\n\n");

    for fil in filer {
        text.push_str(fil);
        text.push('\n');
    }

    text.push_str("\n## YOUR TASK\n\n");
    text.push_str("Propose the change and the test command. Use the format from the instructions.");

    text
}

/// Beskriver innehållet och det felande testet.
///
/// Go: `buildFixTestPrompt`.
pub fn bygg_fixtestprompt(
    mål_titel: &str,
    mål_beskrivning: &str,
    förslag: &Proposal,
    fel: &str,
) -> String {
    let mut text = String::new();

    text.push_str("Goal: ");
    text.push_str(mål_titel);
    text.push('\n');
    text.push_str(mål_beskrivning);
    text.push_str("\n\nThis test failed on your own proposed content:\n");
    text.push_str(&förslag.test_command);
    text.push_str("\n\nThe failure was:\n");
    text.push_str(fel);
    text.push_str("\n\nThe content stays as it is:\n");

    for ändring in &förslag.changes {
        text.push_str("\nFILE: ");
        text.push_str(&ändring.path);
        text.push_str("\nCONTENT:\n");
        text.push_str(&ändring.content);
        text.push_str("\nEND_FILE\n");
    }

    text.push_str("\nWrite a new test command that passes on exactly this content.");

    text
}

/// Beskriver det förslag som inte gick igenom.
///
/// Go: `buildRevisionPrompt`.
pub fn bygg_revisionsprompt(
    mål_titel: &str,
    mål_beskrivning: &str,
    förslag: &Proposal,
    fel: &str,
) -> String {
    let mut text = String::new();

    text.push_str("Goal: ");
    text.push_str(mål_titel);
    text.push('\n');
    text.push_str(mål_beskrivning);
    text.push_str("\n\nYour previous proposal was:\n");
    text.push_str("TITLE: ");
    text.push_str(&förslag.title);
    text.push_str("\nTEST_COMMAND: ");
    text.push_str(&förslag.test_command);
    text.push('\n');

    for ändring in &förslag.changes {
        text.push_str("\nFILE: ");
        text.push_str(&ändring.path);
        text.push_str("\nCONTENT:\n");
        text.push_str(&ändring.content);
        text.push_str("\nEND_FILE\n");
    }

    text.push_str("\nThe verification failed with:\n");
    text.push_str(fel);
    text.push_str(
        "\n\nGive a corrected proposal. The test command must pass on the content you propose.",
    );

    text
}

/// Gör om sökvägarna till relativa mot repots rot.
///
/// Go: `relativize`.
pub fn relativisera(root: &Path, förslag: &mut Proposal) {
    for ändring in &mut förslag.changes {
        ändring.path = relativ_sökväg(root, &ändring.path);
    }
}

/// Gör en sökväg relativ mot root om den ligger under den.
///
/// Go: `relativePath`.
pub fn relativ_sökväg(root: &Path, sökväg: &str) -> String {
    if sökväg.is_empty() {
        return sökväg.to_string();
    }

    let sökväg = Path::new(sökväg);

    if !sökväg.is_absolute() {
        return sökväg.to_string_lossy().to_string();
    }

    let Ok(relativ) = sökväg.strip_prefix(root) else {
        return sökväg.to_string_lossy().to_string();
    };

    relativ.to_string_lossy().replace('\\', "/")
}

/// Läser modellens svar och gör om det till ett förslag.
///
/// Go: `parseProposal`.
pub fn tolka_förslag(svar: &str) -> Result<Proposal, String> {
    let mut förslag = Proposal::default();

    let mut aktuell_sökväg = String::new();
    let mut aktuellt_innehåll = String::new();
    let mut testkommando = String::new();
    let mut i_fil = false;
    let mut i_testkommando = false;
    let mut heredoc_slut = String::new();

    for rad in svar.split('\n') {
        let trimmed = rad.trim();

        // Ett testkommando kan bestå av flera rader - exempelvis ett
        // heredoc-skript. Det läses tills heredoc-markören kommer.
        if i_testkommando {
            testkommando.push_str(rad);
            testkommando.push('\n');

            if trimmed == heredoc_slut {
                i_testkommando = false;
                heredoc_slut.clear();
                förslag.test_command = format!(
                    "{}\n{}",
                    förslag.test_command,
                    testkommando.trim_end_matches('\n')
                );
                testkommando.clear();
            }

            continue;
        }

        if let Some(titel) = trimmed.strip_prefix("TITLE:") {
            förslag.title = titel.trim().to_string();
            i_fil = false;
            continue;
        }

        if let Some(beskrivning) = trimmed.strip_prefix("DESCRIPTION:") {
            förslag.description = beskrivning.trim().to_string();
            i_fil = false;
            continue;
        }

        if let Some(kommando) = trimmed.strip_prefix("TEST_COMMAND:") {
            förslag.test_command = kommando.trim().to_string();
            i_fil = false;

            if let Some(slut) = heredoc_terminator(&förslag.test_command) {
                i_testkommando = true;
                heredoc_slut = slut;
            }

            continue;
        }

        if let Some(sökväg) = trimmed.strip_prefix("FILE:") {
            if i_fil && !aktuell_sökväg.is_empty() {
                förslag.changes.push(Change {
                    path: aktuell_sökväg.clone(),
                    content: aktuellt_innehåll.trim_end_matches('\n').to_string(),
                    reason: String::new(),
                });
            }

            aktuell_sökväg = sökväg.trim().to_string();
            aktuellt_innehåll.clear();
            i_fil = true;
            continue;
        }

        if trimmed == "END_FILE" {
            if i_fil && !aktuell_sökväg.is_empty() {
                förslag.changes.push(Change {
                    path: aktuell_sökväg.clone(),
                    content: aktuellt_innehåll.trim_end_matches('\n').to_string(),
                    reason: String::new(),
                });
            }

            aktuell_sökväg.clear();
            aktuellt_innehåll.clear();
            i_fil = false;
            continue;
        }

        if i_fil && !aktuell_sökväg.is_empty() {
            // "CONTENT:" är en markör, inte en del av filen.
            if trimmed == "CONTENT:" {
                continue;
            }

            aktuellt_innehåll.push_str(rad);
            aktuellt_innehåll.push('\n');
        }
    }

    if i_fil && !aktuell_sökväg.is_empty() {
        förslag.changes.push(Change {
            path: aktuell_sökväg.clone(),
            content: aktuellt_innehåll.trim_end_matches('\n').to_string(),
            reason: String::new(),
        });
    }

    if i_testkommando {
        // heredoc som aldrig avslutades - behåll det vi läste.
        förslag.test_command = testkommando.trim_end_matches('\n').to_string();
    }

    if förslag.changes.is_empty() {
        return Err("failed to parse any file change from the response".to_string());
    }

    if förslag.title.is_empty() {
        förslag.title = "Föreslagen ändring".to_string();
    }

    Ok(förslag)
}

/// Läser avslutningen av ett heredoc.
///
/// Go: `heredocTerminator`.
pub fn heredoc_terminator(kommando: &str) -> Option<String> {
    let index = kommando.find("<<")?;

    let rest = kommando[index + 2..].trim().trim_matches(['\'', '"']);

    if rest.is_empty() {
        return None;
    }

    rest.split_whitespace().next().map(|s| s.to_string())
}

/// Berättar om en fil är relevant för ett kodmål.
///
/// Go: `isRelevantSource`.
pub fn är_relevant_källkod(namn: &str) -> bool {
    namn.ends_with(".go")
        || namn.ends_with(".ts")
        || namn.ends_with(".vue")
        || namn.ends_with(".md")
        || namn.ends_with(".yaml")
        || namn.ends_with(".yml")
        || namn.ends_with(".sh")
        || namn.ends_with("_test.go")
}

/// Läser de filer som är mest relevanta för ett förslag.
///
/// Go: `collectFiles`.
pub fn samla_filer(ws: &Workspace, gräns: usize) -> Vec<String> {
    if let Ok(ändrade) = ws.ändrade_filer() {
        if !ändrade.is_empty() {
            return ändrade;
        }
    }

    let mut filer: Vec<String> = Vec::new();
    gå(ws.root(), ws.root(), &mut filer, gräns);

    filer
}

/// Samlar filer i en katalog och prioriterar källkod.
///
/// Go: `walk`.
fn gå(root: &Path, katalog: &Path, filer: &mut Vec<String>, gräns: usize) {
    if filer.len() >= gräns {
        return;
    }

    let Ok(poster) = std::fs::read_dir(katalog) else {
        return;
    };

    let mut poster: Vec<PathBuf> = poster.flatten().map(|post| post.path()).collect();
    poster.sort();

    for sökväg in poster {
        if filer.len() >= gräns {
            return;
        }

        let Some(namn) = sökväg.file_name().and_then(|n| n.to_str()) else {
            continue;
        };

        if namn.starts_with('.') {
            continue;
        }

        if sökväg.is_dir() {
            if namn == "node_modules" || namn == "vendor" {
                continue;
            }

            gå(root, &sökväg, filer, gräns);

            continue;
        }

        if !är_relevant_källkod(namn) {
            continue;
        }

        let Ok(innehåll) = std::fs::read_to_string(&sökväg) else {
            continue;
        };

        filer.push(format!(
            "### {}\n{}",
            relativ_sökväg(root, &sökväg.to_string_lossy()),
            innehåll
        ));
    }
}

/// Systemprompten som används.
pub const SYSTEMPROMPT: &str = r#"You are a software engineer proposing a minimal code change.

You get a goal and the files of a repository. Propose the smallest change that
fulfils the goal, together with a test that covers it.

Rules:
- Change as little as possible. Do not refactor unrelated code.
- Always include a test that verifies the change.
- Only propose files that exist in the repository, or new files whose path you
  state explicitly.
- Never propose changes to secrets, CI configuration or dependency locks.
- Keep the existing content of every file you change. You propose the full
  new content of the file, so you must repeat everything that is already
  there - only add, change or remove what the goal actually requires.
- If you change a file, keep its existing structure, headings and tone.

Answer in exactly this format:

TITLE: [one line]
DESCRIPTION: [why this change fulfils the goal]
TEST_COMMAND: [the shell command that verifies the change]

FILE: [path]
CONTENT:
[the full new content of the file]
END_FILE
"#;

/// Systemprompten för revidering.
pub const REVISION_SYSTEMPROMPT: &str = r#"You are a software engineer correcting a proposed change.

Your previous proposal was verified with its own test command and the test
failed. The most common reason is that the test looks for text that the
proposed content does not contain, or that the text is broken across lines.

Rules:
- Keep the goal and the intent of the previous proposal.
- Make the test command pass for the content you propose. Check every string
  the test looks for against the content, and be careful with line breaks.
- Prefer a test that checks structure over one that checks exact sentences.
- Keep the existing content of every file you change.

Answer in exactly this format:

TITLE: [one line]
DESCRIPTION: [why this change fulfils the goal]
TEST_COMMAND: [the shell command that verifies the change]

FILE: [path]
CONTENT:
[the full new content of the file]
END_FILE
"#;

/// Systemprompten för att laga ett testkommando.
pub const FIXTEST_SYSTEMPROMPT: &str = r#"You are a software engineer fixing a test command.

The content of the change stays exactly as it is. Only the test command is
wrong - it checks for something the content does not contain, or it matches
text that is broken across lines.

Rules:
- The new test command must pass on the given content.
- Check every string the test looks for against the content, word by word.
- Prefer a test that checks structure - a heading, a function, a section -
  over one that checks exact sentences.
- The test must still verify the intent of the goal.
- Keep the test a single shell command that can be run from the repository
  root.

Answer with exactly one line:

TEST_COMMAND: [the shell command]
"#;

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn förslag_läses_ur_svaret() {
        let svar = "TITLE: Lägg till test\nDESCRIPTION: För att testa\nTEST_COMMAND: go test ./...\n\nFILE: src/main.go\nCONTENT:\npackage main\n\nfunc main() {}\nEND_FILE\n";

        let förslag = tolka_förslag(svar).unwrap();

        assert_eq!(förslag.title, "Lägg till test");
        assert_eq!(förslag.description, "För att testa");
        assert_eq!(förslag.test_command, "go test ./...");
        assert_eq!(förslag.changes.len(), 1);
        assert_eq!(förslag.changes[0].path, "src/main.go");
        assert_eq!(förslag.changes[0].content, "package main\n\nfunc main() {}");
    }

    #[test]
    fn flera_filer_läses() {
        let svar = "TITLE: Två filer\nFILE: a.txt\nCONTENT:\nA\nEND_FILE\nFILE: b.txt\nCONTENT:\nB\nEND_FILE\n";

        let förslag = tolka_förslag(svar).unwrap();

        assert_eq!(förslag.changes.len(), 2);
        assert_eq!(förslag.changes[1].path, "b.txt");
        assert_eq!(förslag.changes[1].content, "B");
    }

    #[test]
    fn utan_filer_avvisas() {
        assert!(tolka_förslag("TITLE: Inget\n").is_err());
    }

    #[test]
    fn tom_titel_får_standardvärde() {
        let svar = "FILE: a.txt\nCONTENT:\nA\nEND_FILE\n";

        let förslag = tolka_förslag(svar).unwrap();

        assert_eq!(förslag.title, "Föreslagen ändring");
    }

    #[test]
    fn testkommandot_läses() {
        assert_eq!(
            tolka_testkommando("TEXT\nTEST_COMMAND: go test ./...\n"),
            "go test ./..."
        );
        assert_eq!(tolka_testkommando("inget"), "");
    }

    #[test]
    fn heredoc_läses() {
        assert_eq!(
            heredoc_terminator("python3 - <<'PY'"),
            Some("PY".to_string())
        );
        assert_eq!(heredoc_terminator("go test ./..."), None);
    }

    #[test]
    fn källkod_filtreras() {
        assert!(är_relevant_källkod("main.go"));
        assert!(är_relevant_källkod("README.md"));
        assert!(!är_relevant_källkod("bild.png"));
    }
}
