//go:build !integration

package upcloud_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/UpCloudLtd/packer-plugin-upcloud/builder/upcloud"
	"github.com/UpCloudLtd/packer-plugin-upcloud/internal/driver"
)

func TestConfig_Prepare_ValidUsernamePassword(t *testing.T) {
	t.Parallel()
	c := &upcloud.Config{}
	raws := []interface{}{
		map[string]interface{}{
			"username":     "testuser",
			"password":     "testpass",
			"zone":         "fi-hel1",
			"storage_uuid": "01000000-0000-4000-8000-000030060200",
		},
	}

	warns, err := c.Prepare(raws...)
	assert.NoError(t, err)
	assert.Empty(t, warns)
	assert.Equal(t, "testuser", c.Username)
	assert.Equal(t, "testpass", c.Password)
	assert.Empty(t, c.Token)
}

func TestConfig_Prepare_ValidAPIToken(t *testing.T) {
	t.Parallel()
	c := &upcloud.Config{}
	raws := []interface{}{
		map[string]interface{}{
			"token":        "test-api-token",
			"zone":         "fi-hel1",
			"storage_uuid": "01000000-0000-4000-8000-000030060200",
		},
	}

	warns, err := c.Prepare(raws...)
	assert.NoError(t, err)
	assert.Empty(t, warns)
	assert.Equal(t, "test-api-token", c.Token)
	assert.Empty(t, c.Username)
	assert.Empty(t, c.Password)
}

func TestConfig_Prepare_BothAuthMethods(t *testing.T) {
	t.Parallel()
	c := &upcloud.Config{}
	raws := []interface{}{
		map[string]interface{}{
			"username":     "testuser",
			"password":     "testpass",
			"token":        "test-api-token",
			"zone":         "fi-hel1",
			"storage_uuid": "01000000-0000-4000-8000-000030060200",
		},
	}

	warns, err := c.Prepare(raws...)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "you cannot specify both username/password and token")
	assert.Empty(t, warns)
}

func TestConfig_Prepare_NoAuthMethods(t *testing.T) {
	t.Parallel()
	c := &upcloud.Config{}
	raws := []interface{}{
		map[string]interface{}{
			"zone":         "fi-hel1",
			"storage_uuid": "01000000-0000-4000-8000-000030060200",
		},
	}

	warns, err := c.Prepare(raws...)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication required: specify either username and password, or token")
	assert.Empty(t, warns)
}

func TestConfig_Prepare_OnlyUsername(t *testing.T) {
	t.Parallel()
	c := &upcloud.Config{}
	raws := []interface{}{
		map[string]interface{}{
			"username":     "testuser",
			"zone":         "fi-hel1",
			"storage_uuid": "01000000-0000-4000-8000-000030060200",
		},
	}

	warns, err := c.Prepare(raws...)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'password' must be specified when using username/password authentication")
	assert.Empty(t, warns)
}

func TestConfig_Prepare_OnlyPassword(t *testing.T) {
	t.Parallel()
	c := &upcloud.Config{}
	raws := []interface{}{
		map[string]interface{}{
			"password":     "testpass",
			"zone":         "fi-hel1",
			"storage_uuid": "01000000-0000-4000-8000-000030060200",
		},
	}

	warns, err := c.Prepare(raws...)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'username' must be specified when using username/password authentication")
	assert.Empty(t, warns)
}

func TestConfig_Prepare_MissingZone(t *testing.T) {
	t.Parallel()
	c := &upcloud.Config{}
	raws := []interface{}{
		map[string]interface{}{
			"username":     "testuser",
			"password":     "testpass",
			"storage_uuid": "01000000-0000-4000-8000-000030060200",
		},
	}

	warns, err := c.Prepare(raws...)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'zone' must be specified")
	assert.Empty(t, warns)
}

func TestConfig_Prepare_MissingStorage(t *testing.T) {
	t.Parallel()
	c := &upcloud.Config{}
	raws := []interface{}{
		map[string]interface{}{
			"username": "testuser",
			"password": "testpass",
			"zone":     "fi-hel1",
		},
	}

	warns, err := c.Prepare(raws...)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'storage_uuid' or 'storage_name' must be specified")
	assert.Empty(t, warns)
}

func TestConfig_Prepare_TemplatePrefix(t *testing.T) {
	t.Parallel()
	c := &upcloud.Config{}
	raws := []interface{}{
		map[string]interface{}{
			"username":        "testuser",
			"password":        "testpass",
			"zone":            "fi-hel1",
			"storage_uuid":    "01000000-0000-4000-8000-000030060200",
			"template_prefix": "custom-prefix",
		},
	}

	warns, err := c.Prepare(raws...)
	assert.NoError(t, err)
	assert.Empty(t, warns)
	assert.Equal(t, "custom-prefix", c.TemplatePrefix)
	assert.Empty(t, c.TemplateName)
}

func TestConfig_Prepare_TemplateName(t *testing.T) {
	t.Parallel()
	c := &upcloud.Config{}
	raws := []interface{}{
		map[string]interface{}{
			"username":      "testuser",
			"password":      "testpass",
			"zone":          "fi-hel1",
			"storage_uuid":  "01000000-0000-4000-8000-000030060200",
			"template_name": "my-custom-template",
		},
	}

	warns, err := c.Prepare(raws...)
	assert.NoError(t, err)
	assert.Empty(t, warns)
	assert.Equal(t, "my-custom-template", c.TemplateName)
	assert.Empty(t, c.TemplatePrefix)
}

func TestConfig_Prepare_BothTemplateFields(t *testing.T) {
	t.Parallel()
	c := &upcloud.Config{}
	raws := []interface{}{
		map[string]interface{}{
			"username":        "testuser",
			"password":        "testpass",
			"zone":            "fi-hel1",
			"storage_uuid":    "01000000-0000-4000-8000-000030060200",
			"template_prefix": "custom-prefix",
			"template_name":   "my-custom-template",
		},
	}

	warns, err := c.Prepare(raws...)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "you can either use 'template_prefix' or 'template_name' in your configuration")
	assert.Empty(t, warns)
}

func TestConfig_Prepare_LongTemplatePrefix(t *testing.T) {
	t.Parallel()
	c := &upcloud.Config{}
	longPrefix := "this-is-a-very-long-template-prefix-that-exceeds-the-limit"
	raws := []interface{}{
		map[string]interface{}{
			"username":        "testuser",
			"password":        "testpass",
			"zone":            "fi-hel1",
			"storage_uuid":    "01000000-0000-4000-8000-000030060200",
			"template_prefix": longPrefix,
		},
	}

	warns, err := c.Prepare(raws...)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'template_prefix' must be 0-40 characters")
	assert.Empty(t, warns)
}

func TestConfig_Prepare_LongTemplateName(t *testing.T) {
	t.Parallel()
	c := &upcloud.Config{}
	longName := "this-is-a-very-long-template-name-that-exceeds-the-limit"
	raws := []interface{}{
		map[string]interface{}{
			"username":      "testuser",
			"password":      "testpass",
			"zone":          "fi-hel1",
			"storage_uuid":  "01000000-0000-4000-8000-000030060200",
			"template_name": longName,
		},
	}

	warns, err := c.Prepare(raws...)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'template_name' is limited to 40 characters")
	assert.Empty(t, warns)
}

func TestConfig_Prepare_Defaults(t *testing.T) {
	t.Parallel()
	c := &upcloud.Config{}
	raws := []interface{}{
		map[string]interface{}{
			"username":     "testuser",
			"password":     "testpass",
			"zone":         "fi-hel1",
			"storage_uuid": "01000000-0000-4000-8000-000030060200",
		},
	}

	warns, err := c.Prepare(raws...)
	assert.NoError(t, err)
	assert.Empty(t, warns)

	// Check defaults
	assert.Equal(t, "custom-image", c.TemplatePrefix)
	assert.Equal(t, 25, c.StorageSize)
	assert.Equal(t, "maxiops", c.StorageTier)
	assert.Equal(t, upcloud.DefaultTimeout, c.Timeout)
	assert.Equal(t, "ssh", c.Comm.Type)
	assert.Equal(t, "root", c.Comm.SSHUsername)
}

func TestConfig_setEnv_APIToken(t *testing.T) {
	t.Setenv(driver.EnvConfigAPIToken, "test-token")

	c := &upcloud.Config{}
	c.SetEnv()
	assert.Equal(t, "test-token", c.Token)
}

func TestConfig_setEnv_DoesNotOverrideExisting(t *testing.T) {
	t.Setenv(driver.EnvConfigUsername, "env-user")
	t.Setenv(driver.EnvConfigPassword, "env-pass")
	t.Setenv(driver.EnvConfigAPIToken, "env-token")

	c := &upcloud.Config{
		Username: "existing-user",
		Password: "existing-pass",
		Token:    "existing-token",
	}
	c.SetEnv()

	// Should not override existing values
	assert.Equal(t, "existing-user", c.Username)
	assert.Equal(t, "existing-pass", c.Password)
	assert.Equal(t, "existing-token", c.Token)
}

func TestConfig_DefaultIPaddress(t *testing.T) {
	t.Parallel()
	c := &upcloud.Config{
		NetworkInterfaces: []upcloud.NetworkInterface{
			{
				Type: upcloud.InterfaceTypePublic,
				IPAddresses: []upcloud.IPAddress{
					{
						Family:  "IPv4",
						Default: false,
					},
					{
						Family:  "IPv4",
						Default: true,
					},
				},
			},
			{
				Type: upcloud.InterfaceTypePrivate,
				IPAddresses: []upcloud.IPAddress{
					{
						Family:  "IPv4",
						Default: false,
					},
				},
			},
		},
	}

	ip, ifaceType := c.DefaultIPaddress()
	assert.NotNil(t, ip)
	assert.Equal(t, "IPv4", ip.Family)
	assert.True(t, ip.Default)
	assert.Equal(t, upcloud.InterfaceTypePublic, ifaceType)
}

func TestConfig_StorageName(t *testing.T) {
	t.Parallel()
	c := &upcloud.Config{}
	raws := []interface{}{
		map[string]interface{}{
			"username":     "testuser",
			"password":     "testpass",
			"zone":         "fi-hel1",
			"storage_name": "Ubuntu Server 20.04",
		},
	}

	warns, err := c.Prepare(raws...)
	assert.NoError(t, err)
	assert.Empty(t, warns)
	assert.Equal(t, "Ubuntu Server 20.04", c.StorageName)
	assert.Empty(t, c.StorageUUID)
}

func TestConfig_Prepare_CustomValues(t *testing.T) {
	t.Parallel()
	c := &upcloud.Config{}
	raws := []interface{}{
		map[string]interface{}{
			"token":                  "test-token",
			"zone":                   "de-fra1",
			"storage_uuid":           "01000000-0000-4000-8000-000030060200",
			"template_name":          "my-custom-template",
			"storage_size":           50,
			"storage_tier":           "standard",
			"state_timeout_duration": "10m",
			"boot_wait":              "30s",
			"clone_zones":            []string{"nl-ams1", "us-nyc1"},
		},
	}

	warns, err := c.Prepare(raws...)
	assert.NoError(t, err)
	assert.Empty(t, warns)

	assert.Equal(t, "test-token", c.Token)
	assert.Equal(t, "de-fra1", c.Zone)
	assert.Equal(t, "my-custom-template", c.TemplateName)
	assert.Equal(t, 50, c.StorageSize)
	assert.Equal(t, "standard", c.StorageTier)
	assert.Equal(t, 10*time.Minute, c.Timeout)
	assert.Equal(t, 30*time.Second, c.BootWait)
	assert.Equal(t, []string{"nl-ams1", "us-nyc1"}, c.CloneZones)
}
