//! Kontrakstestet mellan Plan och agenten, från Plan:s sida.
//!
//! Varje anrop i kontraktet (`src/contracts/plan-agent.yaml`) ska finnas i
//! Plan API:ts route-tabell - annars har agenten något att anropa som inte
//! finns. Det här är Plan:s motsvarighet till agentens kontrakttest.

use std::collections::BTreeSet;

use plan_api::routes::table::ROUTES;

/// Kontraktet mellan Plan och agenten.
#[derive(Debug, serde::Deserialize)]
struct Kontrakt {
    #[serde(default)]
    endpoints: Vec<Endpoint>,
}

#[derive(Debug, serde::Deserialize)]
struct Endpoint {
    method: String,
    path: String,
}

/// Läser kontraktet från `src/contracts/plan-agent.yaml`.
fn kontrakt() -> Kontrakt {
    let sökväg = std::path::PathBuf::from(env!("CARGO_MANIFEST_DIR"))
        .join("../../../contracts/plan-agent.yaml");

    let raw = std::fs::read_to_string(&sökväg).expect("kontraktet ska gå att läsa");
    serde_yaml::from_str(&raw).expect("kontraktet ska gå att tolka")
}

/// Plan API:ts routes.
fn plan_routes() -> BTreeSet<(String, String)> {
    ROUTES
        .iter()
        .map(|(metod, sokvag)| (metod.to_string(), sokvag.to_string()))
        .collect()
}

/// Varje anrop i kontraktet ska finnas i Plan API:t.
#[test]
fn kontraktets_anrop_finns_i_plan() {
    let routes = plan_routes();

    for endpoint in kontrakt().endpoints {
        assert!(
            routes.contains(&(endpoint.method.clone(), endpoint.path.clone())),
            "kontraktet använder {} {}, men den vägen finns inte i Plan",
            endpoint.method,
            endpoint.path
        );
    }
}
