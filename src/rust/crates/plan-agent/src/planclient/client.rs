//! Klienten mot Plan API.
//!
//! Port av `src/agent/internal/planclient` - samma anrop, samma headers,
//! samma felmeddelanden. Agenten autentiserar med `X-API-Key` (Plan läser
//! API-nycklar i den headern; Bearer är sessionstoken i Plan).

use serde::de::DeserializeOwned;
use serde::Serialize;

use crate::planclient::models;
use crate::planclient::models::*;
use crate::planclient::paths;

/// Klienten mot Plan API.
#[derive(Debug, Clone)]
pub struct PlanClient {
    base_url: String,
    api_key: String,
    http: reqwest::Client,
}

/// Fel från ett anrop mot Plan.
#[derive(Debug, thiserror::Error)]
pub enum PlanClientError {
    #[error("PLAN_API_KEY environment variable is required")]
    MissingApiKey,
    #[error("request failed: {0}")]
    Request(#[from] reqwest::Error),
    #[error("failed to encode or decode JSON: {0}")]
    Json(#[from] serde_json::Error),
    #[error("API error {status}: {body}")]
    Api { status: u16, body: String },
    #[error("{0}")]
    Other(String),
}

impl PlanClient {
    /// Skapar en klient från miljövariablerna.
    ///
    /// `PLAN_API_URL` pekar på Plan API:t, `PLAN_API_KEY` är agentens nyckel.
    /// Utan nyckel är agenten inte konfigurerad - då ska tjänsten inte
    /// försöka prata med Plan alls.
    pub fn new() -> Result<Self, PlanClientError> {
        let base_url = std::env::var("PLAN_API_URL")
            .ok()
            .filter(|url| !url.is_empty())
            .unwrap_or_else(|| "http://localhost:8081".to_string());

        let api_key = std::env::var("PLAN_API_KEY")
            .ok()
            .filter(|key| !key.is_empty())
            .ok_or(PlanClientError::MissingApiKey)?;

        Ok(Self {
            base_url: base_url.trim_end_matches('/').to_string(),
            api_key,
            http: reqwest::Client::builder()
                .timeout(std::time::Duration::from_secs(30))
                .build()
                .map_err(|err| PlanClientError::Other(err.to_string()))?,
        })
    }

    /// Sökvägen till Plan API:t.
    pub fn base_url(&self) -> &str {
        &self.base_url
    }

    /// Skickar ett autentiserat anrop och tolkar svaret.
    async fn do_request<B, T>(
        &self,
        method: reqwest::Method,
        path: &str,
        body: Option<&B>,
    ) -> Result<T, PlanClientError>
    where
        B: Serialize + ?Sized,
        T: DeserializeOwned,
    {
        let url = format!("{}{}", self.base_url, path);

        let mut request = self.http.request(method, &url).header(
            reqwest::header::HeaderName::from_static("x-api-key"),
            &self.api_key,
        );

        if let Some(body) = body {
            request = request
                .header(reqwest::header::CONTENT_TYPE, "application/json")
                .body(serde_json::to_vec(body)?);
        }

        let response = request.send().await?;
        let status = response.status();
        let text = response.text().await?;

        if !status.is_success() {
            return Err(PlanClientError::Api {
                status: status.as_u16(),
                body: text,
            });
        }

        if text.trim().is_empty() {
            // Ett svar utan innehåll tolkas som "inget att returnera".
            return serde_json::from_str("null").map_err(PlanClientError::Json);
        }

        Ok(serde_json::from_str(&text)?)
    }

    /// Målen med en viss status.
    ///
    /// API:t filtrerar inte på status - filtreringen sker lokalt, som i
    /// Go-klienten.
    pub async fn goals_by_status(&self, status_id: i32) -> Result<Vec<Goal>, PlanClientError> {
        let goals: Vec<Goal> = self
            .do_request(reqwest::Method::GET, paths::PATH_GOALS, None::<&()>)
            .await?;
        Ok(goals
            .into_iter()
            .filter(|goal| goal.status_id == status_id)
            .collect())
    }

    /// Ett mål.
    pub async fn goal(&self, id: i64) -> Result<Goal, PlanClientError> {
        #[derive(serde::Deserialize)]
        struct Svar {
            goal: Goal,
        }
        let svar: Svar = self
            .do_request(
                reqwest::Method::GET,
                &paths::med_id(paths::PATH_GOAL, id),
                None::<&()>,
            )
            .await?;
        Ok(svar.goal)
    }

    /// Uppdaterar ett måls fält.
    pub async fn update_goal(
        &self,
        id: i64,
        req: UpdateGoalRequest,
    ) -> Result<Goal, PlanClientError> {
        #[derive(serde::Deserialize)]
        struct Svar {
            goal: Goal,
        }
        let svar: Svar = self
            .do_request(
                reqwest::Method::PATCH,
                &paths::med_id(paths::PATH_GOAL, id),
                Some(&req),
            )
            .await?;
        Ok(svar.goal)
    }

    /// Uppdaterar titel, beskrivning och status.
    pub async fn update_goal_with_status(
        &self,
        id: i64,
        title: Option<String>,
        description: Option<String>,
        status_id: i32,
    ) -> Result<(), PlanClientError> {
        self.update_goal(
            id,
            UpdateGoalRequest {
                title,
                description,
                ..Default::default()
            },
        )
        .await
        .map_err(|err| PlanClientError::Other(format!("failed to update goal: {err}")))?;

        if status_id > 0 {
            self.update_goal_status(id, status_id)
                .await
                .map_err(|err| PlanClientError::Other(format!("failed to update status: {err}")))?;
        }

        Ok(())
    }

    /// Uppdaterar titel, beskrivning och fas.
    pub async fn update_goal_with_phase(
        &self,
        id: i64,
        title: Option<String>,
        description: Option<String>,
        phase: &str,
    ) -> Result<(), PlanClientError> {
        self.update_goal(
            id,
            UpdateGoalRequest {
                title,
                description,
                ..Default::default()
            },
        )
        .await
        .map_err(|err| PlanClientError::Other(format!("failed to update goal: {err}")))?;

        self.update_goal_phase(id, phase)
            .await
            .map_err(|err| PlanClientError::Other(format!("failed to update goal phase: {err}")))?;

        Ok(())
    }

    /// Uppdaterar ett måls status.
    ///
    /// Status ändras via en egen endpoint - PATCH mot målet ignorerar
    /// `status_id`.
    pub async fn update_goal_status(&self, id: i64, status_id: i32) -> Result<(), PlanClientError> {
        #[derive(Serialize)]
        struct Body {
            status_id: i32,
        }
        let _: serde_json::Value = self
            .do_request(
                reqwest::Method::PATCH,
                &paths::med_id(paths::PATH_GOAL_STATUS, id),
                Some(&Body { status_id }),
            )
            .await?;
        Ok(())
    }

    /// Alla statusar.
    pub async fn statuses(&self) -> Result<Vec<Status>, PlanClientError> {
        self.do_request(reqwest::Method::GET, paths::PATH_STATUSES, None::<&()>)
            .await
    }

    /// Användarens platser.
    ///
    /// Agenten behöver dem för att kunna föreslå platser på ett mål - den
    /// föreslår bara platser som redan finns.
    pub async fn places(&self) -> Result<Vec<Place>, PlanClientError> {
        self.do_request(reqwest::Method::GET, paths::PATH_PLACES, None::<&()>)
            .await
    }

    /// Ersätter målets taggar med de angivna.
    ///
    /// API:t skapar taggar som inte finns ännu - så agenten kan föreslå en
    /// ny tagg. Befintliga taggar ska skickas med, eftersom anropet ersätter
    /// hela uppsättningen.
    pub async fn set_goal_tags(
        &self,
        goal_id: i64,
        tags: Vec<String>,
    ) -> Result<(), PlanClientError> {
        #[derive(Serialize)]
        struct Body {
            tags: Vec<String>,
        }
        let _: serde_json::Value = self
            .do_request(
                reqwest::Method::PUT,
                &paths::med_id(paths::PATH_GOAL_TAGS, goal_id),
                Some(&Body { tags }),
            )
            .await?;
        Ok(())
    }

    /// Ersätter målets platser med de angivna.
    pub async fn set_goal_places(
        &self,
        goal_id: i64,
        place_ids: Vec<i64>,
    ) -> Result<(), PlanClientError> {
        #[derive(Serialize)]
        struct Body {
            place_ids: Vec<i64>,
        }
        let _: serde_json::Value = self
            .do_request(
                reqwest::Method::PUT,
                &paths::med_id(paths::PATH_GOAL_PLACES, goal_id),
                Some(&Body { place_ids }),
            )
            .await?;
        Ok(())
    }

    /// Status-id:t för ett fasnamn.
    pub async fn status_id_by_name(&self, phase: &str) -> Result<i32, PlanClientError> {
        if !models::GOAL_PHASES.contains(&phase) {
            return Err(PlanClientError::Other(format!(
                "unknown goal phase {phase:?}"
            )));
        }

        for status in self
            .statuses()
            .await
            .map_err(|err| PlanClientError::Other(format!("failed to get goal statuses: {err}")))?
        {
            if status.name == phase {
                return Ok(status.id as i32);
            }
        }

        Err(PlanClientError::Other(format!(
            "goal phase {phase:?} is not available in Plan"
        )))
    }

    /// Fasnamnet för ett status-id.
    ///
    /// Det är väntet att statusen motsvarar en fas - men ett okänt id ger
    /// ett fel, inte en tom sträng.
    pub async fn phase_name(&self, status_id: i32) -> Result<String, PlanClientError> {
        for status in self
            .statuses()
            .await
            .map_err(|err| PlanClientError::Other(format!("failed to get goal statuses: {err}")))?
        {
            if status.id as i32 == status_id {
                return Ok(status.name);
            }
        }

        Err(PlanClientError::Other(format!(
            "status {status_id} is not available in Plan"
        )))
    }

    /// Sätter ett måls fas.
    pub async fn update_goal_phase(
        &self,
        goal_id: i64,
        phase: &str,
    ) -> Result<(), PlanClientError> {
        let status_id = self.status_id_by_name(phase).await?;
        self.update_goal_status(goal_id, status_id).await
    }

    /// Målen som är markerade för agentarbete.
    ///
    /// Endast markerade mål hämtas - det är agentens enda ingång.
    pub async fn agent_goals(&self) -> Result<Vec<Goal>, PlanClientError> {
        self.do_request(
            reqwest::Method::GET,
            &format!("{}?agent=true", paths::PATH_GOALS),
            None::<&()>,
        )
        .await
    }

    /// Skapar ett mål och kopplar det till ett föräldermål.
    pub async fn create_subgoal(
        &self,
        parent_id: i64,
        title: &str,
        description: &str,
    ) -> Result<Goal, PlanClientError> {
        let goal: Goal = self
            .do_request(
                reqwest::Method::POST,
                paths::PATH_GOALS,
                Some(&CreateGoalRequest {
                    title: title.to_string(),
                    icon: None,
                    status_id: Some(models::STATUS_DEFINE),
                }),
            )
            .await
            .map_err(|err| PlanClientError::Other(format!("create subgoal: {err}")))?;

        if !description.is_empty() {
            self.update_goal(
                goal.id,
                UpdateGoalRequest {
                    description: Some(description.to_string()),
                    ..Default::default()
                },
            )
            .await
            .map_err(|err| PlanClientError::Other(format!("set subgoal description: {err}")))?;
        }

        self.do_request::<CreateRelationRequest, ()>(
            reqwest::Method::POST,
            paths::PATH_GOAL_RELATIONS,
            Some(&CreateRelationRequest {
                child_id: goal.id,
                parent_id,
                order: 0,
                weight: 10,
            }),
        )
        .await
        .map_err(|err| PlanClientError::Other(format!("attach subgoal to parent: {err}")))?;

        Ok(goal)
    }

    /// Registrerar att agenten börjar arbeta med ett mål.
    pub async fn create_agent_run(
        &self,
        req: CreateAgentRunRequest,
    ) -> Result<AgentRun, PlanClientError> {
        self.do_request(reqwest::Method::POST, paths::PATH_AGENT_RUNS, Some(&req))
            .await
    }

    /// Rapporterar tokenanvändningen för en körning.
    pub async fn record_agent_run_usage(
        &self,
        id: i64,
        req: AgentRunUsageRequest,
    ) -> Result<(), PlanClientError> {
        let _: serde_json::Value = self
            .do_request(
                reqwest::Method::POST,
                &paths::med_id(paths::PATH_AGENT_RUN_USAGE, id),
                Some(&req),
            )
            .await?;
        Ok(())
    }

    /// Lägger till en strukturerad händelse i en körning.
    pub async fn create_agent_run_event(
        &self,
        id: i64,
        req: CreateAgentRunEventRequest,
    ) -> Result<(), PlanClientError> {
        let _: serde_json::Value = self
            .do_request(
                reqwest::Method::POST,
                &paths::med_id(paths::PATH_AGENT_RUN_EVENTS, id),
                Some(&req),
            )
            .await?;
        Ok(())
    }

    /// Uppdaterar fas, status eller sammanfattning för en körning.
    pub async fn update_agent_run(
        &self,
        id: i64,
        req: UpdateAgentRunRequest,
    ) -> Result<AgentRun, PlanClientError> {
        self.do_request(
            reqwest::Method::PATCH,
            &paths::med_id(paths::PATH_AGENT_RUN, id),
            Some(&req),
        )
        .await
    }

    /// Agentkörningarna som registrerats i Plan.
    pub async fn agent_runs(&self, goal_id: Option<i64>) -> Result<Vec<AgentRun>, PlanClientError> {
        let path = match goal_id {
            Some(id) => format!("{}?goal_id={}", paths::PATH_AGENT_RUNS, id),
            None => paths::PATH_AGENT_RUNS.to_string(),
        };
        self.do_request(reqwest::Method::GET, &path, None::<&()>)
            .await
    }

    /// Registrerar ett förslag som väntar på ett mänskligt beslut.
    pub async fn create_agent_approval(
        &self,
        run_id: i64,
        req: CreateAgentApprovalRequest,
    ) -> Result<AgentApproval, PlanClientError> {
        self.do_request(
            reqwest::Method::POST,
            &paths::med_id(paths::PATH_AGENT_RUN_APPROVALS, run_id),
            Some(&req),
        )
        .await
    }

    /// Godkännandena för en körning.
    pub async fn agent_approvals(
        &self,
        run_id: i64,
    ) -> Result<Vec<AgentApproval>, PlanClientError> {
        self.do_request(
            reqwest::Method::GET,
            &paths::med_id(paths::PATH_AGENT_RUN_APPROVALS, run_id),
            None::<&()>,
        )
        .await
    }
}
