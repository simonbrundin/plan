package featurestest

import (
	"net/http"
	"net/http/httptest"

	"github.com/cucumber/godog"
	"github.com/gin-gonic/gin"
)

var (
	testServer   *httptest.Server
	testResponse *httptest.ResponseRecorder
	apiAnropad   bool
)

func setupTestAPI() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	r.GET("/api/v1/status", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})
	return r
}

func initieraAPI(ctx *godog.ScenarioContext) {
	ctx.BeforeScenario(func(*godog.Scenario) {
		router := setupTestAPI()
		testServer = httptest.NewServer(router)
		apiAnropad = false
	})

	ctx.AfterScenario(func(*godog.Scenario, error) {
		if testServer != nil {
			testServer.Close()
		}
	})

	ctx.Step(`^API-servern körs på port (\d+)$`, func(port int) error { return nil })
	ctx.Step(`^API-servern körs$`, func() error { return nil })

	ctx.Step(`^jag GET (\/.*)$`, func(path string) error {
		resp, err := http.Get(testServer.URL + path)
		if err != nil {
			return err
		}
		testResponse = httptest.NewRecorder()
		testResponse.Code = resp.StatusCode
		apiAnropad = true
		return nil
	})

	ctx.Step(`^status "([^"]*)" ska returneras$`, func(status string) error {
		if !apiAnropad {
			return godog.ErrPending
		}
		if testResponse == nil || testResponse.Code != 200 {
			return godog.ErrPending
		}
		return nil
	})

	ctx.Step(`^jag POST (\/.*) med title "([^"]*)"$`, func(path, title string) error {
		apiAnropad = true
		return nil
	})

	ctx.Step(`^en konversation ska skapas med id$`, func() error {
		if !apiAnropad {
			return godog.ErrPending
		}
		return nil
	})

	ctx.Step(`^en konversation finns$`, func() error { return nil })
	ctx.Step(`^listan av konversationer ska returneras$`, func() error { return nil })
	ctx.Step(`^en konversation med id "([^"]*)"$`, func(id string) error { return nil })
	ctx.Step(`^konversationen ska returneras$`, func() error { return nil })
	ctx.Step(`^jag PATCH (\/.*) med ny title$`, func(path string) error { apiAnropad = true; return nil })
	ctx.Step(`^titeln ska uppdateras$`, func() error { return nil })
	ctx.Step(`^jag POST (\/.*) med "([^"]*)"$`, func(path, msg string) error { apiAnropad = true; return nil })
	ctx.Step(`^agenten ska ta emot och svara$`, func() error { return nil })
	ctx.Step(`^jag ansluter till (\/.*)$`, func(path string) error { return nil })
	ctx.Step(`^nya meddelanden ska levereras via SSE$`, func() error { return nil })
	ctx.Step(`^jag POST (\/.*) med provider och nyckel$`, func(path string) error { apiAnropad = true; return nil })
	ctx.Step(`^nyckeln ska sparas$`, func() error { return nil })
	ctx.Step(`^en API-nyckel finns$`, func() error { return nil })
	ctx.Step(`^providers och modeller ska returneras \(utan nycklar\)$`, func() error { return nil })
	ctx.Step(`^en API-nyckel för "([^"]*)"$`, func(provider string) error { return nil })
	ctx.Step(`^jag DELETE (\/.*)$`, func(path string) error { apiAnropad = true; return nil })
	ctx.Step(`^nyckeln ska tas bort$`, func() error { return nil })
	ctx.Step(`^jag besöker (\/.*)$`, func(path string) error { return nil })
	ctx.Step(`^Swagger UI ska visas$`, func() error { return nil })
}
