//! Kontrakstestet mellan Plan och agenten.
//!
//! Port av `src/agent/internal/planclient/contract_test.go`. Båda sidor
//! testar mot samma fil (`src/contracts/plan-agent.yaml`) - om en sida
//! ändras utan den andra faller testerna.

use plan_agent::contract::Contract;
use plan_agent::planclient::models::{APPROVAL_KINDS, GOAL_PHASES, RUN_STATUSES};
use plan_agent::planclient::paths::ALLA_SOKVAGAR;

/// Läser kontraktet.
fn kontrakt() -> Contract {
    Contract::load().expect("kontraktet ska gå att läsa")
}

/// Kontraktet ska ha en version och innehåll.
#[test]
fn kontraktet_finns() {
    let kontrakt = kontrakt();

    assert!(kontrakt.version > 0, "kontraktet ska ha en version");
    assert!(
        !kontrakt.phases.is_empty()
            && !kontrakt.statuses.is_empty()
            && !kontrakt.endpoints.is_empty(),
        "kontraktet ska innehålla faser, tillstånd och endpoints"
    );
}

/// Agentens tillstånd ska vara de som kontraktet anger - annars skulle
/// agenten få avslag på sin egen statusuppdatering.
#[test]
fn statusar_matchar_kontraktet() {
    let kontrakt = kontrakt();

    assert_eq!(
        RUN_STATUSES.len(),
        kontrakt.statuses.len(),
        "kontraktet har {} tillstånd, agenten har {}",
        kontrakt.statuses.len(),
        RUN_STATUSES.len()
    );

    for (index, status) in kontrakt.statuses.iter().enumerate() {
        assert_eq!(
            RUN_STATUSES[index], status,
            "tillstånd {index}: kontraktet säger {status:?}, agenten säger {:?}",
            RUN_STATUSES[index]
        );
    }
}

/// Agentens faser ska vara de som kontraktet anger.
#[test]
fn faser_matchar_kontraktet() {
    let kontrakt = kontrakt();

    assert_eq!(
        GOAL_PHASES.len(),
        kontrakt.phases.len(),
        "kontraktet har {} faser, agenten har {}",
        kontrakt.phases.len(),
        GOAL_PHASES.len()
    );

    for (index, fas) in kontrakt.phases.iter().enumerate() {
        assert_eq!(
            GOAL_PHASES[index], fas,
            "fas {index}: kontraktet säger {fas:?}, agenten säger {:?}",
            GOAL_PHASES[index]
        );
    }
}

/// Agentens godkännandetyper ska vara de som kontraktet anger.
///
/// Plan avvisar en okänd typ - så om agenten hittar på en typ som inte
/// står i kontraktet blir förslaget stoppat i API:t.
#[test]
fn godkannandetyper_matchar_kontraktet() {
    let kontrakt = kontrakt();

    assert_eq!(
        APPROVAL_KINDS.len(),
        kontrakt.approval_kinds.len(),
        "kontraktet har {} godkännandetyper, agenten har {}",
        kontrakt.approval_kinds.len(),
        APPROVAL_KINDS.len()
    );

    for (index, kind) in kontrakt.approval_kinds.iter().enumerate() {
        assert_eq!(
            APPROVAL_KINDS[index], kind,
            "godkännandetyp {index}: kontraktet säger {kind:?}, agenten säger {:?}",
            APPROVAL_KINDS[index]
        );
    }
}

/// Varje sökväg agenten använder ska finnas i kontraktet.
#[test]
fn sokvagar_matchar_kontraktet() {
    let kontrakt = kontrakt();

    for sokvag in ALLA_SOKVAGAR {
        assert!(
            kontrakt
                .endpoints
                .iter()
                .any(|endpoint| endpoint.path == *sokvag),
            "agenten använder {sokvag}, men den vägen finns inte i kontraktet"
        );
    }
}

/// Varje sökväg agenten använder ska vara markerad som agentens eget anrop.
///
/// Det markerade anropet är Plan:s agent-scope: agenten får bara göra de
/// anrop den är markerad för. Ett anrop agenten gör men som inte är
/// markerat ger 403 i prod, där agentens nyckel har scope 'agent'.
#[test]
fn agentens_anrop_ar_markerade() {
    let kontrakt = kontrakt();

    for sokvag in ALLA_SOKVAGAR {
        assert!(
            kontrakt.is_agent_endpoint(sokvag),
            "agenten använder {sokvag}, men den vägen är inte markerad som agentens eget anrop"
        );
    }
}

/// Varje anrop i kontraktet ska ha en giltig metod och en sökväg.
#[test]
fn kontraktets_anrop_ar_valliga() {
    let kontrakt = kontrakt();

    for endpoint in &kontrakt.endpoints {
        assert!(
            ["GET", "POST", "PUT", "PATCH", "DELETE"].contains(&endpoint.method.as_str()),
            "kontraktet har okänd metod {:?} för {}",
            endpoint.method,
            endpoint.path
        );
        assert!(
            endpoint.path.starts_with('/'),
            "kontraktets sökväg {} börjar inte med /",
            endpoint.path
        );
    }
}
