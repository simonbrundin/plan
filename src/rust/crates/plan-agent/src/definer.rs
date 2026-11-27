//! Agentens förslag på hur ett mål ska se ut.
//!
//! Go: `internal/definer`. Definitionen är agentens egen - den är ett
//! förslag, inte ett facit. Agenten skriver aldrig till målet själv.

use std::sync::Arc;

use serde::{Deserialize, Deserializer, Serialize};

use crate::messages::Message;
use crate::provider::{LlmProvider, Model, StreamEvent};

/// Maximal längd på en titel (antal tecken).
pub const MAX_TITLE_LENGTH: usize = 50;

/// Agentens förslag på målets utseende.
///
/// Fälten är de målfält agenten kan fylla. Ett tomt fält betyder att
/// agenten inte har något att föreslå - då lämnas målets värde orört.
///
/// Modellen svarar ofta med `null` för fält den inte har något till -
/// det läses som tomma värden, precis som i Go.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct Definition {
    /// Målets titel.
    #[serde(default, deserialize_with = "null_som_tom_str")]
    pub title: String,
    /// Målets beskrivning.
    #[serde(default, deserialize_with = "null_som_tom_str")]
    pub description: String,
    /// Ikonen - en emoji eller ett Iconify-namn.
    #[serde(
        default,
        deserialize_with = "null_som_tom_str",
        skip_serializing_if = "String::is_empty"
    )]
    pub icon: String,
    /// Om målet ska göras nu.
    #[serde(
        default,
        deserialize_with = "null_som_false",
        skip_serializing_if = "is_false"
    )]
    pub now: bool,
    /// Om målet aldrig slutar.
    #[serde(
        default,
        deserialize_with = "null_som_false",
        skip_serializing_if = "is_false"
    )]
    pub eternal: bool,
    /// Om målet är avsett för en agent.
    #[serde(
        default,
        deserialize_with = "null_som_false",
        skip_serializing_if = "is_false"
    )]
    pub agent: bool,
    /// Slutdatumet ("YYYY-MM-DD").
    #[serde(
        default,
        deserialize_with = "null_som_tom_str",
        skip_serializing_if = "String::is_empty"
    )]
    pub due: String,
    /// Startdatumet ("YYYY-MM-DD").
    #[serde(
        default,
        deserialize_with = "null_som_tom_str",
        skip_serializing_if = "String::is_empty"
    )]
    pub start_date: String,
    /// Den uppskattade tidsåtgången, i minuter.
    #[serde(
        default,
        deserialize_with = "null_som_noll",
        skip_serializing_if = "is_zero_i32"
    )]
    pub estimated_minutes: i32,
    /// Upprepningsregeln ("FREQ=WEEKLY;BYDAY=MO").
    #[serde(
        default,
        deserialize_with = "null_som_tom_str",
        skip_serializing_if = "String::is_empty"
    )]
    pub recurrence_rule: String,
    /// Schemaläggningsstrategin ("schedule" eller "push").
    #[serde(
        default,
        deserialize_with = "null_som_tom_str",
        skip_serializing_if = "String::is_empty"
    )]
    pub recurrence_mode: String,
    /// Seriens startdatum ("YYYY-MM-DD").
    #[serde(
        default,
        deserialize_with = "null_som_tom_str",
        skip_serializing_if = "String::is_empty"
    )]
    pub recurrence_start_date: String,
    /// Taggar som passar målet.
    #[serde(
        default,
        deserialize_with = "null_som_tom_lista",
        skip_serializing_if = "Vec::is_empty"
    )]
    pub tags: Vec<String>,
    /// Platser där målet kan slutföras.
    #[serde(
        default,
        deserialize_with = "null_som_tom_lista",
        skip_serializing_if = "Vec::is_empty"
    )]
    pub places: Vec<String>,
    /// Antaganden agenten gjort.
    #[serde(
        default,
        deserialize_with = "null_som_tom_lista",
        skip_serializing_if = "Vec::is_empty"
    )]
    pub notes: Vec<String>,
}

/// Läser `null` som en tom sträng.
///
/// Go lämnar fältet oförändrat när det är `null` - modellen skriver ofta
/// ut null för fält den inte har något till.
fn null_som_tom_str<'de, D>(deserializer: D) -> Result<String, D::Error>
where
    D: Deserializer<'de>,
{
    Ok(Option::<String>::deserialize(deserializer)?.unwrap_or_default())
}

/// Läser `null` som falskt.
fn null_som_false<'de, D>(deserializer: D) -> Result<bool, D::Error>
where
    D: Deserializer<'de>,
{
    Ok(Option::<bool>::deserialize(deserializer)?.unwrap_or_default())
}

/// Läser `null` som noll.
fn null_som_noll<'de, D>(deserializer: D) -> Result<i32, D::Error>
where
    D: Deserializer<'de>,
{
    Ok(Option::<i32>::deserialize(deserializer)?.unwrap_or_default())
}

/// Läser `null` som en tom lista.
fn null_som_tom_lista<'de, D>(deserializer: D) -> Result<Vec<String>, D::Error>
where
    D: Deserializer<'de>,
{
    Ok(Option::<Vec<String>>::deserialize(deserializer)?.unwrap_or_default())
}

/// Om värdet är falskt.
fn is_false(värde: &bool) -> bool {
    !*värde
}

/// Om värdet är noll.
fn is_zero_i32(värde: &i32) -> bool {
    *värde == 0
}

/// Målet som ska definieras, tillsammans med värdena som redan finns.
///
/// De befintliga värdena skickas med så att agenten inte föreslår något
/// som redan är satt.
#[derive(Debug, Clone, Default)]
pub struct Input {
    /// Målets titel.
    pub title: String,
    /// Målets beskrivning.
    pub description: String,
    /// Återkoppling från en människa på ett tidigare förslag.
    pub feedback: String,
    /// Målets nuvarande ikon.
    pub icon: String,
    /// Målets nuvarande flaggor.
    pub now: bool,
    /// Om målet är evigt.
    pub eternal: bool,
    /// Om målet är ett agentmål.
    pub agent: bool,
    /// Målets nuvarande datum.
    pub due: String,
    /// Målets nuvarande startdatum.
    pub start_date: String,
    /// Målets nuvarande uppskattning, i minuter.
    pub estimated_minutes: i32,
    /// Målets nuvarande upprepningsregel.
    pub recurrence_rule: String,
    /// Målets nuvarande taggar.
    pub tags: Vec<String>,
    /// Målets nuvarande platser, med namn.
    pub places: Vec<String>,
}

/// Agenten som definierar mål.
pub struct Definer {
    /// Leverantören som används.
    leverantör: Arc<dyn LlmProvider>,
    /// Modellen som används.
    modell: Model,
}

impl Definer {
    /// Skapar en definierare.
    pub fn ny(leverantör: Arc<dyn LlmProvider>, modell: Model) -> Self {
        Self {
            leverantör, modell
        }
    }

    /// Tar fram ett förslag på hur målet ska se ut.
    pub async fn definiera(&self, input: &Input) -> Result<Definition, String> {
        let systemprompt = SYSTEMPROMPT;
        let userprompt = bygg_userprompt(input);

        let meddelanden = vec![Message::system(systemprompt), Message::user(userprompt)];

        let mut händelser = self
            .leverantör
            .stream(
                self.modell.clone(),
                meddelanden,
                Vec::new(),
                crate::provider::StreamingOptions::default(),
            )
            .await
            .map_err(|fel| format!("failed to stream response: {fel}"))?;

        let mut svar = String::new();

        while let Some(händelse) = händelser.recv().await {
            match händelse {
                StreamEvent::TextDelta { delta, .. } => svar.push_str(&delta),
                StreamEvent::Done { .. } => {
                    tracing::debug!("Definitionens svar: {svar:?}");

                    return tolka_definition(&svar);
                }
                StreamEvent::Error { error, .. } => return Err(error),
                _ => {}
            }
        }

        Err("stream ended without completion".to_string())
    }
}

/// Systemprompten som används.
///
/// Go: `definer.Define`.
pub const SYSTEMPROMPT: &str = r#"You are a software requirements analyst. Your job is to take a loose goal or idea and turn it into a well-defined goal.

Answer with ONE JSON object and nothing else. No prose, no markdown fences.

The object has these fields:
{
  "title": "concise title, max 50 characters",
  "description": "what the goal accomplishes, key requirements, success criteria and scope",
  "icon": "one emoji or an iconify name like heroicons:flag",
  "now": false,
  "eternal": false,
  "agent": false,
  "due": "YYYY-MM-DD or null",
  "start_date": "YYYY-MM-DD or null",
  "estimated_minutes": 0,
  "recurrence_rule": "RFC 5545 RRULE such as FREQ=WEEKLY;BYDAY=MO, or null",
  "recurrence_mode": "schedule or push",
  "recurrence_start_date": "YYYY-MM-DD or null",
  "tags": ["short tags"],
  "places": ["names of places the goal applies to"],
  "notes": ["assumptions you made"]
}

Rules:
- Leave a field out or set it to null when you have nothing to say about it. Never invent a value just to fill the field.
- due, start_date and recurrence_start_date are calendar dates in YYYY-MM-DD.
- estimated_minutes is a realistic estimate for one occurrence of the goal. Use 0 when you cannot estimate.
- Only use places that are listed in the goal's existing places, or that you are told about. Use the exact names.
- now is true only for something that should be done right now. eternal is true for goals that never end.
- agent is true when a person should hand the goal over to an agent.
- tags are short and lowercase.

Be specific and actionable. Avoid vague language like "improve" or "optimize" without specifics."#;

/// Bygger frågan till modellen.
///
/// Go: `buildUserPrompt`.
pub fn bygg_userprompt(input: &Input) -> String {
    let mut text = String::new();

    text.push_str("## GOAL TO DEFINE\n\n");
    text.push_str(&input.title);
    text.push_str("\n\n");

    if !input.description.is_empty() {
        text.push_str("## EXISTING DESCRIPTION\n\n");
        text.push_str(&input.description);
        text.push_str("\n\n");
    }

    text.push_str("## WHAT THE GOAL ALREADY HAS\n\n");
    text.push_str(&format!("- ikon: {}\n", eller_ingen(&input.icon)));
    text.push_str(&format!("- ska göras nu: {}\n", ja_eller_nej(input.now)));
    text.push_str(&format!("- evigt mål: {}\n", ja_eller_nej(input.eternal)));
    text.push_str(&format!("- agentmål: {}\n", ja_eller_nej(input.agent)));
    text.push_str(&format!("- slutdatum: {}\n", eller_ingen(&input.due)));
    text.push_str(&format!(
        "- startdatum: {}\n",
        eller_ingen(&input.start_date)
    ));
    text.push_str(&format!(
        "- uppskattad tid: {}\n",
        eller_ingen(&minuter_text(input.estimated_minutes))
    ));
    text.push_str(&format!(
        "- upprepningsregel: {}\n",
        eller_ingen(&input.recurrence_rule)
    ));
    text.push_str(&format!(
        "- taggar: {}\n",
        eller_ingen(&input.tags.join(", "))
    ));
    text.push_str(&format!(
        "- platser: {}\n",
        eller_ingen(&input.places.join(", "))
    ));
    text.push('\n');

    if !input.feedback.is_empty() {
        text.push_str("## PREVIOUS FEEDBACK (please address this)\n\n");
        text.push_str(&input.feedback);
        text.push_str("\n\n");
    }

    text.push_str("## YOUR TASK\n\n");
    text.push_str(
        "Answer with one JSON object. Fill in every field you have a good reason to fill in, ",
    );
    text.push_str("and leave the rest out. If the goal is unclear, make reasonable assumptions and state them in notes.\n");

    text
}

/// Skriver "ingen" när värdet saknas.
///
/// Go: `orNone`.
fn eller_ingen(värde: &str) -> String {
    if värde.trim().is_empty() {
        "ingen".to_string()
    } else {
        värde.to_string()
    }
}

/// Skriver ja eller nej.
///
/// Go: `yesNo`.
fn ja_eller_nej(värde: bool) -> &'static str {
    if värde {
        "ja"
    } else {
        "nej"
    }
}

/// Skriver en uppskattning som text.
///
/// Go: `minutesText`.
fn minuter_text(minuter: i32) -> String {
    if minuter <= 0 {
        return String::new();
    }

    format!("{minuter} minuter")
}

/// Tolkar svaret från modellen.
///
/// Svaret ska vara ett JSON-objekt. Om modellen svarar med något annat
/// faller parsningen tillbaka på det äldre textformatet - en trasig
/// definition ska inte stoppa agenten.
///
/// Go: `parseDefinition`.
pub fn tolka_definition(svar: &str) -> Result<Definition, String> {
    match tolka_json(svar) {
        Ok(definition) => Ok(definition),
        Err(_) => tolka_text(svar),
    }
}

/// Läser ett JSON-objekt ur svaret.
///
/// Go: `parseJSONDefinition`.
pub fn tolka_json(svar: &str) -> Result<Definition, String> {
    let Some(start) = svar.find('{') else {
        return Err("svaret innehåller inget JSON-objekt".to_string());
    };

    let objekt = json_objekt(svar, start).ok_or("JSON-objektet i svaret är inte komplett")?;

    let mut definition: Definition = serde_json::from_str(&objekt)
        .map_err(|fel| format!("kunde inte tolka definitionen: {fel}"))?;

    if definition.title.trim().is_empty() {
        return Err("definitionen saknar en titel".to_string());
    }

    normalisera(&mut definition);

    Ok(definition)
}

/// Hittar det JSON-objekt som börjar på positionen start.
///
/// Strängar och escapes räknas med, så att ett "}" inne i en text inte
/// avslutar objektet.
///
/// Go: `jsonObject`.
pub fn json_objekt(text: &str, start: usize) -> Option<String> {
    let bytes = text.as_bytes();

    let mut djup = 0i32;
    let mut i_sträng = false;
    let mut escaped = false;

    for (position, tecken) in bytes.iter().enumerate().skip(start) {
        let tecken = *tecken;

        if i_sträng {
            if escaped {
                escaped = false;
            } else if tecken == b'\\' {
                escaped = true;
            } else if tecken == b'"' {
                i_sträng = false;
            }

            continue;
        }

        match tecken {
            b'"' => i_sträng = true,
            b'{' => djup += 1,
            b'}' => {
                djup -= 1;

                if djup == 0 {
                    return Some(text[start..position + 1].to_string());
                }
            }
            _ => {}
        }
    }

    None
}

/// Tolkar det äldre textformatet ("TITLE:" och "DESCRIPTION:").
///
/// Go: `parseTextDefinition`.
pub fn tolka_text(svar: &str) -> Result<Definition, String> {
    let mut titel = String::new();
    let mut beskrivning = String::new();
    let mut i_beskrivning = false;

    for rad in svar.split('\n') {
        let trimmed = rad.trim();

        if let Some(rest) = trimmed.strip_prefix("TITLE:") {
            titel = rest.trim().to_string();
            continue;
        }

        if let Some(rest) = trimmed.strip_prefix("DESCRIPTION:") {
            beskrivning = rest.trim().to_string();
            i_beskrivning = true;
            continue;
        }

        if i_beskrivning {
            // En ny sektion avslutar beskrivningen.
            if trimmed.starts_with("TITLE:") || trimmed.starts_with("ANALYSIS:") {
                break;
            }

            beskrivning.push(' ');
            beskrivning.push_str(trimmed);
        }
    }

    if titel.is_empty() {
        return Err("svaret innehåller ingen titel".to_string());
    }

    let mut definition = Definition {
        title: titel,
        description: beskrivning.trim().to_string(),
        ..Default::default()
    };

    normalisera(&mut definition);

    Ok(definition)
}

/// Sätter värdena i den form de ska sparas i.
///
/// Go: `normalizeDefinition`.
pub fn normalisera(definition: &mut Definition) {
    definition.title = definition.title.trim().to_string();
    definition.description = definition.description.trim().to_string();
    definition.icon = definition.icon.trim().to_string();
    definition.due = definition.due.trim().to_string();
    definition.start_date = definition.start_date.trim().to_string();
    definition.recurrence_rule = definition.recurrence_rule.trim().to_string();
    definition.recurrence_mode = definition.recurrence_mode.trim().to_string();
    definition.recurrence_start_date = definition.recurrence_start_date.trim().to_string();

    // Titeln kortas av om den är för lång.
    if definition.title.len() > MAX_TITLE_LENGTH {
        definition.title = format!("{}...", &definition.title[..MAX_TITLE_LENGTH - 3]);
    }
}

impl crate::worker::GoalDefiner for Definer {
    fn definiera<'a>(
        &'a self,
        input: &'a Input,
    ) -> std::pin::Pin<Box<dyn std::future::Future<Output = Result<Definition, String>> + Send + 'a>>
    {
        Box::pin(self.definiera(input))
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn json_läses_ur_svaret() {
        let svar = r#"{"title":"Skriv test","description":"För definitionen","icon":"flag","now":true,"tags":["a","b"]}"#;

        let definition = tolka_definition(svar).unwrap();

        assert_eq!(definition.title, "Skriv test");
        assert_eq!(definition.description, "För definitionen");
        assert_eq!(definition.icon, "flag");
        assert!(definition.now);
        assert_eq!(definition.tags, vec!["a", "b"]);
    }

    #[test]
    fn json_i_kodblock_läses() {
        let svar = "Här är förslaget:\n```json\n{\"title\":\"Titel\"}\n```\n";

        let definition = tolka_definition(svar).unwrap();

        assert_eq!(definition.title, "Titel");
    }

    #[test]
    fn textformatet_används_som_reserv() {
        let svar = "TITLE: Rubrik\nDESCRIPTION: Förklaringen\nfortsättning";

        let definition = tolka_definition(svar).unwrap();

        assert_eq!(definition.title, "Rubrik");
        assert_eq!(definition.description, "Förklaringen fortsättning");
    }

    #[test]
    fn null_läses_som_tomma_värden() {
        let svar = r#"{"title":"Titel","due":null,"start_date":null,"estimated_minutes":null,"tags":null,"places":null,"now":null}"#;

        let definition = tolka_definition(svar).unwrap();

        assert_eq!(definition.title, "Titel");
        assert_eq!(definition.due, "");
        assert_eq!(definition.start_date, "");
        assert_eq!(definition.estimated_minutes, 0);
        assert!(definition.tags.is_empty());
        assert!(definition.places.is_empty());
        assert!(!definition.now);
    }

    #[test]
    fn skräp_avvisas() {
        assert!(tolka_definition("ingen definition här").is_err());
    }

    #[test]
    fn lång_titel_kortas_av() {
        let definition = tolka_definition(&format!(r#"{{"title":"{}"}}"#, "a".repeat(80))).unwrap();

        assert_eq!(definition.title.len(), MAX_TITLE_LENGTH);
        assert!(definition.title.ends_with("..."));
    }

    #[test]
    fn klammer_i_text_räknas_inte() {
        let objekt = json_objekt(r#"{"title":"a } b"}"#, 0);

        assert_eq!(objekt, Some(r#"{"title":"a } b"}"#.to_string()));
    }

    #[test]
    fn frågan_innehåller_målets_värden() {
        let input = Input {
            title: "Målet".to_string(),
            icon: String::new(),
            now: true,
            ..Default::default()
        };

        let fråga = bygg_userprompt(&input);

        assert!(fråga.contains("## GOAL TO DEFINE\n\nMålet\n"));
        assert!(fråga.contains("- ikon: ingen"));
        assert!(fråga.contains("- ska göras nu: ja"));
        assert!(fråga.contains("- evigt mål: nej"));
    }
}
