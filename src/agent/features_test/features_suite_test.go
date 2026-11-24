package featurestest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
	execDir, _ := os.Getwd()
	agentRoot := filepath.Dir(execDir)
	featuresPath := filepath.Join(agentRoot, ".features")

	suite := godog.TestSuite{
		ScenarioInitializer: initieraScenarios,
		Options: &godog.Options{
			Format:   "progress",
			Paths:    []string{featuresPath},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("Godog test suite misslyckades")
	}
}
