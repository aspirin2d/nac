package nac

type Agent interface {
	Name() string
	Description() string
	QueryId() int
}

type AgentConfig struct {
	Agents []Agent `toml:"agents"`
}
