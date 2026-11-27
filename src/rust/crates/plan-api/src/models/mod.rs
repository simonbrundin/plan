//! Modeller.
//!
//! Port av `src/api/internal/models`. Fältnamnen och JSON-formen är
//! identiska med Go-tjänstens, så gränssnittet och agenten ser samma svar.

pub mod agent;
pub mod calendar;
pub mod calendar_event;
pub mod go_num;
pub mod goal;
pub mod place;
pub mod status;
pub mod tag;
