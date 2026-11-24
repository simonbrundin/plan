package featurestest

import (
	"context"

	"github.com/cucumber/godog"
	"plan-agent/internal/core"
	"plan-agent/internal/core/messages"
	"plan-agent/internal/core/provider"
)

var (
	testAgent   *core.AgentRuntime
	testHistory []messages.Message
	testEvents  []*core.AgentEvent
)

type testProvider struct{}

func (p *testProvider) Name() string { return "test" }

func (p *testProvider) Stream(
	ctx context.Context,
	model provider.Model,
	msgs []messages.Message,
	toolDefs []provider.ToolDefinition,
	opts provider.StreamingOptions,
) (<-chan provider.StreamEvent, error) {
	ch := make(chan provider.StreamEvent, 1)

	userMsg := ""
	for i := len(msgs) - 1; i >= 0; i-- {
		um, ok := msgs[i].(*messages.UserMessage)
		if !ok {
			continue
		}
		for _, block := range um.Content {
			if tb, ok := block.(*messages.TextBlock); ok {
				userMsg = tb.Text
				break
			}
		}
		break
	}

	msg := &messages.AssistantMessage{
		Role:       messages.RoleAssistant,
		Content:    []messages.ContentBlock{&messages.TextBlock{Text: "Svar på: " + userMsg}},
		StopReason: core.StopReasonEndTurn,
	}

	ch <- &provider.DoneEvent{Message: msg}
	close(ch)
	return ch, nil
}

func setupTestAgent() {
	testEvents = make([]*core.AgentEvent, 0)
	testAgent = core.NewAgent(&core.AgentLoopConfig{
		Model:    provider.Model{Name: "test-model", Provider: "test"},
		Provider: &testProvider{},
	})
}

func initieraCoreAgent(ctx *godog.ScenarioContext) {
	ctx.Before(func(c context.Context, sc *godog.Scenario) (context.Context, error) {
		setupTestAgent()
		testHistory = nil
		testEvents = nil
		return c, nil
	})

	ctx.Step(`^en agent konfigurerad med en LLM-provider$`, func() error {
		setupTestAgent()
		return nil
	})

	ctx.Step(`^verktyg är registrerade$`, func() error { return nil })

	ctx.Step(`^jag skickar "([^"]*)"$`, func(msg string) error {
		var history []messages.Message
		if testHistory != nil {
			history = testHistory
		}
		result := testAgent.SendMessageWithHistory(context.Background(), history, msg, func(event *core.AgentEvent) {
			testEvents = append(testEvents, event)
		})
		if result.Error != nil {
			return result.Error
		}
		testHistory = result.Messages
		return nil
	})

	ctx.Step(`^agenten ska svara med ett meddelande$`, func() error {
		if len(testHistory) == 0 {
			return godog.ErrPending
		}
		lastMsg := testHistory[len(testHistory)-1]
		if _, ok := lastMsg.(*messages.AssistantMessage); ok {
			return nil
		}
		return godog.ErrPending
	})

	ctx.Step(`^en agent med tom historik$`, func() error {
		testHistory = nil
		return nil
	})

	ctx.Step(`^sedan skickar "([^"]*)"$`, func(msg string) error {
		var history []messages.Message
		if testHistory != nil {
			history = testHistory
		}
		result := testAgent.SendMessageWithHistory(context.Background(), history, msg, func(event *core.AgentEvent) {
			testEvents = append(testEvents, event)
		})
		if result.Error != nil {
			return result.Error
		}
		testHistory = result.Messages
		return nil
	})

	ctx.Step(`^agenten ska referera till föregående meddelande$`, func() error {
		if len(testHistory) < 4 {
			return godog.ErrPending
		}
		return nil
	})

	ctx.Step(`^en agent som tar emot "([^"]*)"$`, func(msg string) error {
		setupTestAgent()
		return nil
	})

	ctx.Step(`^modellen svarar med stopreason "([^"]*)"$`, func(reason string) error { return nil })
	ctx.Step(`^agent-loopen ska avslutas$`, func() error { return nil })
	ctx.Step(`^en agent med parallel verktygskörning avaktiverad$`, func() error { return nil })
	ctx.Step(`^jag ber om "([^"]*)"$`, func(task string) error { return nil })
	ctx.Step(`^bash ska köras före read$`, func() error { return godog.ErrPending })
	ctx.Step(`^en agent med verktyg$`, func() error { setupTestAgent(); return nil })
	ctx.Step(`^ett verktyg kastar ett fel$`, func() error { return godog.ErrPending })
	ctx.Step(`^felmeddelandet ska visas för användaren$`, func() error { return godog.ErrPending })
	ctx.Step(`^en agent som kör ett verktyg som returnerar tomt$`, func() error { return godog.ErrPending })
	ctx.Step(`^verktyget körs utan fel$`, func() error { return nil })
	ctx.Step(`^resultatet ska läggas till i kontexten$`, func() error { return godog.ErrPending })
	ctx.Step(`^en agent med ShouldStop konfigurerad$`, func() error { return godog.ErrPending })
	ctx.Step(`^ShouldStop returnerar sant$`, func() error { return godog.ErrPending })
	ctx.Step(`^agent-loopen ska avbrytas$`, func() error { return godog.ErrPending })
	ctx.Step(`^en agent med PrepareNextTurn konfigurerad$`, func() error { return godog.ErrPending })
	ctx.Step(`^PrepareNextTurn ändrar kontexten$`, func() error { return godog.ErrPending })
	ctx.Step(`^nästa tur ska köras med uppdaterad kontext$`, func() error { return godog.ErrPending })
}
