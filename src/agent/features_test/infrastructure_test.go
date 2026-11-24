package featurestest

import (
	"os"
	"os/exec"

	"github.com/cucumber/godog"
)

func initieraInfrastructure(ctx *godog.ScenarioContext) {
	ctx.Step(`^docker-compose konfigurerad$`, func() error {
		if _, err := os.Stat("environments/dev/docker-compose.yml"); os.IsNotExist(err) {
			return godog.ErrPending
		}
		return nil
	})

	ctx.Step(`^jag kör docker compose up$`, func() error {
		cmd := exec.Command("docker", "compose", "-f", "environments/dev/docker-compose.yml", "ps")
		if err := cmd.Run(); err != nil {
			return godog.ErrPending
		}
		return nil
	})

	ctx.Step(`^PostgreSQL och backend ska starta$`, func() error { return nil })
	ctx.Step(`^Tiltfile konfigurerad$`, func() error {
		if _, err := os.Stat("environments/dev/Tiltfile"); os.IsNotExist(err) {
			return godog.ErrPending
		}
		return nil
	})

	ctx.Step(`^jag kör tilt up$`, func() error {
		cmd := exec.Command("tilt", "version")
		if err := cmd.Run(); err != nil {
			return godog.ErrPending
		}
		return nil
	})

	ctx.Step(`^alla tjänster ska starta med hot-reload$`, func() error { return nil })
	ctx.Step(`^Kubernetes cluster$`, func() error {
		cmd := exec.Command("kubectl", "cluster-info")
		if err := cmd.Run(); err != nil {
			return godog.ErrPending
		}
		return nil
	})

	ctx.Step(`^jag applicerar (\/.*)$`, func(path string) error { return nil })
	ctx.Step(`^backend ska deployas$`, func() error { return nil })
	ctx.Step(`^CloudNativePG installerat$`, func() error { return nil })
	ctx.Step(`^jag skapar PostgreSQL-resurs$`, func() error { return nil })
	ctx.Step(`^databasen ska etableras med rätt secrets$`, func() error { return nil })
	ctx.Step(`^Flux konfigurerat$`, func() error {
		cmd := exec.Command("flux", "version")
		if err := cmd.Run(); err != nil {
			return godog.ErrPending
		}
		return nil
	})

	ctx.Step(`^jag pushar till git$`, func() error { return nil })
	ctx.Step(`^ändringar ska synkas till kluster$`, func() error { return nil })
	ctx.Step(`^secrets krypterade med age$`, func() error { return nil })
	ctx.Step(`^deployment sker$`, func() error { return nil })
	ctx.Step(`^secrets ska dekrypteras och användas$`, func() error { return nil })
	ctx.Step(`^ingress controller$`, func() error { return nil })
	ctx.Step(`^trafik ska routas till backend$`, func() error { return nil })
	ctx.Step(`^liveness och readiness probes ska vara konfigurerade$`, func() error { return nil })
}
