//! Autentiseringstester.
//!
//! Testerna går genom hela kedjan - router, middleware och handler - så
//! att beteendet matchar Go-tjänsten: statuskoder och felmeddelanden.

use std::sync::OnceLock;

use axum::body::Body;
use axum::http::{Request, StatusCode};
use tower::ServiceExt;

use plan_api::config::Config;
use plan_api::routes::router;

/// Miljövariablerna är process-globala, så testerna körs i tur och ordning.
fn lås() -> &'static tokio::sync::Mutex<()> {
    static LÅS: OnceLock<tokio::sync::Mutex<()>> = OnceLock::new();
    LÅS.get_or_init(|| tokio::sync::Mutex::new(()))
}

/// Nollställer variabler som påverkar autentiseringen.
///
/// Testerna ska bete sig oavsett hur utvecklingsmiljön är satt.
fn forbered() {
    std::env::remove_var("AUTH_DISABLED");
}

/// Bygger en router utan databas.
fn app() -> axum::Router {
    router(None, &Config::from_env())
}

/// Skickar ett anrop och ger tillbaka status och body.
async fn anrop(
    app: &axum::Router,
    metod: &str,
    path: &str,
    headers: &[(&str, &str)],
) -> (StatusCode, String) {
    let mut builder = Request::builder().method(metod).uri(path);
    for (namn, värde) in headers {
        builder = builder.header(*namn, *värde);
    }

    let svar = app
        .clone()
        .oneshot(
            builder
                .body(Body::empty())
                .expect("requesten går att bygga"),
        )
        .await
        .expect("servern svarar");

    let status = svar.status();
    let bytes = axum::body::to_bytes(svar.into_body(), 1024 * 64)
        .await
        .expect("body går att läsa");

    (status, String::from_utf8_lossy(&bytes).to_string())
}

#[tokio::test]
async fn utan_inget_tillstand_kraver_man_inte_tilltrad() {
    let _lås = lås().lock().await;
    forbered();
    let app = app();

    let (status, body) = anrop(&app, "GET", "/api/v1/goals", &[]).await;

    assert_eq!(status, StatusCode::UNAUTHORIZED);
    assert_eq!(body, r#"{"message":"Authorization header required"}"#);
}

#[tokio::test]
async fn en_ogiltig_api_nyckel_avvisas() {
    let _lås = lås().lock().await;
    forbered();
    std::env::set_var("AGENT_API_KEY", "agent-nyckel");

    let app = app();
    let (status, body) = anrop(&app, "GET", "/api/v1/goals", &[("X-API-Key", "fel")]).await;

    assert_eq!(status, StatusCode::UNAUTHORIZED);
    assert!(body.contains("Invalid API key"), "body: {body}");
}

#[tokio::test]
async fn agent_nyckeln_far_gora_sina_anrop() {
    let _lås = lås().lock().await;
    forbered();
    std::env::set_var("AGENT_API_KEY", "agent-nyckel");

    let app = app();
    // Ett mål-anrop som agenten faktiskt gör - det når fram till handlern.
    // Utan databas svarar handlern 503, vilket visar att anropet inte
    // stoppades i autentiseringen.
    let (status, _) = anrop(
        &app,
        "GET",
        "/api/v1/goals",
        &[("X-API-Key", "agent-nyckel")],
    )
    .await;

    assert_eq!(status, StatusCode::SERVICE_UNAVAILABLE);
}

#[tokio::test]
async fn agent_nyckeln_stoppas_utanfor_sitt_scope() {
    let _lås = lås().lock().await;
    forbered();
    std::env::set_var("AGENT_API_KEY", "agent-nyckel");

    let app = app();
    let (status, body) = anrop(
        &app,
        "GET",
        "/api/v1/calendars",
        &[("X-API-Key", "agent-nyckel")],
    )
    .await;

    assert_eq!(status, StatusCode::FORBIDDEN);
    assert_eq!(
        body,
        r#"{"error":"Forbidden","message":"Agent API-nyckeln får inte göra det här anropet"}"#
    );
}

#[tokio::test]
async fn agent_nyckeln_far_inte_radera_mal() {
    let _lås = lås().lock().await;
    forbered();
    std::env::set_var("AGENT_API_KEY", "agent-nyckel");

    let app = app();
    let (status, _) = anrop(
        &app,
        "DELETE",
        "/api/v1/goals/1",
        &[("X-API-Key", "agent-nyckel")],
    )
    .await;

    assert_eq!(status, StatusCode::FORBIDDEN);
}

#[tokio::test]
async fn en_okand_vag_svarar_404_precis_som_gin() {
    let _lås = lås().lock().await;
    forbered();
    std::env::set_var("AGENT_API_KEY", "agent-nyckel");

    let app = app();
    let (status, body) = anrop(
        &app,
        "DELETE",
        "/api/v1/goals",
        &[("X-API-Key", "agent-nyckel")],
    )
    .await;

    // Gin svarar "404 page not found" när ingen route matchar. Authen
    // körs inte ens, eftersom den i Go sitter på API-gruppens routes.
    assert_eq!(status, StatusCode::NOT_FOUND);
    assert_eq!(body, "404 page not found");
}

#[tokio::test]
async fn halsokontrollen_kraver_inget_tillstand() {
    let _lås = lås().lock().await;
    forbered();
    let app = app();

    let (status, body) = anrop(&app, "GET", "/health", &[]).await;

    assert_eq!(status, StatusCode::OK);
    assert_eq!(body, r#"{"status":"ok"}"#);
}

#[tokio::test]
async fn inloggningsvagen_ar_publik() {
    let _lås = lås().lock().await;
    forbered();
    let app = app();

    // Autentiseringsendpoints är själva vägen in - de ska inte kräva
    // en session.
    let (status, _) = anrop(&app, "GET", "/api/v1/auth/login", &[]).await;

    assert_ne!(status, StatusCode::UNAUTHORIZED);
}

#[tokio::test]
async fn utvecklingstoken_gor_tillstand() {
    let _lås = lås().lock().await;
    forbered();
    let app = app();

    let (status, _) = anrop(
        &app,
        "GET",
        "/api/v1/goals",
        &[("Authorization", "Bearer user_4")],
    )
    .await;

    assert_eq!(status, StatusCode::SERVICE_UNAVAILABLE);
}

#[tokio::test]
async fn en_oportad_vag_svarar_501() {
    let _lås = lås().lock().await;
    forbered();
    let app = app();

    // Vägar som ännu inte är portade ska ge 501 - inte en tyst 404 som
    // ser ut att fungera. Auth-vägarna är de nästa som portas.
    let (status, _) = anrop(
        &app,
        "GET",
        "/api/v1/auth/session",
        &[("Authorization", "Bearer user_4")],
    )
    .await;

    assert_eq!(status, StatusCode::NOT_IMPLEMENTED);
}

#[tokio::test]
async fn en_ogiltig_utvecklingstoken_avvisas() {
    let _lås = lås().lock().await;
    forbered();
    let app = app();

    let (status, body) = anrop(
        &app,
        "GET",
        "/api/v1/goals",
        &[("Authorization", "Bearer user_abc")],
    )
    .await;

    assert_eq!(status, StatusCode::UNAUTHORIZED);
    assert_eq!(body, r#"{"message":"Invalid user token"}"#);
}
