package nac

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadAgentsConfig(t *testing.T) {
	config := MustLoadAgents("testdata/agents.toml")

	agents := config.Agents
	assert.Equal(t, len(agents), 3)

	agent := agents[0]
	_, _, day := agent.Birthday.Date() // test date parsing
	assert.Equal(t, 17, day)
	assert.Equal(t, 7, len(agent.Descriptions))
}

func TestLoadTemplatesFromConfig(t *testing.T) {
	config := MustLoadTemplates("testdata/templates.toml")
	templates := config.Templates

	assert.Equal(t, len(templates), 2)
	assert.Equal(t, templates[0].Compiled.Name(), "template")
}
