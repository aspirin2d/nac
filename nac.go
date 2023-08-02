// nac is short for Not Animal Crossing.
package nac

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

const mdb_users = "users"

type Config struct {
	OpenAIAPIKey string `toml:"openai_api_key"`

	MongoUri string `toml:"mongo_uri"`
	MongoDb  string `toml:"mongo_db"`

	RedisUri string `toml:"redis_uri"`
	RedisDb  int
}

type Nac struct {
	mdb    *mongo.Database
	rdb    *redis.Client
	logger *zap.SugaredLogger
}

// FromConfig initializes the nac package with the given config file.
func FromConfig(path string) *Nac {
	var config Config = Config{
		MongoUri: "mongodb://localhost:27017",
		MongoDb:  "nac",

		RedisUri: "redis://localhost:6379",
		RedisDb:  0,
	}

	// read config file
	_, err := toml.DecodeFile(path, &config)
	if err != nil {
		panic(err)
	}

	if config.OpenAIAPIKey == "" {
		panic(errors.New("OpenAI API key not set"))
	}

	ctx := context.Background()

	// connect to mongodb
	mdb := mongoClient(ctx, &config).Database(config.MongoDb)

	// setup wx_id index and username for user collection, and make them unique
	mdb.Collection(mdb_users).Indexes().CreateMany(ctx, []mongo.IndexModel{
		// {Keys: bson.M{"wx_id": 1}, Options: options.Index().SetUnique(true)},
		{Keys: bson.M{"username": 1}, Options: options.Index().SetUnique(true)},
	})

	// connect to redis
	rdb := redisClient(ctx, &config)

	// init logger
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()

	return &Nac{logger: sugar, mdb: mdb, rdb: rdb}
}

func mongoClient(ctx context.Context, config *Config) *mongo.Client {
	mc, err := mongo.Connect(ctx, options.Client().ApplyURI(config.MongoUri))
	if err != nil {
		panic(err)
	}

	err = mc.Ping(ctx, nil)
	if err != nil {
		panic(err)
	}

	log.Println("connected to mongo at:", config.MongoUri)
	return mc
}

func redisClient(ctx context.Context, config *Config) *redis.Client {
	opts, err := redis.ParseURL(config.RedisUri)
	if err != nil {
		panic(err)
	}

	opts.DB = config.RedisDb

	rc := redis.NewClient(opts)
	err = rc.Ping(ctx).Err()
	if err != nil {
		panic(err)
	}

	log.Println("connected to redis at:", config.RedisUri)
	return rc
}

func (n *Nac) ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		// only check for the first error if there are any
		if len(c.Errors) > 0 {
			last := c.Errors.Last()
			if last.IsType(gin.ErrorTypePrivate) {
				c.JSON(500, gin.H{"msg": "internal server error"})
			} else if last.IsType(gin.ErrorTypeBind) {
				c.JSON(c.Writer.Status(), gin.H{"msg": fmt.Sprintf("binding error: %s", last.Error())})
			} else if last.IsType(gin.ErrorTypeRender) {
				c.JSON(c.Writer.Status(), gin.H{"msg": fmt.Sprintf("rendering error: %s", last.Error())})
			} else {
				c.JSON(c.Writer.Status(), gin.H{"msg": last.Error()})
			}

			if n.logger != nil {
				n.logger.Error(c.Errors.JSON())
			}
		}
	}
}

// ping endpoint for health checks
func (n *Nac) Ping(c *gin.Context) {
	c.String(200, "pong")
}
