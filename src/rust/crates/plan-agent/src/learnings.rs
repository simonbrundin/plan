//! Lärdomar - det som agenten lärt sig under en körning.
//!
//! Port av `src/agent/internal/core/orchestrator/learnings.go`. Lärdomarna
//! sparas i en fil (standard `~/.config/agent/learnings.md`) och läses tillbaka
//! som kontext till agentens prompt.
//!
//! Filens format är Go:s - en ny lärdom läggs först, och rubriken läggs
//! på varje gång. Då bygger båda implementationerna samma fil.

use std::path::{Path, PathBuf};

use chrono::Local;

/// Rubriken som skrivs överst i filen.
const RUBRIK: &str =
    "# Learnings\n\nThis file contains lessons learned from checkpoint execution.\n\n";

/// Vad som står i filen när ingen lärdom sparats.
const TOM: &str = "# Learnings\n\nNo learnings recorded yet.\n";

/// Antalet lärdomar som tas med i prompten.
const SENASTE: usize = 10;

/// Hanterar lärdomar i en fil.
#[derive(Debug, Clone)]
pub struct LearningsManager {
    /// Filen lärdomarna sparas i.
    fil: PathBuf,
}

impl LearningsManager {
    /// Skapar en hanterare för en given fil.
    pub fn ny(fil: impl Into<PathBuf>) -> Self {
        Self { fil: fil.into() }
    }

    /// Filen lärdomarna sparas i.
    pub fn fil(&self) -> &Path {
        &self.fil
    }

    /// Standardfilen - `~/.config/agent/learnings.md`.
    ///
    /// Go faller tillbaka på samma sökväg när ingen fil är satt.
    pub fn standardfil() -> PathBuf {
        let hem = std::env::var("HOME").unwrap_or_default();

        PathBuf::from(hem)
            .join(".config")
            .join("agent")
            .join("learnings.md")
    }

    /// Sparar en lärdom.
    ///
    /// Den nya lärdomen läggs först i filen - de senaste lärdomarna är de
    /// som är relevanta.
    pub fn add_learning(
        &self,
        kategori: &str,
        titel: &str,
        beskrivning: &str,
        lärdom: &str,
    ) -> std::io::Result<()> {
        let datum = Local::now().format("%Y-%m-%d").to_string();

        let post = format!(
            "## {datum}: {titel}\n\n- **Date**: {datum}\n- **Category**: {kategori}\n- **Description**: {beskrivning}\n- **Lesson**: {lärdom}\n\n"
        );

        let existing = std::fs::read_to_string(&self.fil).unwrap_or_default();

        let innehåll = format!("{RUBRIK}{post}\n{existing}");

        if let Some(katalog) = self.fil.parent() {
            std::fs::create_dir_all(katalog)?;
        }

        std::fs::write(&self.fil, innehåll)
    }

    /// Sparar en lärdom från ett checkpoint.
    pub fn add_checkpoint_learning(
        &self,
        checkpoint: &str,
        problem: &str,
        _fix: &str,
        lärdom: &str,
    ) -> std::io::Result<()> {
        // Go skickar vidare fixen som beskrivning - kategorin blir
        // "checkpoint".
        self.add_learning("checkpoint", checkpoint, problem, lärdom)
    }

    /// Läser alla lärdomar.
    pub fn get_learnings(&self) -> std::io::Result<String> {
        match std::fs::read_to_string(&self.fil) {
            Ok(innehåll) => Ok(innehåll),
            Err(err) if err.kind() == std::io::ErrorKind::NotFound => Ok(TOM.to_string()),
            Err(err) => Err(err),
        }
    }

    /// Bygger kontext till prompten.
    ///
    /// Bara de tio senaste lärdomarna tas med. Om kategorier anges
    /// filtreras de på dem - annars tas alla med.
    pub fn generate_prompt_context(&self, kategorier: &[&str]) -> String {
        let Ok(innehåll) = self.get_learnings() else {
            return String::new();
        };

        let rader: Vec<&str> = innehåll.split('\n').collect();

        let mut relevant: Vec<&str> = Vec::new();
        let mut antal = 0;
        let mut ta_med = false;

        for rad in rader {
            if rad.starts_with("## 20") {
                antal += 1;

                if antal > SENASTE {
                    break;
                }

                ta_med = false;

                for kategori in kategorier {
                    if rad.contains(kategori) {
                        ta_med = true;
                        break;
                    }
                }
            }

            if rad.starts_with("- **Category**:") {
                for kategori in kategorier {
                    if rad.contains(kategori) {
                        ta_med = true;
                        break;
                    }
                }
            }

            // Inga kategorier angivna - då tas allt med.
            if kategorier.is_empty() {
                ta_med = true;
            }

            if ta_med {
                relevant.push(rad);
            }
        }

        if relevant.is_empty() {
            return String::new();
        }

        format!("\n\n## Relevant Learnings\n\n{}", relevant.join("\n"))
    }

    /// Lärdomar som hör till en kategori.
    pub fn get_learnings_by_category(&self, kategori: &str) -> std::io::Result<String> {
        let all = self.get_learnings()?;

        let mut filtrerade: Vec<&str> = Vec::new();
        let mut ta_med = false;

        for rad in all.split('\n') {
            if rad.starts_with("## 20") || rad.starts_with("# Learnings") {
                ta_med = rad.contains(kategori);
            }

            if ta_med || rad.starts_with("# Learnings") {
                filtrerade.push(rad);
            }
        }

        if filtrerade.is_empty() {
            return Ok(String::new());
        }

        Ok(filtrerade.join("\n"))
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    /// Skapar en hanterare i en temporär katalog.
    fn hanterare(i_temp: &tempfile::TempDir) -> LearningsManager {
        LearningsManager::ny(i_temp.path().join("learnings.md"))
    }

    #[test]
    fn tom_fil_ger_tomt_svar() {
        let temp = tempfile::tempdir().unwrap();
        let manager = hanterare(&temp);

        assert_eq!(
            manager.get_learnings().unwrap(),
            "# Learnings\n\nNo learnings recorded yet.\n"
        );
    }

    #[test]
    fn lärdom_sparas_med_go_format() {
        let temp = tempfile::tempdir().unwrap();
        let manager = hanterare(&temp);

        manager
            .add_learning("test", "En titel", "En beskrivning", "En lärdom")
            .unwrap();

        let innehåll = manager.get_learnings().unwrap();
        let datum = chrono::Local::now().format("%Y-%m-%d").to_string();

        // Go skriver rubriken som "## datum: titel" - datumet först.
        let förväntat = format!(
            "# Learnings\n\nThis file contains lessons learned from checkpoint execution.\n\n## {datum}: En titel\n\n- **Date**: {datum}\n- **Category**: test\n- **Description**: En beskrivning\n- **Lesson**: En lärdom\n\n\n"
        );

        assert_eq!(innehåll, förväntat);
    }

    #[test]
    fn senaste_lärdomen_kommer_först() {
        let temp = tempfile::tempdir().unwrap();
        let manager = hanterare(&temp);

        manager.add_learning("test", "Första", "", "En").unwrap();
        manager.add_learning("test", "Andra", "", "Två").unwrap();

        let innehåll = manager.get_learnings().unwrap();

        assert!(innehåll.find(": Andra").unwrap() < innehåll.find(": Första").unwrap());
    }

    #[test]
    fn promptkontext_tar_med_tio_senaste() {
        let temp = tempfile::tempdir().unwrap();
        let manager = hanterare(&temp);

        for i in 0..12 {
            manager
                .add_learning("test", &format!("Lärdom {i}"), "", "Lärdom")
                .unwrap();
        }

        let kontext = manager.generate_prompt_context(&[]);

        assert!(kontext.starts_with("\n\n## Relevant Learnings\n\n"));
        assert!(kontext.contains("Lärdom 11"));
        assert!(!kontext.contains("Lärdom 1 "));
    }

    #[test]
    fn kategorier_filtrerar() {
        let temp = tempfile::tempdir().unwrap();
        let manager = hanterare(&temp);

        manager
            .add_learning("test", "Testlärdom", "", "En")
            .unwrap();
        manager
            .add_learning("bug", "Bugglärdom", "", "Två")
            .unwrap();

        let kontext = manager.generate_prompt_context(&["bug"]);

        // Go börjar först vid den rad som matchar kategorin - rubriken
        // ovanför den kommer inte med.
        assert!(kontext.contains("- **Category**: bug"));
        assert!(!kontext.contains(": Bugglärdom"));
    }

    #[test]
    fn lärdomar_per_kategori() {
        let temp = tempfile::tempdir().unwrap();
        let manager = hanterare(&temp);

        manager
            .add_learning("test", "Testlärdom", "", "En")
            .unwrap();
        manager
            .add_learning("bug", "Bugglärdom", "", "Två")
            .unwrap();

        let filtrerade = manager.get_learnings_by_category("bug").unwrap();

        // Go filtrerar på rubriker, och det som matchar är bara
        // "# Learnings" - lärdomarna själva finns inte med.
        assert_eq!(filtrerade, "# Learnings\n# Learnings");
    }
}
