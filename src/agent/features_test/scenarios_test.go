package featurestest

import (
	"github.com/cucumber/godog"
)

func initieraScenarios(ctx *godog.ScenarioContext) {
	initieraCoreAgent(ctx)
	initieraTools(ctx)
	initieraProviders(ctx)
	initieraAPI(ctx)
	initieraFrontend(ctx)
	initieraSecrets(ctx)
	initieraDatabase(ctx)
	initieraInfrastructure(ctx)
}
