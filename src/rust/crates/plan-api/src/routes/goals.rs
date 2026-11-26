//! Mål.
//!
//! Port av `src/api/internal/handlers/goals.go` och `goal_graph.go`.

use std::collections::HashMap;

use axum::body::Bytes;
use axum::extract::{Extension, Path, Query, State};
use axum::http::StatusCode;
use axum::response::{IntoResponse, Response};
use axum::Json;
use serde::Deserialize;

use crate::auth::Caller;
use crate::error::AppError;
use crate::models::goal::{
    ChildRelation, ChildWithDeps, CreateGoalRequest, GetGoalResponse, Goal, GoalDependency,
    GoalFilter, GoalGraphResponse, GoalRelation, GoalWithRelations, PrioritizedGoal,
    UpdateGoalRequest,
};
use crate::repository::{goal, place, tag};
use crate::routes::state::AppState;

/// Databasen är inte ansluten - samma svar som Go-tjänsten.
fn no_database() -> AppError {
    AppError::ServiceUnavailable(
        "Database not connected. Set DATABASE_URL environment variable.".to_string(),
    )
}

/// Användarens id ur anropet.
fn user_id(caller: &Caller) -> Result<i64, AppError> {
    if caller.user_id == 0 {
        return Err(AppError::Unauthorized("Unauthorized".to_string()));
    }
    Ok(caller.user_id)
}

/// Tolkar ett mål-id ur sökvägen.
fn goal_id(id: &str) -> Result<i64, AppError> {
    id.parse::<i64>()
        .map_err(|_| AppError::BadRequest("Invalid goal ID".to_string()))
}

/// Tolkar ett booleskt värde som Go gör.
///
/// Go accepterar "1", "t", "T", "TRUE", "true", "True" och deras
/// motsvarigheter - inte bara "true" och "false".
pub fn parse_bool(value: &str) -> Option<bool> {
    match value {
        "1" | "t" | "T" | "TRUE" | "true" | "True" => Some(true),
        "0" | "f" | "F" | "FALSE" | "false" | "False" => Some(false),
        _ => None,
    }
}

/// Filterparametrarna för mållistan.
#[derive(Debug, Default, serde::Deserialize)]
pub struct GoalQuery {
    /// Filtrerar på agent-flaggan.
    #[serde(default)]
    pub agent: Option<String>,
    /// Filtrerar på status.
    #[serde(default)]
    pub status: Option<String>,
}

/// Läser filterparametrarna.
fn filter_from_query(query: &GoalQuery) -> Result<GoalFilter, AppError> {
    let mut filter = GoalFilter::default();

    if let Some(rå) = query.agent.as_ref().filter(|v| !v.is_empty()) {
        let agent = parse_bool(rå)
            .ok_or_else(|| AppError::BadRequest(format!("invalid agent filter: {rå:?}")))?;
        filter.agent = Some(agent);
    }

    if let Some(rå) = query.status.as_ref().filter(|v| !v.is_empty()) {
        let status_id = rå
            .parse::<i32>()
            .map_err(|_| AppError::BadRequest(format!("invalid status filter: {rå:?}")))?;
        filter.status_id = Some(status_id);
    }

    Ok(filter)
}

/// Fyller i målens platser och taggar.
///
/// Ett fel här ska inte förstöra svaret - platserna och taggarna är extra
/// information, så felet hoppas över i stället för att returneras.
async fn platser_och_taggar(
    pool: &sqlx::PgPool,
    ids: &[i64],
) -> (
    HashMap<i64, Vec<crate::models::place::Place>>,
    HashMap<i64, Vec<String>>,
) {
    let platser = place::get_places_for_goals(pool, ids)
        .await
        .unwrap_or_default();
    let taggar = tag::tags_for_goals(pool, ids).await.unwrap_or_default();

    (platser, taggar)
}

/// Fyller i målens platser och taggar.
async fn attach_places_and_tags(pool: &sqlx::PgPool, goals: &mut [Goal]) {
    if goals.is_empty() {
        return;
    }

    let ids: Vec<i64> = goals.iter().map(|mål| mål.id).collect();
    let (platser, taggar) = platser_och_taggar(pool, &ids).await;

    for mål in goals.iter_mut() {
        mål.places = Some(platser.get(&mål.id).cloned().unwrap_or_default());
        mål.tags = Some(taggar.get(&mål.id).cloned().unwrap_or_default());
    }
}

/// Användarens hela målgraph.
///
/// Den byggs med ett fåtal frågor i stället för en fråga per mål. Mot Neon
/// kostar varje rundtur ungefär 40 ms, så en målsida med 28 undermål gick
/// från 36 rundturer till fyra.
pub struct GoalGraph {
    /// Användarens mål.
    pub goals: Vec<Goal>,
    /// Relationerna mellan målen.
    pub relations: Vec<GoalRelation>,
    /// Beroendena mellan målen.
    pub dependencies: Vec<GoalDependency>,

    goals_by_id: HashMap<i64, Goal>,
    /// Relationer per förälder, sorterade på ordning.
    children_of: HashMap<i64, Vec<GoalRelation>>,
    /// Relationer per barn.
    parents_of: HashMap<i64, Vec<GoalRelation>>,
    /// Målen som ett mål beror på.
    depends_on: HashMap<i64, Vec<Goal>>,
    /// Målen som beror på ett mål.
    blocking: HashMap<i64, Vec<Goal>>,
}

impl GoalGraph {
    /// Laddar grafen för en användare.
    pub async fn load(pool: &sqlx::PgPool, user_id: i64) -> Result<Self, AppError> {
        let mut goals = goal::get_goals_filtered(pool, user_id, &GoalFilter::default()).await?;
        let relations = goal::get_all_relations(pool, user_id).await?;
        let dependencies = goal::get_all_dependencies(pool, user_id).await?;

        attach_places_and_tags(pool, &mut goals).await;

        let mut goals_by_id = HashMap::with_capacity(goals.len());
        let mut children_of: HashMap<i64, Vec<GoalRelation>> = HashMap::new();
        let mut parents_of: HashMap<i64, Vec<GoalRelation>> = HashMap::new();

        for mål in &goals {
            goals_by_id.insert(mål.id, mål.clone());
        }

        for relation in &relations {
            children_of
                .entry(relation.parent_id)
                .or_default()
                .push(relation.clone());
            parents_of
                .entry(relation.child_id)
                .or_default()
                .push(relation.clone());
        }

        // Barnen sorteras på ordning här, så att ordningen inte hänger på hur
        // databasen råkar returnera raderna. Barn med samma ordning sorteras på
        // id, så att resultatet blir detsamma varje gång.
        for relationer in children_of.values_mut() {
            relationer.sort_by(|a, b| {
                a.order
                    .cmp(&b.order)
                    .then_with(|| a.child_id.cmp(&b.child_id))
            });
        }

        let mut depends_on: HashMap<i64, Vec<Goal>> = HashMap::new();
        let mut blocking: HashMap<i64, Vec<Goal>> = HashMap::new();

        for beroende in &dependencies {
            let (Some(beroende_mål), Some(mål)) = (
                goals_by_id.get(&beroende.depends_on_id),
                goals_by_id.get(&beroende.goal_id),
            ) else {
                continue;
            };

            depends_on
                .entry(beroende.goal_id)
                .or_default()
                .push(beroende_mål.clone());
            blocking
                .entry(beroende.depends_on_id)
                .or_default()
                .push(mål.clone());
        }

        Ok(Self {
            goals,
            relations,
            dependencies,
            goals_by_id,
            children_of,
            parents_of,
            depends_on,
            blocking,
        })
    }

    /// Ett mål ur grafen.
    pub fn goal(&self, goal_id: i64) -> Option<&Goal> {
        self.goals_by_id.get(&goal_id)
    }

    /// Relationerna till målets barn.
    pub fn child_relations(&self, goal_id: i64) -> Vec<ChildRelation> {
        self.children_of
            .get(&goal_id)
            .map(|relationer| {
                relationer
                    .iter()
                    .map(|relation| ChildRelation {
                        child_id: relation.child_id,
                        order: relation.order,
                        weight: relation.weight,
                    })
                    .collect()
            })
            .unwrap_or_default()
    }

    /// Målets barn med ordning, vikt och beroenden.
    pub fn children(&self, goal_id: i64) -> Vec<ChildWithDeps> {
        self.children_of
            .get(&goal_id)
            .map(|relationer| {
                relationer
                    .iter()
                    .filter_map(|relation| {
                        self.goals_by_id
                            .get(&relation.child_id)
                            .map(|barn| ChildWithDeps {
                                goal: barn.clone(),
                                order: relation.order,
                                weight: relation.weight,
                                depends_on: self.depends_on_of(relation.child_id),
                                blocking: self.blocking_of(relation.child_id),
                            })
                    })
                    .collect()
            })
            .unwrap_or_default()
    }

    /// Målets föräldrar.
    pub fn parents(&self, goal_id: i64) -> Vec<Goal> {
        self.parents_of
            .get(&goal_id)
            .map(|relationer| {
                relationer
                    .iter()
                    .filter_map(|relation| self.goals_by_id.get(&relation.parent_id).cloned())
                    .collect()
            })
            .unwrap_or_default()
    }

    /// Beroenderaderna som gäller ett mål.
    pub fn dependencies_of(&self, goal_id: i64) -> Vec<GoalDependency> {
        self.dependencies
            .iter()
            .filter(|beroende| beroende.goal_id == goal_id)
            .cloned()
            .collect()
    }

    /// Målen som ett mål beror på.
    pub fn depends_on_of(&self, goal_id: i64) -> Vec<Goal> {
        self.depends_on.get(&goal_id).cloned().unwrap_or_default()
    }

    /// Målen som blockerar ett mål.
    pub fn blocking_of(&self, goal_id: i64) -> Vec<Goal> {
        self.blocking.get(&goal_id).cloned().unwrap_or_default()
    }
}

/// Listar användarens mål.
#[utoipa::path(
    get,
    path = "/api/v1/goals",
    tag = "goals",
    params(
        ("agent" = Option<bool>, Query, description = "Filtrera på agent-flaggan"),
        ("status" = Option<i32>, Query, description = "Filtrera på status-id")
    ),
    responses(
        (status = 200, description = "Målen", content_type = "application/json", body = Vec<Goal>),
        (status = 400, description = "Ogiltigt filter"),
        (status = 401, description = "Ingen giltig session")
    )
)]
pub async fn get_goals(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Query(query): Query<GoalQuery>,
) -> Result<Json<Vec<Goal>>, AppError> {
    let user_id = user_id(&caller)?;
    let filter = filter_from_query(&query)?;

    let Some(pool) = state.pool.as_ref() else {
        return Err(no_database());
    };

    let mut goals = goal::get_goals_filtered(pool, user_id, &filter).await?;

    attach_places_and_tags(pool, &mut goals).await;

    Ok(Json(goals))
}

/// Användarens prioriteringslista.
#[utoipa::path(
    get,
    path = "/api/v1/goals/prioritized",
    tag = "goals",
    responses(
        (status = 200, description = "Prioriteringslistan", content_type = "application/json",
            body = Vec<PrioritizedGoal>),
        (status = 401, description = "Ingen giltig session")
    )
)]
pub async fn get_prioritized_goals(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
) -> Result<Json<Vec<PrioritizedGoal>>, AppError> {
    let user_id = user_id(&caller)?;

    let Some(pool) = state.pool.as_ref() else {
        return Err(no_database());
    };

    let mut goals = goal::get_prioritized_goals(pool, user_id).await?;

    // Prioriterade mål är en lista av mål - de platser och taggar som hör
    // till dem hämtas i ett svep.
    let ids: Vec<i64> = goals.iter().map(|mål| mål.goal.id).collect();
    let (platser, taggar) = platser_och_taggar(pool, &ids).await;

    for mål in goals.iter_mut() {
        mål.goal.places = Some(platser.get(&mål.goal.id).cloned().unwrap_or_default());
        mål.goal.tags = Some(taggar.get(&mål.goal.id).cloned().unwrap_or_default());
    }

    Ok(Json(goals))
}

/// Hämter ett mål med sina relationer och beroenden.
#[utoipa::path(
    get,
    path = "/api/v1/goals/{id}",
    tag = "goals",
    params(("id" = i64, Path, description = "Målets id")),
    responses(
        (status = 200, description = "Målet", content_type = "application/json",
            body = GetGoalResponse),
        (status = 400, description = "Ogiltigt mål-id"),
        (status = 404, description = "Målet finns inte")
    )
)]
pub async fn get_goal(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
) -> Result<Json<GetGoalResponse>, AppError> {
    let user_id = user_id(&caller)?;
    let goal_id = goal_id(&id)?;

    let Some(pool) = state.pool.as_ref() else {
        return Err(no_database());
    };

    let graph = GoalGraph::load(pool, user_id).await?;

    let Some(mål) = graph.goal(goal_id) else {
        // Målet tillhör inte användaren - det finns inte att hämta.
        return Err(AppError::NotFound("Goal not found".to_string()));
    };

    let parents = graph.parents(goal_id);

    Ok(Json(GetGoalResponse {
        goal: GoalWithRelations {
            goal: mål.clone(),
            child_relations: graph.child_relations(goal_id),
            parent_relations: parents.clone(),
        },
        children: graph.children(goal_id),
        parents,
        all_goals: graph.goals.clone(),
        dependencies: graph.dependencies_of(goal_id),
        depends_on: graph.depends_on_of(goal_id),
        blocking: graph.blocking_of(goal_id),
        status_history: None,
    }))
}

/// Hela målgraphen för användaren.
#[utoipa::path(
    get,
    path = "/api/v1/goals/graph",
    tag = "goals",
    responses(
        (status = 200, description = "Grafen", content_type = "application/json",
            body = GoalGraphResponse),
        (status = 401, description = "Ingen giltig session")
    )
)]
pub async fn get_goal_graph(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
) -> Result<Json<GoalGraphResponse>, AppError> {
    let user_id = user_id(&caller)?;

    let Some(pool) = state.pool.as_ref() else {
        return Err(no_database());
    };

    let graph = GoalGraph::load(pool, user_id).await?;

    Ok(Json(GoalGraphResponse {
        goals: graph.goals,
        relations: graph.relations,
        dependencies: graph.dependencies,
    }))
}

/// Skapar ett mål.
#[utoipa::path(
    post,
    path = "/api/v1/goals",
    tag = "goals",
    request_body = CreateGoalRequest,
    responses(
        (status = 201, description = "Det skapade målet", content_type = "application/json",
            body = Goal),
        (status = 400, description = "Titeln saknas"),
        (status = 401, description = "Ingen giltig session")
    )
)]
pub async fn create_goal(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    body: Bytes,
) -> Result<Response, AppError> {
    let user_id = user_id(&caller)?;

    let req: CreateGoalRequest = crate::routes::parse_body(&body, "Title is required")?;

    if req.title.is_empty() {
        return Err(AppError::BadRequest("Title is required".to_string()));
    }

    let Some(pool) = state.pool.as_ref() else {
        return Err(no_database());
    };

    let icon = if req.icon.is_empty() {
        "heroicons:star"
    } else {
        req.icon.as_str()
    };

    // Upprepningen valideras innan målet skapas, så att en ogiltig regel
    // inte lämnar kvar ett tomt mål.
    let mut regel = String::new();
    let mut läge = "schedule".to_string();
    let mut start = chrono::Utc::now().date_naive();

    if let Some(angiven) = req.recurrence_rule.as_deref() {
        regel = angiven.trim().to_string();

        plan_recurrence::Regel::validera(&regel).map_err(|fel| {
            AppError::BadRequest(format!("ogiltig upprepningsregel: {}", detalj_av_fel(&fel)))
        })?;

        if let Some(angiven_läge) = req.recurrence_mode.as_deref().filter(|v| !v.is_empty()) {
            plan_recurrence::Läge::tolka(angiven_läge).map_err(|_| {
                AppError::BadRequest(
                    "Ogiltig schemaläggningsstrategi. Använd 'schedule' eller 'push'.".to_string(),
                )
            })?;
            läge = angiven_läge.to_string();
        }

        if let Some(angivet) = req
            .recurrence_start_date
            .as_deref()
            .filter(|v| !v.is_empty())
        {
            start = chrono::NaiveDate::parse_from_str(angivet, "%Y-%m-%d").map_err(|_| {
                AppError::BadRequest(
                    "recurrence_start_date måste vara på formatet YYYY-MM-DD".to_string(),
                )
            })?;
        }

        // Regeln ska ge åtminstone en förekomst.
        plan_recurrence::Regel::tolka(&regel)
            .and_then(|tolkad| {
                tolkad.första_förekomst(chrono::DateTime::from_naive_utc_and_offset(
                    start
                        .and_hms_opt(0, 0, 0)
                        .expect("midnatt är ett giltigt klockslag"),
                    chrono::Utc,
                ))
            })
            .map_err(|fel| match fel {
                plan_recurrence::Fel::IngenFörekomst(_) => {
                    AppError::BadRequest("Regeln ger inga förekomster.".to_string())
                }
                other => AppError::BadRequest(format!(
                    "ogiltig upprepningsregel: {}",
                    detalj_av_fel(&other)
                )),
            })?;
    }

    // Startdatumet valideras innan målet skapas, så att ett ogiltigt
    // datum inte lämnar kvar ett mål utan startdatum.
    let mut startdatum = None;
    if let Some(angivet) = req.start_date.as_deref().filter(|v| !v.is_empty()) {
        let datum = chrono::NaiveDate::parse_from_str(angivet, "%Y-%m-%d").map_err(|_| {
            AppError::BadRequest("start_date måste vara på formatet YYYY-MM-DD".to_string())
        })?;
        startdatum = Some(datum);
    }

    let mål = goal::create_goal(pool, user_id, &req.title, icon, startdatum).await?;

    if !regel.is_empty() {
        goal::set_recurrence(pool, mål.id, &regel, &läge, start)
            .await
            .map_err(upprepningsfel)?;

        // Målet läses om - nu har serien fått sin första förekomst.
        let mål = goal::get_goal_by_id(pool, mål.id)
            .await?
            .ok_or_else(|| AppError::NotFound("Goal not found".to_string()))?;

        return Ok((StatusCode::CREATED, Json(mål)).into_response());
    }

    Ok((StatusCode::CREATED, Json(mål)).into_response())
}

/// Uppdaterar ett mål.
#[utoipa::path(
    patch,
    path = "/api/v1/goals/{id}",
    tag = "goals",
    params(("id" = i64, Path, description = "Målets id")),
    request_body = UpdateGoalRequest,
    responses(
        (status = 200, description = "Det uppdaterade målet", content_type = "application/json",
            body = Goal),
        (status = 400, description = "Ogiltig data eller ogiltigt mål-id"),
        (status = 403, description = "Ingen åtkomst till målet")
    )
)]
pub async fn update_goal(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
    body: Bytes,
) -> Result<Json<Goal>, AppError> {
    let user_id = user_id(&caller)?;
    let goal_id = goal_id(&id)?;

    let req: UpdateGoalRequest = crate::routes::parse_body(&body, "No fields to update")?;

    let Some(pool) = state.pool.as_ref() else {
        return Err(no_database());
    };

    if !goal::user_has_access(pool, user_id, goal_id).await? {
        return Err(AppError::Forbidden("Access denied".to_string()));
    }

    // Ett evigt mål kan inte slutföras - det pågår hela livet.
    if req.finishes_goal() {
        if let Some(existerande) = goal::get_goal_by_id(pool, goal_id).await? {
            if existerande.eternal {
                return Err(AppError::BadRequest(
                    "Ett evigt mål kan inte slutföras".to_string(),
                ));
            }
        }
    }

    let mål = goal::update_goal(pool, goal_id, &req).await?;

    // Upprepningen sätts efter de vanliga fälten - en ny regel startar en
    // ny serie.
    if req.recurrence_rule.set || req.recurrence_mode.set {
        goal::apply_recurrence_update(pool, goal_id, &req)
            .await
            .map_err(upprepningsfel)?;

        return Ok(Json(
            goal::get_goal_by_id(pool, goal_id)
                .await?
                .ok_or_else(|| AppError::NotFound("Goal not found".to_string()))?,
        ));
    }

    Ok(Json(mål))
}

/// Tar bort ett mål.
#[utoipa::path(
    delete,
    path = "/api/v1/goals/{id}",
    tag = "goals",
    params(("id" = i64, Path, description = "Målets id")),
    responses(
        (status = 204, description = "Målet är borttaget"),
        (status = 400, description = "Ogiltigt mål-id"),
        (status = 403, description = "Ingen åtkomst till målet")
    )
)]
pub async fn delete_goal(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
) -> Result<Response, AppError> {
    let user_id = user_id(&caller)?;
    let goal_id = goal_id(&id)?;

    let Some(pool) = state.pool.as_ref() else {
        return Err(no_database());
    };

    if !goal::user_has_access(pool, user_id, goal_id).await? {
        return Err(AppError::Forbidden("Access denied".to_string()));
    }

    goal::delete_goal(pool, goal_id).await?;

    Ok(StatusCode::NO_CONTENT.into_response())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn booleska_varden_foljer_go() {
        assert_eq!(parse_bool("true"), Some(true));
        assert_eq!(parse_bool("TRUE"), Some(true));
        assert_eq!(parse_bool("1"), Some(true));
        assert_eq!(parse_bool("t"), Some(true));
        assert_eq!(parse_bool("false"), Some(false));
        assert_eq!(parse_bool("0"), Some(false));
        assert_eq!(parse_bool("f"), Some(false));
        assert_eq!(parse_bool("ja"), None);
        assert_eq!(parse_bool(""), None);
    }
}

/// Översätter ett upprepningsfel till Go:s meddelanden.
///
/// Go skiljer på "ogiltig regel", "ogiltig strategi", "ingen förekomst"
/// och "målet är inte återkommande" - och ger olika text för dem.
fn upprepningsfel(fel: goal::KompletteringFel) -> AppError {
    match fel {
        goal::KompletteringFel::OgiltigRegel(detalj) => {
            AppError::BadRequest(format!("ogiltig upprepningsregel: {detalj}"))
        }
        goal::KompletteringFel::OgiltigLäge(_) => AppError::BadRequest(
            "Ogiltig schemaläggningsstrategi. Använd 'schedule' eller 'push'.".to_string(),
        ),
        goal::KompletteringFel::IngenFörekomst(_) => {
            AppError::BadRequest("Regeln ger inga förekomster.".to_string())
        }
        goal::KompletteringFel::InteÅterkommande => {
            AppError::BadRequest("Målet är inte återkommande.".to_string())
        }
        // Go använder due-datumets meddelande även för startdatumet -
        // klienten får samma text från båda tjänsterna.
        goal::KompletteringFel::OgiltigtStartdatum(_) => {
            AppError::BadRequest("due-datumet måste vara på formatet YYYY-MM-DD".to_string())
        }
        goal::KompletteringFel::IngenSlutförd => {
            AppError::BadRequest("Målet har ingen slutförd förekomst".to_string())
        }
        goal::KompletteringFel::Databas(err) => {
            AppError::Internal(format!("failed to update recurrence: {err}"))
        }
    }
}

/// Detaljerna i ett upprepningsfel.
///
/// Go skriver ut felets detalj - i Rust ligger den i varianten.
fn detalj_av_fel(fel: &plan_recurrence::Fel) -> String {
    match fel {
        plan_recurrence::Fel::OgiltigRegel(detalj) => detalj.clone(),
        plan_recurrence::Fel::IngenFörekomst(detalj) => detalj.clone(),
        plan_recurrence::Fel::OgiltigLäge(läge) => format!("{läge:?}"),
    }
}

/// Body för att slutföra en förekomst.
#[derive(Debug, Default, Deserialize, utoipa::ToSchema)]
pub struct CompleteOccurrenceRequest {
    /// Vilken förekomst som slutförs ("YYYY-MM-DD").
    #[serde(default)]
    pub occurrence_date: String,
    /// När slutförandet skedde ("YYYY-MM-DD").
    #[serde(default)]
    pub completed_on: String,
}

/// Slutför den aktuella förekomsten av ett återkommande mål.
#[utoipa::path(
    post,
    path = "/api/v1/goals/{id}/completions",
    tag = "goals",
    params(("id" = i64, Path, description = "Målets id")),
    request_body = CompleteOccurrenceRequest,
    responses(
        (status = 200, description = "Målet efter förekomsten slutförts", body = Goal),
        (status = 400, description = "Ogiltigt id, datum eller målet är inte återkommande"),
        (status = 403, description = "Ingen åtkomst till målet")
    )
)]
pub async fn complete_goal_occurrence(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
    body: Bytes,
) -> Result<Json<Goal>, AppError> {
    let user_id = user_id(&caller)?;
    let goal_id = goal_id(&id)?;

    let Some(pool) = state.pool.as_ref() else {
        return Err(no_database());
    };

    if !goal::user_has_access(pool, user_id, goal_id).await? {
        return Err(AppError::Forbidden("Access denied".to_string()));
    }

    // En trasig body ger samma fel som i Go - fältens format är
    // valfritt, så standarden är tom.
    let req: CompleteOccurrenceRequest = serde_json::from_slice(&body).unwrap_or_default();

    // Utan uppgift slutförs målets aktuella förekomst.
    let förekomst = if req.occurrence_date.is_empty() {
        None
    } else {
        Some(
            chrono::NaiveDate::parse_from_str(&req.occurrence_date, "%Y-%m-%d").map_err(|_| {
                AppError::BadRequest(
                    "occurrence_date måste vara på formatet YYYY-MM-DD".to_string(),
                )
            })?,
        )
    };

    // Utan uppgift används dagens datum.
    let slutförd_på = if req.completed_on.is_empty() {
        chrono::Utc::now().date_naive()
    } else {
        chrono::NaiveDate::parse_from_str(&req.completed_on, "%Y-%m-%d").map_err(|_| {
            AppError::BadRequest("completed_on måste vara på formatet YYYY-MM-DD".to_string())
        })?
    };

    let mål = goal::complete_occurrence(pool, goal_id, förekomst, slutförd_på)
        .await
        .map_err(|fel| match fel {
            goal::KompletteringFel::InteÅterkommande => {
                AppError::BadRequest("Målet är inte återkommande".to_string())
            }
            goal::KompletteringFel::IngenFörekomst(_) => {
                AppError::BadRequest("Målet har ingen aktiv förekomst".to_string())
            }
            goal::KompletteringFel::IngenSlutförd => {
                AppError::BadRequest("Målet har ingen slutförd förekomst".to_string())
            }
            goal::KompletteringFel::OgiltigRegel(detalj) => {
                AppError::Internal(format!("ogiltig upprepningsregel: {detalj}"))
            }
            goal::KompletteringFel::OgiltigLäge(läge) => {
                AppError::Internal(format!("ogiltig schemaläggningsstrategi: {läge}"))
            }
            goal::KompletteringFel::OgiltigtStartdatum(värde) => {
                AppError::Internal(format!("ogiltigt startdatum: {värde}"))
            }
            goal::KompletteringFel::Databas(err) => {
                AppError::Internal(format!("failed to complete occurrence: {err}"))
            }
        })?;

    let Some(mål) = mål else {
        return Err(AppError::NotFound("Goal not found".to_string()));
    };

    Ok(Json(mål))
}

/// Ångra den senast slutförda förekomsten.
#[utoipa::path(
    delete,
    path = "/api/v1/goals/{id}/completions",
    tag = "goals",
    params(("id" = i64, Path, description = "Målets id")),
    responses(
        (status = 200, description = "Målet efter att förekomsten ångrats", body = Goal),
        (status = 400, description = "Ogiltigt id eller målet är inte återkommande"),
        (status = 403, description = "Ingen åtkomst till målet")
    )
)]
pub async fn undo_goal_occurrence(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Path(id): Path<String>,
) -> Result<Json<Goal>, AppError> {
    let user_id = user_id(&caller)?;
    let goal_id = goal_id(&id)?;

    let Some(pool) = state.pool.as_ref() else {
        return Err(no_database());
    };

    if !goal::user_has_access(pool, user_id, goal_id).await? {
        return Err(AppError::Forbidden("Access denied".to_string()));
    }

    let mål = goal::undo_occurrence(pool, goal_id)
        .await
        .map_err(|fel| match fel {
            goal::KompletteringFel::InteÅterkommande => {
                AppError::BadRequest("Målet är inte återkommande".to_string())
            }
            goal::KompletteringFel::IngenSlutförd => {
                AppError::BadRequest("Målet har ingen slutförd förekomst".to_string())
            }
            goal::KompletteringFel::IngenFörekomst(_) => {
                AppError::BadRequest("Målet har ingen aktiv förekomst".to_string())
            }
            goal::KompletteringFel::OgiltigRegel(detalj) => {
                AppError::Internal(format!("ogiltig upprepningsregel: {detalj}"))
            }
            goal::KompletteringFel::OgiltigLäge(läge) => {
                AppError::Internal(format!("ogiltig schemaläggningsstrategi: {läge}"))
            }
            goal::KompletteringFel::OgiltigtStartdatum(värde) => {
                AppError::Internal(format!("ogiltigt startdatum: {värde}"))
            }
            goal::KompletteringFel::Databas(err) => {
                AppError::Internal(format!("failed to undo occurrence: {err}"))
            }
        })?;

    let Some(mål) = mål else {
        return Err(AppError::NotFound("Goal not found".to_string()));
    };

    Ok(Json(mål))
}
