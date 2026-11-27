//! Agentens runtime - loopen som kör modell och verktyg.
//!
//! Go: `internal/core/agent.go`. Agenten skickar meddelanden till
//! språkmodellen, utför verktygsanropen som modellen begär och fortsätter
//! tills modellen svarar utan verktygsanrop.

use std::sync::Arc;

use serde_json::Value;

use crate::messages::{ContentBlock, Message};
use crate::provider::{
    konvertera_assistent_innehåll, tolka_verktygsargument, LlmProvider, Model, ProviderManager,
    SharedProvider, StreamEvent, ToolResult,
};
use crate::tools::{ToolCall, ToolDefinition};

/// Händelsetyper som agenten skickar ut.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default)]
pub enum EventType {
    /// Agenten startar.
    #[default]
    AgentStart,
    /// Agenten är klar.
    AgentEnd,
    /// En tur börjar.
    TurnStart,
    /// En tur är slut.
    TurnEnd,
    /// Ett meddelande börjar.
    MessageStart,
    /// Ett meddelande är slut.
    MessageEnd,
    /// Ett meddelande uppdateras.
    MessageUpdate,
    /// Ett verktyg börjar köras.
    ToolStart,
    /// Ett verktyg är slut.
    ToolEnd,
    /// Ett verktyg uppdateras.
    ToolUpdate,
}

impl EventType {
    /// Händelsetypen som text - så som Go skriver den.
    pub fn as_str(&self) -> &'static str {
        match self {
            EventType::AgentStart => "agent_start",
            EventType::AgentEnd => "agent_end",
            EventType::TurnStart => "turn_start",
            EventType::TurnEnd => "turn_end",
            EventType::MessageStart => "message_start",
            EventType::MessageEnd => "message_end",
            EventType::MessageUpdate => "message_update",
            EventType::ToolStart => "tool_execution_start",
            EventType::ToolEnd => "tool_execution_end",
            EventType::ToolUpdate => "tool_execution_update",
        }
    }
}

/// En händelse från agenten.
#[derive(Debug, Clone, Default)]
pub struct AgentEvent {
    /// Händelsens typ.
    pub typ: EventType,
    /// Meddelandet händelsen gäller.
    pub message: Option<Message>,
    /// Id:t på verktygsanropet.
    pub tool_call_id: String,
    /// Verktygets namn.
    pub tool_name: String,
    /// Argumenten.
    pub args: Option<Value>,
    /// Resultatet.
    pub result: Option<ToolResult>,
    /// Om resultatet var ett fel.
    pub is_error: bool,
    /// Resultaten av verktygsanropen.
    pub tool_results: Vec<Message>,
    /// Meddelandena i samtalet.
    pub messages: Vec<Message>,
}

/// En framtid som behandlar en händelse.
pub type EventFuture = std::pin::Pin<Box<dyn std::future::Future<Output = ()> + Send + 'static>>;

/// En hanterare av agentens händelser.
///
/// Handlaren är asynkron - så att den kan spara meddelanden i
/// databasen medan svaret strömmas.
pub type EventHandler = Arc<dyn Fn(AgentEvent) -> EventFuture + Send + Sync>;

/// Sammanhanget för en tur - används av krokar.
pub struct TurnContext {
    /// Modellens svar.
    pub message: Option<Message>,
    /// Resultaten av verktygsanropen.
    pub tool_results: Vec<Message>,
    /// Agentens aktuella tillstånd.
    pub context: AgentContext,
}

/// Agentens aktuella tillstånd.
#[derive(Clone)]
pub struct AgentContext {
    /// Systemprompten.
    pub system_prompt: String,
    /// Meddelandena.
    pub messages: Vec<Message>,
    /// Verktygen.
    pub tools: Vec<ToolDefinition>,
}

/// En uppdatering av nästa tur.
#[derive(Clone, Default)]
pub struct TurnUpdate {
    /// Nytt sammanhang.
    pub context: Option<AgentContext>,
    /// Ny modell.
    pub model: Option<Model>,
}

/// Resultatet av en körning.
#[derive(Debug, Clone, Default)]
pub struct RunResult {
    /// Meddelandena i samtalet.
    pub messages: Vec<Message>,
    /// Felet, om körningen misslyckades.
    pub error: Option<String>,
}

/// En krok som avgör om agenten får sluta.
pub type ShouldStopHook = Arc<dyn Fn(&TurnContext) -> bool + Send + Sync>;

/// En krok som förbereder nästa tur.
pub type PrepareNextTurnHook = Arc<dyn Fn(&TurnContext) -> Option<TurnUpdate> + Send + Sync>;

/// En krok som anropas när en tur är klar.
pub type OnTurnCompleteHook = Arc<dyn Fn(&Message, &str) + Send + Sync>;

/// Standardinställningarna för agentens loop.
#[derive(Default)]
pub struct AgentLoopConfig {
    /// Modellen som används.
    pub model: Model,
    /// Verktygen agenten har.
    pub tools: Vec<ToolDefinition>,
    /// Systemprompten.
    pub system_prompt: String,
    /// Leverantören som används.
    pub provider: Option<SharedProvider>,
    /// Chefen för flera leverantörer - används för reserv.
    pub provider_manager: Option<ProviderManager>,
    /// Krok som avgör om agenten får sluta.
    pub should_stop: Option<ShouldStopHook>,
    /// Krok som förbereder nästa tur.
    pub prepare_next_turn: Option<PrepareNextTurnHook>,
    /// Krok som anropas när en tur är klar.
    pub on_turn_complete: Option<OnTurnCompleteHook>,
    /// Det maximala antalet turer.
    pub max_turns: usize,
}

/// Agentens runtime.
pub struct AgentRuntime {
    /// Inställningarna.
    config: AgentLoopConfig,
    /// Leverantören som användes i senaste turen.
    nuvarande_leverantör: String,
}

impl AgentRuntime {
    /// Skapar en agent.
    pub fn ny(config: AgentLoopConfig) -> Self {
        Self {
            config,
            nuvarande_leverantör: String::new(),
        }
    }

    /// Modellen som används.
    pub fn modell(&self) -> &Model {
        &self.config.model
    }

    /// Verktygen agenten har.
    pub fn verktyg(&self) -> &[ToolDefinition] {
        &self.config.tools
    }

    /// Skickar ett meddelande utan historik.
    pub async fn skicka_meddelande(&mut self, prompt: &str, handler: EventHandler) -> RunResult {
        self.skicka_meddelande_med_historik(Vec::new(), prompt, handler)
            .await
    }

    /// Skickar ett meddelande med historiken bevarad.
    pub async fn skicka_meddelande_med_historik(
        &mut self,
        historik: Vec<Message>,
        prompt: &str,
        handler: EventHandler,
    ) -> RunResult {
        let mut meddelanden = historik;

        meddelanden.push(Message::User {
            content: vec![ContentBlock::Text {
                text: prompt.to_string(),
            }],
            timestamp: crate::messages::nu(),
        });

        self.kör(meddelanden, handler).await
    }

    /// Kör agenten med givna meddelanden.
    async fn kör(&mut self, meddelanden: Vec<Message>, handler: EventHandler) -> RunResult {
        let mut kontext = AgentContext {
            system_prompt: self.config.system_prompt.clone(),
            messages: meddelanden.clone(),
            tools: self.config.tools.clone(),
        };

        // Systemprompten läggs först.
        if !kontext.system_prompt.is_empty() {
            kontext.messages.insert(
                0,
                Message::System {
                    content: vec![ContentBlock::Text {
                        text: kontext.system_prompt.clone(),
                    }],
                    timestamp: crate::messages::nu(),
                },
            );
        }

        handler(AgentEvent {
            typ: EventType::AgentStart,
            ..Default::default()
        })
        .await;

        let max_turns = if self.config.max_turns == 0 {
            20
        } else {
            self.config.max_turns
        };

        let mut stoppades_normalt = false;

        for _ in 0..max_turns {
            handler(AgentEvent {
                typ: EventType::TurnStart,
                ..Default::default()
            })
            .await;

            let meddelande = match self.strömma_svar(&mut kontext, &handler).await {
                Ok(meddelande) => meddelande,
                Err(fel) => {
                    handler(AgentEvent {
                        typ: EventType::AgentEnd,
                        messages: kontext.messages.clone(),
                        ..Default::default()
                    })
                    .await;

                    return RunResult {
                        messages: kontext.messages,
                        error: Some(fel),
                    };
                }
            };

            kontext.messages.push(meddelande.clone());

            // Kontrollera om agenten får sluta. Checkpointens grindar
            // avgör det - utan grindar får agenten sluta som vanligt.
            let stopp = meddelande
                .stop_reason()
                .is_some_and(|reason| reason == "end_turn" || reason == "stop_sequence");

            if stopp {
                let tur = TurnContext {
                    message: Some(meddelande.clone()),
                    tool_results: Vec::new(),
                    context: kontext.clone(),
                };

                if let Some(should_stop) = &self.config.should_stop {
                    if !should_stop(&tur) {
                        // Grindarna blockerade stoppet - fortsätt arbeta.
                        handler(AgentEvent {
                            typ: EventType::TurnEnd,
                            message: Some(meddelande.clone()),
                            ..Default::default()
                        })
                        .await;

                        continue;
                    }
                }

                stoppades_normalt = true;
                break;
            }

            // Kör verktygsanropen.
            let verktygsresultat = self
                .kör_verktygsanrop(&mut kontext, &meddelande, &handler)
                .await;

            for resultat in &verktygsresultat {
                kontext.messages.push(resultat.clone());
            }

            let tur = TurnContext {
                message: Some(meddelande.clone()),
                tool_results: verktygsresultat.clone(),
                context: kontext.clone(),
            };

            if let Some(should_stop) = &self.config.should_stop {
                if should_stop(&tur) {
                    stoppades_normalt = true;
                    break;
                }
            }

            if let Some(prepare) = &self.config.prepare_next_turn {
                if let Some(uppdatering) = prepare(&tur) {
                    if let Some(nytt) = uppdatering.context {
                        kontext = nytt;
                    }

                    if let Some(modell) = uppdatering.model {
                        self.config.model = modell;
                    }
                }
            }

            handler(AgentEvent {
                typ: EventType::TurnEnd,
                message: Some(meddelande.clone()),
                tool_results: verktygsresultat,
                ..Default::default()
            })
            .await;

            if let Some(komplete) = &self.config.on_turn_complete {
                komplete(&meddelande, &self.nuvarande_leverantör);
            }
        }

        handler(AgentEvent {
            typ: EventType::AgentEnd,
            messages: kontext.messages.clone(),
            ..Default::default()
        })
        .await;

        if stoppades_normalt {
            return RunResult {
                messages: kontext.messages,
                error: None,
            };
        }

        RunResult {
            messages: kontext.messages,
            error: Some(format!(
                "agent stopped after reaching the maximum of {max_turns} turns"
            )),
        }
    }

    /// Strömmar modellens svar.
    async fn strömma_svar(
        &self,
        kontext: &mut AgentContext,
        handler: &EventHandler,
    ) -> Result<Message, String> {
        let verktyg: Vec<crate::provider::ToolDefinition> = kontext
            .tools
            .iter()
            .map(|verktyg| verktyg.till_leverantör())
            .collect();

        let (leverantörsnamn, mut mottagare) = if let Some(chef) = &self.config.provider_manager {
            chef.stream(
                &self.config.model,
                &kontext.messages,
                &verktyg,
                &crate::provider::StreamingOptions {
                    system_prompt: kontext.system_prompt.clone(),
                    ..Default::default()
                },
            )
            .await
            .map_err(|fel| format!("stream error: {fel}"))?
        } else if let Some(leverantör) = &self.config.provider {
            let händelser = leverantör
                .stream(
                    self.config.model.clone(),
                    kontext.messages.clone(),
                    verktyg,
                    crate::provider::StreamingOptions {
                        system_prompt: kontext.system_prompt.clone(),
                        ..Default::default()
                    },
                )
                .await
                .map_err(|fel| format!("stream error: {fel}"))?;

            ("default".to_string(), händelser)
        } else {
            return Err("stream error: no provider configured".to_string());
        };

        if !leverantörsnamn.is_empty() {
            tracing::info!("Agenten använder leverantören {leverantörsnamn}");
        }

        let mut partial;
        let mut meddelande = Message::Assistant {
            content: Vec::new(),
            stop_reason: String::new(),
            timestamp: crate::messages::nu(),
            usage: None,
        };

        while let Some(händelse) = mottagare.recv().await {
            match händelse {
                StreamEvent::Start { partial: ny } => {
                    partial = ny;

                    handler(AgentEvent {
                        typ: EventType::MessageStart,
                        message: Some(partial.to_message()),
                        ..Default::default()
                    })
                    .await;
                }
                StreamEvent::TextDelta { delta, partial: ny } => {
                    partial = ny;

                    partial.content.push(ContentBlock::Text { text: delta });

                    meddelande = partial.to_message();
                }
                StreamEvent::ToolCallDelta {
                    id,
                    arguments,
                    partial: ny,
                    ..
                } => {
                    partial = ny;

                    // Argumenten lagras rått tills anropet är klart.
                    if let Some(ContentBlock::ToolCall {
                        arguments: args, ..
                    }) = hitta_verktygsanrop(&mut partial.content, &id)
                    {
                        args["_raw_args"] = Value::String(arguments);
                    }

                    meddelande = partial.to_message();
                }
                StreamEvent::Done { message } => {
                    meddelande = message;

                    handler(AgentEvent {
                        typ: EventType::MessageEnd,
                        message: Some(meddelande.clone()),
                        ..Default::default()
                    })
                    .await;

                    return Ok(meddelande);
                }
                StreamEvent::Error {
                    error: fel,
                    message,
                } => {
                    meddelande = message;

                    handler(AgentEvent {
                        typ: EventType::MessageEnd,
                        message: Some(meddelande.clone()),
                        ..Default::default()
                    })
                    .await;

                    return Err(fel);
                }
            }
        }

        Ok(meddelande)
    }

    /// Kör verktygsanropen i modellens svar.
    async fn kör_verktygsanrop(
        &self,
        kontext: &mut AgentContext,
        meddelande: &Message,
        handler: &EventHandler,
    ) -> Vec<Message> {
        let mut resultat = Vec::new();

        for block in meddelande.content() {
            let ContentBlock::ToolCall { id, name, .. } = block else {
                continue;
            };

            resultat.push(
                self.kör_ett_anrop(id, name, block, &kontext.tools, handler)
                    .await,
            );
        }

        resultat
    }

    /// Kör ett verktygsanrop.
    async fn kör_ett_anrop(
        &self,
        id: &str,
        namn: &str,
        block: &ContentBlock,
        verktyg: &[ToolDefinition],
        handler: &EventHandler,
    ) -> Message {
        let mut block = block.clone();

        // De råa argumenten skickas med i händelsen - de tolkas först
        // när verktyget körs.
        let råa_argument = match &block {
            ContentBlock::ToolCall { arguments, .. } => arguments.clone(),
            _ => Value::Null,
        };

        handler(AgentEvent {
            typ: EventType::ToolStart,
            tool_call_id: id.to_string(),
            tool_name: namn.to_string(),
            args: Some(råa_argument),
            ..Default::default()
        })
        .await;

        let argument = tolka_verktygsargument(&mut block);

        let manager = crate::tools::cancel::global();
        let signal = manager.register(id);

        let (uppdateringssändare, mut uppdateringsmottagare) = tokio::sync::mpsc::channel(16);

        let anrop = ToolCall {
            id: id.to_string(),
            args: argument.clone(),
            signal: signal.clone(),
            on_update: Some(uppdateringssändare),
        };

        let verktyg = verktyg.iter().find(|verktyg| verktyg.name == namn).cloned();

        let resultat = match verktyg {
            None => ToolResult {
                content: vec![ContentBlock::Text {
                    text: format!("Tool '{namn}' not found"),
                }],
                is_error: true,
                ..Default::default()
            },
            Some(verktyg) => {
                // Argumenten valideras mot schemat.
                if let Err(fel) =
                    crate::tools::validator::validera_argument(&argument, &verktyg.input_schema)
                {
                    ToolResult {
                        content: vec![ContentBlock::Text {
                            text: fel.to_string(),
                        }],
                        is_error: true,
                        ..Default::default()
                    }
                } else {
                    // Resultatet hämtas medan uppdateringar skickas.
                    let mut utförande = tokio::spawn(async move { (verktyg.execute)(anrop).await });

                    let svar = loop {
                        tokio::select! {
                            resultat = &mut utförande => {
                                break resultat;
                            }
                            uppdatering = uppdateringsmottagare.recv() => {
                                let Some(uppdatering) = uppdatering else {
                                    continue;
                                };

                                handler(AgentEvent {
                                    typ: EventType::ToolUpdate,
                                    tool_call_id: id.to_string(),
                                    tool_name: namn.to_string(),
                                    result: Some(uppdatering),
                                    ..Default::default()
                                }).await;
                            }
                        }
                    };

                    // Ett fel i verktyget blir ett felresultat - inte
                    // ett fel i hela körningen.
                    match svar {
                        Ok(Ok(resultat)) => resultat,
                        Ok(Err(fel)) => ToolResult {
                            content: vec![ContentBlock::Text { text: fel }],
                            is_error: true,
                            ..Default::default()
                        },
                        Err(fel) => ToolResult {
                            content: vec![ContentBlock::Text {
                                text: fel.to_string(),
                            }],
                            is_error: true,
                            ..Default::default()
                        },
                    }
                }
            }
        };

        manager.unregister(id);

        handler(AgentEvent {
            typ: EventType::ToolEnd,
            tool_call_id: id.to_string(),
            tool_name: namn.to_string(),
            result: Some(resultat.clone()),
            is_error: resultat.is_error,
            ..Default::default()
        })
        .await;

        Message::ToolResult {
            tool_call_id: id.to_string(),
            tool_name: namn.to_string(),
            content: resultat.content,
            is_error: resultat.is_error,
            details: resultat.details,
            usage: resultat.usage,
            timestamp: crate::messages::nu(),
        }
    }
}

/// Hittar ett verktygsanrop i en innehållslista.
fn hitta_verktygsanrop<'a>(
    innehåll: &'a mut [ContentBlock],
    id: &str,
) -> Option<&'a mut ContentBlock> {
    innehåll.iter_mut().find(|block| match block {
        ContentBlock::ToolCall { id: anrops_id, .. } => anrops_id == id,
        _ => false,
    })
}

impl Message {
    /// Varför meddelandet slutade.
    pub fn stop_reason(&self) -> Option<&str> {
        match self {
            Message::Assistant { stop_reason, .. } => Some(stop_reason),
            _ => None,
        }
    }
}

/// Gör om assistentens innehåll till text - används av chatten.
pub fn assistent_text(meddelande: &Message) -> String {
    let (text, _) = konvertera_assistent_innehåll(meddelande.content());

    text.unwrap_or_default()
}

/// Typen på en leverantör - används av testerna.
pub fn leverantörstyp(p: &dyn LlmProvider) -> &str {
    p.name()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn händelsetyper_matchar_go() {
        assert_eq!(EventType::AgentStart.as_str(), "agent_start");
        assert_eq!(EventType::ToolStart.as_str(), "tool_execution_start");
        assert_eq!(EventType::ToolEnd.as_str(), "tool_execution_end");
    }

    #[test]
    fn stoppreason_lases_från_meddelandet() {
        let meddelande = Message::Assistant {
            content: Vec::new(),
            stop_reason: "end_turn".to_string(),
            timestamp: 0,
            usage: None,
        };

        assert_eq!(meddelande.stop_reason(), Some("end_turn"));
    }
}
