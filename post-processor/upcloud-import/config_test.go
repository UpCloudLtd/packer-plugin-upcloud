//go:build !integration

package upcloudimport_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/UpCloudLtd/packer-plugin-upcloud/internal/driver"
	upcloudimport "github.com/UpCloudLtd/packer-plugin-upcloud/post-processor/upcloud-import"
)

func TestNewConfig_ValidUsernamePassword(t *testing.T) {
	t.Parallel()
	c, err := upcloudimport.NewConfig([]interface{}{map[string]interface{}{
		"username":      "testuser",
		"password":      "testpass",
		"zones":         []string{"fi-hel1", "us-nyc1"},
		"template_name": "my-template",
	}}...)

	assert.NoError(t, err)
	require.NotNil(t, c)
	assert.Equal(t, "testuser", c.Username)
	assert.Equal(t, "testpass", c.Password)
	assert.Empty(t, c.Token)
	assert.Equal(t, []string{"fi-hel1", "us-nyc1"}, c.Zones)
	assert.Equal(t, "my-template", c.TemplateName)
}

func TestNewConfig_ValidAPIToken(t *testing.T) {
	t.Parallel()
	c, err := upcloudimport.NewConfig([]interface{}{map[string]interface{}{
		"token":         "test-api-token",
		"zones":         []string{"fi-hel1", "us-nyc1"},
		"template_name": "my-template",
	}}...)

	assert.NoError(t, err)
	require.NotNil(t, c)
	assert.Equal(t, "test-api-token", c.Token)
	assert.Empty(t, c.Username)
	assert.Empty(t, c.Password)
	assert.Equal(t, []string{"fi-hel1", "us-nyc1"}, c.Zones)
	assert.Equal(t, "my-template", c.TemplateName)
}

func TestNewConfig_BothAuthMethods(t *testing.T) {
	t.Parallel()
	c, err := upcloudimport.NewConfig([]interface{}{map[string]interface{}{
		"username":      "testuser",
		"password":      "testpass",
		"token":         "test-api-token",
		"zones":         []string{"fi-hel1"},
		"template_name": "my-template",
	}}...)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "you cannot specify both username/password and token")
	require.NotNil(t, c)
}

func TestNewConfig_NoAuthMethods(t *testing.T) {
	t.Parallel()
	c, err := upcloudimport.NewConfig([]interface{}{map[string]interface{}{
		"zones":         []string{"fi-hel1"},
		"template_name": "my-template",
	}}...)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication required: specify either username and password, or token")
	require.NotNil(t, c)
}

func TestNewConfig_OnlyUsername(t *testing.T) {
	t.Parallel()
	c, err := upcloudimport.NewConfig([]interface{}{map[string]interface{}{
		"username":      "testuser",
		"zones":         []string{"fi-hel1"},
		"template_name": "my-template",
	}}...)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'password' must be specified when using username/password authentication")
	require.NotNil(t, c)
}

func TestNewConfig_OnlyPassword(t *testing.T) {
	t.Parallel()
	c, err := upcloudimport.NewConfig([]interface{}{map[string]interface{}{
		"password":      "testpass",
		"zones":         []string{"fi-hel1"},
		"template_name": "my-template",
	}}...)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'username' must be specified when using username/password authentication")
	require.NotNil(t, c)
}

func TestNewConfig_EmptyZones(t *testing.T) {
	t.Parallel()
	c, err := upcloudimport.NewConfig([]interface{}{map[string]interface{}{
		"username":      "testuser",
		"password":      "testpass",
		"zones":         []string{},
		"template_name": "my-template",
	}}...)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "list of zones is empty")
	require.NotNil(t, c)
}

func TestNewConfig_MissingZones(t *testing.T) {
	t.Parallel()
	c, err := upcloudimport.NewConfig([]interface{}{map[string]interface{}{
		"username":      "testuser",
		"password":      "testpass",
		"template_name": "my-template",
	}}...)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "list of zones is empty")
	require.NotNil(t, c)
}

func TestNewConfig_MissingTemplateName(t *testing.T) {
	t.Parallel()
	c, err := upcloudimport.NewConfig([]interface{}{map[string]interface{}{
		"username": "testuser",
		"password": "testpass",
		"zones":    []string{"fi-hel1"},
	}}...)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'template_name' must be specified")
	require.NotNil(t, c)
}

func TestNewConfig_Defaults(t *testing.T) {
	t.Parallel()
	c, err := upcloudimport.NewConfig([]interface{}{map[string]interface{}{
		"username":      "testuser",
		"password":      "testpass",
		"zones":         []string{"fi-hel1"},
		"template_name": "my-template",
	}}...)

	assert.NoError(t, err)
	require.NotNil(t, c)

	// Check defaults
	assert.Equal(t, upcloudimport.DefaultTimeout, c.Timeout)
	assert.Equal(t, "maxiops", c.StorageTier)
	assert.False(t, c.ReplaceExisting)
}

func TestNewConfig_CustomValues(t *testing.T) {
	t.Parallel()
	c, err := upcloudimport.NewConfig([]interface{}{map[string]interface{}{
		"token":                  "test-token",
		"zones":                  []string{"fi-hel1", "de-fra1"},
		"template_name":          "my-custom-template",
		"replace_existing":       true,
		"storage_tier":           "standard",
		"state_timeout_duration": "30m",
	}}...)

	assert.NoError(t, err)
	require.NotNil(t, c)

	assert.Equal(t, "test-token", c.Token)
	assert.Equal(t, []string{"fi-hel1", "de-fra1"}, c.Zones)
	assert.Equal(t, "my-custom-template", c.TemplateName)
	assert.True(t, c.ReplaceExisting)
	assert.Equal(t, "standard", c.StorageTier)
	assert.Equal(t, 30*time.Minute, c.Timeout)
}

func TestConfig_fromEnv_APIToken(t *testing.T) {
	t.Setenv(driver.EnvConfigAPIToken, "test-token")

	c, err := upcloudimport.NewConfig([]interface{}{map[string]interface{}{
		"zones":         []string{"fi-hel1"},
		"template_name": "my-template",
	}}...)

	assert.NoError(t, err)
	require.NotNil(t, c)
	assert.Equal(t, "test-token", c.Token)
}

func TestConfig_fromEnv_Username(t *testing.T) {
	t.Setenv(driver.EnvConfigUsername, "env-user")
	t.Setenv(driver.EnvConfigPassword, "env-pass")

	c, err := upcloudimport.NewConfig([]interface{}{map[string]interface{}{
		"zones":         []string{"fi-hel1"},
		"template_name": "my-template",
	}}...)

	assert.NoError(t, err)
	require.NotNil(t, c)
	assert.Equal(t, "env-user", c.Username)
	assert.Equal(t, "env-pass", c.Password)
}

func TestConfig_fromEnv_DoesNotOverrideExisting(t *testing.T) {
	t.Setenv(driver.EnvConfigUsername, "env-user")
	t.Setenv(driver.EnvConfigPassword, "env-pass")

	c, err := upcloudimport.NewConfig([]interface{}{map[string]interface{}{
		"username":      "config-user",
		"password":      "config-pass",
		"zones":         []string{"fi-hel1"},
		"template_name": "my-template",
	}}...)

	assert.NoError(t, err)
	require.NotNil(t, c)
	// Config values should override environment variables
	assert.Equal(t, "config-user", c.Username)
	assert.Equal(t, "config-pass", c.Password)
	assert.Empty(t, c.Token) // No API token in config or env
}

func TestNewConfig_WithEnvironmentVariables(t *testing.T) {
	t.Setenv(driver.EnvConfigAPIToken, "env-token")

	c, err := upcloudimport.NewConfig([]interface{}{map[string]interface{}{
		"zones":         []string{"fi-hel1"},
		"template_name": "my-template",
	}}...)

	assert.NoError(t, err)
	require.NotNil(t, c)
	assert.Equal(t, "env-token", c.Token)
	assert.Empty(t, c.Username)
	assert.Empty(t, c.Password)
}

func TestNewConfig_ConfigOverridesEnvironment(t *testing.T) {
	t.Setenv(driver.EnvConfigAPIToken, "env-token")

	c, err := upcloudimport.NewConfig([]interface{}{map[string]interface{}{
		"token":         "config-token",
		"zones":         []string{"fi-hel1"},
		"template_name": "my-template",
	}}...)

	assert.NoError(t, err)
	require.NotNil(t, c)
	// Config values should override environment variables
	assert.Equal(t, "config-token", c.Token)
}

func TestNewConfig_TimeoutHandling(t *testing.T) {
	t.Parallel()
	// Test that timeout less than 1 gets set to default
	c, err := upcloudimport.NewConfig([]interface{}{map[string]interface{}{
		"token":                  "test-token",
		"zones":                  []string{"fi-hel1"},
		"template_name":          "my-template",
		"state_timeout_duration": "0s",
	}}...)

	assert.NoError(t, err)
	require.NotNil(t, c)
	assert.Equal(t, upcloudimport.DefaultTimeout, c.Timeout)
}

func TestNewConfig_StorageTierDefault(t *testing.T) {
	t.Parallel()
	c, err := upcloudimport.NewConfig([]interface{}{map[string]interface{}{
		"token":         "test-token",
		"zones":         []string{"fi-hel1"},
		"template_name": "my-template",
		// Don't specify storage_tier
	}}...)

	assert.NoError(t, err)
	require.NotNil(t, c)
	assert.Equal(t, "maxiops", c.StorageTier)
}
