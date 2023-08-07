package nac

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Memory struct {
	Content string    `json:"content" bson:"content"`
	Created time.Time `json:"created,omitempty" bson:"created,omitempty"`
}

// AddMemory adds memory to user's redis memory list
// POST "/u/:uid/a/:aid/m/add"
func (n *Nac) AddMemories(c *gin.Context) {
	ctx := c.Request.Context()
	uid := c.MustGet("uid").(primitive.ObjectID)
	aid := c.MustGet("aid").(primitive.ObjectID)

	var mems []Memory
	err := c.BindJSON(&mems)
	if err != nil {
		return // bind error already handled by gin
	}

	// encodeing memory to bson
	var bs []interface{}
	for _, m := range mems {
		bytes, err := bson.Marshal(m)
		if err != nil {
			return
		}
		bs = append(bs, bytes)
	}

	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	// push to redis agent memory list
	key := "m:" + uid.Hex() + ":" + aid.Hex()
	_, err = n.rdb.LPush(ctx, key, bs...).Result() // newest first
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	// trim memory list size with the config
	_, err = n.rdb.LTrim(ctx, key, 0, int64(n.Config().MemoryLimit)).Result()
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
}

// GetMemories gets memories from user's agent
// POST "/u/:uid/a/:aid/m"
func (n *Nac) GetMemories(c *gin.Context) {
	ctx := c.Request.Context()
	uid := c.MustGet("uid").(primitive.ObjectID)
	aid := c.MustGet("aid").(primitive.ObjectID)
	mems, err := n.rdb.LRange(ctx, "m:"+uid.Hex()+":"+aid.Hex(), 0, -1).Result()
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	var res []Memory
	for _, m := range mems {
		var mem Memory
		err := bson.Unmarshal([]byte(m), &mem)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		res = append(res, mem)
	}
	c.JSON(200, res)
}
