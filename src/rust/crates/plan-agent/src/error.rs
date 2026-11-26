//! Felhantering för agent-API:et.

use axum::http::StatusCode;
use axum::response::{IntoResponse, Response};
use axum::Json;
use serde_json::json;

#[derive(Debug, thiserror::Error)]
pub enum AgentError {
    /// Endpointen finns i route-tabellen men är inte portad ännu.
    #[error("{0}")]
    NotImplemented(String),

    /// Klienten mot Plan API:fick fel.
    #[error("{0}")]
    Plan(String),

    /// Något gick fel internt.
    #[error("{0}")]
    Internal(String),

    /// Objektet hittades inte.
    #[error("{0}")]
    NotFound(String),
}

impl AgentError {
    /// Statuskod som ska skickas med svaret.
    pub fn status(&self) -> StatusCode {
        match self {
            AgentError::NotImplemented(_) => StatusCode::NOT_IMPLEMENTED,
            AgentError::NotFound(_) => StatusCode::NOT_FOUND,
            AgentError::Plan(_) | AgentError::Internal(_) => StatusCode::INTERNAL_SERVER_ERROR,
        }
    }
}

impl From<crate::planclient::PlanClientError> for AgentError {
    fn from(err: crate::planclient::PlanClientError) -> Self {
        AgentError::Plan(err.to_string())
    }
}

impl IntoResponse for AgentError {
    fn into_response(self) -> Response {
        let status = self.status();

        let message = match &self {
            AgentError::Internal(err) => {
                tracing::error!("Internt fel: {err}");
                "Ett internt fel inträffade".to_string()
            }
            other => other.to_string(),
        };

        let error = match &self {
            AgentError::NotImplemented(_) => "Not implemented",
            AgentError::NotFound(_) => "Not found",
            _ => "Internal server error",
        };

        (status, Json(json!({ "error": error, "message": message }))).into_response()
    }
}
