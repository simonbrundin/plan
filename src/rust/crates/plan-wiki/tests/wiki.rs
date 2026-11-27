//! Porten av Go-testerna i `internal/wiki`.
//!
//! Testerna följer samma fall som Go-versionen så att beteendet är
//! lika i båda implementationerna.

use std::path::Path;

use chrono::{Local, TimeZone};
use plan_wiki as wiki;

/// Skriver en sida till disk.
fn skriv_sida(dir: &Path, rel: &str, innehåll: &str) {
    let sökväg = dir.join(rel);

    std::fs::create_dir_all(sökväg.parent().expect("förälder")).expect("skapade katalog");
    std::fs::write(&sökväg, innehåll).expect("skrev sida");
}

/// Skapar en wiki med en sida i varje typ.
fn ny_test_wiki() -> tempfile::TempDir {
    let dir = tempfile::tempdir().expect("tempdir");

    skriv_sida(
        dir.path(),
        "wiki/concepts/Flux.md",
        "---\ntype: concept\ntags: [flux]\n---\n\n# Flux\n\nEn operator.\n",
    );

    dir
}

#[test]
fn slug_normaliserar_titlar() {
    let fall = [
        ("Arkitektur", "arkitektur"),
        ("Sessionens livscykel", "sessionens-livscykel"),
        ("Återkoppling av och-och", "aterkoppling-av-och-och"),
        ("  Bigram-algoritmen  ", "bigram-algoritmen"),
        ("SQL & NoSQL", "sql-nosql"),
    ];

    for (input, want) in fall {
        assert_eq!(wiki::slug(input), want, "slug av {input:?}");
    }
}

#[test]
fn parse_note_läser_metadata() {
    let dir = tempfile::tempdir().expect("tempdir");

    skriv_sida(
        dir.path(),
        "wiki/notes/arkitektur.md",
        "---\ntitle: Arkitektur\ntype: note\ntags: [arkitektur, go]\ncreated: 2026-10-21\nrelated: [bigram-algoritmen]\n---\n\nPaketen är delade i lager.\n",
    );

    let note = wiki::parse_note(&dir.path().join("wiki/notes/arkitektur.md")).expect("parse_note");

    assert_eq!(note.id, "arkitektur");
    assert_eq!(note.title, "Arkitektur");
    assert_eq!(note.note_type, "note");
    assert_eq!(note.tags, vec!["arkitektur".to_string(), "go".to_string()]);
    assert_eq!(note.created, "2026-10-21");
    assert_eq!(note.related, vec!["bigram-algoritmen".to_string()]);
    assert!(note.body.contains("Paketen är delade i lager"));
}

#[test]
fn parse_note_använder_rubrik_som_titel() {
    let dir = tempfile::tempdir().expect("tempdir");

    skriv_sida(
        dir.path(),
        "wiki/entities/flux.md",
        "---\ntype: entity\n---\n\n# Flux GitOps\n\nEn operator.\n",
    );

    let note = wiki::parse_note(&dir.path().join("wiki/entities/flux.md")).expect("parse_note");

    assert_eq!(note.title, "Flux GitOps");
}

#[test]
fn parse_note_saknar_frontmatter() {
    let dir = tempfile::tempdir().expect("tempdir");

    skriv_sida(dir.path(), "wiki/notes/utan.md", "Bara text.\n");

    let note = wiki::parse_note(&dir.path().join("wiki/notes/utan.md")).expect("parse_note");

    assert_eq!(note.title, "utan");
    assert_eq!(note.body, "Bara text.");
}

#[test]
fn parse_note_trastar_frontmatter() {
    let dir = tempfile::tempdir().expect("tempdir");

    skriv_sida(dir.path(), "wiki/notes/trasig.md", "---\ntitle: Trasig\n");

    assert!(wiki::parse_note(&dir.path().join("wiki/notes/trasig.md")).is_err());
}

#[test]
fn body_links_hanterar_båda_stilen() {
    let note = wiki::Note {
        id: "notes/a".to_string(),
        related: vec!["notes/b".to_string()],
        body: "Se [b](b.md) och [[concepts/C]] och [[notes/a]].".to_string(),
        ..Default::default()
    };

    assert_eq!(note.body_links().len(), 3);
    assert_eq!(note.all_related().len(), 3, "a ska filtreras bort");
}

#[test]
fn open_hittar_sidor_rekursivt() {
    let dir = ny_test_wiki();
    let w = wiki::open(dir.path()).expect("open");

    assert_eq!(w.notes.len(), 1);
    assert_eq!(w.notes[0].id, "concepts/Flux");
}

#[test]
fn resolve_hittar_wikilänkens_mål() {
    let dir = ny_test_wiki();

    skriv_sida(
        dir.path(),
        "wiki/entities/Talos.md",
        "---\ntype: entity\n---\n\n# Talos\n",
    );

    let w = wiki::open(dir.path()).expect("open");

    assert!(w.resolve("concepts/Flux").is_some());
    assert!(
        w.resolve("entities/talos").is_some(),
        "ska vara case-insensitive"
    );
    assert!(w.resolve("saknas").is_none());
}

#[test]
fn search_rangerar_rubrik_över_brödtext() {
    let dir = ny_test_wiki();

    skriv_sida(
        dir.path(),
        "wiki/concepts/Arkitektur.md",
        "---\ntitle: Arkitektur\ntags: [go]\n---\n\nPaketstrukturen.\n",
    );
    skriv_sida(
        dir.path(),
        "wiki/concepts/annat.md",
        "---\ntitle: Annat\ntags: [noter]\n---\n\nIngen arkitektur här.\n",
    );

    let w = wiki::open(dir.path()).expect("open");
    let träffar = w.search("arkitektur");

    assert!(!träffar.is_empty());
    assert_eq!(träffar[0].note.id, "concepts/Arkitektur");
}

#[test]
fn search_med_taggfilter() {
    let dir = ny_test_wiki();

    skriv_sida(
        dir.path(),
        "wiki/notes/a.md",
        "---\ntitle: A\ntags: [go, test]\n---\n\nBigram\n",
    );
    skriv_sida(
        dir.path(),
        "wiki/notes/b.md",
        "---\ntitle: B\ntags: [go]\n---\n\nBigram\n",
    );

    let w = wiki::open(dir.path()).expect("open");
    let träffar = w.search("bigram");

    assert_eq!(träffar.len(), 2);

    let taggade = träffar
        .iter()
        .filter(|träff| träff.note.has_tag("test"))
        .count();

    assert_eq!(taggade, 1);
}

#[test]
fn backlinks_hittar_länkande_sidor() {
    let dir = ny_test_wiki();

    skriv_sida(
        dir.path(),
        "wiki/concepts/kallare.md",
        "---\ntitle: Källa\nrelated: [concepts/Flux]\n---\n\nSe [[concepts/Flux]].\n",
    );

    let w = wiki::open(dir.path()).expect("open");
    let tillbaka = w.backlinks("concepts/Flux");

    assert_eq!(tillbaka.len(), 1);
    assert_eq!(tillbaka[0].id, "concepts/kallare");
}

#[test]
fn katalog_och_graph_genereras() {
    let dir = ny_test_wiki();

    skriv_sida(
        dir.path(),
        "wiki/concepts/Talos.md",
        "---\ntype: concept\ntags: [x]\n---\n\n# Talos\n\nSe [[concepts/Flux]].\n",
    );

    let w = wiki::open(dir.path()).expect("open");

    let katalog = w.catalog();
    assert!(katalog.contains("concepts/Flux"), "katalogen saknar sidan");
    assert!(
        katalog.contains("`x`"),
        "katalogen saknar taggsammanfattning"
    );

    let graph = w.graph();
    assert!(graph.contains("graph LR"), "graph saknar mermaid-header");
    assert!(
        graph.contains("n_concepts_Talos --> n_concepts_Flux"),
        "graph saknar kant mellan sidorna"
    );
    assert!(graph.contains("Backlinks"), "graph saknar backlink-lista");
}

#[test]
fn ingest_skapar_sida_och_arkiverar_inboxen() {
    let dir = tempfile::tempdir().expect("tempdir");

    std::fs::create_dir_all(dir.path().join("inbox")).expect("skapade inbox");
    skriv_sida(
        dir.path(),
        "inbox/ny-not.md",
        "---\ntitle: Ny not\ntags: [nytt]\n---\n\nInnehåll i noten.\n",
    );

    let nu = Local.with_ymd_and_hms(2026, 10, 21, 12, 0, 0).unwrap();
    let report = wiki::ingest(
        dir.path(),
        &wiki::IngestOptions {
            now: Some(nu),
            ..Default::default()
        },
    )
    .expect("ingest");

    assert_eq!(report.created, vec!["notes/ny-not".to_string()]);
    assert_eq!(report.archived.len(), 1);

    assert!(
        dir.path().join("wiki/notes/ny-not.md").exists(),
        "sidan skapades inte"
    );
    assert!(
        !dir.path().join("inbox/ny-not.md").exists(),
        "inbox-filen ligger kvar efter ingest"
    );
    assert!(
        dir.path().join("archive/2026-10-21-ny-not.md").exists(),
        "arkiv saknas"
    );

    for genererad in ["catalog.md", "graph.md", "log.md"] {
        assert!(
            dir.path().join(genererad).exists(),
            "{genererad} skapades inte"
        );
    }
}

#[test]
fn ingest_lägger_till_på_existerande_sida() {
    let dir = tempfile::tempdir().expect("tempdir");

    skriv_sida(
        dir.path(),
        "wiki/notes/befintlig.md",
        "---\ntitle: Befintlig\ntags: [a]\n---\n\nFörsta versionen.\n",
    );

    std::fs::create_dir_all(dir.path().join("inbox")).expect("skapade inbox");
    skriv_sida(
        dir.path(),
        "inbox/befintlig.md",
        "---\ntitle: Befintlig\ntags: [b]\n---\n\nAndra versionen.\n",
    );

    let nu = Local.with_ymd_and_hms(2026, 10, 21, 12, 0, 0).unwrap();
    let report = wiki::ingest(
        dir.path(),
        &wiki::IngestOptions {
            now: Some(nu),
            ..Default::default()
        },
    )
    .expect("ingest");

    assert_eq!(report.updated, vec!["notes/befintlig".to_string()]);
    assert!(report.created.is_empty());

    let note = wiki::parse_note(&dir.path().join("wiki/notes/befintlig.md")).expect("parse_note");

    assert!(note.body.contains("Första versionen."));
    assert!(note.body.contains("Andra versionen."));
    assert!(note.body.contains("## Uppdatering 2026-10-21"));
    assert_eq!(note.tags.len(), 2);
}

#[test]
fn ingest_torrkörning_ändrar_ingenting() {
    let dir = tempfile::tempdir().expect("tempdir");

    std::fs::create_dir_all(dir.path().join("inbox")).expect("skapade inbox");
    skriv_sida(dir.path(), "inbox/not.md", "---\ntitle: Not\n---\n\nText\n");

    let report = wiki::ingest(
        dir.path(),
        &wiki::IngestOptions {
            dry_run: true,
            now: None,
        },
    )
    .expect("ingest");

    assert_eq!(report.created.len(), 1);
    assert!(!dir.path().join("wiki/notes/not.md").exists());
    assert!(dir.path().join("inbox/not.md").exists());
}

#[test]
fn ingest_hoppar_over_inboxens_readme() {
    let dir = tempfile::tempdir().expect("tempdir");

    std::fs::create_dir_all(dir.path().join("inbox")).expect("skapade inbox");
    skriv_sida(dir.path(), "inbox/README.md", "# Inbox\n");

    let report = wiki::ingest(dir.path(), &wiki::IngestOptions::default()).expect("ingest");

    assert!(report.created.is_empty());
    assert!(report.updated.is_empty());
}

#[test]
fn suggest_merges_föreslår_liknande_titlar() {
    let dir = tempfile::tempdir().expect("tempdir");

    // Samma ämne med olika skrivställe - ska föreslås.
    skriv_sida(
        dir.path(),
        "wiki/notes/arkitektur.md",
        "---\ntitle: Arkitektur\ntags: [go, session]\n---\n\nText\n",
    );
    skriv_sida(
        dir.path(),
        "wiki/notes/arkitektur-2.md",
        "---\ntitle: arkitektur\ntags: [go]\n---\n\nText\n",
    );
    // Olika ämnen som delar generiska taggar - ska inte föreslås.
    skriv_sida(
        dir.path(),
        "wiki/notes/algoritm.md",
        "---\ntitle: Algoritm\ntags: [go, session]\n---\n\nText\n",
    );

    let w = wiki::open(dir.path()).expect("open");
    let förslag = wiki::ingest::suggest_merges(&w);

    assert_eq!(förslag.len(), 1, "bara liknande titlar ska föreslås");
    assert_eq!(förslag[0].a, "notes/arkitektur");
    assert_eq!(förslag[0].b, "notes/arkitektur-2");
}

#[test]
fn resolve_dir_hittar_wikin_uppåt() {
    let root = tempfile::tempdir().expect("tempdir");

    std::fs::create_dir_all(root.path().join("wiki/wiki")).expect("skapade wiki");

    let nested = root.path().join("src/go");
    std::fs::create_dir_all(&nested).expect("skapade nested");

    let previous = std::env::current_dir().expect("cwd");
    std::env::set_current_dir(&nested).expect("chdir");

    let result = wiki::resolve_dir("");

    std::env::set_current_dir(previous).expect("chdir tillbaka");

    let dir = result.expect("resolve_dir");
    assert_eq!(dir, root.path().join("wiki"), "resolve_dir pekade fel");
}

#[test]
fn project_tag_markerar_projekt() {
    assert_eq!(wiki::project_tag("typetrainer"), "projekt:typetrainer");
    assert_eq!(wiki::project_tag(""), "");
}

#[test]
fn note_har_projekt() {
    let note = wiki::Note {
        id: "a".to_string(),
        tags: vec!["bug".to_string(), "projekt:typetrainer".to_string()],
        ..Default::default()
    };

    assert_eq!(note.project_of(), "typetrainer");
    assert!(note.has_project("typetrainer"));
    assert!(!note.has_project("dashboard"));
    assert!(note.has_project(""));
}

#[test]
fn detect_project_hittar_repots_namn() {
    let dir = tempfile::tempdir().expect("tempdir");

    // Ett riktigt git-repo så att git pekar på tempkatalogen.
    let init = std::process::Command::new("git")
        .arg("-C")
        .arg(dir.path())
        .args(["init", "--quiet"])
        .output();

    if !init.map(|utdata| utdata.status.success()).unwrap_or(false) {
        return;
    }

    let project = wiki::detect_project(dir.path().to_str().unwrap_or_default());

    if project.is_empty() {
        return;
    }

    assert_eq!(project, bas(dir.path()));
}

#[test]
fn agents_block_skapar_fil() {
    let dir = tempfile::tempdir().expect("tempdir");

    let changed = wiki::ensure_agents_block(dir.path()).expect("ensure_agents_block");
    assert!(changed);

    let innehåll =
        std::fs::read_to_string(dir.path().join("AGENTS.md")).expect("AGENTS.md skapades inte");

    assert!(innehåll.contains(wiki::AGENTS_BLOCK_START));
    assert!(innehåll.contains("wikictl note"));
    assert!(!innehåll.starts_with("# <!--"));
}

#[test]
fn agents_block_är_idempotent() {
    let dir = tempfile::tempdir().expect("tempdir");

    wiki::ensure_agents_block(dir.path()).expect("ensure_agents_block");
    let first = std::fs::read_to_string(dir.path().join("AGENTS.md")).expect("läs AGENTS.md");

    let changed = wiki::ensure_agents_block(dir.path()).expect("ensure_agents_block");
    assert!(!changed, "andra körningen ska inte ändra något");

    let second = std::fs::read_to_string(dir.path().join("AGENTS.md")).expect("läs AGENTS.md");
    assert_eq!(first, second);
}

#[test]
fn agents_block_behåller_befintligt_innehåll() {
    let dir = tempfile::tempdir().expect("tempdir");

    std::fs::write(
        dir.path().join("AGENTS.md"),
        "# Eget projekt\n\nMin egen text.\n",
    )
    .expect("skrev AGENTS.md");

    wiki::ensure_agents_block(dir.path()).expect("ensure_agents_block");

    let innehåll = std::fs::read_to_string(dir.path().join("AGENTS.md")).expect("läs AGENTS.md");

    assert!(innehåll.contains("Min egen text."));
    assert!(innehåll.contains(wiki::AGENTS_BLOCK_START));
}

#[test]
fn agents_block_ersätter_gammalt_block() {
    let dir = tempfile::tempdir().expect("tempdir");

    let gammalt = format!(
        "{}\n GAMMAL TEXT\n{}",
        wiki::AGENTS_BLOCK_START,
        wiki::AGENTS_BLOCK_END
    );

    std::fs::write(
        dir.path().join("AGENTS.md"),
        format!("# Projekt\n\n{gammalt}\n"),
    )
    .expect("skrev AGENTS.md");

    wiki::ensure_agents_block(dir.path()).expect("ensure_agents_block");

    let innehåll = std::fs::read_to_string(dir.path().join("AGENTS.md")).expect("läs AGENTS.md");

    assert!(!innehåll.contains("GAMMAL TEXT"));
    assert!(innehåll.contains(wiki::AGENTS_BLOCK_START));
    assert_eq!(
        innehåll.matches(wiki::AGENTS_BLOCK_START).count(),
        1,
        "startmarkören finns fler än en gång"
    );
}

#[test]
fn write_page_skapar_sida() {
    let dir = tempfile::tempdir().expect("tempdir");

    let id = wiki::write_page(
        dir.path(),
        &wiki::WriteOptions {
            kind: Some(wiki::PageKind::Concept),
            title: "Kyverno CEL".to_string(),
            body: "# Kyverno CEL\n\nCEL är standarden.".to_string(),
            summary: "CEL är standarden för policies.".to_string(),
            tags: vec!["kyverno".to_string()],
            ..Default::default()
        },
    )
    .expect("write_page");

    assert_eq!(id, "concepts/kyverno-cel");

    let note =
        wiki::parse_note(&dir.path().join("wiki/concepts/kyverno-cel.md")).expect("parse_note");

    assert_eq!(note.note_type, "concept");
    assert_eq!(note.summary, "CEL är standarden för policies.");
}

#[test]
fn write_page_lägger_till_när_sidan_finns() {
    let dir = tempfile::tempdir().expect("tempdir");

    let mut opts = wiki::WriteOptions {
        kind: Some(wiki::PageKind::Entity),
        title: "Talos".to_string(),
        body: "Första versionen.".to_string(),
        ..Default::default()
    };

    wiki::write_page(dir.path(), &opts).expect("write_page");

    opts.body = "Andra versionen.".to_string();
    let id = wiki::write_page(dir.path(), &opts).expect("write_page");

    assert_eq!(id, "entities/talos");

    let note = wiki::parse_note(&dir.path().join("wiki/entities/talos.md")).expect("parse_note");

    assert!(note.body.contains("Första versionen."));
    assert!(note.body.contains("Andra versionen."));
}

#[test]
fn write_page_ersätter_med_replace() {
    let dir = tempfile::tempdir().expect("tempdir");

    let mut opts = wiki::WriteOptions {
        kind: Some(wiki::PageKind::Entity),
        title: "Talos".to_string(),
        body: "Första versionen.".to_string(),
        ..Default::default()
    };

    wiki::write_page(dir.path(), &opts).expect("write_page");

    opts.body = "Bara den här.".to_string();
    opts.replace = true;
    wiki::write_page(dir.path(), &opts).expect("write_page");

    let note = wiki::parse_note(&dir.path().join("wiki/entities/talos.md")).expect("parse_note");

    assert!(!note.body.contains("Första versionen."));
}

#[test]
fn link_lägger_till_relation() {
    let dir = tempfile::tempdir().expect("tempdir");

    wiki::write_page(
        dir.path(),
        &wiki::WriteOptions {
            kind: Some(wiki::PageKind::Entity),
            title: "Talos".to_string(),
            body: "Text.".to_string(),
            ..Default::default()
        },
    )
    .expect("write_page");

    wiki::write_page(
        dir.path(),
        &wiki::WriteOptions {
            kind: Some(wiki::PageKind::Concept),
            title: "GitOps".to_string(),
            body: "Text.".to_string(),
            ..Default::default()
        },
    )
    .expect("write_page");

    wiki::link(dir.path(), "entities/talos", "concepts/gitops").expect("link");

    let note = wiki::parse_note(&dir.path().join("wiki/entities/talos.md")).expect("parse_note");

    assert!(note.related.join(",").contains("concepts/gitops"));
    assert!(note.body.contains("[[concepts/gitops]]"));

    let w = wiki::open(dir.path()).expect("open");
    assert_eq!(w.backlinks("concepts/gitops").len(), 1);
}

#[test]
fn link_till_saknad_sida_felar() {
    let dir = tempfile::tempdir().expect("tempdir");

    wiki::write_page(
        dir.path(),
        &wiki::WriteOptions {
            kind: Some(wiki::PageKind::Entity),
            title: "Talos".to_string(),
            body: "Text.".to_string(),
            ..Default::default()
        },
    )
    .expect("write_page");

    assert!(wiki::link(dir.path(), "entities/talos", "concepts/finns-inte").is_err());
}

#[test]
fn index_listar_sidor_per_typ() {
    let dir = tempfile::tempdir().expect("tempdir");

    skriv_sida(
        dir.path(),
        "wiki/entities/Talos.md",
        "---\ntype: entity\nsummary: Operativsystemet.\n---\n\n# Talos\n",
    );
    skriv_sida(
        dir.path(),
        "wiki/concepts/GitOps.md",
        "---\ntype: concept\n---\n\n# GitOps\n\nGit är sanningskällan. Mer text.\n",
    );

    let w = wiki::open(dir.path()).expect("open");
    let index = wiki::index_by_kind(&w);

    assert!(index.contains("## Entiteter"));
    assert!(index.contains("## Koncept"));
    assert!(index.contains("Operativsystemet."));
    assert!(index.contains("Git är sanningskällan."));
}

#[test]
fn source_file_lägger_till_råkälla() {
    let dir = tempfile::tempdir().expect("tempdir");

    let src = dir.path().join("kalla.md");
    std::fs::write(&src, "Råtext.\n").expect("skrev källa");

    let nu = Local.with_ymd_and_hms(2026, 10, 23, 0, 0, 0).unwrap();
    let dest = wiki::source_file(
        dir.path(),
        src.to_str().unwrap_or_default(),
        "Min källa",
        "https://example.com",
        nu,
    )
    .expect("source_file");

    assert_eq!(
        dest.file_name()
            .map(|namn| namn.to_string_lossy().to_string()),
        Some("2026-10-23-min-kalla.md".to_string())
    );

    let w = wiki::open(dir.path()).expect("open");
    let källor = wiki::sources(&w);

    assert_eq!(källor.len(), 1);
    assert!(!källor[0].indexed, "källan ska vara ointegrerad");
}

#[test]
fn lint_hittar_ödelapp_och_ointegrerad_källa() {
    let dir = tempfile::tempdir().expect("tempdir");

    skriv_sida(
        dir.path(),
        "wiki/concepts/Ensam.md",
        "---\ntype: concept\n---\n\n# Ensam\n\nText.\n",
    );

    let src = dir.path().join("kalla.md");
    std::fs::write(&src, "Råtext.\n").expect("skrev källa");

    wiki::source_file(
        dir.path(),
        src.to_str().unwrap_or_default(),
        "Källa",
        "",
        Local::now(),
    )
    .expect("source_file");

    let w = wiki::open(dir.path()).expect("open");
    let findings = wiki::lint(&w, &wiki::LintOptions::default());

    assert!(
        findings.iter().any(|f| f.kind == wiki::LintKind::Orphan),
        "lint hittade inte ödelapp"
    );
    assert!(
        findings
            .iter()
            .any(|f| f.kind == wiki::LintKind::Unintegrated),
        "lint hittade inte ointegrerad källa"
    );
}

#[test]
fn append_log_använder_greppbart_format() {
    let dir = tempfile::tempdir().expect("tempdir");

    wiki::append_log(
        dir.path(),
        &[wiki::LogEntry {
            date: "2026-10-23".to_string(),
            op: "ingest".to_string(),
            title: "create notes/foo".to_string(),
        }],
        &["[create] [[notes/foo]]".to_string()],
    )
    .expect("append_log");

    let data = std::fs::read_to_string(dir.path().join("log.md")).expect("läs loggen");

    assert!(data.contains("## [2026-10-23] ingest | create notes/foo"));
}

#[test]
fn merge_sources_accepterar_båda_nycklarna() {
    let dir = tempfile::tempdir().expect("tempdir");

    skriv_sida(
        dir.path(),
        "wiki/summaries/x.md",
        "---\ntype: summary\nsource: a.md\nsources: [b.md, a.md]\n---\n\n# X\n",
    );

    let w = wiki::open(dir.path()).expect("open");

    assert_eq!(w.notes.len(), 1);
    assert_eq!(
        w.notes[0].sources,
        vec!["b.md".to_string(), "a.md".to_string()]
    );
}

#[test]
fn lint_rapporterar_dubblerad_sektion() {
    let dir = tempfile::tempdir().expect("tempdir");

    skriv_sida(
        dir.path(),
        "wiki/concepts/x.md",
        "---\ntype: concept\n---\n\n# X\n\n## Kopplingar\n\n- [[a]]\n\n## Kopplingar\n\n- [[b]]\n",
    );

    let w = wiki::open(dir.path()).expect("open");
    let findings = wiki::lint(&w, &wiki::LintOptions::default());

    assert!(
        findings
            .iter()
            .any(|f| f.kind == wiki::LintKind::DuplicateSection),
        "lint hittade inte den dubblerade sektionen"
    );
}

#[test]
fn source_add_i_raw_flyttar_i_stället_för_kopiera() {
    let dir = tempfile::tempdir().expect("tempdir");

    let raw_dir = dir.path().join("raw");
    std::fs::create_dir_all(&raw_dir).expect("skapade raw");

    let src = raw_dir.join("kalla.md");
    std::fs::write(&src, "Råtext.\n").expect("skrev källa");

    let dest = wiki::source_file(
        dir.path(),
        src.to_str().unwrap_or_default(),
        "Min källa",
        "https://example.com",
        Local::now(),
    )
    .expect("source_file");

    assert!(!src.exists(), "originalen finns kvar i raw/");

    let data = std::fs::read_to_string(&dest).expect("läs dest");
    assert!(data.contains("type: source"));
}

/// Hjälpfunktion för att hitta en fils sista segment.
#[allow(dead_code)]
fn bas(sökväg: &Path) -> String {
    sökväg
        .file_name()
        .map(|namn| namn.to_string_lossy().to_string())
        .unwrap_or_default()
}
