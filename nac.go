// nac is short for Not Animal Crossing.
package nac

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"github.com/nitishm/go-rejson/v4"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

const mdb_users = "users"
const mdb_agent_types = "agent_types"

type Config struct {
	OpenAIAPIKey string `toml:"openai_api_key"`

	MongoUri string `toml:"mongo_uri"`
	MongoDb  string `toml:"mongo_db"`

	RedisUri string `toml:"redis_uri"`
	RedisDb  int    `toml:"redis_db"`

	MemoryLimit int `toml:"memory_limit"`

	Agents []Agent `toml:"agents" bson:"agents"`
}

type Nac struct {
	mdb    *mongo.Database
	rdb    *redis.Client
	rejson *rejson.Handler
	logger *zap.SugaredLogger
	config *Config
}

// FromConfig initializes the nac package with the given config file.
func FromConfig(path string) *Nac {
	var config Config = Config{
		MongoUri: "mongodb://localhost:27017",
		MongoDb:  "nac",

		RedisUri: "redis://localhost:6379",
		RedisDb:  0,

		MemoryLimit: 100,
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
	mdb := mongoClient(ctx, config).Database(config.MongoDb)

	// setup wx_id index and username for user collection, and make them unique
	mdb.Collection(mdb_users).Indexes().CreateMany(ctx, []mongo.IndexModel{
		// {Keys: bson.M{"wx_id": 1}, Options: options.Index().SetUnique(true)},
		{Keys: bson.M{"username": 1}, Options: options.Index().SetUnique(true)},
	})

	// save agent types to mongodb
	var models []mongo.WriteModel
	for _, ag := range config.Agents {
		models = append(models, mongo.NewUpdateOneModel().SetFilter(bson.M{"_id": ag.Id}).SetUpdate(bson.M{"$set": ag}).SetUpsert(true))
	}

	_, err = mdb.Collection(mdb_agent_types).BulkWrite(ctx, models)
	if err != nil {
		panic(err)
	}

	// connect to redis
	rdb := redisClient(ctx, config)

	// init logger
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()

	return &Nac{logger: sugar, mdb: mdb, rdb: rdb, config: &config}
}

func mongoClient(ctx context.Context, config Config) *mongo.Client {
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

func redisClient(ctx context.Context, config Config) *redis.Client {
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

func (n *Nac) Config() *Config {
	return n.config
}

func (n *Nac) ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		// only check for the first error if there are any
		if len(c.Errors) > 0 {
			last := c.Errors.Last()
			if last.IsType(gin.ErrorTypePrivate) {
				c.JSON(500, gin.H{"msg": "internal server error"})
				if n.logger != nil {
					n.logger.Error(c.Errors.JSON())
				}
			} else if last.IsType(gin.ErrorTypeBind) {
				c.JSON(c.Writer.Status(), gin.H{"msg": fmt.Sprintf("binding error: %s", last.Error())})
			} else if last.IsType(gin.ErrorTypeRender) {
				c.JSON(c.Writer.Status(), gin.H{"msg": fmt.Sprintf("rendering error: %s", last.Error())})
			} else {
				c.JSON(c.Writer.Status(), gin.H{"msg": last.Error()})
			}
		}
	}
}

// ping endpoint for health checks
func (n *Nac) Ping(c *gin.Context) {
	c.String(200, "pong")
}

func (n *Nac) ClearData(ctx context.Context) {
	err := n.mdb.Drop(ctx)
	if err != nil {
		panic(err)
	}

	err = n.rdb.FlushAll(ctx).Err()
	if err != nil {
		panic(err)
	}
}

func (n *Nac) OK(c *gin.Context) {
	c.JSON(200, gin.H{"msg": "ok"})
}

func (n *Nac) GetObjectId(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		oid, err := primitive.ObjectIDFromHex(c.Param(name))
		if err != nil {
			c.AbortWithError(400, fmt.Errorf("invalid object id: %v", oid)).SetType(gin.ErrorTypePublic)
			return
		}
		c.Set(name, oid)
	}
}
