package nac

import (
	"text/template"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/robfig/cron/v3"
)

var (
	cronParser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
)

type CompiledTemplate struct {
	*template.Template
}

func (t *CompiledTemplate) UnmarshalText(text []byte) error {
	compiled, err := template.New("template").Parse(string(text))
	if err != nil {
		return err
	}
	t.Template = compiled
	return nil
}

type CompiledSchedule struct {
	Schedule cron.Schedule
}

func (s *CompiledSchedule) UnmarshalText(text []byte) error {
	schedule, err := cronParser.Parse(string(text))
	if err != nil {
		return err
	}
	s.Schedule = schedule
	return nil
}

type Agent struct {
	Id int `toml:"id" bson:"_id"`
	// agents Name
	Name string `toml:"name" bson:"name"`
	// birthday of the agent, example: "1987-07-05T00:00:00Z"
	Birthday time.Time `toml:"birthday" bson:"birthday"`
	// initial description and memories
	Descriptions []string `toml:"descriptions" bson:"descriptions"`
	// templates id that the agent could use
	Templates []int `toml:"templates" bson:"templates"`
}

type Template struct {
	Id int `toml:"id" bson:"_id"`
	// template content
	Compiled *CompiledTemplate `toml:"content" bson:"content"`
	// Crontab
	Crontab *CompiledSchedule `toml:"crontab,omitempty" bson:"crontab,omitempty"`
}

type AgentsConfig struct {
	Agents []Agent `toml:"agents" bson:"agents"`
}

type TemplatesConfig struct {
	Templates []Template `toml:"templates" bson:"templates"`
}

// MustLoadAgents loads the agents config from the given path
// if failed, it will panic
func MustLoadAgents(path string) AgentsConfig {
	var config AgentsConfig
	_, err := toml.DecodeFile(path, &config)
	if err != nil {
		panic(err)
	}
	return config
}

func MustLoadTemplates(path string) TemplatesConfig {
	var templates TemplatesConfig
	_, err := toml.DecodeFile(path, &templates)
	if err != nil {
		panic(err)
	}
	return templates
}

type TemplateInput struct {
	Now          time.Time      // current time
	UserName     string         // username
	Background   string         // background desc
	Descriptions []string       // agents' descriptions list
	Memories     []string       // agents' memories list
	Meta         map[string]any // additional meta data
}
