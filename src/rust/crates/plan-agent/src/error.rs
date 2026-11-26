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

    /// Klienten skickade något agenten inte kan tolka.
    #[error("{0}")]
    BadRequest(String),
}

impl AgentError {
    /// Statuskod som ska skickas med svaret.
    pub fn status(&self) -> StatusCode {
        match self {
            AgentError::NotImplemented(_) => StatusCode::NOT_IMPLEMENTED,
            AgentError::NotFound(_) => StatusCode::NOT_FOUND,
            AgentError::BadRequest(_) => StatusCode::BAD_REQUEST,
            AgentError::Plan(_) | AgentError::Internal(_) => StatusCode::INTERNAL_SERVER_ERROR,
        }
    }
}

impl From<std::io::Error> for AgentError {
    fn from(err: std::io::Error) -> Self {
        AgentError::Internal(err.to_string())
    }
}

impl From<sqlx::Error> for AgentError {
    fn from(err: sqlx::Error) -> Self {
        AgentError::Internal(err.to_string())
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

        // Go-agentens svar har bara en nyckel - "error". Meddelandet är
        // därför texten i felet, utan en extra "message"-nyckel.
        if let AgentError::Internal(err) = &self {
            tracing::error!("Internt fel: {err}");
        }

        (status, Json(json!({ "error": self.to_string() }))).into_response()
    }
}
