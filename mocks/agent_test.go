package mock

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sleep2death/nac"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAddAgent(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	n := nac.FromConfig("../example.config.toml")
	defer n.ClearData(context.Background())

	assert.NotZero(t, len(n.Config().AgentTypes))
	assert.NotNil(t, n.Config())

	r.Use(n.ErrorHandler())
	r.POST("/u/add", n.BindUser, n.InsertUser)
	r.POST("/u/:uid/a/add/:tid", n.GetObjectId("uid"), n.GetAgentType, n.NewAgent)

	req := httptest.NewRequest("POST", "/u/add", strings.NewReader("{\"username\":\"aspirin2d\"}"))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var res map[string]string
	json.Unmarshal(w.Body.Bytes(), &res)
	uid, err := primitive.ObjectIDFromHex(res["id"])
	assert.NoError(t, err)

	req = httptest.NewRequest("POST", "/u/"+uid.Hex()+"/a/add/1", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "id")

	req = httptest.NewRequest("POST", "/u/"+uid.Hex()+"/a/add/1", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
	// t.Log(w.Body.String())

	req = httptest.NewRequest("POST", "/u/"+uid.Hex()+"/a/add/2", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
	// t.Log(w.Body.String())
}
