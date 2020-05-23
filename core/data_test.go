package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetActionProperty(t *testing.T) {
	prop, err := GetActionProperty("name")
	assert.Nil(t, err)
	assert.Equal(t, ActionName, prop)
	prop, err = GetActionProperty("sender")
	assert.Nil(t, err)
	assert.Equal(t, ActionSender, prop)
	prop, err = GetActionProperty("other")
	assert.NotNil(t, err)
}
