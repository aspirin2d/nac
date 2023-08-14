package nac

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPing(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	// setup routes
	n := &Nac{}

	r.Use(n.ErrorHandler())
	r.GET("/ping", n.Ping)

	req, _ := http.NewRequest("GET", "/ping", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "pong", w.Body.String())
}

func TestLoadConfig(t *testing.T) {
	var conf Config
	_, err := toml.DecodeFile(".config.toml", &conf)
	assert.NoError(t, err)

	assert.Equal(t, 3, len(conf.Agents))
	assert.Equal(t, 1, len(conf.Templates))
}
