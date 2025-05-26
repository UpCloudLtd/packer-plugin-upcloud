package upcloudimport_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	upcloudimport "github.com/UpCloudLtd/packer-plugin-upcloud/post-processor/upcloud-import"
)

func TestNewConfig(t *testing.T) {
	t.Parallel()
	_, err := upcloudimport.NewConfig()
	assert.Error(t, err)
	c, err := upcloudimport.NewConfig([]interface{}{map[string]interface{}{
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
	assert.False(t, c.ReplaceExisting)
	assert.Equal(t, upcloudimport.DefaultTimeout, c.Timeout)
}
