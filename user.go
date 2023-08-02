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
func (n *Nac) AddUser(c *gin.Context) {
	usr := bindUser(c)

	if usr == nil {
		return
	}

	ctx := c.Request.Context()
	res, err := n.mdb.Collection(mdb_users).InsertOne(ctx, usr)
	if err != nil {
		// if username already exists
		if mongo.IsDuplicateKeyError(err) {
			c.AbortWithError(400, errors.New("username already occupied")).SetType(gin.ErrorTypePublic)
		} else {
			c.AbortWithError(500, err).SetType(gin.ErrorTypePrivate)
		}
		return
	}

	c.JSON(200, gin.H{"id": res.InsertedID})
}

func bindUser(c *gin.Context) *User {
	// username validation
	var usr User
	err := c.Bind(&usr)
	if err != nil {
		return nil
	}

	if matched, err := usernameReg.MatchString(usr.Username); matched == false || err != nil {
		c.AbortWithError(400, errors.New("username invalid")).SetType(gin.ErrorTypePublic)
		return nil
	}

	usr.Created = time.Now()
	return &usr
}
