//! Felhantering.
//!
//! Go-tjänsten svarar med `models.ErrorResponse`, som bara har fältet
//! `message` (och ett valfritt `details`). Samma format används här, så
//! gränssnittet behöver inte ändras.

use axum::http::StatusCode;
use axum::response::{IntoResponse, Response};
use axum::Json;
use serde_json::json;

#[derive(Debug, thiserror::Error)]
// Varianterna används i takt med att handlerna portas.
#[allow(dead_code)]
pub enum AppError {
    /// Databasfel - skickas med samma text som Go-tjänsten gör.
    #[error("{0}")]
    Database(#[from] sqlx::Error),

    /// Hittade inte raden som efterfrågades.
    #[error("{0}")]
    NotFound(String),

    /// Klienten skickade ogiltig data.
    #[error("{0}")]
    BadRequest(String),

    /// Klienten har inte rätt att göra det den försöker.
    #[error("{0}")]
    Forbidden(String),

    /// En konflikt - t.ex. att en budget stoppar en körning.
    ///
    /// Svaret innehåller även `details`, så att klienten vet vad som
    /// stoppade - inte bara att något gick snett.
    #[error("{message}")]
    Conflict {
        /// Meddelandet som skickas.
        message: String,
        /// Detaljerna om vad som hände.
        details: serde_json::Value,
    },

    /// Ingen giltig autentisering.
    #[error("{0}")]
    Unauthorized(String),

    /// Något gick fel internt.
    #[error("{0}")]
    Internal(String),

    /// Endpointen finns i route-tabellen men är inte portad ännu.
    #[error("{0}")]
    NotImplemented(String),

    /// Databasen är inte ansluten - samma svar som Go-tjänsten ger.
    #[error("{0}")]
    ServiceUnavailable(String),

    /// En tidsstämpel kunde inte tolkas som RFC3339.
    #[error("{0}")]
    InvalidTimestamp(String),

    /// Ett due-datum kunde inte tolkas som "YYYY-MM-DD".
    #[error("due-datumet måste vara på formatet YYYY-MM-DD")]
    InvalidDueDate,

    /// Ett startdatum kunde inte tolkas som "YYYY-MM-DD".
    #[error("startdatumet måste vara på formatet YYYY-MM-DD")]
    InvalidStartDate,

    /// Den uppskattade tidsåtgången är inte ett positivt antal minuter.
    #[error("tidsåtgången måste vara ett positivt antal minuter")]
    InvalidEstimate,
}

impl AppError {
    /// Statuskod som ska skickas med svaret.
    pub fn status(&self) -> StatusCode {
        match self {
            AppError::NotFound(_) => StatusCode::NOT_FOUND,
            AppError::BadRequest(_) => StatusCode::BAD_REQUEST,
            AppError::Forbidden(_) => StatusCode::FORBIDDEN,
            AppError::Conflict { .. } => StatusCode::CONFLICT,
            AppError::Unauthorized(_) => StatusCode::UNAUTHORIZED,
            AppError::NotImplemented(_) => StatusCode::NOT_IMPLEMENTED,
            AppError::ServiceUnavailable(_) => StatusCode::SERVICE_UNAVAILABLE,
            AppError::InvalidTimestamp(_)
            | AppError::InvalidDueDate
            | AppError::InvalidStartDate
            | AppError::InvalidEstimate => StatusCode::BAD_REQUEST,
            AppError::Database(_) | AppError::Internal(_) => StatusCode::INTERNAL_SERVER_ERROR,
        }
    }
}

impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        let status = self.status();

        // Interna fel loggas - klienten ska inte se databasdetaljer.
        let message = match &self {
            AppError::Database(err) => {
                tracing::error!("Databasfel: {err}");
                err.to_string()
            }
            AppError::Internal(err) => {
                tracing::error!("Internt fel: {err}");
                err.clone()
            }
            other => other.to_string(),
        };

        // Konflikter har ett extra fält - därför byggs svaret särskilt.
        if let AppError::Conflict { message, details } = &self {
            return (
                status,
                Json(json!({ "message": message, "details": details })),
            )
                .into_response();
        }

        (status, Json(json!({ "message": message }))).into_response()
    }
}

/// Resultat-typ för handlers.
#[allow(dead_code)]
pub type AppResult<T> = Result<T, AppError>;
