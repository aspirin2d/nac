package nac

import (
	"errors"
	"fmt"
	"time"

	"github.com/dlclark/regexp2"
	redis "github.com/redis/go-redis/v9"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var usernameReg = regexp2.MustCompile("^(?=[a-zA-Z0-9._]{6,20}$)(?!.*[_.]{2})[^_.].*[^_.]$", 0)

type User struct {
	Id       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Username string             `bson:"username" json:"username"`
	Created  time.Time          `bson:"created" json:"created"`

	AddedTypes []int   `bson:"added_types" json:"added_types"`
	Agents     []Agent `bson:"agents" json:"agents"`
}

// BindUser will add an user to the database.
func (n *Nac) InsertUser(c *gin.Context) {
	usr := c.MustGet("user").(*User)

	ctx := c.Request.Context()
	res, err := n.mdb.Collection(mdb_users).InsertOne(ctx, usr)
	if err != nil {
		// if username already exists
		if mongo.IsDuplicateKeyError(err) {
			c.AbortWithError(400, errors.New("username already exists")).SetType(gin.ErrorTypePublic)
		} else {
			c.AbortWithError(500, err).SetType(gin.ErrorTypePrivate)
		}
		return
	}

	c.JSON(200, gin.H{"id": res.InsertedID})
}

// BindUser binds request body to User struct
// and it creates new ObjectID and sets time.Now() to Created field
func (n *Nac) BindUser(c *gin.Context) {
	// username validation
	var usr User
	err := c.Bind(&usr)
	if err != nil {
		return
	}

	if matched, err := usernameReg.MatchString(usr.Username); matched == false || err != nil {
		c.AbortWithError(400, errors.New("username invalid")).SetType(gin.ErrorTypePublic)
		return
	}

	usr.Created = time.Now()
	usr.Id = primitive.NewObjectID()
	usr.Agents = []Agent{}
	usr.AddedTypes = []int{}

	c.Set("user", &usr)
}

func (n *Nac) GetUserId(c *gin.Context) {
	id := c.Param("uid")
	uid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.AbortWithError(400, errors.New("invalid uid")).SetType(gin.ErrorTypePublic)
		return
	}

	c.Set("uid", uid)
}

func (n *Nac) GetUserById(c *gin.Context) {
	ctx := c.Request.Context()
	uid := c.MustGet("uid").(primitive.ObjectID)

	var usr User

	bytes, err := n.rdb.Get(ctx, "u:"+uid.Hex()).Bytes()

	// if not found in redis, then find it in mongo
	if err == redis.Nil {
		res := n.mdb.Collection(mdb_users).FindOne(ctx, bson.M{"_id": uid})
		err = res.Err()
		if err != nil && err == mongo.ErrNoDocuments {
			c.AbortWithError(400, fmt.Errorf("user not found"))
			return
		}
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		err = res.Decode(&usr)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		bytes, err = res.DecodeBytes()
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		// set redis cache
		// log.Print("not cached by redis: ", usr)
		n.rdb.Set(ctx, "u:"+uid.Hex(), string(bytes), time.Hour*36)
		c.Set("user", &usr)
		return
	}

	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	err = bson.Unmarshal(bytes, &usr)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	// log.Print("cached by redis: ", usr)
	c.Set("user", &usr)
}
