//! Tester som kör mot riktiga git-repomed i en temporär katalog.
//!
//! Testerna täcker det som agenten faktiskt gör: kopiera ett repo, ändra
//! filer, diffa, köra tester och committa på en branch.

use std::path::{Path, PathBuf};
use std::process::Command;

use plan_agent::executor::{Change, Executor, Request};
use plan_agent::workspace::Workspace;

/// Skapar ett git-repo med en fil.
fn skapa_repo(namn: &str) -> PathBuf {
    let sökväg =
        std::env::temp_dir().join(format!("agent-testrepo-{namn}-{}", rand::random::<u64>()));

    std::fs::create_dir_all(&sökväg).unwrap();

    let git = |argument: &[&str]| {
        let utdata = Command::new("git")
            .args(argument)
            .current_dir(&sökväg)
            .output()
            .unwrap();

        assert!(
            utdata.status.success(),
            "git {:?} misslyckades: {}",
            argument,
            String::from_utf8_lossy(&utdata.stderr)
        );
    };

    git(&["init", "-q", "-b", "main"]);
    git(&["config", "user.email", "test@plan.local"]);
    git(&["config", "user.name", "Test"]);

    std::fs::write(sökväg.join("README.md"), "# Hej\n").unwrap();
    git(&["add", "-A"]);
    git(&["commit", "-q", "-m", "start"]);

    sökväg
}

/// Kör ett git-kommando i ett repo.
fn git(sökväg: &Path, argument: &[&str]) -> String {
    let utdata = Command::new("git")
        .args(argument)
        .current_dir(sökväg)
        .output()
        .unwrap();

    assert!(utdata.status.success());

    String::from_utf8_lossy(&utdata.stdout).to_string()
}

#[test]
fn workspace_kopierar_repomed_och_ändringen_stannar_i_kopian() {
    let repo = skapa_repo("kopiering");

    let ws = Workspace::ny(repo.to_str().unwrap()).unwrap();

    assert!(ws.root().starts_with(std::env::temp_dir()));
    assert_eq!(ws.source(), repo.canonicalize().unwrap());

    // Originalet ska vara orört.
    assert_eq!(
        std::fs::read_to_string(repo.join("README.md")).unwrap(),
        "# Hej\n"
    );

    ws.skriv("README.md", "# Hej hej\n").unwrap();

    assert_eq!(ws.läs("README.md").unwrap(), "# Hej hej\n");
    assert_eq!(
        std::fs::read_to_string(repo.join("README.md")).unwrap(),
        "# Hej\n"
    );

    assert_eq!(ws.ändrade_filer().unwrap(), vec!["README.md".to_string()]);

    let diff = ws.diff().unwrap();

    assert!(diff.contains("README.md"), "diffen saknar filen: {diff}");

    ws.städa().unwrap();
}

#[test]
fn workspace_avvisar_sökvägar_som_lämner_kopian() {
    let repo = skapa_repo("traversal");

    let ws = Workspace::ny(repo.to_str().unwrap()).unwrap();

    assert!(ws.skriv("../utomhus.txt", "nej").is_err());
    assert!(ws.skriv("/etc/passwd", "nej").is_err());
    assert!(ws.skriv("a/../../utomhus.txt", "nej").is_err());
    assert!(ws.läs("../utomhus.txt").is_err());
    assert!(!ws.finns("../utomhus.txt"));

    ws.städa().unwrap();
}

#[test]
fn workspace_kör_kommandon_i_kopian() {
    let repo = skapa_repo("kommandon");

    let ws = Workspace::ny(repo.to_str().unwrap()).unwrap();

    let utdata = ws.kör("pwd", &[]).unwrap();

    assert!(utdata.trim().contains("agent-workspace-"));

    // Ett felande kommando ska ge ett fel med utdatan.
    let fel = ws.kör("sh", &["-c", "echo fel >&2; exit 3"]).unwrap_err();

    assert!(fel.contains("exit status 3"), "felmeddelandet: {fel}");
    assert!(fel.contains("fel"), "felmeddelandet: {fel}");

    ws.städa().unwrap();
}

#[tokio::test]
async fn executor_applicerar_ändringen_och_kör_testen() {
    let repo = skapa_repo("exekvering");

    let exekverare = Executor::ny(repo.to_str().unwrap()).unwrap();

    let resultat = exekverare
        .execute(&Request {
            changes: vec![Change {
                path: "README.md".to_string(),
                content: "# Hej hej\n".to_string(),
                reason: "test".to_string(),
            }],
            test_command: "grep -q hej README.md".to_string(),
        })
        .await
        .unwrap();

    assert!(
        resultat.tests_passed,
        "testet gick inte igenom: {resultat:?}"
    );
    assert_eq!(resultat.changed_files, vec!["README.md".to_string()]);
    assert!(
        resultat.diff.contains("README.md"),
        "diffen: {}",
        resultat.diff
    );
    assert!(resultat.test_error.is_empty());
}

#[tokio::test]
async fn executor_rapporterar_misslyckade_tester() {
    let repo = skapa_repo("misslyckade");

    let exekverare = Executor::ny(repo.to_str().unwrap()).unwrap();

    let resultat = exekverare
        .execute(&Request {
            changes: vec![Change {
                path: "fil.txt".to_string(),
                content: "hej\n".to_string(),
                reason: String::new(),
            }],
            test_command: "grep -q saknas fil.txt".to_string(),
        })
        .await
        .unwrap();

    assert!(!resultat.tests_passed);
    assert!(!resultat.test_error.is_empty());
}

#[tokio::test]
async fn committa_lämnar_huvudgrenen_orörd() {
    let repo = skapa_repo("commit");

    let exekverare = Executor::ny(repo.to_str().unwrap()).unwrap();

    exekverare
        .execute(&Request {
            changes: vec![Change {
                path: "fil.txt".to_string(),
                content: "hej\n".to_string(),
                reason: String::new(),
            }],
            test_command: String::new(),
        })
        .await
        .unwrap();

    // Ändringen landar på en egen branch - huvudgrenen rörs aldrig.
    exekverare.committa("agent/test", "test").unwrap();

    let brancher = git(&repo, &["branch", "--list", "agent/test"]);

    assert!(brancher.contains("agent/test"), "brancherna: {brancher}");

    let innehåll = git(&repo, &["show", "agent/test:fil.txt"]);

    assert_eq!(innehåll, "hej\n");

    // Huvudgrenen ska vara oförändrad - ändringen finns bara på branchen.
    let huvud = git(&repo, &["show", "main:README.md"]);

    assert_eq!(huvud, "# Hej\n");
}
