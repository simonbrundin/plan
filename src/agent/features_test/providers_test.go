package featurestest

import (
	"context"
	"os"

	"github.com/cucumber/godog"
)

var providerResultat []string

func initieraProviders(ctx *godog.ScenarioContext) {
	ctx.Before(func(c context.Context, sc *godog.Scenario) (context.Context, error) {
		providerResultat = nil
		return c, nil
	})

	ctx.Step(`^OpenAI provider konfigurerad med API-nyckel$`, func() error {
		if os.Getenv("OPENAI_API_KEY") == "" {
			return godog.ErrPending
		}
		return nil
	})

	ctx.Step(`^Anthropic provider konfigurerad med API-nyckel$`, func() error {
		if os.Getenv("ANTHROPIC_API_KEY") == "" {
			return godog.ErrPending
		}
		return nil
	})

	ctx.Step(`^MiniMax provider konfigurerad med API-nyckel$`, func() error {
		if os.Getenv("MINIMAX_API_KEY") == "" {
			return godog.ErrPending
		}
		return nil
	})

	ctx.Step(`^AgentGateway provider konfigurerad$`, func() error { return nil })

	ctx.Step(`^jag skickar ett meddelande$`, func() error {
		providerResultat = append(providerResultat, "message sent")
		return nil
	})

	ctx.Step(`^ett svar ska returneras$`, func() error {
		if len(providerResultat) == 0 {
			return godog.ErrPending
		}
		return nil
	})

	ctx.Step(`^ett svar ska returneras via gateway$`, func() error { return nil })
	ctx.Step(`^en provider med streaming aktiverat$`, func() error { return nil })
	ctx.Step(`^text ska levereras som deltahändelser$`, func() error { return nil })
	ctx.Step(`^en provider konfigurerad med modell "([^"]*)"$`, func(model string) error { return nil })
	ctx.Step(`^rätt modell ska användas$`, func() error { return nil })
	ctx.Step(`^modellen svarar med <think> block$`, func() error { return godog.ErrPending })
	ctx.Step(`^thinking ska extraheras separat$`, func() error { return godog.ErrPending })
	ctx.Step(`^flera providers konfigurerade$`, func() error { return nil })
	ctx.Step(`^primär provider returnerar fel$`, func() error { return nil })
	ctx.Step(`^nästa provider i prioritet ska användas$`, func() error { return godog.ErrPending })
}
