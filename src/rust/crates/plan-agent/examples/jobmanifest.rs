//! Skriver ut jobbmanifestet som JSON - för jämförelse mot Go.

use plan_agent::jobrunner::{Config, Runner, Spec};
use plan_agent::k8s::Client;

#[tokio::main]
async fn main() {
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
                name: plan_agent::jobrunner::job_name(7),
                profile: "coding".to_string(),
            },
            "plan/agent:test",
        )
        .unwrap();

    println!("{}", serde_json::to_string_pretty(&job).unwrap());
}
