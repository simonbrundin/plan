//! Repository-lager.
//!
//! SQL:en är densamma som i Go-tjänsten (`src/api/internal/repository`),
//! så båda tjänsterna läser och skriver samma rader.

pub mod calendar;
pub mod calendar_event;
pub mod goal;
pub mod place;
pub mod tag;
