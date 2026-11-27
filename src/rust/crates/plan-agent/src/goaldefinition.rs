//! Agentens förslag på hur ett mål ska se ut.
//!
//! Go: `internal/goaldefinition`. Paketet bär policyn: agenten får fylla
//! luckor, aldrig skriva över värden som användaren redan angett.
//! Förslaget lämnas till en människa som godkänner det, och det är exakt
//! de godkända värdena som sedan appliceras - inte en ny generering.

use serde::{Deserialize, Serialize};

use crate::definer::Definition;
use crate::planclient::models::{Goal, Place, UpdateGoalRequest};
use crate::planclient::PlanClient;

/// Godkännandets typ för en måldefinition.
pub const KIND: &str = "goal_definition";

/// De exakta värden som föreslås.
///
/// Ett fält som inte finns här lämnas orört. `Option` gör att "sätt
/// värdet" går att skilja från "lämna oförändrat".
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct Values {
    /// Föreslagen titel.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub title: Option<String>,
    /// Föreslagen beskrivning.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub description: Option<String>,
    /// Föreslagen ikon.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub icon: Option<String>,
    /// Om målet ska göras nu.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub now: Option<bool>,
    /// Om målet är evigt.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub eternal: Option<bool>,
    /// Om målet är ett agentmål.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub agent: Option<bool>,
    /// Föreslaget slutdatum.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub due: Option<String>,
    /// Föreslaget startdatum.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub start_date: Option<String>,
    /// Den uppskattade tidsåtgången, i minuter.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub estimated_minutes: Option<i32>,
    /// Föreslagen upprepningsregel.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub recurrence_rule: Option<String>,
    /// Föreslagen schemaläggningsstrategi.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub recurrence_mode: Option<String>,
    /// Seriens startdatum.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub recurrence_start_date: Option<String>,
    /// Målets fulla tagglista efter ändringen.
    ///
    /// API:t ersätter hela uppsättningen, så listan måste innehålla de
    /// taggar målet redan har.
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub tags: Vec<String>,
    /// Målets fulla platslista efter ändringen.
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub place_ids: Vec<i64>,
    /// Fasen målet flyttas till.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub status: Option<String>,
}

/// En ändring som föreslås - den visas för människan.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct Field {
    /// Fältets namn, så att det går att hitta det i värdena.
    #[serde(rename = "field")]
    pub key: String,
    /// Fältets namn på svenska.
    pub label: String,
    /// Målets nuvarande värde. Tomt = inget värde.
    pub from: String,
    /// Det föreslagna värdet.
    pub to: String,
}

/// Agentens förslag på hur ett mål ska se ut.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct Proposal {
    /// Godkännandets typ.
    pub kind: String,
    /// Målet förslaget gäller.
    pub goal_id: i64,
    /// Ändringarna som föreslås, i läsbar form.
    pub fields: Vec<Field>,
    /// De exakta värdena som appliceras om förslaget godkänns.
    pub values: Values,
    /// Antaganden agenten gjort.
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub notes: Vec<String>,
}

/// Sammanhanget ett förslag byggs i.
pub struct Context {
    /// Målet som ska definieras.
    pub goal: Goal,
    /// Agentens förslag.
    pub definition: Definition,
    /// Användarens platser. Agenten föreslår bara platser som finns
    /// bland dem.
    pub places: Vec<Place>,
    /// Målets nuvarande fas.
    pub current_phase: String,
    /// Fasen målet flyttas till när definitionen är godkänd.
    pub next_phase: String,
}

impl Proposal {
    /// Lägger till en ändring i förslaget.
    ///
    /// Go: `Proposal.add`.
    fn lägg_till(&mut self, key: &str, label: &str, from: &str, to: &str) {
        self.fields.push(Field {
            key: key.to_string(),
            label: label.to_string(),
            from: from.to_string(),
            to: to.to_string(),
        });
    }

    /// Lägger till en notering om något som hoppades över.
    ///
    /// Go: `Proposal.note`.
    fn notera(&mut self, text: &str) {
        if text.is_empty() {
            return;
        }

        self.notes.push(text.to_string());
    }
}

/// Bygger förslaget utifrån målet och agentens definition.
///
/// Policyn är att agenten fyller luckor. Ett fält som redan har ett värde
/// lämnas orört - även om agenten föreslår något annat.
///
/// Go: `Build`.
pub fn bygg(ctx: &Context) -> Proposal {
    let mut proposal = Proposal {
        kind: KIND.to_string(),
        goal_id: ctx.goal.id,
        fields: Vec::new(),
        values: Values::default(),
        notes: ctx.definition.notes.clone(),
    };

    let definition = &ctx.definition;

    // Titel och beskrivning är definitionens kärna - de föreslås alltid,
    // men bara om de faktiskt ändrar något.
    if !definition.title.is_empty() && definition.title != ctx.goal.title {
        proposal.values.title = Some(definition.title.clone());
        proposal.lägg_till("title", "Titel", &ctx.goal.title, &definition.title.clone());
    }

    if !definition.description.is_empty() && definition.description != ctx.goal.description {
        proposal.values.description = Some(definition.description.clone());
        proposal.lägg_till(
            "description",
            "Beskrivning",
            &ctx.goal.description,
            &definition.description.clone(),
        );
    }

    // Ikon - bara om målet saknar en.
    if ctx.goal.icon.is_empty() && !definition.icon.is_empty() {
        if giltig_ikon(&definition.icon) {
            proposal.values.icon = Some(definition.icon.clone());
            proposal.lägg_till("icon", "Ikon", "", &definition.icon.clone());
        } else {
            proposal.notera(&format!(
                "ikonen {:?} är inte ett giltigt ikonnamn och hoppades över",
                definition.icon
            ));
        }
    }

    // Operativa flaggor - bara sätts på, aldrig av.
    if !ctx.goal.now && definition.now {
        proposal.values.now = Some(true);
        proposal.lägg_till("now", "Gör nu", "nej", "ja");
    }

    if !ctx.goal.eternal && definition.eternal {
        proposal.values.eternal = Some(true);
        proposal.lägg_till("eternal", "Evigt mål", "nej", "ja");
    }

    if !ctx.goal.agent && definition.agent {
        proposal.values.agent = Some(true);
        proposal.lägg_till("agent", "Agentmål", "nej", "ja");
    }

    // Datum - bara om målet saknar dem.
    if ctx.goal.due.is_none() && !definition.due.is_empty() {
        match giltigt_datum(&definition.due) {
            Some(datum) => {
                proposal.values.due = Some(datum.clone());
                proposal.lägg_till("due", "Slutdatum", "", &datum);
            }
            None => proposal.notera(&format!(
                "slutdatumet {:?} är inte ett giltigt datum (YYYY-MM-DD) och hoppades över",
                definition.due
            )),
        }
    }

    if ctx.goal.start_date.is_none() && !definition.start_date.is_empty() {
        match giltigt_datum(&definition.start_date) {
            Some(datum) => {
                proposal.values.start_date = Some(datum.clone());
                proposal.lägg_till("start_date", "Startdatum", "", &datum);
            }
            None => proposal.notera(&format!(
                "startdatumet {:?} är inte ett giltigt datum (YYYY-MM-DD) och hoppades över",
                definition.start_date
            )),
        }
    }

    // Uppskattad tid - bara om målet saknar en.
    if ctx.goal.estimated_minutes.is_none() && definition.estimated_minutes > 0 {
        let värde = definition.estimated_minutes;
        proposal.values.estimated_minutes = Some(värde);
        proposal.lägg_till(
            "estimated_minutes",
            "Uppskattad tid",
            "",
            &format!("{värde} min"),
        );
    }

    // Upprepning - bara om målet saknar en regel.
    if ctx.goal.recurrence_rule.is_none() && !definition.recurrence_rule.is_empty() {
        if let Err(fel) = validera_regel(&definition.recurrence_rule) {
            proposal.notera(&format!(
                "upprepningsregeln {:?} är inte giltig och hoppades över",
                definition.recurrence_rule
            ));
            let _ = fel;
        } else {
            proposal.values.recurrence_rule = Some(definition.recurrence_rule.clone());
            proposal.lägg_till(
                "recurrence_rule",
                "Upprepning",
                "",
                &definition.recurrence_rule.clone(),
            );

            let mut läge = definition.recurrence_mode.clone();

            if läge.is_empty() {
                läge = "schedule".to_string();
            }

            if läge != "schedule" && läge != "push" {
                proposal.notera(&format!(
                    "upprepningsstrategin {:?} är okänd - \"schedule\" användes i stället",
                    läge
                ));
                läge = "schedule".to_string();
            }

            proposal.values.recurrence_mode = Some(läge.clone());
            proposal.lägg_till("recurrence_mode", "Upprepningsstrategi", "", &läge);

            let mut start = definition.recurrence_start_date.clone();

            if start.is_empty() {
                start = definition.start_date.clone();
            }

            if start.is_empty() {
                start = definition.due.clone();
            }

            if !start.is_empty() {
                match giltigt_datum(&start) {
                    Some(datum) => {
                        proposal.values.recurrence_start_date = Some(datum.clone());
                        proposal.lägg_till(
                            "recurrence_start_date",
                            "Upprepning startar",
                            "",
                            &datum,
                        );
                    }
                    None => proposal.notera(&format!(
                        "startdatumet för upprepningen {:?} är ogiltigt och hoppades över",
                        start
                    )),
                }
            }
        }
    }

    // Taggar - slås ihop med målets egna.
    if let Some((taggar, tillagda)) = slå_ihop_taggar(&ctx.goal.tags, &definition.tags) {
        if tillagda {
            proposal.values.tags = taggar.clone();
            proposal.lägg_till(
                "tags",
                "Taggar",
                &ctx.goal.tags.join(", "),
                &taggar.join(", "),
            );
        }
    }

    // Platser - slås ihop med målets egna.
    if let Some((idn, namn, tillagda)) =
        slå_ihop_platser(&ctx.goal.places, &definition.places, &ctx.places)
    {
        if tillagda {
            proposal.values.place_ids = idn.clone();
            proposal.lägg_till(
                "places",
                "Platser",
                &platsnamn(&ctx.goal.places).join(", "),
                &namn.join(", "),
            );
        }
    }

    // Fasen - målet går vidare när definitionen är godkänd.
    if !ctx.next_phase.is_empty() {
        proposal.values.status = Some(ctx.next_phase.clone());
        proposal.lägg_till("status", "Fas", &ctx.current_phase, &ctx.next_phase);
    }

    proposal
}

/// Gör om förslaget till något som kan sparas i Plan.
///
/// Go: `Proposal.Payload`.
pub fn payload(proposal: &Proposal) -> serde_json::Value {
    serde_json::to_value(proposal).unwrap_or(serde_json::Value::Object(Default::default()))
}

/// Läser ett förslag ur ett godkännandes payload.
///
/// Det är så här exakt det som människan godkände läses tillbaka - utan
/// att något genereras om.
///
/// Go: `FromPayload`.
pub fn från_payload(payload: &serde_json::Value) -> Result<Proposal, String> {
    if payload.is_null() {
        return Err("godkännandet saknar det föreslagna innehållet".to_string());
    }

    let mut proposal: Proposal = serde_json::from_value(payload.clone())
        .map_err(|fel| format!("kunde inte tolka förslaget: {fel}"))?;

    if proposal.kind.is_empty() {
        proposal.kind = KIND.to_string();
    }

    Ok(proposal)
}

/// Skriver de föreslagna värdena till målet.
///
/// Exakt de värden som stod i förslaget appliceras - inget genereras om.
///
/// Go: `Proposal.Apply`.
pub async fn applicera(
    proposal: &Proposal,
    client: &PlanClient,
    goal_id: i64,
) -> Result<(), String> {
    let values = &proposal.values;

    let request = UpdateGoalRequest {
        title: values.title.clone(),
        description: values.description.clone(),
        icon: values.icon.clone(),
        now: values.now,
        eternal: values.eternal,
        agent: values.agent,
        due: values.due.clone(),
        start_date: values.start_date.clone(),
        estimated_minutes: values.estimated_minutes,
        recurrence_rule: values.recurrence_rule.clone(),
        recurrence_mode: values.recurrence_mode.clone(),
        recurrence_start_date: values.recurrence_start_date.clone(),
        status_id: None,
    };

    if !är_tom(&request) {
        client
            .update_goal(goal_id, request)
            .await
            .map_err(|fel| format!("failed to update goal: {fel}"))?;
    }

    // Taggar och platser ersätter hela uppsättningen, så de kommer
    // efter målets övriga fält.
    if !values.tags.is_empty() {
        client
            .set_goal_tags(goal_id, values.tags.clone())
            .await
            .map_err(|fel| format!("failed to set goal tags: {fel}"))?;
    }

    if !values.place_ids.is_empty() {
        client
            .set_goal_places(goal_id, values.place_ids.clone())
            .await
            .map_err(|fel| format!("failed to set goal places: {fel}"))?;
    }

    // Statusen ändras via den dedikerade endpointen - den vanliga
    // målendpointen ignorerar status_id.
    if let Some(status) = &values.status {
        if !status.is_empty() {
            client
                .update_goal_phase(goal_id, status)
                .await
                .map_err(|fel| format!("failed to update goal phase: {fel}"))?;
        }
    }

    Ok(())
}

/// Berättar om en uppdatering innehåller något alls.
///
/// Go: `isEmpty`.
fn är_tom(request: &UpdateGoalRequest) -> bool {
    request.title.is_none()
        && request.description.is_none()
        && request.icon.is_none()
        && request.now.is_none()
        && request.eternal.is_none()
        && request.agent.is_none()
        && request.due.is_none()
        && request.start_date.is_none()
        && request.estimated_minutes.is_none()
        && request.recurrence_rule.is_none()
        && request.recurrence_mode.is_none()
        && request.recurrence_start_date.is_none()
}

/// Berättar om förslaget innehåller andra data än fasen.
///
/// Fasen är agentens eget steg i flödet - den behöver inte godkännas av
/// en människa.
///
/// Go: `Proposal.HasDataFields`.
pub fn har_datafält(proposal: &Proposal) -> bool {
    proposal.fields.iter().any(|fält| fält.key != "status")
}

/// Beskriver förslaget för människan som ska ta ställning.
///
/// Go: `Proposal.Summary`.
pub fn sammanfattning(proposal: &Proposal) -> String {
    if proposal.fields.is_empty() {
        return "Agenten har inget nytt att föreslå för målet.".to_string();
    }

    let mut text = String::from("Agenten föreslår att målet får följande ändringar:\n\n");

    for fält in &proposal.fields {
        let från = if fält.from.is_empty() {
            "ingen"
        } else {
            &fält.from
        };

        text.push_str(&format!("- {}: {} → {}\n", fält.label, från, fält.to));
    }

    if !proposal.notes.is_empty() {
        text.push_str("\nNoteringar:\n");

        for notering in &proposal.notes {
            text.push_str(&format!("- {notering}\n"));
        }
    }

    text.trim_end_matches('\n').to_string()
}

/// Beskriver vad förslaget innebär.
///
/// Go: `Proposal.Consequence`.
pub fn konsekvens(proposal: &Proposal) -> String {
    if proposal.fields.is_empty() {
        return "Inget ändras.".to_string();
    }

    format!(
        "{} fält ändras på målet. Inget värde tas bort.",
        proposal.fields.len()
    )
}

/// Slår ihop målets taggar med agentens förslag.
///
/// Agentens taggar läggs till - målets egna tas aldrig bort.
///
/// Go: `mergeTags`.
pub fn slå_ihop_taggar(
    befintliga: &[String],
    föreslagna: &[String],
) -> Option<(Vec<String>, bool)> {
    let mut sammanslagna: Vec<String> = Vec::with_capacity(befintliga.len() + föreslagna.len());
    let mut sedda: Vec<String> = Vec::new();

    for tagg in befintliga {
        let namn = normalisera_tagg(tagg);

        if namn.is_empty() || sedda.contains(&namn) {
            continue;
        }

        sedda.push(namn.clone());
        sammanslagna.push(namn);
    }

    let mut tillagda = false;

    for tagg in föreslagna {
        let namn = normalisera_tagg(tagg);

        if namn.is_empty() || sedda.contains(&namn) {
            continue;
        }

        sedda.push(namn.clone());
        sammanslagna.push(namn);
        tillagda = true;
    }

    Some((sammanslagna, tillagda))
}

/// Gör om ett taggnamn till den form Plan sparar det i.
///
/// Go: `normalizeTag`.
pub fn normalisera_tagg(rå: &str) -> String {
    let namn = rå.trim();
    let namn = namn.strip_prefix('#').unwrap_or(namn);
    let namn = namn.to_lowercase();

    namn.split_whitespace().collect::<Vec<_>>().join(" ")
}

/// Slår ihop målets platser med agentens förslag.
///
/// Agenten föreslår platsnamn. Bara platser som redan finns hos
/// användaren godtas - agenten skapar inga nya platser.
///
/// Go: `mergePlaces`.
pub fn slå_ihop_platser(
    befintliga: &[Place],
    föreslagna: &[String],
    kända: &[Place],
) -> Option<(Vec<i64>, Vec<String>, bool)> {
    let mut idn: Vec<i64> = Vec::with_capacity(befintliga.len() + föreslagna.len());
    let mut namn: Vec<String> = Vec::with_capacity(befintliga.len() + föreslagna.len());
    let mut sedda: Vec<i64> = Vec::new();

    for plats in befintliga {
        if sedda.contains(&plats.id) {
            continue;
        }

        sedda.push(plats.id);
        idn.push(plats.id);
        namn.push(plats.name.clone());
    }

    let mut tillagda = false;

    for namn_förslag in föreslagna {
        let trimmad = namn_förslag.trim();

        if trimmad.is_empty() {
            continue;
        }

        let Some(plats) = hitta_plats(trimmad, kända) else {
            continue;
        };

        if sedda.contains(&plats.id) {
            continue;
        }

        sedda.push(plats.id);
        idn.push(plats.id);
        namn.push(plats.name.clone());
        tillagda = true;
    }

    Some((idn, namn, tillagda))
}

/// Letar upp en plats på sitt namn.
///
/// Go: `findPlace`.
pub fn hitta_plats<'a>(namn: &str, kända: &'a [Place]) -> Option<&'a Place> {
    let nål = namn.trim().to_lowercase();

    kända
        .iter()
        .find(|plats| plats.name.trim().to_lowercase() == nål)
}

/// Listar platsernas namn.
///
/// Go: `placeNames`.
pub fn platsnamn(platser: &[Place]) -> Vec<String> {
    platser.iter().map(|plats| plats.name.clone()).collect()
}

/// Kontrollerar att ett datum går att tolka som "YYYY-MM-DD".
///
/// Go: `validDate`.
pub fn giltigt_datum(värde: &str) -> Option<String> {
    let trimmad = värde.trim();

    if trimmad.is_empty() {
        return None;
    }

    chrono::NaiveDate::parse_from_str(trimmad, "%Y-%m-%d")
        .ok()
        .map(|_| trimmad.to_string())
}

/// Kontrollerar att ikonen går att visa.
///
/// En ikon är antingen en emoji eller ett Iconify-namn ("kollektion:namn").
///
/// Go: `validIcon`.
pub fn giltig_ikon(värde: &str) -> bool {
    let trimmad = värde.trim();

    if trimmad.is_empty() {
        return false;
    }

    if trimmad.contains(':') {
        let mut delar = trimmad.splitn(2, ':');

        let Some(kollektion) = delar.next() else {
            return false;
        };

        let Some(namn) = delar.next() else {
            return false;
        };

        return är_ikonnamn(kollektion) && är_ikonnamn(namn);
    }

    // En emoji - högst några tecken, varav minst ett är utanför ASCII.
    let tecken: Vec<char> = trimmad.chars().collect();

    if tecken.len() > 4 {
        return false;
    }

    tecken.iter().any(|tecken| *tecken > '\u{7f}')
}

/// Kontrollerar att en del av ett Iconify-namn är giltig.
///
/// Go: `isIconName`.
pub fn är_ikonnamn(värde: &str) -> bool {
    if värde.is_empty() {
        return false;
    }

    värde
        .chars()
        .all(|tecken| tecken.is_ascii_lowercase() || tecken.is_ascii_digit() || tecken == '-')
}

/// Kontrollerar att en upprepningsregel är giltig.
///
/// Go: `validateRule`.
pub fn validera_regel(regel: &str) -> Result<(), String> {
    let normaliserad = regel.trim().to_uppercase();

    if normaliserad.is_empty() {
        return Err("regeln är tom".to_string());
    }

    let normaliserad = normaliserad
        .strip_prefix("RRULE:")
        .unwrap_or(&normaliserad)
        .trim()
        .to_string();

    for del in normaliserad.split(';') {
        let Some((nyckel, värde)) = del.split_once('=') else {
            continue;
        };

        match nyckel.trim() {
            "INTERVAL" | "COUNT" => {
                let Ok(tal) = värde.trim().parse::<i64>() else {
                    return Err(format!("{nyckel} har ogiltigt värde {värde:?}"));
                };

                if tal < 1 {
                    return Err(format!("{nyckel} måste vara minst 1"));
                }
            }
            _ => {}
        }
    }

    plan_recurrence::Regel::validera(&normaliserad)
        .map_err(|fel| format!("regeln kunde inte tolkas: {fel}"))
}

#[cfg(test)]
mod tests {
    use super::*;

    fn mål() -> Goal {
        Goal {
            id: 7,
            title: "Gammal titel".to_string(),
            description: "Gammal beskrivning".to_string(),
            icon: String::new(),
            status_id: 1,
            agent: true,
            tags: vec!["hus".to_string()],
            places: Vec::new(),
            created: chrono::Utc::now(),
            started: None,
            finished: None,
            inbox: 0,
            now: false,
            eternal: false,
            due: None,
            start_date: None,
            estimated_minutes: None,
            recurrence_rule: None,
            recurrence_mode: None,
        }
    }

    #[test]
    fn befintliga_värden_lämnas_orörda() {
        let definition = Definition {
            title: "Gammal titel".to_string(),
            description: "Gammal beskrivning".to_string(),
            now: true,
            ..Default::default()
        };

        let mut goal = mål();
        goal.now = true;

        let ctx = Context {
            goal,
            definition,
            places: Vec::new(),
            current_phase: "define".to_string(),
            next_phase: "plan".to_string(),
        };

        let proposal = bygg(&ctx);

        assert!(proposal.values.title.is_none());
        assert!(proposal.values.now.is_none());
        assert_eq!(proposal.values.status.as_deref(), Some("plan"));
    }

    #[test]
    fn luckor_fylls_och_noteras() {
        let definition = Definition {
            title: "Ny titel".to_string(),
            description: "Ny beskrivning".to_string(),
            icon: "heroicons:flag".to_string(),
            due: "2026-12-01".to_string(),
            tags: vec!["arbete".to_string()],
            notes: vec!["antagande".to_string()],
            ..Default::default()
        };

        let ctx = Context {
            goal: mål(),
            definition,
            places: Vec::new(),
            current_phase: "define".to_string(),
            next_phase: "plan".to_string(),
        };

        let proposal = bygg(&ctx);

        assert_eq!(proposal.values.title.as_deref(), Some("Ny titel"));
        assert_eq!(proposal.values.icon.as_deref(), Some("heroicons:flag"));
        assert_eq!(proposal.values.due.as_deref(), Some("2026-12-01"));
        assert_eq!(proposal.values.tags, vec!["hus", "arbete"]);
        assert_eq!(proposal.notes, vec!["antagande"]);
    }

    #[test]
    fn ogiltig_datum_noteras_och_skips() {
        let definition = Definition {
            title: "Ny titel".to_string(),
            due: "2026-13-45".to_string(),
            ..Default::default()
        };

        let ctx = Context {
            goal: mål(),
            definition,
            places: Vec::new(),
            current_phase: "define".to_string(),
            next_phase: "plan".to_string(),
        };

        let proposal = bygg(&ctx);

        assert!(proposal.values.due.is_none());
        assert!(proposal.notes.iter().any(|n| n.contains("slutdatumet")));
    }

    #[test]
    fn ogiltig_upprepningsregel_noteras() {
        let definition = Definition {
            title: "Ny titel".to_string(),
            recurrence_rule: "FREQ=BOGUS".to_string(),
            ..Default::default()
        };

        let ctx = Context {
            goal: mål(),
            definition,
            places: Vec::new(),
            current_phase: "define".to_string(),
            next_phase: "plan".to_string(),
        };

        let proposal = bygg(&ctx);

        assert!(proposal.values.recurrence_rule.is_none());
        assert!(proposal
            .notes
            .iter()
            .any(|n| n.contains("upprepningsregeln")));
    }

    #[test]
    fn endast_fasen_är_inga_datafält() {
        let ctx = Context {
            goal: mål(),
            definition: Definition::default(),
            places: Vec::new(),
            current_phase: "define".to_string(),
            next_phase: "plan".to_string(),
        };

        let proposal = bygg(&ctx);

        assert!(!har_datafält(&proposal));
        assert_eq!(
            konsekvens(&proposal),
            "1 fält ändras på målet. Inget värde tas bort."
        );
    }

    #[test]
    fn payload_roundtrip_behåller_värdena() {
        let definition = Definition {
            title: "Ny titel".to_string(),
            due: "2026-12-01".to_string(),
            ..Default::default()
        };

        let ctx = Context {
            goal: mål(),
            definition,
            places: Vec::new(),
            current_phase: "define".to_string(),
            next_phase: "plan".to_string(),
        };

        let proposal = bygg(&ctx);
        let raw = payload(&proposal);
        let tillbaka = från_payload(&raw).unwrap();

        assert_eq!(tillbaka.values.title, proposal.values.title);
        assert_eq!(tillbaka.values.due, proposal.values.due);
        assert_eq!(tillbaka.kind, KIND);
    }

    #[test]
    fn taggar_normaliseras() {
        assert_eq!(normalisera_tagg("  #Hus  Arbete "), "hus arbete");
    }
}
