package nac

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestBindUser(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	nac := &Nac{}

	req := httptest.NewRequest("POST", "/u/add", strings.NewReader("{\"username\":\"aspirin2d\"}"))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	nac.BindUser(c)
	assert.Zero(t, len(c.Errors))

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	req = httptest.NewRequest("POST", "/u/add", strings.NewReader("{\"username\":\"aspirin 2d\"}"))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	nac.BindUser(c)
	assert.Equal(t, c.Errors[0].Err.Error(), "username invalid")

	w = httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	// error handling middleware testing
	r.Use(nac.ErrorHandler())
	r.POST("/u/add", nac.BindUser)

	req = httptest.NewRequest("POST", "/u/add", strings.NewReader("{\"username\":\"aspirin2d\"")) // missing closing bracket
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
	assert.Contains(t, w.Body.String(), "binding error")
}
