//go:generate packer-sdc mapstructure-to-hcl2 -type Config
//go:generate packer-sdc struct-markdown
package upcloudimport

import (
	"errors"
	"time"

	"github.com/UpCloudLtd/packer-plugin-upcloud/internal/driver"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

const defaultTimeout time.Duration = 60 * time.Minute

type Config struct {

	// The username to use when interfacing with the UpCloud API.
	Username string `mapstructure:"username" required:"true"`

	// The password to use when interfacing with the UpCloud API.
	Password string `mapstructure:"password" required:"true"`

	// The list of zones in which the template should be imported
	Zones []string `mapstructure:"zones" required:"true"`

	// The name of the template. Use `replace_existing` to replace existing template
	// with same name or suffix template name with e.g. timestamp to avoid errors during import
	TemplateName string `mapstructure:"template_name" required:"true"`

	// Replace existing template if one exists with the same name. Defaults to `false`.
	ReplaceExisting bool `mapstructure:"replace_existing"`

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
		return &c, err
	}

	c.fromEnv()

	errs := new(packer.MultiError)

	if len(c.Zones) == 0 {
		errs = packer.MultiErrorAppend(
			errs, errors.New("list of zones is empty"))
	}

	if c.Username == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("'username' must be specified"),
		)
	}

	if c.Password == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("'password' must be specified"),
		)
	}

	if c.TemplateName == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("'template_name' must be specified"),
		)
	}

	if len(errs.Errors) > 0 {
		return &c, errs
	}

	if c.Timeout < 1 {
		c.Timeout = defaultTimeout
	}

	return &c, nil
}

func (c *Config) fromEnv() {
	if c.Username == "" {
		c.Username = driver.UsernameFromEnv()
	}
	if c.Password == "" {
		c.Password = driver.PasswordFromEnv()
	}
}
