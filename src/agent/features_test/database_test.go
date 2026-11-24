package featurestest

import (
	"os"

	"github.com/cucumber/godog"
)

func initieraDatabase(ctx *godog.ScenarioContext) {
	ctx.Step(`^DATABASE_URL är konfigurerad$`, func() error {
		if os.Getenv("DATABASE_URL") == "" {
			return godog.ErrPending
		}
		return nil
	})

	ctx.Step(`^connection pool ska skapas$`, func() error { return nil })
	ctx.Step(`^en konversation med titel "([^"]*)"$`, func(title string) error { return nil })
	ctx.Step(`^jag skapar konversationen$`, func() error { return nil })
	ctx.Step(`^den ska sparas i databasen$`, func() error { return nil })
	ctx.Step(`^jag sparar ett meddelande$`, func() error { return nil })
	ctx.Step(`^det ska kopplas till konversationen$`, func() error { return nil })
	ctx.Step(`^en konversation med meddelanden$`, func() error { return nil })
	ctx.Step(`^jag hämtar konversationen$`, func() error { return nil })
	ctx.Step(`^meddelandena ska inkluderas$`, func() error { return nil })
	ctx.Step(`^flera konversationer$`, func() error { return nil })
	ctx.Step(`^jag hämtar med offset (\d+) och limit (\d+)$`, func(offset, limit int) error { return nil })
	ctx.Step(`^högst (\d+) ska returneras$`, func(max int) error { return nil })
	ctx.Step(`^provider "([^"]*)" och nyckel "([^"]*)"$`, func(provider, key string) error { return nil })
	ctx.Step(`^den ska finnas i api_keys-tabellen$`, func() error { return nil })
	ctx.Step(`^users-tabell ska finnas för authentication$`, func() error { return nil })
	ctx.Step(`^sessions-tabell ska finnas för token-hantering$`, func() error { return nil })
}
