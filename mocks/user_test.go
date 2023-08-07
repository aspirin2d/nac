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
)

func TestUser(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	// setup routes
	n := nac.FromConfig("../example.config.toml")
	defer func() {
		n.ClearData(context.TODO())
	}()

	r.Use(n.ErrorHandler())
	r.POST("/u/add", n.BindUser, n.InsertUser)
	r.GET("/u/:uid", n.GetUserId, n.GetUserById)

	req := httptest.NewRequest("POST", "/u/add", strings.NewReader("{\"username\":\"aspirin2d\"}"))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var usr nac.User
	json.Unmarshal(w.Body.Bytes(), &usr)

	req = httptest.NewRequest("POST", "/u/add", strings.NewReader("{\"username\":\"aspirin2d\"}"))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
	assert.Contains(t, w.Body.String(), "username already exists")

	req = httptest.NewRequest("GET", "/u/"+usr.Id.Hex(), nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	// test redis cache
	req = httptest.NewRequest("GET", "/u/"+usr.Id.Hex(), nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
