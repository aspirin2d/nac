package nac

import (
	"time"
)

type Agent struct {
	Id   int    `toml:"id" bson:"_id"`
	Name string `toml:"name" bson:"name"`

	// Birthday of the agent, example: "1987-07-05T05:45:00Z"
	Birthday time.Time `toml:"birthday" bson:"birthday"`

	// initial description and memories
	Desc []string `toml:"desc" bson:"desc"`

	// posts that the agent could make
	Posts []string `toml:"posts" bson:"posts"`
}
