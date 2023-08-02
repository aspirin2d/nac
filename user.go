package nac

import (
	"errors"
	"time"

	"github.com/dlclark/regexp2"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var usernameReg = regexp2.MustCompile("^(?=[a-zA-Z0-9._]{6,20}$)(?!.*[_.]{2})[^_.].*[^_.]$", 0)

type User struct {
	Id       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Username string             `bson:"username" json:"username"`
	Created  time.Time          `bson:"created" json:"created"`
}

// AddUser will add an user to the database.
// never mock testing
func (n *Nac) InsertUser(c *gin.Context) {
	usr, ok := c.MustGet("user").(*User)
	if !ok {
		c.AbortWithError(400, errors.New("user type casting error")).SetType(gin.ErrorTypePrivate)
		return
	}

	ctx := c.Request.Context()
	res, err := n.mdb.Collection(mdb_users).InsertOne(ctx, usr)
	if err != nil {
		// if username already exists
		if mongo.IsDuplicateKeyError(err) {
			c.AbortWithError(400, errors.New("user already exists"))
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
		c.AbortWithError(400, errors.New("username invalid"))
		return
	}

	usr.Created = time.Now()
	usr.Id = primitive.NewObjectID()

	c.Set("user", usr)
}

func (n *Nac) GetUser(c *gin.Context) {
	id, exists := c.Params.Get("uid")
	if !exists {
		c.AbortWithError(400, errors.New("user_id not found"))
		return
	}

	uid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.AbortWithError(400, errors.New("invalid user_id"))
		return
	}

	c.Set("uid", uid)
}
