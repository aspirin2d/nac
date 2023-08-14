package template

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/gin-gonic/gin"
)

var (
	ErrAgentNotFound      = fmt.Errorf("agent not found")
	ErrTemplateNotMatched = fmt.Errorf("template not matched")
)

type Bootstrap struct {
	template *template.Template
}

func (temp *Bootstrap) Init(desc TemplateDesc) {
	t, err := template.New(desc.Name).Parse(desc.Content)
	if err != nil {
		panic(err)
	}
	temp.template = t
}

func (temp Bootstrap) Exec(c *gin.Context) (result string, err error) {
	var buf bytes.Buffer
	agent, exists := c.Get("agent")
	if !exists {
		return "", ErrAgentNotFound
	}

	err = temp.template.Execute(&buf, map[string]any{
		"Meta":  map[string]any{},
		"Agent": agent,
	})
	if err != nil {
		return
	}
	result = buf.String()
	return
}
