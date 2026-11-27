//! `wikictl` - hanterar wikin: sök, lista, skriv inlägg och
//! ingestera inboxen in i kunskapsgrafen.
//!
//! Porten av `cmd/wikictl/main.go`.

use std::collections::HashMap;
use std::path::PathBuf;

use chrono::Local;
use plan_wiki as wiki;

/// Skriver till stdout utan att panika när pipeen stängs.
///
/// Go ignorerar skrivfel mot en stängd pipe - det ska Rust också göra,
/// annars avbryter `wikictl list | head` med en panik.
macro_rules! ut {
    ($($arg:tt)*) => {{
        use std::io::Write;
        let mut utdata = std::io::stdout().lock();
        let _ = writeln!(utdata, $($arg)*);
    }};
}

/// Skriver till stdout utan radbrytning.
macro_rules! ut_rad {
    ($($arg:tt)*) => {{
        use std::io::Write;
        let mut utdata = std::io::stdout().lock();
        let _ = write!(utdata, $($arg)*);
    }};
}

fn main() {
    if let Err(err) = run(std::env::args().skip(1).collect()) {
        eprintln!("fel: {err}");
        std::process::exit(1);
    }
}

/// Flaggornas typer.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
enum Kind {
    Bool,
    Str,
    Int,
}

/// Beskrivning av en flagga.
struct Spec {
    namn: &'static str,
    typ: Kind,
}

impl Spec {
    const fn ny(namn: &'static str, typ: Kind) -> Self {
        Self { namn, typ }
    }
}

/// Tolkade flaggor och positionella argument.
struct Flags {
    värden: HashMap<String, String>,
    args: Vec<String>,
}

impl Flags {
    /// Värdet på en strängflagga.
    fn str(&self, namn: &str) -> String {
        self.värden.get(namn).cloned().unwrap_or_default()
    }

    /// Värdet på en heltalsflagga.
    fn int(&self, namn: &str) -> i64 {
        self.värden
            .get(namn)
            .and_then(|värde| värde.parse::<i64>().ok())
            .unwrap_or(0)
    }

    /// Värdet på en boolflagga.
    fn flag(&self, namn: &str) -> bool {
        matches!(self.värden.get(namn).map(String::as_str), Some("true"))
    }
}

/// Tolkar flaggor på Go:s sätt: `-flag`, `--flag`, `-flag=värde`.
fn tolka(args: &[String], specs: &[Spec]) -> Result<Flags, String> {
    let mut värden: HashMap<String, String> = HashMap::new();
    let mut positionella: Vec<String> = Vec::new();

    let mut index = 0;

    while index < args.len() {
        let arg = &args[index];

        if arg == "--" {
            index += 1;
            positionella.extend(args[index..].iter().cloned());
            break;
        }

        if !arg.starts_with('-') || arg == "-" {
            break;
        }

        let utan = arg.trim_start_matches('-');
        let (namn, inline) = match utan.split_once('=') {
            Some((namn, värde)) => (namn, Some(värde.to_string())),
            None => (utan, None),
        };

        let Some(spec) = specs.iter().find(|spec| spec.namn == namn) else {
            return Err(format!("flag provided but not defined: -{namn}"));
        };

        let värde = match spec.typ {
            Kind::Bool => inline.unwrap_or_else(|| "true".to_string()),
            Kind::Str | Kind::Int => match inline {
                Some(värde) => värde,
                None => {
                    index += 1;

                    if index >= args.len() {
                        return Err(format!("flag needs an argument: -{namn}"));
                    }

                    args[index].clone()
                }
            },
        };

        värden.insert(namn.to_string(), värde);
        index += 1;
    }

    positionella.extend(args[index..].iter().cloned());

    Ok(Flags {
        värden,
        args: positionella,
    })
}

/// Plockar ut `--wiki` ur argumentlistan.
fn split_wiki_flag(args: &[String]) -> (String, Vec<String>) {
    let mut wiki_dir = String::new();
    let mut rest = Vec::with_capacity(args.len());

    let mut index = 0;

    while index < args.len() {
        let arg = &args[index];

        if arg == "--wiki" && index + 1 < args.len() {
            wiki_dir = args[index + 1].clone();
            index += 2;
            continue;
        }

        if let Some(värde) = arg.strip_prefix("--wiki=") {
            wiki_dir = värde.to_string();
            index += 1;
            continue;
        }

        rest.push(arg.clone());
        index += 1;
    }

    (wiki_dir, rest)
}

/// Dispatchar till rätt delkommando.
fn run(args: Vec<String>) -> Result<(), String> {
    let (wiki_dir, args) = split_wiki_flag(&args);

    if args.is_empty() {
        usage();
        return Ok(());
    }

    let cmd = args[0].clone();
    let rest = &args[1..];

    match cmd.as_str() {
        "search" => run_search(rest, &wiki_dir),
        "list" => run_list(rest, &wiki_dir),
        "show" => run_show(rest, &wiki_dir),
        "graph" => run_graph(rest, &wiki_dir),
        "inbox" => run_inbox(&wiki_dir),
        "note" => run_note(rest, &wiki_dir),
        "page" => run_page(rest, &wiki_dir),
        "link" => run_link(rest, &wiki_dir),
        "source" => run_source(rest, &wiki_dir),
        "ingest" => run_ingest(rest, &wiki_dir),
        "index" => run_index(&wiki_dir),
        "lint" => run_lint(rest, &wiki_dir),
        "init" => run_init(rest),
        "help" | "-h" | "--help" => {
            usage();
            Ok(())
        }
        annan => Err(format!(
            "okänt kommando: {annan:?} (kör 'wikictl help')"
        )),
    }
}

/// Visar hjälpen.
fn usage() {
    ut_rad!(
        r#"wikictl - hantera kunskaps-wikin

Användning:
  wikictl <kommando> [flaggor]

Kommandon:
  search <fråga>      Sök i wikin (--tag, --project, --limit)
  list                Lista alla sidor (--project)
  show <id>           Visa en sida
  graph               Visa kunskapsgrafen (--mermaid)
  inbox               Visa inlägg som väntar på att ingesteras
  note                Skriv en not till inboxen (--title, --tags, --body, --project)
  page                Skriv en sida (--type, --title, --body, --file, --tags, ...)
  link                Länka en sida till en annan (--from, --to, --remove)
  source              Hantera råkällor (add, list)
  ingest              Ingestera inboxen (--dry-run)
  index               Bygg om index.md, catalog.md och graph.md
  lint                Kolla wikin (trasiga länkar, ödelappar, dubbler, gammal info)
  init [repo]         Lägg in wikinreglerna i ett repos AGENTS.md

Globala flaggor:
  --wiki <sökväg>     Wiki-katalogen (default: hittas uppifrån, eller $WIKI_DIR)
"#
    );
}

/// Öppnar wikin.
fn open_wiki(wiki_dir: &str) -> Result<wiki::Wiki, String> {
    let dir = wiki::resolve_dir(wiki_dir)?;

    wiki::open(&dir)
}

/// Delar upp en kommaseparerad tagglista.
fn split_tags(raw: &str) -> Vec<String> {
    if raw.is_empty() {
        return Vec::new();
    }

    raw.split(',')
        .map(str::trim)
        .filter(|del| !del.is_empty())
        .map(str::to_string)
        .collect()
}

/// Svarar om listan redan innehåller projekttaggen.
fn has_project_tag(tags: &[String], project: &str) -> bool {
    let want = wiki::project_tag(project);

    tags.iter().any(|tagg| tagg.eq_ignore_ascii_case(&want))
}

/// Söker i wikin.
fn run_search(args: &[String], wiki_dir: &str) -> Result<(), String> {
    let flags = tolka(
        args,
        &[
            Spec::ny("tag", Kind::Str),
            Spec::ny("project", Kind::Str),
            Spec::ny("limit", Kind::Int),
        ],
    )?;

    if flags.args.is_empty() {
        return Err("sökning kräver en fråga".to_string());
    }

    let fråga = flags.args.join(" ");
    let w = open_wiki(wiki_dir)?;

    let mut limit = flags.int("limit");
    if limit == 0 {
        limit = 10;
    }

    let mut visade = 0;

    for träff in w.search(&fråga) {
        if !flags.str("tag").is_empty() && !träff.note.has_tag(&flags.str("tag")) {
            continue;
        }

        if !flags.str("project").is_empty() && !träff.note.has_project(&flags.str("project")) {
            continue;
        }

        if limit > 0 && visade >= limit {
            break;
        }

        visade += 1;

        ut!("{}  (poäng {})", träff.note.id, träff.score);

        if !träff.note.tags.is_empty() {
            ut!("  taggar: {}", träff.note.tags.join(", "));
        }

        for rad in &träff.lines {
            ut!("  {rad}");
        }

        ut!();
    }

    if visade == 0 {
        ut!("Inga träffar. Skriv en not i inboxen om detta borde finnas i grafen.");
    }

    Ok(())
}

/// Listar alla sidor.
fn run_list(args: &[String], wiki_dir: &str) -> Result<(), String> {
    let flags = tolka(args, &[Spec::ny("project", Kind::Str)])?;

    let w = open_wiki(wiki_dir)?;

    if w.notes.is_empty() {
        ut!("Wikin är tom.");
        return Ok(());
    }

    for anteckning in &w.notes {
        if !flags.str("project").is_empty() && !anteckning.has_project(&flags.str("project")) {
            continue;
        }

        ut!(
            "{:<40} {:<35} {}",
            anteckning.id,
            anteckning.title,
            anteckning.tags.join(", ")
        );
    }

    Ok(())
}

/// Visar en sida.
fn run_show(args: &[String], wiki_dir: &str) -> Result<(), String> {
    let flags = tolka(args, &[])?;

    if flags.args.is_empty() {
        return Err("show kräver en id".to_string());
    }

    let w = open_wiki(wiki_dir)?;

    let Some(anteckning) = w.note(&flags.args[0]) else {
        return Err(format!("sidan {:?} finns inte", flags.args[0]));
    };

    ut_rad!("{}", anteckning.render());

    Ok(())
}

/// Visar länkgraphen.
fn run_graph(args: &[String], wiki_dir: &str) -> Result<(), String> {
    let flags = tolka(args, &[Spec::ny("mermaid", Kind::Bool)])?;

    let w = open_wiki(wiki_dir)?;
    let graph = w.graph();

    if flags.flag("mermaid") {
        ut_rad!("{}", mermaid_block(&graph));
        return Ok(());
    }

    ut_rad!("{graph}");

    Ok(())
}

/// Visar inlägg som väntar på att ingesteras.
fn run_inbox(wiki_dir: &str) -> Result<(), String> {
    let dir = wiki::resolve_dir(wiki_dir)?;
    let inbox = dir.join(wiki::wiki::INBOX_DIR);

    let poster = match std::fs::read_dir(&inbox) {
        Ok(poster) => poster,
        Err(err) if err.kind() == std::io::ErrorKind::NotFound => {
            ut!("Inboxen är tom.");
            return Ok(());
        }
        Err(err) => return Err(format!("kunde inte läsa {}: {err}", inbox.display())),
    };

    let mut namn: Vec<String> = poster
        .filter_map(|poster| poster.ok())
        .map(|poster| poster.path())
        .filter(|sökväg| sökväg.is_file())
        .filter_map(|sökväg| {
            sökväg
                .file_name()
                .map(|namn| namn.to_string_lossy().to_string())
        })
        .filter(|namn| namn.ends_with(".md"))
        .filter(|namn| namn != "README.md" && !namn.starts_with('_'))
        .collect();

    namn.sort();

    if namn.is_empty() {
        ut!("Inboxen är tom.");
        return Ok(());
    }

    for namn in namn {
        ut!("{namn}");
    }

    Ok(())
}

/// Skriver en not till inboxen.
fn run_note(args: &[String], wiki_dir: &str) -> Result<(), String> {
    let flags = tolka(
        args,
        &[
            Spec::ny("title", Kind::Str),
            Spec::ny("tags", Kind::Str),
            Spec::ny("body", Kind::Str),
            Spec::ny("project", Kind::Str),
        ],
    )?;

    let title = flags.str("title");

    if title.is_empty() {
        return Err("--title krävs".to_string());
    }

    let dir = wiki::resolve_dir(wiki_dir)?;
    let inbox = dir.join(wiki::wiki::INBOX_DIR);

    std::fs::create_dir_all(&inbox)
        .map_err(|err| format!("kunde inte skapa {}: {err}", inbox.display()))?;

    let mut anteckning = wiki::Note {
        id: wiki::slug(&title),
        title: title.clone(),
        note_type: "note".to_string(),
        tags: split_tags(&flags.str("tags")),
        body: flags.str("body").trim().to_string(),
        ..Default::default()
    };

    // Tagga med projektet om det inte redan finns en sådan tagg.
    let mut project = flags.str("project");

    if project.is_empty() {
        project = wiki::detect_project("");
    }

    if !project.is_empty() && anteckning.project_of().is_empty() {
        anteckning.tags.push(wiki::project_tag(&project));
    }

    let path = inbox.join(format!("{}.md", anteckning.id));

    if path.exists() {
        return Err(format!(
            "det finns redan en not med samma rubrik i inboxen: {}",
            path.display()
        ));
    }

    std::fs::write(&path, anteckning.render())
        .map_err(|err| format!("kunde inte skriva {}: {err}", path.display()))?;

    ut!("Skrev {}", path.display());
    ut!("Kör 'wikictl ingest' för att lägga noten i grafen.");

    Ok(())
}

/// Skriver en sida i wiki-lagret.
fn run_page(args: &[String], wiki_dir: &str) -> Result<(), String> {
    let flags = tolka(
        args,
        &[
            Spec::ny("type", Kind::Str),
            Spec::ny("title", Kind::Str),
            Spec::ny("body", Kind::Str),
            Spec::ny("file", Kind::Str),
            Spec::ny("summary", Kind::Str),
            Spec::ny("tags", Kind::Str),
            Spec::ny("related", Kind::Str),
            Spec::ny("sources", Kind::Str),
            Spec::ny("replace", Kind::Bool),
            Spec::ny("project", Kind::Str),
        ],
    )?;

    let title = flags.str("title");

    if title.is_empty() {
        return Err("--title krävs".to_string());
    }

    let kind = wiki::parse_kind(&flags.str("type"))?;

    let mut text = flags.str("body");
    let fil = flags.str("file");

    if !fil.is_empty() {
        text = std::fs::read_to_string(&fil)
            .map_err(|err| format!("kunde inte läsa {fil}: {err}"))?;
    }

    if text.trim().is_empty() {
        return Err("--body eller --file krävs".to_string());
    }

    let dir = wiki::resolve_dir(wiki_dir)?;

    let mut page_tags = split_tags(&flags.str("tags"));
    let mut project = flags.str("project");

    if project.is_empty() {
        project = wiki::detect_project("");
    }

    if !project.is_empty() && !has_project_tag(&page_tags, &project) {
        page_tags.push(wiki::project_tag(&project));
    }

    let id = wiki::write_page(
        &dir,
        &wiki::WriteOptions {
            kind: Some(kind),
            title: title.clone(),
            body: text,
            summary: flags.str("summary"),
            tags: page_tags,
            related: split_tags(&flags.str("related")),
            sources: split_tags(&flags.str("sources")),
            status: String::new(),
            replace: flags.flag("replace"),
            now: None,
        },
    )?;

    refresh_index(&dir)?;

    wiki::append_log(
        &dir,
        &[wiki::LogEntry {
            date: Local::now().format("%Y-%m-%d").to_string(),
            op: "page".to_string(),
            title: id.clone(),
        }],
        &[],
    )?;

    ut!("Skrev [[{id}]]");

    Ok(())
}

/// Lägger till eller tar bort en länk mellan två sidor.
fn run_link(args: &[String], wiki_dir: &str) -> Result<(), String> {
    let flags = tolka(
        args,
        &[
            Spec::ny("from", Kind::Str),
            Spec::ny("to", Kind::Str),
            Spec::ny("remove", Kind::Bool),
        ],
    )?;

    let from = flags.str("from");
    let to = flags.str("to");

    if from.is_empty() || to.is_empty() {
        return Err("--from och --to krävs".to_string());
    }

    let dir = wiki::resolve_dir(wiki_dir)?;

    if flags.flag("remove") {
        wiki::unlink(&dir, &from, &to)?;
    } else {
        wiki::link(&dir, &from, &to)?;
    }

    refresh_index(&dir)?;

    let (op, verb) = if flags.flag("remove") {
        ("unlink", "Tog bort länk")
    } else {
        ("link", "Länkade")
    };

    wiki::append_log(
        &dir,
        &[wiki::LogEntry {
            date: Local::now().format("%Y-%m-%d").to_string(),
            op: op.to_string(),
            title: format!("{from} -> {to}"),
        }],
        &[],
    )?;

    ut!("{verb} {from} -> {to}");

    Ok(())
}

/// Hanterar råkällager.
fn run_source(args: &[String], wiki_dir: &str) -> Result<(), String> {
    if args.is_empty() {
        return Err("source kräver add eller list".to_string());
    }

    let sub = args[0].clone();
    let rest = &args[1..];

    let flags = tolka(
        rest,
        &[
            Spec::ny("file", Kind::Str),
            Spec::ny("title", Kind::Str),
            Spec::ny("origin", Kind::Str),
        ],
    )?;

    let dir = wiki::resolve_dir(wiki_dir)?;

    match sub.as_str() {
        "add" => {
            let mut fil = flags.str("file");

            if fil.is_empty() && !flags.args.is_empty() {
                fil = flags.args[0].clone();
            }

            if fil.is_empty() {
                return Err("--file krävs".to_string());
            }

            let dest = wiki::source_file(
                &dir,
                &fil,
                &flags.str("title"),
                &flags.str("origin"),
                Local::now(),
            )?;

            wiki::append_log(
                &dir,
                &[wiki::LogEntry {
                    date: Local::now().format("%Y-%m-%d").to_string(),
                    op: "source".to_string(),
                    title: dest
                        .file_name()
                        .map(|namn| namn.to_string_lossy().to_string())
                        .unwrap_or_default(),
                }],
                &[],
            )?;

            ut!("Lade in råkälla {}", dest.display());
            ut!("Skriv nu en sammanfattning i wikin och länka den till källan.");
        }
        "list" => {
            let w = wiki::open(&dir)?;
            let källor = wiki::sources(&w);

            if källor.is_empty() {
                ut!("Inga råkällor.");
                return Ok(());
            }

            for källa in källor {
                let state = if källa.indexed {
                    "integrerad"
                } else {
                    "ointegrerad"
                };

                let origin = if källa.origin.is_empty() {
                    "-".to_string()
                } else {
                    källa.origin.clone()
                };

                ut!(
                    "{:<40} {:<12} {:<14} {}",
                    källa.id, state, källa.added, origin
                );
            }
        }
        annan => {
            return Err(format!(
                "okänd source-kommando: {annan:?} (giltiga: add, list)"
            ))
        }
    }

    Ok(())
}

/// Ingestera inboxen in i grafen.
fn run_ingest(args: &[String], wiki_dir: &str) -> Result<(), String> {
    let flags = tolka(args, &[Spec::ny("dry-run", Kind::Bool)])?;

    let dir = wiki::resolve_dir(wiki_dir)?;

    let report = wiki::ingest(
        &dir,
        &wiki::IngestOptions {
            dry_run: flags.flag("dry-run"),
            now: None,
        },
    )?;

    if flags.flag("dry-run") {
        ut!("Torrkörning - inget skrevs.");
    }

    ut!("Skapade:     {}", join_or_none(&report.created));
    ut!("Uppdaterade: {}", join_or_none(&report.updated));
    ut!("Arkiverade:  {}", join_or_none(&report.archived));

    if !report.suggest.is_empty() {
        ut!("\nFörslag att slå ihop (kräver agent):");

        for förslag in &report.suggest {
            ut!("  {förslag}");
        }
    }

    Ok(())
}

/// Bygger om index.md, catalog.md och graph.md.
fn run_index(wiki_dir: &str) -> Result<(), String> {
    let dir = wiki::resolve_dir(wiki_dir)?;

    refresh_index(&dir)?;

    ut!("index.md, catalog.md och graph.md är uppdaterade.");

    Ok(())
}

/// Lägger in wikinreglerna i ett repos AGENTS.md.
fn run_init(args: &[String]) -> Result<(), String> {
    let flags = tolka(args, &[])?;

    let repo_dir = if flags.args.is_empty() {
        std::env::current_dir()
            .map_err(|err| format!("kunde inte läsa arbetskatalogen: {err}"))?
    } else {
        PathBuf::from(&flags.args[0])
    };

    let changed = wiki::ensure_agents_block(&repo_dir)?;

    if changed {
        ut!("Uppdaterade {}", repo_dir.join(wiki::init::AGENTS_FILE_NAME).display());
    } else {
        ut!(
            "{} var redan uppdaterad",
            repo_dir.join(wiki::init::AGENTS_FILE_NAME).display()
        );
    }

    Ok(())
}

/// Kolla wikin.
fn run_lint(args: &[String], wiki_dir: &str) -> Result<(), String> {
    let flags = tolka(args, &[Spec::ny("stale-days", Kind::Int)])?;

    let w = open_wiki(wiki_dir)?;

    let stale_days = flags.int("stale-days");

    let opts = wiki::LintOptions {
        stale_after: if stale_days > 0 {
            Some(std::time::Duration::from_secs(
                (stale_days as u64) * 24 * 60 * 60,
            ))
        } else {
            None
        },
        now: None,
    };

    let findings = wiki::lint(&w, &opts);

    for finding in &findings {
        ut!("{finding}");
    }

    if findings.is_empty() {
        ut!("Inga problem hittade.");
    } else {
        ut!("\n{} problem hittade.", findings.len());
    }

    Ok(())
}

/// Bygger om de genererade filerna.
fn refresh_index(dir: &std::path::Path) -> Result<(), String> {
    let w = wiki::open(dir)?;

    w.write_index()?;
    w.write_catalog()?;
    w.write_graph()
}

/// Extraherar mermaid-diagrammet ur den genererade graph.md.
fn mermaid_block(graph: &str) -> String {
    const FENCE: &str = "```";

    let Some(start) = graph.find(&format!("{FENCE}mermaid")) else {
        return String::new();
    };

    let body = start + FENCE.len() + "mermaid".len();
    let Some(end) = graph[body..].find(FENCE) else {
        return String::new();
    };

    format!("{}\n", graph[body..body + end].trim())
}

/// Joinar en lista, eller "-" om den är tom.
fn join_or_none(items: &[String]) -> String {
    if items.is_empty() {
        "-".to_string()
    } else {
        items.join(", ")
    }
}
