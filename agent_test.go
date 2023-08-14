package nac

import (
	"net/http/httptest"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAgentsStep(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var conf Config
	_, err := toml.DecodeFile(".config.toml", &conf)
	assert.Nil(t, err)
	assert.Equal(t, len(conf.Agents), 3)

	n := &Nac{config: &conf}

	// parse template descriptions into templates
	for _, t := range conf.Templates {
		t.Check()
	}

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.GET("/u/:uid/step", n.GetObjectId("uid"), n.Step)

	req := httptest.NewRequest("GET", "/u/5f0d9e9d9d0a5d1c5f9c7e1a/step", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}
