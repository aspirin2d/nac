// nac is short for Not Animal Crossing.
package nac

import "github.com/BurntSushi/toml"

type Config struct {
	OpenAIAPIKey string `toml:"openai_api_key"`

	MongoUri string `toml:"mongo_uri"`
	MongoDb  string `toml:"mongo_db"`

	RedisUri string `toml:"redis_uri"`
}

// Init initializes the nac package with the given config file.
func Init(path string) {
	var config Config = Config{
		MongoUri: "mongodb://localhost:27017",
		MongoDb:  "memo",
	}
	var err error

	_, err = toml.DecodeFile(path, &config)
	if err != nil {
		panic(err)
	}
}
