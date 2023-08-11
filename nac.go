package nac

import (
	"bytes"
)

type Nac struct {
	tc *TemplateConfig // template config
}

func (n *Nac) Prompt(agent Agent, meta any) (prompt string, err error) {
	var buf bytes.Buffer
	err = n.tc.Template.Execute(&buf, map[string]any{"Agent": agent, "Meta": meta})
	return buf.String(), err
}
