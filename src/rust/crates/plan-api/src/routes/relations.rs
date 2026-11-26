//! Relationer, beroenden och omordning.
//!
//! Port av relationsektionerna i `src/api/internal/handlers/goals.go`.
//! Felmeddelandena är desamma som i Go - gränssnittet visar dem direkt.

use axum::body::Bytes;
use axum::extract::State;
use axum::Extension;
use axum::Json;
use serde_json::{json, Value};

use crate::auth::Caller;
use crate::error::AppError;
use crate::models::goal::{
    ChildRelation, CreateDependencyRequest, CreateRelationRequest, DeleteDependencyRequest,
    DeleteRelationRequest, GoalDependency, ReorderGoalsRequest, ReorderPriorityGoalsRequest,
    ReorderRelationsRequest, UpdateRelationRequest, UpdateRelationWeightsRequest,
};
use crate::repository::goal;
use crate::routes::state::AppState;

/// Användarens id ur anropet.
fn user_id(caller: &Caller) -> Result<i64, AppError> {
    if caller.user_id == 0 {
        return Err(AppError::Unauthorized("Unauthorized".to_string()));
    }
    Ok(caller.user_id)
}

/// Databasen, eller ett fel som säger att den saknas.
fn pool(state: &AppState) -> Result<&sqlx::PgPool, AppError> {
    state.pool.as_ref().ok_or_else(|| {
        AppError::ServiceUnavailable(
            "Database not connected. Set DATABASE_URL environment variable.".to_string(),
        )
    })
}

/// Skapar en relation mellan ett föräldermål och ett barnmål.
#[utoipa::path(
    post,
    path = "/api/v1/goals/relations",
    tag = "relations",
    request_body = CreateRelationRequest,
    responses(
        (status = 200, description = "Relationen", body = ChildRelation),
        (status = 400, description = "childId och parentId saknas"),
        (status = 401, description = "Ingen giltig session"),
        (status = 403, description = "Access denied")
    )
)]
pub async fn create_relation(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    body: Bytes,
) -> Result<Json<ChildRelation>, AppError> {
    let req: CreateRelationRequest =
        crate::routes::parse_body(&body, "childId and parentId are required")?;

    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;

    // Go binder "required" - ett nollvärde är alltså ett fel.
    if req.child_id == 0 || req.parent_id == 0 {
        return Err(AppError::BadRequest(
            "childId and parentId are required".to_string(),
        ));
    }

    let har_barn = goal::user_has_access(pool, user_id, req.child_id).await?;
    let har_förälder = goal::user_has_access(pool, user_id, req.parent_id).await?;
    if !har_barn || !har_förälder {
        return Err(AppError::Forbidden("Access denied".to_string()));
    }

    let relation =
        goal::create_relation(pool, req.parent_id, req.child_id, req.order, req.weight).await?;

    Ok(Json(relation))
}

/// Uppdaterar en relations ordning och vikt.
#[utoipa::path(
    patch,
    path = "/api/v1/goals/relations",
    tag = "relations",
    request_body = UpdateRelationRequest,
    responses(
        (status = 200, description = "Uppdaterad"),
        (status = 400, description = "childId och parentId saknas"),
        (status = 401, description = "Ingen giltig session"),
        (status = 403, description = "Access denied")
    )
)]
pub async fn update_relation(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    body: Bytes,
) -> Result<Json<Value>, AppError> {
    let req: UpdateRelationRequest =
        crate::routes::parse_body(&body, "childId and parentId are required")?;

    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;

    if req.child_id == 0 || req.parent_id == 0 {
        return Err(AppError::BadRequest(
            "childId and parentId are required".to_string(),
        ));
    }

    let har_tillgång = goal::user_has_access(pool, user_id, req.parent_id).await?;
    if !har_tillgång {
        return Err(AppError::Forbidden("Access denied".to_string()));
    }

    goal::update_relation(pool, req.parent_id, req.child_id, req.order, req.weight).await?;

    Ok(Json(json!({ "updated": true })))
}

/// Ersätter alla vikter för en förälder.
///
/// Vikterna måste vara unika och summa till 100.
#[utoipa::path(
    patch,
    path = "/api/v1/goals/relations/weights",
    tag = "relations",
    request_body = UpdateRelationWeightsRequest,
    responses(
        (status = 200, description = "Uppdaterad"),
        (status = 400, description = "Vikterna är ogiltiga"),
        (status = 401, description = "Ingen giltig session"),
        (status = 403, description = "Access denied")
    )
)]
pub async fn update_relation_weights(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    body: Bytes,
) -> Result<Json<Value>, AppError> {
    let req: UpdateRelationWeightsRequest =
        crate::routes::parse_body(&body, "parentId and weights are required")?;

    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;

    if req.parent_id == 0 || req.weights.is_empty() {
        return Err(AppError::BadRequest(
            "parentId and weights are required".to_string(),
        ));
    }

    let har_tillgång = goal::user_has_access(pool, user_id, req.parent_id).await?;
    if !har_tillgång {
        return Err(AppError::Forbidden("Access denied".to_string()));
    }

    let mut sedda: Vec<i64> = Vec::new();
    let mut total = 0;
    for vikt in &req.weights {
        if sedda.contains(&vikt.child_id) || vikt.weight < 0 || vikt.weight > 100 {
            return Err(AppError::BadRequest(
                "weights must be unique and between 0 and 100".to_string(),
            ));
        }
        sedda.push(vikt.child_id);
        total += vikt.weight;
    }

    if total != 100 {
        return Err(AppError::BadRequest("weights must total 100".to_string()));
    }

    let vikter: Vec<(i64, i32)> = req
        .weights
        .iter()
        .map(|vikt| (vikt.child_id, vikt.weight))
        .collect();

    goal::update_relation_weights(pool, req.parent_id, &vikter).await?;

    Ok(Json(json!({ "updated": true })))
}

/// Tar bort en relation.
#[utoipa::path(
    delete,
    path = "/api/v1/goals/relations",
    tag = "relations",
    request_body = DeleteRelationRequest,
    responses(
        (status = 200, description = "Borttagen"),
        (status = 400, description = "childId och parentId saknas"),
        (status = 401, description = "Ingen giltig session"),
        (status = 403, description = "Access denied")
    )
)]
pub async fn delete_relation(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    body: Bytes,
) -> Result<Json<Value>, AppError> {
    let req: DeleteRelationRequest =
        crate::routes::parse_body(&body, "childId and parentId are required")?;

    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;

    if req.child_id == 0 || req.parent_id == 0 {
        return Err(AppError::BadRequest(
            "childId and parentId are required".to_string(),
        ));
    }

    let har_tillgång = goal::user_has_access(pool, user_id, req.parent_id).await?;
    if !har_tillgång {
        return Err(AppError::Forbidden("Access denied".to_string()));
    }

    goal::delete_relation(pool, req.parent_id, req.child_id).await?;

    Ok(Json(json!({ "deleted": true })))
}

/// Ordnar om barnen under ett mål.
#[utoipa::path(
    patch,
    path = "/api/v1/goals/relations/reorder",
    tag = "relations",
    request_body = ReorderRelationsRequest,
    responses(
        (status = 200, description = "Omordnad"),
        (status = 400, description = "parentId och childIds saknas"),
        (status = 401, description = "Ingen giltig session"),
        (status = 403, description = "Access denied")
    )
)]
pub async fn reorder_relations(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    body: Bytes,
) -> Result<Json<Value>, AppError> {
    let req: ReorderRelationsRequest =
        crate::routes::parse_body(&body, "parentId and childIds are required")?;

    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;

    if req.parent_id == 0 {
        return Err(AppError::BadRequest(
            "parentId and childIds are required".to_string(),
        ));
    }

    if req.child_ids.is_empty() {
        return Err(AppError::BadRequest(
            "childIds must contain at least one child".to_string(),
        ));
    }

    let har_tillgång = goal::user_has_access(pool, user_id, req.parent_id).await?;
    if !har_tillgång {
        return Err(AppError::Forbidden("Access denied".to_string()));
    }

    goal::reorder_relations(pool, req.parent_id, &req.child_ids).await?;

    Ok(Json(json!({ "updated": true })))
}

/// Ordnar om grundmålen.
#[utoipa::path(
    patch,
    path = "/api/v1/goals/reorder",
    tag = "goals",
    request_body = ReorderGoalsRequest,
    responses(
        (status = 200, description = "Omordnad"),
        (status = 400, description = "goalIds saknas"),
        (status = 401, description = "Ingen giltig session")
    )
)]
pub async fn reorder_goals(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    body: Bytes,
) -> Result<Json<Value>, AppError> {
    let req: ReorderGoalsRequest = crate::routes::parse_body(&body, "goalIds is required")?;

    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;

    if req.goal_ids.is_empty() {
        return Err(AppError::BadRequest(
            "goalIds must contain at least one goal".to_string(),
        ));
    }

    goal::reorder_root_goals(pool, user_id, &req.goal_ids).await?;

    Ok(Json(json!({ "updated": true })))
}

/// Ordnar om prioriteringslistan.
#[utoipa::path(
    patch,
    path = "/api/v1/goals/priority/reorder",
    tag = "goals",
    request_body = ReorderPriorityGoalsRequest,
    responses(
        (status = 200, description = "Omordnad"),
        (status = 400, description = "goalIds saknas"),
        (status = 401, description = "Ingen giltig session")
    )
)]
pub async fn reorder_priority_goals(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    body: Bytes,
) -> Result<Json<Value>, AppError> {
    // Go svarar med samma meddelande både när bindningen misslyckas och
    // när listan är tom.
    let req: ReorderPriorityGoalsRequest =
        crate::routes::parse_body(&body, "goalIds must contain at least one goal")?;

    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;

    if req.goal_ids.is_empty() {
        return Err(AppError::BadRequest(
            "goalIds must contain at least one goal".to_string(),
        ));
    }

    goal::reorder_priority_goals(pool, user_id, &req.goal_ids).await?;

    Ok(Json(json!({ "updated": true })))
}

/// Skapar ett beroende mellan två mål.
///
/// Ett beroende som bildar en cykel avvisas - annars skulle målen aldrig
/// bli klara.
#[utoipa::path(
    post,
    path = "/api/v1/goals/dependencies",
    tag = "dependencies",
    request_body = CreateDependencyRequest,
    responses(
        (status = 200, description = "Beroendet", body = GoalDependency),
        (status = 400, description = "Ogiltig request eller cykel"),
        (status = 401, description = "Ingen giltig session"),
        (status = 403, description = "Access denied")
    )
)]
pub async fn create_dependency(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    body: Bytes,
) -> Result<Json<GoalDependency>, AppError> {
    let req: CreateDependencyRequest =
        crate::routes::parse_body(&body, "goalId and dependsOnId are required")?;

    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;

    if req.goal_id == 0 || req.depends_on_id == 0 {
        return Err(AppError::BadRequest(
            "goalId and dependsOnId are required".to_string(),
        ));
    }

    if req.goal_id == req.depends_on_id {
        return Err(AppError::BadRequest(
            "A goal cannot depend on itself".to_string(),
        ));
    }

    let har_målet = goal::user_has_access(pool, user_id, req.goal_id).await?;
    let har_beroendet = goal::user_has_access(pool, user_id, req.depends_on_id).await?;
    if !har_målet || !har_beroendet {
        return Err(AppError::Forbidden("Access denied".to_string()));
    }

    let (har_cykel, väg) = goal::check_for_cycle(pool, req.goal_id, req.depends_on_id).await?;
    if har_cykel {
        return Err(AppError::BadRequest(format!(
            "Cykel upptäckt: Detta skulle skapa en cykel genom: {}",
            format_path(&väg)
        )));
    }

    let beroende = goal::create_dependency(pool, req.goal_id, req.depends_on_id).await?;

    Ok(Json(beroende))
}

/// Tar bort ett beroende.
#[utoipa::path(
    delete,
    path = "/api/v1/goals/dependencies",
    tag = "dependencies",
    request_body = DeleteDependencyRequest,
    responses(
        (status = 200, description = "Borttaget"),
        (status = 400, description = "goalId och dependsOnId saknas"),
        (status = 401, description = "Ingen giltig session"),
        (status = 403, description = "Access denied")
    )
)]
pub async fn delete_dependency(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    body: Bytes,
) -> Result<Json<Value>, AppError> {
    let req: DeleteDependencyRequest =
        crate::routes::parse_body(&body, "goalId and dependsOnId are required")?;

    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;

    if req.goal_id == 0 || req.depends_on_id == 0 {
        return Err(AppError::BadRequest(
            "goalId and dependsOnId are required".to_string(),
        ));
    }

    let har_tillgång = goal::user_has_access(pool, user_id, req.goal_id).await?;
    if !har_tillgång {
        return Err(AppError::Forbidden("Access denied".to_string()));
    }

    goal::delete_dependency(pool, req.goal_id, req.depends_on_id).await?;

    Ok(Json(json!({ "deleted": true })))
}

/// Formaterar en sökväg som "1 → 2 → 3".
fn format_path(path: &[i64]) -> String {
    path.iter()
        .map(|id| id.to_string())
        .collect::<Vec<String>>()
        .join(" → ")
}
