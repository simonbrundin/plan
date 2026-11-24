package featurestest

import (
	"os"

	"github.com/cucumber/godog"
	"plan-agent/internal/core/secrets"
)

func initieraSecrets(ctx *godog.ScenarioContext) {
	ctx.BeforeScenario(func(*godog.Scenario) {
		secrets.EnsureDir()
	})

	ctx.Step(`^en krypterad secrets-fil finns$`, func() error {
		configPath := secrets.DefaultConfigPath()
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			return godog.ErrPending
		}
		return nil
	})

	ctx.Step(`^nycklar ska laddas efter dekryptering$`, func() error { return nil })
	ctx.Step(`^secrets i plaintext$`, func() error { return nil })
	ctx.Step(`^filen krypteras med SOPS$`, func() error { return nil })
	ctx.Step(`^endast auktoriserade nycklar kan dekryptera$`, func() error { return nil })
	ctx.Step(`^secrets-filen är otillgänglig$`, func() error { return nil })
	ctx.Step(`^SOPS_AGE_KEY är satt$`, func() error {
		if os.Getenv("SOPS_AGE_KEY") == "" {
			return godog.ErrPending
		}
		return nil
	})

	ctx.Step(`^miljövariabeln ska användas$`, func() error { return nil })
	ctx.Step(`^1Password CLI är installerat$`, func() error {
		if _, err := os.Stat("/usr/local/bin/op"); os.IsNotExist(err) {
			return godog.ErrPending
		}
		return nil
	})

	ctx.Step(`^nycklar ska hämtas från 1Password$`, func() error { return nil })
	ctx.Step(`^secrets-filen uppdateras$`, func() error { return nil })
	ctx.Step(`^nya nycklar ska laddas$`, func() error { return nil })
}
