//! Orchestratorns tillstånd.
//!
//! Orchestratorn själv portas i fas 4d - här ligger bara tillståndet, så
//! att health och status kan visa det som Go gör. Initieringsfelet sparas
//! här så att det syns i båda svaren - ett fel som bara finns i loggen är
//! lätt att missa.

use std::sync::RwLock;

/// Initieringsfelet, om orchestratorn inte startade.
static INIT_ERROR: RwLock<Option<String>> = RwLock::new(None);

/// Om orchestratorn är redo.
static REDO: RwLock<bool> = RwLock::new(false);

/// Sparar initieringsfelet.
pub fn set_init_error(err: impl Into<String>) {
    if let Ok(mut fel) = INIT_ERROR.write() {
        *fel = Some(err.into());
    }
}

/// Initieringsfelet, om det finns.
pub fn init_error() -> Option<String> {
    INIT_ERROR.read().ok().and_then(|fel| fel.clone())
}

/// Markera orchestratorn som redo.
pub fn set_ready() {
    if let Ok(mut redo) = REDO.write() {
        *redo = true;
    }
}

/// Om orchestratorn är redo.
pub fn is_ready() -> bool {
    REDO.read().map(|redo| *redo).unwrap_or(false)
}
