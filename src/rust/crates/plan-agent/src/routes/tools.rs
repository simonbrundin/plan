//! Verktygskörningar - avbrytning och status.
//!
//! Go-agentens `internal/api/handlers/cancel.go`.

use axum::Json;
use serde::{Deserialize, Serialize};

use crate::error::AgentError;
use crate::tools::cancel;

/// Request för att avbryta ett verktygsanrop.
#[derive(Debug, Default, Deserialize)]
pub struct CancelRequest {
    /// Id:t för verktygsanropet.
    #[serde(rename = "toolCallId")]
    pub tool_call_id: String,
}

/// Svar på en avbrytning.
#[derive(Debug, Serialize)]
pub struct CancelResponse {
    /// Om avbrytningen gick igenom.
    pub success: bool,
    /// Meddelandet.
    pub message: String,
    /// Id:t för verktygsanropet.
    #[serde(rename = "toolCallId")]
    pub tool_call_id: String,
}

/// Svar från "avbryt alla".
///
/// Go bygger svaret med `gin.H` - en karta - och kmapans nycklar
/// sorteras. Fälten står därför i alfabetisk ordning här.
#[derive(Debug, Serialize)]
pub struct CancelAllResponse {
    /// Antalet anrop som avbröts.
    #[serde(rename = "cancelledCount")]
    pub cancelled_count: usize,
    /// Meddelandet.
    pub message: String,
    /// Om avbrytningen gick igenom.
    pub success: bool,
}

/// Svar med statusen för verktygen.
#[derive(Debug, Serialize)]
pub struct ToolStatusResponse {
    /// Antal pågående verktygsanrop.
    #[serde(rename = "runningCount")]
    pub running_count: usize,
}

/// Avbryter ett pågående verktygsanrop.
pub async fn cancel(body: axum::body::Bytes) -> Result<Json<CancelResponse>, AgentError> {
    let req: CancelRequest = serde_json::from_slice(&body)
        .map_err(|_| AgentError::BadRequest("toolCallId is required".to_string()))?;

    if req.tool_call_id.is_empty() {
        return Err(AgentError::BadRequest("toolCallId is required".to_string()));
    }

    let manager = cancel::global();

    // Finns inget anrop alls - då finns det inget att avbryta.
    if !manager.is_cancelled(&req.tool_call_id) && manager.count() == 0 {
        return Err(AgentError::ToolNotFound);
    }

    manager.cancel(&req.tool_call_id);

    Ok(Json(CancelResponse {
        success: true,
        message: "Cancellation signal sent".to_string(),
        tool_call_id: req.tool_call_id,
    }))
}

/// Avbryter alla pågående verktygsanrop.
pub async fn cancel_all() -> Json<CancelAllResponse> {
    let manager = cancel::global();
    let count = manager.count();

    manager.cancel_all();

    Json(CancelAllResponse {
        success: true,
        message: "Cancellation signals sent to all running tools".to_string(),
        cancelled_count: count,
    })
}

/// Status för pågående verktygsanrop.
pub async fn status() -> Json<ToolStatusResponse> {
    Json(ToolStatusResponse {
        running_count: cancel::global().count(),
    })
}
