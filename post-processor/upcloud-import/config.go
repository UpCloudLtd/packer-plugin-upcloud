//go:generate packer-sdc mapstructure-to-hcl2 -type Config
//go:generate packer-sdc struct-markdown
package upcloudimport

import (
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"

	"github.com/UpCloudLtd/packer-plugin-upcloud/internal/driver"
)

const DefaultTimeout time.Duration = 60 * time.Minute

type Config struct {
	// The username to use when interfacing with the UpCloud API.
	Username string `mapstructure:"username"`

	// The password to use when interfacing with the UpCloud API.
	Password string `mapstructure:"password"`

	// The API token to use when interfacing with the UpCloud API. This is mutually exclusive with username and password.
	Token string `mapstructure:"token"`

	// The list of zones in which the template should be imported
	Zones []string `mapstructure:"zones" required:"true"`

	// The name of the template. Use `replace_existing` to replace existing template
	// with same name or suffix template name with e.g. timestamp to avoid errors during import
	TemplateName string `mapstructure:"template_name" required:"true"`

	// Replace existing template if one exists with the same name. Defaults to `false`.
	ReplaceExisting bool `mapstructure:"replace_existing"`

	// The storage tier to use. Available options are `maxiops`, `archive`, and `standard`. Defaults to `maxiops`.
	StorageTier string `mapstructure:"storage_tier"`

	// The amount of time to wait for resource state changes. Defaults to `60m`.
	Timeout time.Duration `mapstructure:"state_timeout_duration"`

	ctx interpolate.Context

	common.PackerConfig `mapstructure:",squash"`
}

func NewConfig(raws ...interface{}) (*Config, error) {
	var c Config
	if err := config.Decode(&c, &config.DecodeOpts{
		PluginType:         BuilderID,
		Interpolate:        true,
		InterpolateContext: &c.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...); err != nil {
		return &c, fmt.Errorf("failed to decode configuration: %w", err)
	}

	c.fromEnv()

	if errs := c.validate(); len(errs.Errors) > 0 {
		return &c, errs
	}

	c.setDefaults()

	return &c, nil
}

// validate validates the configuration and returns any errors.
func (c *Config) validate() *packer.MultiError {
	errs := new(packer.MultiError)

	if len(c.Zones) == 0 {
		errs = packer.MultiErrorAppend(
			errs, errors.New("list of zones is empty"))
	}

	if c.TemplateName == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("'template_name' must be specified"),
		)
	}

	// Validate authentication
	if authErrs := c.validateAuthentication(); authErrs != nil {
		errs = packer.MultiErrorAppend(errs, authErrs.Errors...)
	}

	return errs
}

// validateAuthentication checks authentication configuration.
func (c *Config) validateAuthentication() *packer.MultiError {
	var errs *packer.MultiError

	// Check authentication: either username/password OR token, but not both
	hasUsernamePassword := c.Username != "" || c.Password != ""
	hasAPIToken := c.Token != ""

	if hasUsernamePassword && hasAPIToken {
		errs = packer.MultiErrorAppend(
			errs, errors.New("you cannot specify both username/password and token. Use either username/password or token for authentication"),
		)
	}

	if !hasUsernamePassword && !hasAPIToken {
		errs = packer.MultiErrorAppend(
			errs, errors.New("authentication required: specify either username and password, or token"),
		)
	}

	// If using username/password, both must be provided
	if hasUsernamePassword && (c.Username == "" || c.Password == "") {
		if c.Username == "" {
			errs = packer.MultiErrorAppend(
				errs, errors.New("'username' must be specified when using username/password authentication"),
			)
		}
		if c.Password == "" {
			errs = packer.MultiErrorAppend(
				errs, errors.New("'password' must be specified when using username/password authentication"),
			)
		}
	}

	return errs
}

// setDefaults sets default values for configuration fields.
func (c *Config) setDefaults() {
	if c.Timeout < 1 {
		c.Timeout = DefaultTimeout
	}

	// Set the default storage tier to maxiops if not specified
	if c.StorageTier == "" {
		c.StorageTier = "maxiops"
	}
}

func (c *Config) fromEnv() {
	if c.Username == "" {
		c.Username = driver.UsernameFromEnv()
	}
	if c.Password == "" {
		c.Password = driver.PasswordFromEnv()
	}
	if c.Token == "" {
		c.Token = driver.TokenFromEnv()
	}
}
