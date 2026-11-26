//! Mål.
//!
//! Port av `src/api/internal/models/goal.go`. Fältnamn och JSON-form är
//! identiska med Go-tjänstens - gränssnittet och agenten läser samma svar.

use serde::{Deserialize, Serialize};
use utoipa::ToSchema;

use crate::models::place::Place;

/// Ett mål i systemet.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, ToSchema)]
pub struct Goal {
    /// Målets id.
    pub id: i64,
    /// Målets titel.
    pub title: String,
    /// Längre beskrivning av målet.
    #[serde(default)]
    pub description: String,
    /// Ikonen som visas.
    pub icon: String,
    /// Statusens id.
    pub status_id: i32,
    /// När målet skapades.
    pub created: chrono::DateTime<chrono::Utc>,
    /// När arbetet började - null om det inte börjat.
    pub started: Option<chrono::DateTime<chrono::Utc>>,
    /// När målet slutfördes - null om det inte är klart.
    pub finished: Option<chrono::DateTime<chrono::Utc>>,
    /// Datumet målet ska vara klart - null om det saknas.
    pub due: Option<chrono::NaiveDate>,
    /// Datumet målet tidigast får dyka upp i prioriteringslistan.
    pub start_date: Option<chrono::NaiveDate>,
    /// Uppskattad tidsåtgång i minuter.
    pub estimated_minutes: Option<i32>,
    /// Om målet ligger i inboxen.
    pub inbox: i32,
    /// Om målet är evigt - det kan då inte slutföras.
    pub eternal: bool,
    /// Om målet är avsett att arbeta på av en agent.
    pub agent: bool,
    /// Om målet ska göras nu.
    pub now: bool,
    /// Om målet har en förälder.
    pub has_parent: bool,
    /// Målets placering bland grundmål.
    pub order: i32,
    /// Platser där målet kan slutföras - skickas alltid ut, även tom.
    pub places: Vec<Place>,
    /// Taggar på målet - skickas alltid ut, även tom.
    pub tags: Vec<String>,
    /// Upprepningsregeln som RRULE-sträng - null om målet inte är
    /// återkommande.
    pub recurrence_rule: Option<String>,
    /// Schemaläggningsstrategin - "schedule" eller "push".
    pub recurrence_mode: Option<String>,
    /// Datumet för nästa förekomst.
    pub recurrence_next: Option<chrono::NaiveDate>,
    /// Antalet slutförda förekomster.
    pub recurrence_count: i32,
}

/// Ett fält som antingen sätts, rensas eller lämnas oförändrat.
///
/// Tre tillstånd behövs, och en vanlig `Option` räcker inte - JSON `null`
/// och ett utelämnat fält är samma sak för en pekare:
///
/// * fältet saknas i request-bodyn -> lämnas oförändrat,
/// * fältet finns med ett värde -> sätts,
/// * fältet finns som null -> rensas.
#[derive(Debug, Clone, Default, PartialEq)]
pub struct Nullable<T> {
    /// Finns fältet i requesten?
    pub set: bool,
    /// Värdet - `None` betyder "null" i requesten.
    pub value: Option<T>,
}

impl<T> Nullable<T> {
    /// Finns fältet, och har det ett värde?
    pub fn is_some(&self) -> bool {
        self.set && self.value.is_some()
    }
}

impl<T> utoipa::PartialSchema for Nullable<T>
where
    T: utoipa::PartialSchema,
{
    fn schema() -> utoipa::openapi::RefOr<utoipa::openapi::schema::Schema> {
        T::schema()
    }
}

impl<T> utoipa::ToSchema for Nullable<T> where T: utoipa::ToSchema + utoipa::PartialSchema {}

impl<'de, T> Deserialize<'de> for Nullable<T>
where
    T: Deserialize<'de>,
{
    fn deserialize<D>(deserializer: D) -> Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        // Fältet finns - det är bara värdet som kan vara null.
        Ok(Self {
            set: true,
            value: Option::deserialize(deserializer)?,
        })
    }
}

/// Filter för mållistan.
///
/// Fält som inte sätts filtrerar inte alls. Det gör att `agent=false` kan
/// skiljas från "ingen filtrering".
#[derive(Debug, Clone, Default)]
pub struct GoalFilter {
    /// Filtrerar på agent-flaggan.
    pub agent: Option<bool>,
    /// Filtrerar på status.
    pub status_id: Option<i32>,
}

/// Request-body för att skapa ett mål.
#[derive(Debug, Clone, Deserialize, ToSchema)]
pub struct CreateGoalRequest {
    /// Målets titel - krävs.
    pub title: String,
    /// Ikonen som visas.
    #[serde(default)]
    pub icon: String,
    /// Statusens id.
    #[serde(default)]
    pub status_id: i32,
    /// Upprepningsregeln - gör målet återkommande.
    #[serde(default)]
    pub recurrence_rule: Option<String>,
    /// Schemaläggningsstrategin - "schedule" eller "push".
    #[serde(default)]
    pub recurrence_mode: Option<String>,
    /// Datumet målet tidigast får dyka upp i prioriteringslistan.
    #[serde(default)]
    pub start_date: Option<String>,
    /// Datumet serien startar på.
    #[serde(default)]
    pub recurrence_start_date: Option<String>,
}

/// Request-body för att uppdatera ett mål.
///
/// Fält som sätts till `null` rensas, fält som saknas lämnas oförändrade.
/// Det är därför de är `Option<Option<...>>` - en vanlig `Option` rader
/// inte skillnaden mellan "saknas" och "null".
#[derive(Debug, Clone, Default, Deserialize, ToSchema)]
pub struct UpdateGoalRequest {
    /// Ny titel.
    #[serde(default)]
    pub title: Option<String>,
    /// Ny beskrivning.
    #[serde(default)]
    pub description: Option<String>,
    /// Ny ikon.
    #[serde(default)]
    pub icon: Option<String>,
    /// Ny status.
    #[serde(default)]
    pub status_id: Option<i32>,
    /// När arbetet började - null rensar tidsstämpeln.
    #[serde(default)]
    #[schema(value_type = Option<String>)]
    pub started: Nullable<String>,
    /// När målet slutfördes - null rensar tidsstämpeln.
    #[serde(default)]
    #[schema(value_type = Option<String>)]
    pub finished: Nullable<String>,
    /// Nytt due-datum - null rensar datumet.
    #[serde(default)]
    #[schema(value_type = Option<String>)]
    pub due: Nullable<String>,
    /// Nytt startdatum - null rensar datumet.
    #[serde(default)]
    #[schema(value_type = Option<String>)]
    pub start_date: Nullable<String>,
    /// Ny uppskattad tidsåtgång - null rensar den.
    #[serde(default)]
    #[schema(value_type = Option<i32>)]
    pub estimated_minutes: Nullable<i32>,
    /// Sätter evigt-flaggan.
    #[serde(default)]
    pub eternal: Option<bool>,
    /// Sätter agent-flaggan.
    #[serde(default)]
    pub agent: Option<bool>,
    /// Sätter nu-flaggan.
    #[serde(default)]
    pub now: Option<bool>,
    /// Ny upprepningsregel - null tar bort den.
    #[serde(default)]
    #[schema(value_type = Option<String>)]
    pub recurrence_rule: Nullable<String>,
    /// Ny schemaläggningsstrategi.
    #[serde(default)]
    #[schema(value_type = Option<String>)]
    pub recurrence_mode: Nullable<String>,
    /// Nytt startdatum för serien.
    #[serde(default)]
    pub recurrence_start_date: Option<String>,
}

impl UpdateGoalRequest {
    /// Försöker requesten markera målet som klart?
    ///
    /// Att rensa `finished` (null) räknas inte som att slutföra målet.
    pub fn finishes_goal(&self) -> bool {
        matches!(&self.finished.value, Some(värde) if !värde.is_empty())
    }
}

/// En relation mellan ett barnmål och ett föräldermål.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, ToSchema)]
pub struct ChildRelation {
    /// Barnmålets id.
    pub child_id: i64,
    /// Barnets placering under föräldern.
    pub order: i32,
    /// Barnets andel av förälderns vikt.
    pub weight: i32,
}

/// En relation mellan mål.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, ToSchema)]
pub struct GoalRelation {
    /// Föräldermålets id.
    pub parent_id: i64,
    /// Barnmålets id.
    pub child_id: i64,
    /// Barnets placering under föräldern.
    pub order: i32,
    /// Barnets andel av förälderns vikt.
    pub weight: i32,
}

/// Ett beroende mellan två mål.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, ToSchema)]
pub struct GoalDependency {
    /// Radens id.
    pub id: i64,
    /// Målet som beror.
    pub goal_id: i64,
    /// Målet som det beror på.
    pub depends_on_id: i64,
    /// När beroendet skapades.
    pub created: chrono::DateTime<chrono::Utc>,
}

/// En förälder till ett mål.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, ToSchema)]
pub struct GoalParent {
    /// Föräldermålets id.
    pub id: i64,
    /// Föräldermålets titel.
    pub title: String,
}

/// Ett mål med sina relationer.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, ToSchema)]
pub struct GoalWithRelations {
    /// Målet.
    #[serde(flatten)]
    pub goal: Goal,
    /// Relationerna till barnen.
    #[serde(rename = "childRelations")]
    pub child_relations: Vec<ChildRelation>,
    /// Föräldrarna.
    #[serde(rename = "parentRelations")]
    pub parent_relations: Vec<Goal>,
}

/// Ett barnmål med beroendeinformation.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, ToSchema)]
pub struct ChildWithDeps {
    /// Målet.
    #[serde(flatten)]
    pub goal: Goal,
    /// Placeringen under föräldern.
    pub order: i32,
    /// Vikten under föräldern.
    pub weight: i32,
    /// Målen som barnet beror på.
    #[serde(rename = "dependsOn")]
    pub depends_on: Vec<Goal>,
    /// Målen som blockerar barnet.
    pub blocking: Vec<Goal>,
}

/// Svaret för ett enskilt mål.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, ToSchema)]
pub struct GetGoalResponse {
    /// Målet med sina relationer.
    pub goal: GoalWithRelations,
    /// Barnen.
    pub children: Vec<ChildWithDeps>,
    /// Föräldrarna.
    pub parents: Vec<Goal>,
    /// Alla användarens mål.
    #[serde(rename = "allGoals")]
    pub all_goals: Vec<Goal>,
    /// Beroendena.
    pub dependencies: Vec<GoalDependency>,
    /// Målen som målet beror på.
    #[serde(rename = "dependsOn")]
    pub depends_on: Vec<Goal>,
    /// Målen som målet blockerar.
    pub blocking: Vec<Goal>,
    /// Statushistoriken - null i Go, eftersom handlaren inte fyller i den.
    #[serde(rename = "statusHistory")]
    pub status_history: Option<Vec<crate::models::status::StatusUpdate>>,
}

/// Ett mål med prioritetsinformation.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, ToSchema)]
pub struct PrioritizedGoal {
    /// Målet.
    #[serde(flatten)]
    pub goal: Goal,
    /// Vikten i prioriteringslistan.
    pub weight: i32,
    /// Titeln på föräldern med högst vikt.
    #[serde(rename = "parentTitle", skip_serializing_if = "Option::is_none")]
    pub parent_title: Option<String>,
    /// Id:t på föräldern med högst vikt.
    #[serde(rename = "parentId", skip_serializing_if = "Option::is_none")]
    pub parent_id: Option<i64>,
    /// Placeringen i prioriteringslistan.
    #[serde(rename = "priorityOrder")]
    pub priority_order: i32,
    /// Alla föräldrar som målet ligger under.
    pub parents: Vec<GoalParent>,
}

/// Request-body för att skapa en relation.
#[derive(Debug, Clone, Deserialize, ToSchema)]
pub struct CreateRelationRequest {
    /// Barnmålets id.
    #[serde(rename = "childId")]
    pub child_id: i64,
    /// Föräldermålets id.
    #[serde(rename = "parentId")]
    pub parent_id: i64,
    /// Placeringen under föräldern.
    #[serde(default)]
    pub order: i32,
    /// Vikten under föräldern.
    #[serde(default)]
    pub weight: i32,
}

/// Request-body för att uppdatera en relation.
#[derive(Debug, Clone, Deserialize, ToSchema)]
pub struct UpdateRelationRequest {
    /// Barnmålets id.
    #[serde(rename = "childId")]
    pub child_id: i64,
    /// Föräldermålets id.
    #[serde(rename = "parentId")]
    pub parent_id: i64,
    /// Ny placering.
    #[serde(default)]
    pub order: Option<i32>,
    /// Ny vikt.
    #[serde(default)]
    pub weight: Option<i32>,
}

/// Ett barns andel av förälderns vikt.
#[derive(Debug, Clone, Deserialize, ToSchema)]
pub struct RelationWeight {
    /// Barnmålets id.
    #[serde(rename = "childId")]
    pub child_id: i64,
    /// Vikten.
    pub weight: i32,
}

/// Request-body för att sätta alla vikter under en förälder.
#[derive(Debug, Clone, Deserialize, ToSchema)]
pub struct UpdateRelationWeightsRequest {
    /// Föräldermålets id.
    #[serde(rename = "parentId")]
    pub parent_id: i64,
    /// Vikterna.
    pub weights: Vec<RelationWeight>,
}

/// Request-body för att ordna om barnen under ett mål.
#[derive(Debug, Clone, Deserialize, ToSchema)]
pub struct ReorderRelationsRequest {
    /// Föräldermålets id.
    #[serde(rename = "parentId")]
    pub parent_id: i64,
    /// Barnen i sin nya ordning.
    #[serde(rename = "childIds")]
    pub child_ids: Vec<i64>,
}

/// Request-body för att ordna om grundmålen.
#[derive(Debug, Clone, Deserialize, ToSchema)]
pub struct ReorderGoalsRequest {
    /// Grundmålen i sin nya ordning.
    #[serde(rename = "goalIds")]
    pub goal_ids: Vec<i64>,
}

/// Request-body för att ordna om prioriteringslistan.
#[derive(Debug, Clone, Deserialize, ToSchema)]
pub struct ReorderPriorityGoalsRequest {
    /// Målen i sin nya ordning.
    #[serde(rename = "goalIds")]
    pub goal_ids: Vec<i64>,
}

/// Request-body för att ta bort en relation.
#[derive(Debug, Clone, Deserialize, ToSchema)]
pub struct DeleteRelationRequest {
    /// Barnmålets id.
    #[serde(rename = "childId")]
    pub child_id: i64,
    /// Föräldermålets id.
    #[serde(rename = "parentId")]
    pub parent_id: i64,
}

/// Request-body för att skapa ett beroende.
#[derive(Debug, Clone, Deserialize, ToSchema)]
pub struct CreateDependencyRequest {
    /// Målet som beror.
    #[serde(rename = "goalId")]
    pub goal_id: i64,
    /// Målet som det beror på.
    #[serde(rename = "dependsOnId")]
    pub depends_on_id: i64,
}

/// Request-body för att ta bort ett beroende.
#[derive(Debug, Clone, Deserialize, ToSchema)]
pub struct DeleteDependencyRequest {
    /// Målet som beror.
    #[serde(rename = "goalId")]
    pub goal_id: i64,
    /// Målet som det beror på.
    #[serde(rename = "dependsOnId")]
    pub depends_on_id: i64,
}

/// Hela målgraphen för en användare.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, ToSchema)]
pub struct GoalGraphResponse {
    /// Användarens mål.
    pub goals: Vec<Goal>,
    /// Relationerna mellan målen.
    pub relations: Vec<GoalRelation>,
    /// Beroendena mellan målen.
    pub dependencies: Vec<GoalDependency>,
}

/// Request-body för att slutföra en förekomst av ett återkommande mål.
#[derive(Debug, Clone, Default, Deserialize, ToSchema)]
pub struct CompleteOccurrenceRequest {
    /// Förekomsten som slutförs ("YYYY-MM-DD"). Utan den slutförs målets
    /// aktuella förekomst.
    #[serde(default)]
    pub occurrence_date: String,
    /// Datumet slutförandet skedde ("YYYY-MM-DD").
    #[serde(default)]
    pub completed_on: String,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn uppdatering_kan_rensa_och_satta_falt() {
        // Fältet saknas -> lämnas oförändrat.
        let req: UpdateGoalRequest = serde_json::from_str("{}").unwrap();
        assert!(req.title.is_none());
        assert!(!req.finished.set);
        assert!(!req.finishes_goal());

        // Fältet sätts -> ändras.
        let req: UpdateGoalRequest = serde_json::from_str(r#"{"title":"Ny"}"#).unwrap();
        assert_eq!(req.title.as_deref(), Some("Ny"));

        // Fältet är null -> rensas.
        let req: UpdateGoalRequest = serde_json::from_str(r#"{"finished":null}"#).unwrap();
        assert!(req.finished.set);
        assert!(req.finished.value.is_none());
        assert!(!req.finishes_goal());

        // Fältet har ett värde -> sätts.
        let req: UpdateGoalRequest =
            serde_json::from_str(r#"{"finished":"2026-01-01T10:00:00Z"}"#).unwrap();
        assert!(req.finishes_goal());

        // Ett nullat datum rensas - det sätter ingenting.
        let req: UpdateGoalRequest = serde_json::from_str(r#"{"due":null}"#).unwrap();
        assert!(req.due.set);
        assert!(req.due.value.is_none());

        // Ett datum sätts.
        let req: UpdateGoalRequest = serde_json::from_str(r#"{"due":"2026-01-05"}"#).unwrap();
        assert_eq!(req.due.value.as_deref(), Some("2026-01-05"));
    }

    #[test]
    fn skapar_mal_kraver_titel() {
        let req: CreateGoalRequest = serde_json::from_str(r#"{"title":"Mål"}"#).unwrap();
        assert_eq!(req.title, "Mål");
        assert_eq!(req.icon, "");
    }
}
