package nac

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sleep2death/nac/template"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AgentDesc struct {
	Id           int      `toml:"id" bson:"_id"`
	Name         string   `toml:"name" bson:"name"`
	Descriptions []string `toml:"descriptions" bson:"descriptions"`
	Templates    [][]int  `toml:"templates" bson:"templates"`
}

func (ad AgentDesc) Step(c *gin.Context) (string, error) {
	return "", template.ErrTemplateNotMatched
}

func (at AgentDesc) NewAgent() Agent {
	return Agent{AgentDesc: at.Id, Created: time.Now(), Id: primitive.NewObjectID()}
}

type Agent struct {
	Id        primitive.ObjectID `bson:"_id" json:"id"`
	AgentDesc int                `bson:"agent_type" json:"agent_type"`
	Created   time.Time          `bson:"created" json:"created"`
}

// Step is the main function of an agent.
// "/u/:uid/:aid/"
func (a *Agent) Step(ctx *gin.Context) {
}

// GET "/u/:uid/step"
func (n *Nac) Step(c *gin.Context) {
	ctx := c.Request.Context()
	uid := c.MustGet("uid").(primitive.ObjectID)

	var res string
	var err error

	for _, a := range n.config.Agents {
		res, err = a.Step(c)
		// if not found, continue for next agent
		if err == template.ErrTemplateNotMatched {
			continue
		}
		// if error, abort
		if err != nil && err != template.ErrTemplateNotMatched {
			c.AbortWithError(500, err) // nolint: errcheck
			return
		}
		// finally found a matched template
		err = n.rdb.Set(ctx, fmt.Sprintf("q:%s:%s", uid, a.Name), res, 0).Err()
		if err != nil {
			c.AbortWithError(500, err) // nolint: errcheck
			return
		}
		// we can call this api multiple times to get all available templates
		break
	}

	if err == template.ErrTemplateNotMatched {
		c.AbortWithStatusJSON(200, gin.H{"msg": "template not found"})
	}
}

// NewAgent creates new agent from url params:
// POST "/u/:uid/a/add/:tid"
func (n *Nac) NewAgent(c *gin.Context) {
	ctx := c.Request.Context()
	uid := c.MustGet("uid").(primitive.ObjectID)
	agentType := c.MustGet("agent_type").(uint8)

	// get agent_type by "agent_type"
	var at AgentDesc
	err := n.mdb.Collection(mdb_agent_types).FindOne(ctx, bson.M{"_id": agentType}).Decode(&at)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.AbortWithError(400, fmt.Errorf("agent type not found")).SetType(gin.ErrorTypePublic) // nolint: errcheck
		} else {
			c.AbortWithError(500, err) // nolint: errcheck
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
		c.AbortWithError(500, err) // nolint: errcheck
		return
	}

	if res.MatchedCount == 0 {
		c.AbortWithError(400, fmt.Errorf("user not found, or agent type duplicated")).SetType(gin.ErrorTypePublic) // nolint: errcheck
		return
	}

	c.JSON(200, gin.H{"id": agent.Id})
}

// GetAgentTypes get agent type with the given tid
func (n *Nac) GetAgentType(c *gin.Context) {
	tid, err := strconv.ParseUint(c.Param("tid"), 10, 8)
	if err != nil {
		c.AbortWithError(400, fmt.Errorf("invalid agent type, should be uint8, but got %v", tid)).SetType(gin.ErrorTypePublic) // nolint: errcheck
		return
	}
	c.Set("agent_type", uint8(tid))
}
