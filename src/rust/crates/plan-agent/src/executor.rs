//! Gör kodändringar i en isolerad workspace och verifierar dem med
//! projektets egna tester.
//!
//! Go: `internal/executor`. Paketet är medvetet litet: agenten föreslår
//! en ändring, workern utför den i en kopia av repot och rapporterar
//! resultatet tillbaka till Plan. Originalet rörs aldrig - det är en
//! människa som beslutar om ändringen ska appliceras.

use serde_json::json;

use crate::runusage::Recorder;
use crate::workspace::Workspace;

/// En föreslagen ändring i en fil.
#[derive(Debug, Clone, Default)]
pub struct Change {
    /// Sökvägen relativt repo-roten.
    pub path: String,
    /// Det nya innehållet.
    pub content: String,
    /// Varför ändringen görs.
    pub reason: String,
}

/// En kodändring som ska utföras.
#[derive(Debug, Clone, Default)]
pub struct Request {
    /// De ändringar som ska göras.
    pub changes: Vec<Change>,
    /// Kommandot som verifierar ändringen.
    pub test_command: String,
}

/// Resultatet av en utförd ändring.
///
/// Go: `executor.Result`.
#[derive(Debug, Clone, Default)]
pub struct Utfall {
    /// De filer som ändrats.
    pub changed_files: Vec<String>,
    /// Skillnaden mot originalet.
    pub diff: String,
    /// Utdata från testkörningen.
    pub test_output: String,
    /// Om testerna gick igenom.
    pub tests_passed: bool,
    /// Varför testerna inte gick igenom.
    pub test_error: String,
    /// Branchen ändringen ligger på i källrepoet.
    pub branch: String,
}

/// Utför kodändringar i en isolerad workspace.
pub struct Executor {
    /// Arbetsytan.
    ws: Workspace,
    /// Recordern för tokenanvändning.
    usage: Option<std::sync::Arc<Recorder>>,
}

impl Executor {
    /// Skapar en executor som arbetar i en kopia av `source`.
    ///
    /// Go: `executor.New`.
    pub fn ny(source: &str) -> Result<Self, String> {
        let ws =
            Workspace::ny(source).map_err(|fel| format!("failed to create workspace: {fel}"))?;

        Ok(Self { ws, usage: None })
    }

    /// Kopplar en recorder för tokenanvändning till exekveringen.
    ///
    /// Go: `Executor.WithUsage`.
    pub fn med_användning(mut self, recorder: std::sync::Arc<Recorder>) -> Self {
        self.usage = Some(recorder);
        self
    }

    /// Arbetsytan - till exempel för att läsa filer.
    ///
    /// Go: `Executor.Workspace`.
    pub fn workspace(&self) -> &Workspace {
        &self.ws
    }

    /// Applicerar ändringarna och kör verifieringen.
    ///
    /// Ändringarna appliceras först, sedan körs testerna. Ett fel i
    /// testerna är inte ett fel i utförandet - resultatet ska beskriva
    /// utfallet, inte kasta bort det.
    ///
    /// Go: `Executor.Execute`.
    pub async fn execute(&self, req: &Request) -> Result<Utfall, String> {
        for ändring in &req.changes {
            self.ws
                .skriv(&ändring.path, &ändring.content)
                .map_err(|fel| format!("failed to apply change to {}: {fel}", ändring.path))?;

            self.registrera_verktyg("write_file");
        }

        let ändrade = self
            .ws
            .ändrade_filer()
            .map_err(|fel| format!("failed to list changed files: {fel}"))?;

        let diff = self
            .ws
            .diff()
            .map_err(|fel| format!("failed to produce diff: {fel}"))?;

        let mut resultat = Utfall {
            changed_files: ändrade,
            diff,
            ..Default::default()
        };

        if req.test_command.is_empty() {
            return Ok(resultat);
        }

        self.registrera_verktyg("run_test");

        let utdata = self.ws.kör_shell(&req.test_command);

        match utdata {
            Ok(text) => {
                resultat.test_output = text;
                resultat.tests_passed = true;
            }
            Err(fel) => {
                resultat.tests_passed = false;
                resultat.test_error = fel.clone();
                // Go sparar både utdata och felet - felet innehåller
                // utdatan, så ingen information går förlorad.
                resultat.test_output = fel;
            }
        }

        Ok(resultat)
    }

    /// Registrerar ett verktygsanrop, om en recorder är kopplad.
    fn registrera_verktyg(&self, namn: &str) {
        if let Some(recorder) = &self.usage {
            recorder.registrera_verktyg(namn);
        }
    }

    /// Lägger ändringen på en branch i källrepoet.
    ///
    /// Go: `Executor.Commit`.
    pub fn committa(&self, branch: &str, meddelande: &str) -> Result<(), String> {
        self.ws
            .committa_till_branch(branch, meddelande)
            .map_err(|fel| format!("failed to commit the change: {fel}"))
    }

    /// Skapar en idempotent Pull Request från den publicerade branchen.
    ///
    /// Go: `Executor.CreatePullRequest`.
    pub fn skapa_pull_request(
        &self,
        titel: &str,
        brödtext: &str,
        bas: &str,
    ) -> Result<String, String> {
        self.ws.skapa_pull_request(titel, brödtext, bas)
    }

    /// Tar bort arbetsytan.
    ///
    /// Go: `Executor.Cleanup`.
    pub fn städa(&self) -> Result<(), String> {
        self.ws.städa()
    }
}

impl Utfall {
    /// Beskriver resultatet på ett sätt som passar i en händelse.
    ///
    /// Go: `Result.Summary`.
    pub fn sammanfattning(&self) -> serde_json::Value {
        let mut summary = json!({
            "changed_files": self.changed_files,
            "tests_passed": self.tests_passed,
        });

        if !self.test_error.is_empty() {
            summary["test_error"] = json!(self.test_error);
        }

        if !self.diff.is_empty() {
            summary["diff_size"] = json!(self.diff.len());
        }

        summary
    }

    /// Returnerar de första raderna av diffen - en hel diff kan vara för
    /// stor för ett händelse-payload.
    ///
    /// Go: `Result.DiffPreview`.
    pub fn diff_förhandsvisning(&self, max_rader: usize) -> String {
        if max_rader == 0 {
            return String::new();
        }

        let rader: Vec<&str> = self.diff.split('\n').collect();

        if rader.len() <= max_rader {
            return rader.join("\n");
        }

        rader[..max_rader].join("\n")
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn sammanfattningen_innehåller_testerna() {
        let resultat = Utfall {
            changed_files: vec!["a.rs".to_string()],
            tests_passed: false,
            test_error: "fel".to_string(),
            diff: "diff".to_string(),
            ..Default::default()
        };

        let summary = resultat.sammanfattning();

        assert_eq!(summary["changed_files"][0], "a.rs");
        assert_eq!(summary["tests_passed"], false);
        assert_eq!(summary["test_error"], "fel");
        assert_eq!(summary["diff_size"], 4);
    }

    #[test]
    fn diff_förhandsvisningen_kapats() {
        let resultat = Utfall {
            diff: "a\nb\nc\nd".to_string(),
            ..Default::default()
        };

        assert_eq!(resultat.diff_förhandsvisning(2), "a\nb");
        assert_eq!(resultat.diff_förhandsvisning(0), "");
    }
}
