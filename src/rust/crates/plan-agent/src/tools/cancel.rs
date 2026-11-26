//! Avbrytning av pågående verktygskörningar.
//!
//! Go-agentens `internal/core/tools/cancel.go`. En global manager håller
//! en avbrytningskanal per verktygsanrop - verktyget lyssnar på kanalen
//! och avbryter sig själv.
//!
//! Semantiken följer Go: en avbruten kanal ligger kvar i kartan tills den
//! avbryts en andra gång. Det är därför `Count` räknar avbrutna anrop
//! så länge de inte är borttagna.

use std::collections::HashMap;
use std::sync::{Mutex, OnceLock};

/// En avbrytningskanal för ett verktygsanrop.
///
/// Mottagaren väntar på att värdet blir `true` - då ska verktyget
/// avbrytas.
pub type CancelChannel = tokio::sync::watch::Receiver<bool>;

/// Ett registrerat verktygsanrop.
#[derive(Clone)]
struct ToolCancel {
    avbrytning: tokio::sync::watch::Sender<bool>,
}

impl ToolCancel {
    /// Om anropet är avbrutet.
    fn avbruten(&self) -> bool {
        *self.avbrytning.borrow()
    }

    /// Signalerar avbrytning.
    ///
    /// `send_replace` används i stället för `send` - `send` ändrar inte
    /// värdet om ingen mottagare lyssnar, och då skulle avbrytningen
    /// försvinna.
    fn avbryt(&self) {
        self.avbrytning.send_replace(true);
    }

    /// En mottagare som väntar på avbrytning.
    fn mottagare(&self) -> CancelChannel {
        self.avbrytning.subscribe()
    }
}

/// Manager som håller avbrytningskanaler per verktygsanrop.
#[derive(Default)]
pub struct CancellationManager {
    verktyg: Mutex<HashMap<String, ToolCancel>>,
}

impl CancellationManager {
    /// Skapar en tom manager.
    pub fn new() -> Self {
        Self::default()
    }

    /// Registrerar ett verktygsanrop och returnerar dess kanal.
    ///
    /// Registreras anropet redan returneras samma kanal igen.
    pub fn register(&self, tool_call_id: &str) -> CancelChannel {
        let mut verktyg = self.lock();

        if let Some(anrop) = verktyg.get(tool_call_id) {
            return anrop.mottagare();
        }

        let (avbrytning, _) = tokio::sync::watch::channel(false);
        let anrop = ToolCancel { avbrytning };

        let mottagare = anrop.mottagare();
        verktyg.insert(tool_call_id.to_string(), anrop);

        mottagare
    }

    /// Returnerar kanalen för ett verktygsanrop, om den finns.
    pub fn get(&self, tool_call_id: &str) -> Option<CancelChannel> {
        self.lock().get(tool_call_id).map(ToolCancel::mottagare)
    }

    /// Signalerar avbrytning för ett verktygsanrop.
    ///
    /// Ett redan avbrutet anrop tas bort - precis som i Go, där den
    /// stängda kanalen raderas när den avbryts igen.
    pub fn cancel(&self, tool_call_id: &str) {
        let mut verktyg = self.lock();

        let Some(anrop) = verktyg.get(tool_call_id) else {
            return;
        };

        if anrop.avbruten() {
            verktyg.remove(tool_call_id);
            return;
        }

        anrop.avbryt();
    }

    /// Tar bort ett verktygsanrop.
    pub fn unregister(&self, tool_call_id: &str) {
        let mut verktyg = self.lock();

        if let Some(anrop) = verktyg.remove(tool_call_id) {
            anrop.avbryt();
        }
    }

    /// Avbryter alla pågående verktygsanrop.
    ///
    /// Redan avbrutna anrop ligger kvar - de avbryts inte igen.
    pub fn cancel_all(&self) {
        let mut verktyg = self.lock();

        let öppna: Vec<String> = verktyg
            .iter()
            .filter(|(_, anrop)| !anrop.avbruten())
            .map(|(id, _)| id.clone())
            .collect();

        for id in öppna {
            if let Some(anrop) = verktyg.remove(&id) {
                anrop.avbryt();
            }
        }
    }

    /// Om ett verktygsanrop är avbrutet.
    pub fn is_cancelled(&self, tool_call_id: &str) -> bool {
        self.lock()
            .get(tool_call_id)
            .map(ToolCancel::avbruten)
            .unwrap_or(false)
    }

    /// Antal pågående verktygsanrop.
    pub fn count(&self) -> usize {
        self.lock().len()
    }

    /// Låser kartan.
    ///
    /// Ett redan upptaget låst läses ändå - annars skulle agenten
    /// krascha av en trådlös tråd.
    fn lock(&self) -> std::sync::MutexGuard<'_, HashMap<String, ToolCancel>> {
        self.verktyg
            .lock()
            .unwrap_or_else(|poisoned| poisoned.into_inner())
    }
}

/// Den globala managern.
static MANAGER: OnceLock<CancellationManager> = OnceLock::new();

/// Den globala managern.
pub fn global() -> &'static CancellationManager {
    MANAGER.get_or_init(CancellationManager::new)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn registering_två_gånger_ger_samma_anrop() {
        let manager = CancellationManager::new();

        manager.register("anrop-1");
        manager.register("anrop-1");

        assert_eq!(manager.count(), 1);
    }

    #[test]
    fn avbrytning_markerar_anropet_utan_att_taga_bort_det() {
        let manager = CancellationManager::new();
        manager.register("anrop-1");

        assert!(!manager.is_cancelled("anrop-1"));

        manager.cancel("anrop-1");

        assert!(manager.is_cancelled("anrop-1"));
        // Go raderar kanalen först vid den andra avbrytningen.
        assert_eq!(manager.count(), 1);

        manager.cancel("anrop-1");

        assert_eq!(manager.count(), 0);
    }

    #[test]
    fn avbryta_alla_stänger_öppna_anrop() {
        let manager = CancellationManager::new();
        manager.register("anrop-1");
        manager.register("anrop-2");

        manager.cancel_all();

        // Go tar bort de öppna kanalerna ur kartan när den avbryter dem
        // alla - därför är de inte "avbrutna" efteråt, de är borta.
        assert_eq!(manager.count(), 0);
        assert!(!manager.is_cancelled("anrop-1"));
        assert!(!manager.is_cancelled("anrop-2"));
    }

    #[tokio::test]
    async fn mottagaren_ser_avbrytningen() {
        let manager = CancellationManager::new();
        let mottagare = manager.register("anrop-1");

        manager.cancel("anrop-1");

        assert!(*mottagare.borrow());
    }
}
