package template

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

var (
	registry = map[string]Template{
		"bootstrap": &Bootstrap{},
	}
)

type Template interface {
	Init(desc TemplateDesc)
	Exec(c *gin.Context) (result string, err error)
}

type TemplateDesc struct {
	Id      int    `toml:"id"`
	Name    string `toml:"name"`
	Content string `toml:"content"`
}

func (td TemplateDesc) Check() {
	temp, ok := registry[td.Name]
	if !ok {
		panic(fmt.Errorf("template not registered: %s", td.Name))
	}
	temp.Init(td)
}
