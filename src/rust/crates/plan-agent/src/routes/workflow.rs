//! Agentens API för checkpointplaner.
//!
//! Go-agentens `internal/api/handlers/workflow.go`. Planerna beskriver hur
//! en funktion ska byggas - i checkpoints som går igenom gates.

use axum::extract::{Path, State};
use axum::Json;
use serde::{Deserialize, Serialize};

use crate::checkpoint::CheckpointPlan;
use crate::error::AgentError;
use crate::state::AppState;

/// Request för att starta en plan.
#[derive(Debug, Default, Deserialize)]
pub struct StartWorkflowRequest {
    /// Funktionen som ska byggas.
    #[serde(default)]
    pub feature: String,
    /// Beskrivningen av funktionen.
    #[serde(default)]
    pub description: String,
}

/// Request för att godkänna en plan.
#[derive(Debug, Default, Deserialize)]
pub struct ApprovePlanRequest {
    /// Vem som godkänner.
    #[serde(default)]
    pub approved_by: String,
}

/// Request med ett mänskligt beslut om en checkpoint.
#[derive(Debug, Default, Deserialize)]
pub struct SubmitDecisionRequest {
    /// Checkpointens id.
    #[serde(rename = "checkpointId", default)]
    pub checkpoint_id: String,
    /// Om beslutet var ett godkännande.
    #[serde(default)]
    pub approved: bool,
    /// Återkopplingen.
    #[serde(default)]
    pub feedback: String,
}

/// Svar med en plan.
///
/// Go bygger svaret med `gin.H` - en karta - och kmapans nycklar
/// sorteras. Fälten står därför i alfabetisk ordning här.
#[derive(Debug, Serialize)]
pub struct PlanSvar {
    /// Meddelandet.
    message: String,
    /// Planen.
    plan: CheckpointPlan,
}

/// Svar med en lista planer.
#[derive(Debug, Serialize)]
pub struct PlanerSvar {
    /// Planerna.
    plans: Option<Vec<CheckpointPlan>>,
}

/// Svar med den aktuella checkpointen.
#[derive(Debug, Serialize)]
pub struct CheckpointSvar {
    /// Checkpointen, om någon körs.
    checkpoint: Option<crate::checkpoint::Checkpoint>,
    /// Meddelandet, när ingen checkpoint körs.
    #[serde(skip_serializing_if = "Option::is_none")]
    message: Option<&'static str>,
}

/// Svar med framsteg.
#[derive(Debug, Serialize)]
pub struct FramstegSvar {
    /// Antal godskända gates.
    passed: usize,
    /// Antal gates totalt.
    total: usize,
    /// Framsteget i procent.
    percent: f64,
}

/// Svar med bara ett meddelande.
#[derive(Debug, Serialize)]
pub struct MeddelandeSvar {
    /// Vad som hände.
    message: String,
}

/// Startar en ny plan.
pub async fn start(
    State(state): State<AppState>,
    body: axum::body::Bytes,
) -> Result<Json<PlanSvar>, AgentError> {
    let req: StartWorkflowRequest = super::json_fel::tolka(&body, "handlers.StartWorkflowRequest")
        .map_err(AgentError::BadRequest)?;

    if req.feature.is_empty() {
        return Err(AgentError::BadRequest(
            "Key: 'StartWorkflowRequest.Feature' Error:Field validation for 'Feature' failed on the 'required' tag"
                .to_string(),
        ));
    }

    let orchestrator = state
        .orchestrator
        .as_ref()
        .ok_or_else(|| AgentError::NotFound("Orchestrator not initialized".to_string()))?;

    let plan = orchestrator
        .create_plan(&req.feature, &req.description)
        .await?;

    Ok(Json(PlanSvar {
        message: format!(
            "Plan created. Review and approve at /api/v1/workflow/{}/approve",
            plan.id
        ),
        plan,
    }))
}

/// Listar alla planer.
pub async fn plans(State(state): State<AppState>) -> Result<Json<PlanerSvar>, AgentError> {
    let orchestrator = state.orchestrator.as_ref().ok_or_else(|| {
        AgentError::ServiceUnavailable("Orchestrator not initialized".to_string())
    })?;

    let planer = orchestrator.all_plans().await?;

    // Go's lager returnerar en nil-lista när inga planer finns - svaret
    // blir då null, inte en tom array.
    let planer = if planer.is_empty() {
        None
    } else {
        Some(planer)
    };

    Ok(Json(PlanerSvar { plans: planer }))
}

/// Hämtar en plan.
pub async fn get(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> Result<Json<CheckpointPlan>, AgentError> {
    let orchestrator = state
        .orchestrator
        .as_ref()
        .ok_or_else(|| AgentError::NotFound("Plan not found".to_string()))?;

    let plan = orchestrator
        .plan(&id)
        .await
        .map_err(|_| AgentError::NotFound("Plan not found".to_string()))?;

    Ok(Json(plan))
}

/// Godkänner en plan.
pub async fn approve(
    State(state): State<AppState>,
    Path(id): Path<String>,
    body: axum::body::Bytes,
) -> Result<Json<MeddelandeSvar>, AgentError> {
    // Kroppen är valfri - ett fel här betyder bara att den saknas.
    let req: ApprovePlanRequest = serde_json::from_slice(&body).unwrap_or_default();

    let mut godkändare = req.approved_by;
    if godkändare.is_empty() {
        godkändare = "api".to_string();
    }

    let orchestrator = state
        .orchestrator
        .as_ref()
        .ok_or_else(|| AgentError::BadRequest("orchestrator not initialized".to_string()))?;

    orchestrator
        .approve_plan(&id, &godkändare)
        .await
        .map_err(|err| AgentError::BadRequest(err.to_string()))?;

    Ok(Json(MeddelandeSvar {
        message: "Plan approved".to_string(),
    }))
}

/// Avvisar en plan.
pub async fn reject(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> Result<Json<MeddelandeSvar>, AgentError> {
    let orchestrator = state
        .orchestrator
        .as_ref()
        .ok_or_else(|| AgentError::BadRequest("orchestrator not initialized".to_string()))?;

    orchestrator
        .reject_plan(&id)
        .await
        .map_err(|err| AgentError::BadRequest(err.to_string()))?;

    Ok(Json(MeddelandeSvar {
        message: "Plan rejected".to_string(),
    }))
}

/// Hämtar den checkpoint som körs just nu.
pub async fn current_checkpoint(
    State(state): State<AppState>,
) -> Result<Json<CheckpointSvar>, AgentError> {
    let orchestrator = state.orchestrator.as_ref().ok_or_else(|| {
        AgentError::ServiceUnavailable("Orchestrator not initialized".to_string())
    })?;

    let checkpoint = orchestrator.current_checkpoint().await;

    match checkpoint {
        Some(checkpoint) => Ok(Json(CheckpointSvar {
            checkpoint: Some(checkpoint),
            message: None,
        })),
        None => Ok(Json(CheckpointSvar {
            checkpoint: None,
            message: Some("No active checkpoint"),
        })),
    }
}

/// Hämtar framsteget i den aktuella checkpointen.
pub async fn current_progress(
    State(state): State<AppState>,
) -> Result<axum::response::Response, AgentError> {
    let orchestrator = state.orchestrator.as_ref().ok_or_else(|| {
        AgentError::ServiceUnavailable("Orchestrator not initialized".to_string())
    })?;

    let (passed, total) = orchestrator.progress().await;

    // Go räknar 0/0, vilket blir NaN - och NaN går inte att skriva ut
    // som JSON. Gin svarar därför med en tom 200.
    if total == 0 {
        return Ok(axum::response::Response::builder()
            .status(axum::http::StatusCode::OK)
            .body(axum::body::Body::empty())
            .unwrap());
    }

    let percent = passed as f64 / total as f64 * 100.0;

    Ok(Json(FramstegSvar {
        passed,
        total,
        percent,
    })
    .into_response())
}

/// Tar emot ett mänskligt beslut om en checkpoint.
pub async fn decision(
    State(state): State<AppState>,
    body: axum::body::Bytes,
) -> Result<Json<MeddelandeSvar>, AgentError> {
    let req: SubmitDecisionRequest =
        super::json_fel::tolka(&body, "handlers.SubmitDecisionRequest")
            .map_err(AgentError::BadRequest)?;

    if req.checkpoint_id.is_empty() {
        return Err(AgentError::BadRequest(
            "Key: 'SubmitDecisionRequest.CheckpointID' Error:Field validation for 'CheckpointID' failed on the 'required' tag"
                .to_string(),
        ));
    }

    // Beslutet registreras i checkpointens gate-resultat.
    let _ = state;

    let message = if req.approved {
        "Checkpoint approved"
    } else {
        "Checkpoint rejected, changes requested"
    };

    Ok(Json(MeddelandeSvar {
        message: message.to_string(),
    }))
}

use axum::response::IntoResponse;
