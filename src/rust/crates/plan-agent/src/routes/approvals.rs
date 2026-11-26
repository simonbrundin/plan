//! Godkännanden - förslag som väntar på ett mänskligt beslut.
//!
//! Go-agentens `internal/api/handlers/approvals.go`. Beslutet tas i Plan,
//! där målet finns - Slack används inte.

use axum::extract::{Path, State};
use axum::Json;
use serde::{Deserialize, Serialize};

use crate::approvals::{PendingApproval, STATUS_PENDING};
use crate::error::AgentError;
use crate::state::AppState;

/// Request för ett beslut.
#[derive(Debug, Default, Deserialize)]
pub struct DecisionRequest {
    /// Beslutet - "approve" eller "reject".
    pub decision: String,
    /// Återkopplingen till agenten.
    #[serde(default)]
    pub feedback: String,
}

/// Svar för listan av förslag.
#[derive(Debug, Serialize)]
pub struct GodkännandenSvar {
    /// Förslagen som väntar på ett beslut.
    approvals: Vec<PendingApproval>,
}

/// Svar efter ett beslut.
#[derive(Debug, Serialize)]
pub struct BeslutsSvar {
    /// Vad som beslutades.
    status: &'static str,
    /// Målets id.
    goal_id: i64,
    /// Återkopplingen, vid avslag.
    #[serde(skip_serializing_if = "Option::is_none")]
    feedback: Option<String>,
}

/// Listar förslagen som väntar på ett beslut.
pub async fn list(State(state): State<AppState>) -> Result<Json<GodkännandenSvar>, AgentError> {
    let Some(store) = state.approval_store.as_ref() else {
        return Ok(Json(GodkännandenSvar {
            approvals: Vec::new(),
        }));
    };

    let alla = store.list()?;
    let väntande: Vec<PendingApproval> = alla
        .into_iter()
        .filter(|godkännande| godkännande.status.is_empty() || godkännande.status == STATUS_PENDING)
        .collect();

    Ok(Json(GodkännandenSvar {
        approvals: väntande,
    }))
}

/// Hämtar ett förslag.
pub async fn get(
    State(state): State<AppState>,
    Path(goal_id): Path<String>,
) -> Result<Json<PendingApproval>, AgentError> {
    let goal_id = tolka_goal_id(&goal_id)?;

    let Some(store) = state.approval_store.as_ref() else {
        return Err(AgentError::NotFound(
            "No pending approval found".to_string(),
        ));
    };

    let godkännande = store
        .get(goal_id)?
        .ok_or_else(|| AgentError::NotFound("No pending approval found".to_string()))?;

    Ok(Json(godkännande))
}

/// Godkänner eller avvisar ett förslag.
pub async fn decide(
    State(state): State<AppState>,
    Path(goal_id): Path<String>,
    body: axum::body::Bytes,
) -> Result<Json<BeslutsSvar>, AgentError> {
    let goal_id = tolka_goal_id(&goal_id)?;

    let Some(store) = state.approval_store.as_ref() else {
        return Err(AgentError::NotFound(
            "No pending approval found".to_string(),
        ));
    };

    let req: DecisionRequest = serde_json::from_slice(&body)
        .map_err(|_| AgentError::BadRequest("Invalid request body".to_string()))?;

    let Some(godkännande) = store.get(goal_id)? else {
        return Err(AgentError::NotFound(
            "No pending approval found".to_string(),
        ));
    };

    match req.decision.as_str() {
        "approve" => {
            godkänn(&state, goal_id, &godkännande).await;

            // Beslutet skrivs till körningens histori i Plan - det är
            // där det syns att en människa var med.
            spara_beslut(&state, goal_id, "approve", &godkännande.feedback).await;

            if let Err(err) = store.delete(goal_id) {
                tracing::warn!("Kunde inte ta bort godkännandet för mål {goal_id}: {err}");
            }

            Ok(Json(BeslutsSvar {
                status: "approved",
                goal_id,
                feedback: None,
            }))
        }
        "reject" => {
            let mut avslaget = godkännande.clone();
            avslaget.status = crate::approvals::STATUS_REJECTED.to_string();
            avslaget.feedback = req.feedback.clone();

            if let Err(err) = store.save(&avslaget) {
                tracing::warn!("Kunde inte spara avslaget för mål {goal_id}: {err}");
            }

            // Agenten plockar upp målet igen vid nästa poll och definierar
            // om det med återkopplingen.
            markera_körning(&state, goal_id, &avslaget, "", "queued").await;

            spara_beslut(&state, goal_id, "reject", &req.feedback).await;

            Ok(Json(BeslutsSvar {
                status: "rejected",
                goal_id,
                feedback: Some(req.feedback),
            }))
        }
        _ => Err(AgentError::BadRequest("Unknown decision".to_string())),
    }
}

/// Skriver det föreslagna till Plan och flyttar målet till plan-fasen.
async fn godkänn(state: &AppState, goal_id: i64, godkännande: &PendingApproval) {
    let Some(plan) = state.plan_client.as_ref() else {
        return;
    };

    if let Err(err) = plan
        .update_goal_with_phase(
            goal_id,
            Some(godkännande.title.clone()),
            Some(godkännande.description.clone()),
            "plan",
        )
        .await
    {
        tracing::error!("Kunde inte uppdatera mål {goal_id} i Plan: {err}");
        return;
    }

    tracing::info!("Mål {goal_id} godkänt, status -> plan");

    markera_körning(state, goal_id, godkännande, "plan", "succeeded").await;
}

/// Markerar körningen i Plan utifrån beslutet.
async fn markera_körning(
    state: &AppState,
    goal_id: i64,
    godkännande: &PendingApproval,
    fas: &str,
    status: &str,
) {
    let Some(plan) = state.plan_client.as_ref() else {
        return;
    };

    let Ok(körningar) = plan.agent_runs(Some(goal_id)).await else {
        return;
    };

    let Some(körning) = körningar.first() else {
        return;
    };

    let mut sammanfattning = körning.summary.clone();
    if sammanfattning.is_empty() {
        sammanfattning = godkännande.title.clone();
    }

    let mut begäran = crate::planclient::models::UpdateAgentRunRequest {
        status: Some(status.to_string()),
        summary: Some(sammanfattning),
        ..Default::default()
    };

    if !fas.is_empty() {
        begäran.phase = Some(fas.to_string());
    }

    if let Err(err) = plan.update_agent_run(körning.id, begäran).await {
        tracing::warn!("Kunde inte uppdatera agentkörning {}: {err}", körning.id);
    }
}

/// Skriver det mänskliga beslutet till körningens histori.
///
/// Beslutet fattas av en människa - agenten får aldrig godkänna sig själv,
/// så det är här historiken visar att en människa var med.
async fn spara_beslut(state: &AppState, goal_id: i64, beslut: &str, feedback: &str) {
    let Some(plan) = state.plan_client.as_ref() else {
        return;
    };

    let Ok(körningar) = plan.agent_runs(Some(goal_id)).await else {
        return;
    };

    let Some(körning) = körningar.first() else {
        return;
    };

    let mut nycklar = serde_json::Map::new();
    nycklar.insert(
        "decision".to_string(),
        serde_json::Value::String(beslut.to_string()),
    );
    if !feedback.is_empty() {
        nycklar.insert(
            "feedback".to_string(),
            serde_json::Value::String(feedback.to_string()),
        );
    }

    let begäran = crate::planclient::models::CreateAgentRunEventRequest {
        r#type: "approval_decided".to_string(),
        phase: None,
        payload: Some(serde_json::Value::Object(nycklar)),
    };

    if let Err(err) = plan.create_agent_run_event(körning.id, begäran).await {
        tracing::warn!("Kunde inte spara beslutet för mål {goal_id}: {err}");
    }
}

/// Läser och kontrollerar målets id.
///
/// Go svarar 400 med "Invalid goal ID: ..." när id:t inte är ett tal.
/// Go:s text kommer från `strconv.ParseInt` - den skrivs därför efter
/// samma mall, så att klienten ser samma meddelande.
fn tolka_goal_id(goal_id: &str) -> Result<i64, AgentError> {
    goal_id.parse::<i64>().map_err(|err| {
        let orsak = match err.kind() {
            std::num::IntErrorKind::PosOverflow | std::num::IntErrorKind::NegOverflow => {
                "value out of range"
            }
            _ => "invalid syntax",
        };

        AgentError::BadRequest(format!(
            "Invalid goal ID: strconv.ParseInt: parsing {goal_id:?}: {orsak}"
        ))
    })
}
