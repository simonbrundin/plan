//! Isolerar agentens kodändringar från originalet.
//!
//! Go: `internal/workspace`. Agenten får aldrig skriva direkt i det
//! repo som utvecklas. Istället arbetar den i en kopia, och ändringen
//! diffas tillbaka så att en människa kan granska den innan den
//! appliceras.

use std::path::{Path, PathBuf};
use std::process::Command;

/// Ändringen kom inte ut till källrepoet.
///
/// Ändringen är då gjord och verifierad, men ligger bara lokalt.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct ErrNotPublished;

impl std::fmt::Display for ErrNotPublished {
    fn fmt(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(formatter, "change was not published")
    }
}

impl std::error::Error for ErrNotPublished {}

/// En isolerad kopia av ett repository.
pub struct Workspace {
    /// Sökvägen till kopian.
    root: PathBuf,
    /// Repot som kopierades.
    source: PathBuf,
}

impl Workspace {
    /// Skapar en isolerad kopia av `source` i en temporär katalog.
    ///
    /// Go: `workspace.New`.
    pub fn ny(source: &str) -> Result<Self, String> {
        let abs_source = PathBuf::from(source)
            .canonicalize()
            .map_err(|fel| format!("failed to resolve source path: {fel}"))?;

        if !abs_source.is_dir() {
            return Err(format!(
                "source {} is not a directory",
                abs_source.display()
            ));
        }

        let root = std::env::temp_dir().join(format!("agent-workspace-{}", rand::random::<u64>()));

        std::fs::create_dir_all(&root)
            .map_err(|fel| format!("failed to create workspace directory: {fel}"))?;

        let mut ws = Self {
            root,
            source: abs_source,
        };

        if let Err(fel) = ws.kopiera_källan() {
            // Katalogen ska inte lämnas kvar - den innehåller koden.
            let _ = std::fs::remove_dir_all(&ws.root);

            return Err(fel);
        }

        Ok(ws)
    }

    /// Kopierar repot till arbetsytan.
    ///
    /// git används när källan är ett git-repository - annars kopieras
    /// filerna rakt av.
    ///
    /// Go: `Workspace.copySource`.
    fn kopiera_källan(&mut self) -> Result<(), String> {
        if är_git_repository(&self.source) {
            let utdata = Command::new("git")
                .args(["clone", "--local", "--no-hardlinks"])
                .arg(&self.source)
                .arg(&self.root)
                .output()
                .map_err(|fel| format!("failed to clone repository: {fel}"))?;

            if !utdata.status.success() {
                return Err(format!(
                    "failed to clone repository: {}",
                    String::from_utf8_lossy(&utdata.stderr).trim()
                ));
            }

            self.sätt_identitet()?;
            self.återställ_remote()?;

            return Ok(());
        }

        kopiera_träd(&self.source, &self.root)
            .map_err(|fel| format!("failed to copy repository: {fel}"))
    }

    /// Sätter git-identiteten för arbetsytan.
    ///
    /// Go: `Workspace.setIdentity`.
    fn sätt_identitet(&self) -> Result<(), String> {
        self.kör("git", &["config", "user.email", "agent@plan.local"])
            .map_err(|fel| format!("failed to set git email: {fel}"))?;

        self.kör("git", &["config", "user.name", "Plan Agent"])
            .map_err(|fel| format!("failed to set git name: {fel}"))?;

        Ok(())
    }

    /// Sätter tillbaka källans remote-adress.
    ///
    /// Go: `Workspace.restoreRemote`.
    fn återställ_remote(&self) -> Result<(), String> {
        let utdata = Command::new("git")
            .args(["remote", "get-url", "origin"])
            .current_dir(&self.source)
            .output();

        let Ok(utdata) = utdata else {
            // Källan kan sakna remote - då behålls den lokala adressen.
            return Ok(());
        };

        if !utdata.status.success() {
            return Ok(());
        }

        let remote = String::from_utf8_lossy(&utdata.stdout).trim().to_string();

        if remote.is_empty() || remote == self.source.display().to_string() {
            return Ok(());
        }

        self.kör("git", &["remote", "set-url", "origin", &remote])
            .map(|_| ())
            .map_err(|fel| format!("failed to set remote origin: {fel}"))
    }

    /// Sökvägen till arbetsytan.
    pub fn root(&self) -> &Path {
        &self.root
    }

    /// Sökvägen till repot som kopierades.
    pub fn source(&self) -> &Path {
        &self.source
    }

    /// Skriver en fil i arbetsytan och skapar kataloger vid behov.
    ///
    /// Go: `Workspace.Write`.
    pub fn skriv(&self, relativ_sökväg: &str, innehåll: &str) -> Result<(), String> {
        let mål = self.lös(relativ_sökväg)?;

        if let Some(katalog) = mål.parent() {
            std::fs::create_dir_all(katalog)
                .map_err(|fel| format!("failed to create directory: {fel}"))?;
        }

        std::fs::write(&mål, innehåll).map_err(|fel| format!("failed to write file: {fel}"))
    }

    /// Läser en fil från arbetsytan.
    ///
    /// Go: `Workspace.Read`.
    pub fn läs(&self, relativ_sökväg: &str) -> Result<String, String> {
        let mål = self.lös(relativ_sökväg)?;

        std::fs::read_to_string(&mål).map_err(|fel| format!("failed to read file: {fel}"))
    }

    /// Berättar om en fil finns i arbetsytan.
    ///
    /// Go: `Workspace.Exists`.
    pub fn finns(&self, relativ_sökväg: &str) -> bool {
        match self.lös(relativ_sökväg) {
            Ok(mål) => mål.exists(),
            Err(_) => false,
        }
    }

    /// Gör om en relativ sökväg till en sökväg i arbetsytan.
    ///
    /// Sökvägar som försöker lämna arbetsytan avvisas - agenten ska inte
    /// kunna skriva till originalet eller till systemkataloger.
    ///
    /// Go: `Workspace.resolve`.
    fn lös(&self, relativ_sökväg: &str) -> Result<PathBuf, String> {
        if relativ_sökväg.is_empty() {
            return Err("empty path".to_string());
        }

        if Path::new(relativ_sökväg).is_absolute() {
            return Err(format!("absolute paths are not allowed: {relativ_sökväg}"));
        }

        let mål = normalisera(&self.root.join(relativ_sökväg));
        let rot = normalisera(&self.root);

        if !mål.starts_with(&rot) {
            return Err(format!("path escapes the workspace: {relativ_sökväg}"));
        }

        Ok(mål)
    }

    /// Returnerar skillnaden mellan originalet och arbetsytan.
    ///
    /// Go: `Workspace.Diff`.
    pub fn diff(&self) -> Result<String, String> {
        if är_git_repository(&self.root) {
            let utdata = Command::new("git")
                .args(["diff", "--no-color"])
                .current_dir(&self.root)
                .output()
                .map_err(|fel| format!("failed to produce diff: {fel}"))?;

            if !utdata.status.success() {
                return Err(format!(
                    "failed to produce diff: {}",
                    String::from_utf8_lossy(&utdata.stderr).trim()
                ));
            }

            return Ok(String::from_utf8_lossy(&utdata.stdout).to_string());
        }

        if which("git").is_some() {
            let utdata = Command::new("git")
                .args(["diff", "--no-color", "--no-index"])
                .arg(&self.source)
                .arg(&self.root)
                .output();

            if let Ok(utdata) = utdata {
                return Ok(String::from_utf8_lossy(&utdata.stdout).to_string());
            }
        }

        diff_träd(&self.source, &self.root)
    }

    /// Returnerar de filer som ändrats i arbetsytan.
    ///
    /// Go: `Workspace.ChangedFiles`.
    pub fn ändrade_filer(&self) -> Result<Vec<String>, String> {
        if är_git_repository(&self.root) {
            let utdata = Command::new("git")
                .args(["status", "--porcelain"])
                .current_dir(&self.root)
                .output()
                .map_err(|fel| format!("failed to list changed files: {fel}"))?;

            if !utdata.status.success() {
                return Err(format!(
                    "failed to list changed files: {}",
                    String::from_utf8_lossy(&utdata.stderr).trim()
                ));
            }

            return Ok(tolka_git_status(&String::from_utf8_lossy(&utdata.stdout)));
        }

        ändrade_filer(&self.source, &self.root)
    }

    /// Utför ett kommando i arbetsytan och returnerar utdata.
    ///
    /// Go: `Workspace.Run`.
    pub fn kör(&self, kommando: &str, argument: &[&str]) -> Result<String, String> {
        let utdata = Command::new(kommando)
            .args(argument)
            .current_dir(&self.root)
            .output()
            .map_err(|fel| format!("command {kommando:?} failed: {fel}"))?;

        let text = String::from_utf8_lossy(&utdata.stdout).to_string()
            + &String::from_utf8_lossy(&utdata.stderr);

        if !utdata.status.success() {
            return Err(format!(
                "command {kommando:?} failed: exit status {}: {}",
                utdata.status.code().unwrap_or(-1),
                text.trim()
            ));
        }

        Ok(text)
    }

    /// Utför ett kommando med skalning i arbetsytan.
    ///
    /// Go: `Workspace.RunShell`.
    pub fn kör_shell(&self, kommando: &str) -> Result<String, String> {
        let utdata = Command::new("sh")
            .arg("-c")
            .arg(kommando)
            .current_dir(&self.root)
            .output()
            .map_err(|fel| format!("command {kommando:?} failed: {fel}"))?;

        let text = String::from_utf8_lossy(&utdata.stdout).to_string()
            + &String::from_utf8_lossy(&utdata.stderr);

        if !utdata.status.success() {
            return Err(format!(
                "command {kommando:?} failed: exit status {}: {}",
                utdata.status.code().unwrap_or(-1),
                text.trim()
            ));
        }

        Ok(text)
    }

    /// Committer ändringarna i arbetsytan och skickar branchen till
    /// källrepoet.
    ///
    /// Go: `Workspace.CommitToBranch`.
    pub fn committa_till_branch(&self, branch: &str, meddelande: &str) -> Result<(), String> {
        if !är_git_repository(&self.root) {
            return Err("workspace is not a git repository".to_string());
        }

        self.kör("git", &["checkout", "-b", branch])
            .map_err(|fel| format!("failed to create branch {branch}: {fel}"))?;

        self.kör("git", &["add", "-A"])
            .map_err(|fel| format!("failed to stage changes: {fel}"))?;

        self.kör("git", &["commit", "-m", meddelande])
            .map_err(|fel| format!("failed to commit changes: {fel}"))?;

        self.pusha_branch(branch)
            .map_err(|fel| format!("{ErrNotPublished}: {fel}"))
    }

    /// Skickar branchen till källrepoet.
    ///
    /// Go: `Workspace.pushBranch`.
    fn pusha_branch(&self, branch: &str) -> Result<(), String> {
        let mut argument: Vec<String> = Vec::new();

        if let Ok(header) = std::env::var("GIT_AUTH_HEADER") {
            if !header.is_empty() {
                argument.push("-c".to_string());
                argument.push(format!("http.extraheader={header}"));
            }
        }

        argument.push("push".to_string());
        argument.push("origin".to_string());
        argument.push(branch.to_string());

        let argument: Vec<&str> = argument.iter().map(|v| v.as_str()).collect();

        self.kör("git", &argument).map(|_| ())
    }

    /// Skapar eller återanvänder en Pull Request för den publicerade
    /// branchen.
    ///
    /// Go: `Workspace.CreatePullRequest`.
    pub fn skapa_pull_request(
        &self,
        titel: &str,
        brödtext: &str,
        bas: &str,
    ) -> Result<String, String> {
        let branch = self
            .kör("git", &["branch", "--show-current"])
            .map_err(|fel| format!("failed to determine pull request branch: {fel}"))?
            .trim()
            .to_string();

        if branch.is_empty() {
            return Err("cannot create pull request without a branch".to_string());
        }

        let bas = if bas.is_empty() { "main" } else { bas };

        let mut env: Vec<(String, String)> = Vec::new();

        let har_token = std::env::var("GH_TOKEN").is_ok_and(|v| !v.is_empty())
            || std::env::var("GITHUB_TOKEN").is_ok_and(|v| !v.is_empty());

        if !har_token {
            if let Ok(header) = std::env::var("GIT_AUTH_HEADER") {
                if let Some(token) = token_från_git_header(&header) {
                    env.push(("GH_TOKEN".to_string(), token));
                }
            }
        }

        // En PR som redan finns för samma branch återanvänds - det gör
        // leveransen idempotent vid retries. Ett fel från listningen
        // ignoreras - då får skapandet försöka.
        if let Ok(url) = self.kör_gh(
            "gh",
            &[
                "pr", "list", "--head", &branch, "--base", bas, "--state", "all", "--json", "url",
                "--jq", ".[0].url",
            ],
            &env,
        ) {
            let url = url.trim();

            if !url.is_empty() {
                return Ok(url.to_string());
            }
        }

        let utdata = self
            .kör_gh(
                "gh",
                &[
                    "pr", "create", "--head", &branch, "--base", bas, "--title", titel, "--body",
                    brödtext,
                ],
                &env,
            )
            .map_err(|fel| format!("failed to create pull request: {fel}"))?;

        Ok(utdata.trim().to_string())
    }

    /// Kör ett kommando med extra miljövariabler.
    ///
    /// Go: `Workspace.CreatePullRequest` kör `gh` direkt, så felmeddelandet
    /// skiljer sig från `Workspace.Run`.
    fn kör_gh(
        &self,
        kommando: &str,
        argument: &[&str],
        env: &[(String, String)],
    ) -> Result<String, String> {
        let mut cmd = Command::new(kommando);
        cmd.args(argument).current_dir(&self.root);

        for (nyckel, värde) in env {
            cmd.env(nyckel, värde);
        }

        let utdata = cmd
            .output()
            .map_err(|_| format!("exec: {kommando:?}: executable file not found in $PATH: "))?;

        let text = String::from_utf8_lossy(&utdata.stdout).to_string()
            + &String::from_utf8_lossy(&utdata.stderr);

        if !utdata.status.success() {
            return Err(format!(
                "exit status {}: {}",
                utdata.status.code().unwrap_or(-1),
                text.trim()
            ));
        }

        Ok(text)
    }

    /// Tar bort arbetsytan.
    ///
    /// Go: `Workspace.Cleanup`.
    pub fn städa(&self) -> Result<(), String> {
        if self.root.as_os_str().is_empty() {
            return Ok(());
        }

        std::fs::remove_dir_all(&self.root).map_err(|fel| fel.to_string())
    }
}

impl Drop for Workspace {
    fn drop(&mut self) {
        // Arbetsytan tas bort när den inte längre behövs - koden i den
        // ska inte ligga kvar i poden.
        let _ = std::fs::remove_dir_all(&self.root);
    }
}

/// Berättar om källan är ett git-repository.
///
/// Go: `isGitRepository`.
pub fn är_git_repository(sökväg: &Path) -> bool {
    if which("git").is_none() {
        return false;
    }

    Command::new("git")
        .args(["rev-parse", "--git-dir"])
        .current_dir(sökväg)
        .status()
        .map(|status| status.success())
        .unwrap_or(false)
}

/// Sökvägen till ett program i PATH.
fn which(program: &str) -> Option<PathBuf> {
    let path = std::env::var_os("PATH")?;

    std::env::split_paths(&path)
        .map(|katalog| katalog.join(program))
        .find(|kandidat| kandidat.is_file())
}

/// Normaliserar en sökväg utan att kräva att den finns.
fn normalisera(sökväg: &Path) -> PathBuf {
    let mut delar: Vec<std::ffi::OsString> = Vec::new();

    for del in sökväg.components() {
        match del {
            std::path::Component::ParentDir => {
                delar.pop();
            }
            std::path::Component::CurDir => {}
            annan => delar.push(annan.as_os_str().to_os_string()),
        }
    }

    let mut resultat = PathBuf::new();

    for del in delar {
        resultat.push(del);
    }

    resultat
}

/// Kopierar en katalog rekursivt.
///
/// Go: `copyTree`.
pub fn kopiera_träd(källa: &Path, mål: &Path) -> Result<(), String> {
    let poster = std::fs::read_dir(källa).map_err(|fel| fel.to_string())?;

    for post in poster {
        let post = post.map_err(|fel| fel.to_string())?;
        let typ = post.file_type().map_err(|fel| fel.to_string())?;
        let mål_fil = mål.join(post.file_name());

        if typ.is_dir() {
            std::fs::create_dir_all(&mål_fil).map_err(|fel| fel.to_string())?;

            kopiera_träd(&post.path(), &mål_fil)?;

            continue;
        }

        if !typ.is_file() {
            continue;
        }

        // Rättigheterna följer med - ett skript som ska kunna köras ska
        // fortfarande kunna det i arbetsytan.
        #[cfg(unix)]
        {
            if let Ok(metadata) = post.metadata() {
                std::fs::set_permissions(&mål_fil, metadata.permissions())
                    .map_err(|fel| fel.to_string())?;
            }
        }

        std::fs::copy(post.path(), &mål_fil).map_err(|fel| fel.to_string())?;
    }

    Ok(())
}

/// Tolkar `git status --porcelain` och ger filnamnen.
///
/// Go: `parseGitStatus`.
pub fn tolka_git_status(utdata: &str) -> Vec<String> {
    let mut filer = Vec::new();

    for rad in utdata.split('\n') {
        if rad.len() < 4 {
            continue;
        }

        // Radens inledande blankstug hör till statusen och ska inte tas
        // bort före slicing, då blir sökvägen förkortad.
        filer.push(rad.get(3..).unwrap_or_default().trim().to_string());
    }

    filer
}

/// Jämför två kataloger och returnerar en enkel diff.
///
/// Go: `diffTrees`.
pub fn diff_träd(källa: &Path, mål: &Path) -> Result<String, String> {
    let ändrade = ändrade_filer(källa, mål)?;

    let mut text = String::new();

    for fil in ändrade {
        text.push_str(&format!("--- a/{fil}\n"));
        text.push_str(&format!("+++ b/{fil}\n"));

        let original = std::fs::read(källa.join(&fil)).unwrap_or_default();
        let uppdaterad = std::fs::read(mål.join(&fil)).unwrap_or_default();

        text.push_str(&format!(
            "@@ original {} bytes, updated {} bytes @@\n",
            original.len(),
            uppdaterad.len()
        ));
    }

    Ok(text)
}

/// Jämför filerna i två kataloger.
///
/// Go: `changedFiles`.
pub fn ändrade_filer(källa: &Path, mål: &Path) -> Result<Vec<String>, String> {
    let mut ändrade: Vec<String> = Vec::new();

    samla_filer(mål, &mut |sökväg| {
        let relativ = sökväg
            .strip_prefix(mål)
            .map(|p| p.to_string_lossy().to_string())
            .unwrap_or_default();

        let original = std::fs::read(källa.join(&relativ));
        let uppdaterad = std::fs::read(sökväg);

        match (original, uppdaterad) {
            (Ok(original), Ok(uppdaterad)) => {
                if original != uppdaterad {
                    ändrade.push(relativ);
                }
            }
            _ => ändrade.push(relativ),
        }
    })?;

    Ok(ändrade)
}

/// Går igenom en katalog rekursivt.
fn samla_filer(katalog: &Path, besökare: &mut impl FnMut(&Path)) -> Result<(), String> {
    let poster = std::fs::read_dir(katalog).map_err(|fel| fel.to_string())?;

    for post in poster {
        let post = post.map_err(|fel| fel.to_string())?;
        let typ = post.file_type().map_err(|fel| fel.to_string())?;

        if typ.is_dir() {
            samla_filer(&post.path(), besökare)?;

            continue;
        }

        besökare(&post.path());
    }

    Ok(())
}

/// Hämtar tokenen ur en git-auth-header.
///
/// Go: `tokenFromGitAuthHeader`.
pub fn token_från_git_header(header: &str) -> Option<String> {
    const PREFIX: &str = "Authorization: Basic ";

    let rest = header.strip_prefix(PREFIX)?.trim();

    let avkodad = String::from_utf8(base64_avkod(rest).ok()?).ok()?;
    let (_, token) = avkodad.split_once(':')?;

    Some(token.to_string())
}

/// Avkodar base64.
fn base64_avkod(text: &str) -> Result<Vec<u8>, String> {
    const TECKEN: &[u8] = b"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";

    let mut värden: Vec<u8> = Vec::new();
    let mut buffert: u32 = 0;
    let mut antal_bitar = 0u32;

    for tecken in text.bytes() {
        if tecken == b'=' {
            break;
        }

        let Some(position) = TECKEN.iter().position(|t| *t == tecken) else {
            continue;
        };

        buffert = (buffert << 6) | position as u32;
        antal_bitar += 6;

        if antal_bitar >= 8 {
            antal_bitar -= 8;
            värden.push(((buffert >> antal_bitar) & 0xff) as u8);
        }
    }

    Ok(värden)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn git_status_tolkas() {
        let status = " M src/main.rs\n?? nyfil.txt\nA  tillagd.txt\n";

        assert_eq!(
            tolka_git_status(status),
            vec!["src/main.rs", "nyfil.txt", "tillagd.txt"]
        );
    }

    #[test]
    fn sökvägar_får_inte_lämna_arbetsytan() {
        let arbetsyta = Workspace {
            root: PathBuf::from("/tmp/agent-workspace-test"),
            source: PathBuf::from("/tmp/källa"),
        };

        assert!(arbetsyta.lös("../utomhus").is_err());
        assert!(arbetsyta.lös("/etc/passwd").is_err());
        assert!(arbetsyta.lös("").is_err());
        assert!(arbetsyta.lös("src/main.rs").is_ok());
    }

    #[test]
    fn git_header_avkodas() {
        let header = "Authorization: Basic dXNlcjpoZW1saWd0";

        assert_eq!(token_från_git_header(header), Some("hemligt".to_string()));
        assert_eq!(token_från_git_header("annat"), None);
    }
}
