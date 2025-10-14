//go:build !integration

package upcloud_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	assert.NoError(t, err)
	assert.Equal(t, "test-api-token", c.Token)
	assert.Empty(t, c.Username)
	assert.Empty(t, c.Password)
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
	require.Error(t, err)
	assert.Contains(t, err.Error(), "credentials not found")
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
	require.Error(t, err)
	assert.Contains(t, err.Error(), "credentials not found")
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
	require.Error(t, err)
	assert.Contains(t, err.Error(), "credentials not found")
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
	err := c.SetEnv()
	assert.NoError(t, err)
	assert.Equal(t, "test-token", c.Token)
}

func TestConfig_setEnv_DoesNotOverrideExisting_basic(t *testing.T) {
	t.Setenv(driver.EnvConfigUsername, "env-user")
	t.Setenv(driver.EnvConfigPassword, "env-pass")
	t.Setenv(driver.EnvConfigAPIToken, "env-token")

	c := &upcloud.Config{
		Username: "existing-user",
		Password: "existing-pass",
	}
	err := c.SetEnv()
	assert.NoError(t, err)

	// Should not override existing values
	assert.Equal(t, "existing-user", c.Username)
	assert.Equal(t, "existing-pass", c.Password)
	assert.Empty(t, c.Token)
}

func TestConfig_setEnv_DoesNotOverrideExisting_token(t *testing.T) {
	t.Setenv(driver.EnvConfigUsername, "env-user")
	t.Setenv(driver.EnvConfigPassword, "env-pass")
	t.Setenv(driver.EnvConfigAPIToken, "env-token")

	c := &upcloud.Config{
		Username: "existing-user",
		Password: "existing-pass",
		Token:    "existing-token",
	}
	err := c.SetEnv()
	assert.NoError(t, err)

	// Should not override existing values
	assert.Empty(t, c.Username)
	assert.Empty(t, c.Password)
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
			"storage_name": "Ubuntu Server 24.04",
		},
	}

	warns, err := c.Prepare(raws...)
	assert.NoError(t, err)
	assert.Empty(t, warns)
	assert.Equal(t, "Ubuntu Server 24.04", c.StorageName)
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

func TestConfig_validateNetworkInterfaces(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		interfaces     []upcloud.NetworkInterface
		expectError    bool
		errorSubstring string
	}{
		{
			name: "valid public interface",
			interfaces: []upcloud.NetworkInterface{
				{
					Type: upcloud.InterfaceTypePublic,
					IPAddresses: []upcloud.IPAddress{
						{Family: "IPv4", Default: true},
					},
				},
			},
			expectError: false,
		},
		{
			name: "valid private interface with UUID",
			interfaces: []upcloud.NetworkInterface{
				{
					Type:    upcloud.InterfaceTypePrivate,
					Network: "01234567-89ab-cdef-0123-456789abcdef",
					IPAddresses: []upcloud.IPAddress{
						{Family: "IPv4", Address: "192.168.1.100"},
					},
				},
			},
			expectError: false,
		},
		{
			name: "invalid interface type",
			interfaces: []upcloud.NetworkInterface{
				{Type: upcloud.InterfaceType("invalid")},
			},
			expectError:    true,
			errorSubstring: "has invalid type: invalid",
		},
		{
			name: "private interface without UUID",
			interfaces: []upcloud.NetworkInterface{
				{Type: upcloud.InterfaceTypePrivate},
			},
			expectError:    true,
			errorSubstring: "private network requires network UUID",
		},
		{
			name: "private interface with invalid UUID",
			interfaces: []upcloud.NetworkInterface{
				{
					Type:    upcloud.InterfaceTypePrivate,
					Network: "not-a-uuid",
				},
			},
			expectError:    true,
			errorSubstring: "invalid network UUID",
		},
		{
			name: "invalid IP family",
			interfaces: []upcloud.NetworkInterface{
				{
					Type: upcloud.InterfaceTypePublic,
					IPAddresses: []upcloud.IPAddress{
						{Family: "IPv5"},
					},
				},
			},
			expectError:    true,
			errorSubstring: "invalid IP family 'IPv5'",
		},
		{
			name: "invalid IP address",
			interfaces: []upcloud.NetworkInterface{
				{
					Type: upcloud.InterfaceTypePublic,
					IPAddresses: []upcloud.IPAddress{
						{Family: "IPv4", Address: "invalid-ip"},
					},
				},
			},
			expectError:    true,
			errorSubstring: "invalid IP address 'invalid-ip'",
		},
		{
			name: "IP family mismatch - IPv4 family with IPv6 address",
			interfaces: []upcloud.NetworkInterface{
				{
					Type: upcloud.InterfaceTypePublic,
					IPAddresses: []upcloud.IPAddress{
						{Family: "IPv4", Address: "2001:db8::1"},
					},
				},
			},
			expectError:    true,
			errorSubstring: "IP family is IPv4 but address is IPv6",
		},
		{
			name: "IP family mismatch - IPv6 family with IPv4 address",
			interfaces: []upcloud.NetworkInterface{
				{
					Type: upcloud.InterfaceTypePublic,
					IPAddresses: []upcloud.IPAddress{
						{Family: "IPv6", Address: "192.168.1.1"},
					},
				},
			},
			expectError:    true,
			errorSubstring: "IP family is IPv6 but address is IPv4",
		},
		{
			name: "multiple interfaces with mixed errors",
			interfaces: []upcloud.NetworkInterface{
				{
					Type: upcloud.InterfaceTypePublic,
					IPAddresses: []upcloud.IPAddress{
						{Family: "IPv4", Default: true},
					},
				},
				{
					Type: upcloud.InterfaceType("invalid"),
				},
				{
					Type:    upcloud.InterfaceTypePrivate,
					Network: "bad-uuid",
				},
			},
			expectError:    true,
			errorSubstring: "has invalid type",
		},
		{
			name:        "empty interfaces slice",
			interfaces:  []upcloud.NetworkInterface{},
			expectError: false,
		},
		{
			name: "interface with empty IP addresses",
			interfaces: []upcloud.NetworkInterface{
				{
					Type:        upcloud.InterfaceTypeUtility,
					IPAddresses: []upcloud.IPAddress{},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &upcloud.Config{
				Username:          "testuser",
				Password:          "testpass",
				Zone:              "fi-hel1",
				StorageUUID:       "01000000-0000-4000-8000-000030060200",
				NetworkInterfaces: tt.interfaces,
			}

			_, err := c.Prepare(map[string]interface{}{})

			if tt.expectError {
				assert.Error(t, err, "expected validation errors")
				if tt.errorSubstring != "" {
					assert.Contains(t, err.Error(), tt.errorSubstring)
				}
			} else {
				assert.NoError(t, err, "unexpected validation errors")
			}
		})
	}
}
