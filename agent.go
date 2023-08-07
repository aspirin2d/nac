package nac

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AgentType struct {
	Id   uint8    `toml:"id" bson:"_id"`
	Name string   `toml:"name" bson:"name"`
	Desc []string `toml:"desc" bson:"desc"`
}

func (at AgentType) NewAgent() Agent {
	return Agent{AgentType: at.Id, Created: time.Now(), Id: primitive.NewObjectID()}
}

type Agent struct {
	Id        primitive.ObjectID `bson:"_id" json:"id"`
	AgentType uint8              `bson:"agent_type" json:"agent_type"`
	Created   time.Time          `bson:"created" json:"created"`
}

// NewAgent creates new agent from url params:
// POST "/u/:uid/a/add/:tid"
func (n *Nac) NewAgent(c *gin.Context) {
	ctx := c.Request.Context()
	uid := c.MustGet("uid").(primitive.ObjectID)
	agentType := c.MustGet("agent_type").(uint8)

	// get agent_type by "agent_type"
	var at AgentType
	err := n.mdb.Collection(mdb_agent_types).FindOne(ctx, bson.M{"_id": agentType}).Decode(&at)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.AbortWithError(400, fmt.Errorf("agent type not found")).SetType(gin.ErrorTypePublic)
		} else {
			c.AbortWithError(500, err)
		}
		return
	}

	// insert new agent
	agent := at.NewAgent()
	filter := bson.M{"_id": uid, "agents.agent_type": bson.M{"$ne": agentType}}
	res, err := n.mdb.Collection(mdb_users).UpdateOne(ctx, filter, bson.M{"$push": bson.M{
		"agents": agent,
	}})

	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	if res.MatchedCount == 0 {
		c.AbortWithError(400, fmt.Errorf("user not found, or agent type duplicated")).SetType(gin.ErrorTypePublic)
		return
	}

	c.JSON(200, gin.H{"id": agent.Id})
}

// GetAgentTypes get agent type with the given tid
func (n *Nac) GetAgentType(c *gin.Context) {
	tid, err := strconv.ParseUint(c.Param("tid"), 10, 8)
	if err != nil {
		c.AbortWithError(400, fmt.Errorf("invalid agent type, should be uint8, but got %v", tid)).SetType(gin.ErrorTypePublic)
		return
	}
	c.Set("agent_type", uint8(tid))
}
