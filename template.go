package nac

import (
	"html/template"

	"github.com/BurntSushi/toml"
)

type Query struct {
	Id      int    `toml:"id"`
	Name    string `toml:"name"`
	Content string `toml:"content"`
}

type TemplateConfig struct {
	Interference string
	Queries      []Query

	Template *template.Template
}

func LoadTemplates(path string) TemplateConfig {
	var config TemplateConfig
	_, err := toml.DecodeFile(path, &config)
	if err != nil {
		panic(err)
	}

	config.Template, err = template.New("interference").Parse(config.Interference)
	if err != nil {
		panic(err)
	}

	for _, q := range config.Queries {
		_, err = config.Template.New(q.Name).Parse(q.Content)
		if err != nil {
			panic(err)
		}
	}
	return config
}
