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

func TestAddMemories(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	n := nac.FromConfig("../example.config.toml")
	defer n.ClearData(context.Background())

	r.Use(n.ErrorHandler())

	r.POST("/u/add", n.BindUser, n.InsertUser)
	r.POST("/u/:uid/a/add/:tid", n.GetUserId, n.GetAgentType, n.NewAgent)
	r.POST("/u/:uid/a/:aid/m/add", n.GetObjectId("uid"), n.GetObjectId("aid"), n.AddMemories, n.OK)
	r.GET("/u/:uid/a/:aid/m", n.GetObjectId("uid"), n.GetObjectId("aid"), n.GetMemories)

	req := httptest.NewRequest("POST", "/u/add", strings.NewReader("{\"username\":\"aspirin2d\"}"))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var res map[string]string
	json.Unmarshal(w.Body.Bytes(), &res) // nolint: errcheck
	uid, err := primitive.ObjectIDFromHex(res["id"])
	assert.NoError(t, err)

	req = httptest.NewRequest("POST", "/u/"+uid.Hex()+"/a/add/1", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	json.Unmarshal(w.Body.Bytes(), &res) // nolint: errcheck
	aid, err := primitive.ObjectIDFromHex(res["id"])
	assert.NoError(t, err)

	req = httptest.NewRequest("POST", "/u/"+uid.Hex()+"/a/"+aid.Hex()+"/m/add", strings.NewReader(
		"[{\"content\":\"this is a memory.\"}, {\"content\":\"this is another memory.\"}]",
	))
	req.Header.Add("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	req = httptest.NewRequest("GET", "/u/"+uid.Hex()+"/a/"+aid.Hex()+"/m", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var mems []nac.Memory
	err = json.Unmarshal(w.Body.Bytes(), &mems)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(mems))
	assert.Equal(t, mems[0].Content, "this is another memory.") // first in first out
}
