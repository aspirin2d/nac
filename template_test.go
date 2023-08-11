package nac

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadTemplates(t *testing.T) {
	tc := LoadTemplates("./testdata/templates.toml")
	nac := &Nac{tc: &tc}
	assert.NotNil(t, nac.tc)
}
