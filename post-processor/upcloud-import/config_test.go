package upcloudimport

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	_, err := NewConfig()
	assert.Error(t, err)
	c, err := NewConfig([]interface{}{map[string]interface{}{
		"username":      "test",
		"password":      "passwd",
		"template_name": "my-template",
		"zones":         []string{"a", "b"},
	}}...)
	t.Logf("%+v", c)
	assert.NoError(t, err)
	assert.NotNil(t, c)
	assert.Len(t, c.Zones, 2)
	assert.Equal(t, "test", c.Username)
	assert.Equal(t, "passwd", c.Password)
	assert.Equal(t, "my-template", c.TemplateName)
	assert.Equal(t, false, c.ReplaceExisting)
	assert.Equal(t, defaultTimeout, c.Timeout)
}
