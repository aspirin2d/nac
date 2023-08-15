package nac

import (
	"net/http/httptest"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"github.com/sleep2death/nac/template"
	"github.com/stretchr/testify/assert"
)

type TemplatesConfig struct {
	Templates []template.TemplateDesc `toml:"templates"`
}

type AgentsConfig struct {
	Agents []AgentDesc `toml:"agents"`
}

func TestLoadTemplate(t *testing.T) {
	var tConf TemplatesConfig
	_, err := toml.DecodeFile("./config/templates.toml", &tConf)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(tConf.Templates))

	for _, t := range tConf.Templates {
		t.Check()
	}

	var aConf AgentsConfig
	_, err = toml.DecodeFile("./config/agents.toml", &aConf)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(aConf.Agents))

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ag := aConf.Agents[0].NewAgent()
	ag.Step(ctx)
}
