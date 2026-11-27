//! Kunskaps-wikin - ett filbaserat kunskapsgraf för agenterna.
//!
//! Porten av Go-paketet `internal/wiki` i agenten. Wikin är tre lager:
//! råkällor (`raw/`), syntes (`wiki/`) och schema (`AGENTS.md`).
//!
//! ```text
//! wiki/
//! ├── raw/          immutabla källor
//! ├── inbox/        inlägg som väntar på att ingesteras
//! ├── wiki/         den genererade grafen
//! │   ├── entities/
//! │   ├── concepts/
//! │   ├── summaries/
//! │   ├── synthesis/
//! │   └── notes/
//! ├── archive/      ingesterade inlägg
//! ├── catalog.md    genererad katalog
//! ├── graph.md      genererad länkgraph
//! └── log.md        append-only logg
//! ```

pub mod index;
pub mod ingest;
pub mod init;
pub mod lint;
pub mod log;
pub mod note;
pub mod page;
pub mod project;
pub mod sources;
pub mod wiki;

pub use index::{derive_summary, index_by_kind, kind_av_id, kind_of, summary_line, INDEX_FILE};
pub use ingest::{append_update, ingest, read_inbox, IngestOptions, IngestReport};
pub use init::{ensure_agents_block, AGENTS_BLOCK, AGENTS_BLOCK_END, AGENTS_BLOCK_START};
pub use lint::{lint, LintFinding, LintKind, LintOptions};
pub use log::{append_log, LogEntry};
pub use note::{merge_tags, parse_note, slug, Note};
pub use page::{
    find_page_by_id, link, parse_kind, unlink, write_page, PageKind, WriteOptions, KINDS,
};
pub use project::{detect_project, project_tag, PROJECT_TAG_PREFIX};
pub use sources::{source_file, sources, Source};
pub use wiki::{open, resolve_dir, write_file, Wiki};
