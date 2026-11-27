//! Jobbmanifestet ska vara likvärdigt med Go:s.
//!
//! Go bygger manifestet med `jobrunner.Build` i `src/agent`. Utdata finns
//! i `fixtures/jobmanifest-go.json` - samma manifest som Go skrev ut.
//! Utan den här testen kan Rust-manifestet glida isär från Go:s utan att
//! någon märker det.

use serde_json::Value;

use plan_agent::jobrunner::{job_name, Config, Runner, Spec};
use plan_agent::k8s::Client;

/// Bygger manifestet med samma uppgifter som Go-fixturen.
fn manifest() -> Value {
    let config = Config {
        service_account: "plan-agent-worker".to_string(),
        plan_api_url: "http://plan.plan.svc.cluster.local".to_string(),
        api_key_secret_name: "plan-api-key".to_string(),
        api_key_secret_key: "PLAN_API_KEY".to_string(),
        llm_secret_name: "plan-llm".to_string(),
        llm_secret_key: "MINIMAX_API_KEY".to_string(),
        git_secret_name: "plan-git".to_string(),
        git_secret_key: "GIT_AUTH_HEADER".to_string(),
        repo_url: "https://github.com/simon/plan.git".to_string(),
        repo_branch: "main".to_string(),
        ..Default::default()
    };

    let runner = Runner::ny(Client::ny("https://127.0.0.1:6443", None).unwrap(), config);

    let job = runner
        .build(
            &Spec {
                goal_id: 42,
                run_id: 7,
                name: job_name(7),
                profile: "coding".to_string(),
            },
            "plan/agent:test",
        )
        .expect("manifestet ska kunna byggas");

    serde_json::to_value(job).expect("manifestet ska serialiseras")
}

#[test]
fn manifestet_matchar_go() {
    let go: Value = serde_json::from_str(include_str!("fixtures/jobmanifest-go.json"))
        .expect("fixturen ska vara giltig JSON");

    assert_eq!(manifest(), go);
}
