//go:generate packer-sdc mapstructure-to-hcl2 -type Config,NetworkInterface,IPAddress
package upcloud

import (
	"errors"
	"os"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

const (
	DefaultTemplatePrefix = "custom-image"
	DefaultSSHUsername    = "root"
	DefaultStorageSize    = 25
	DefaultTimeout        = 5 * time.Minute
)

var (
	DefaultNetworking = []request.CreateServerInterface{
		{
			IPAddresses: []request.CreateServerIPAddress{
				{
					Family: upcloud.IPAddressFamilyIPv4,
				},
			},
			Type: upcloud.IPAddressAccessPublic,
		},
	}
)

// for config type convertion
type NetworkInterface struct {
	IPAddresses []IPAddress `mapstructure:"ip_addresses"`
	Type        string      `mapstructure:"type"`
	Network     string      `mapstructure:"network,omitempty"`
}

type IPAddress struct {
	Family  string `mapstructure:"family"`
	Address string `mapstructure:"address,omitempty"`
}

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`

	// Required configuration values
	Username    string `mapstructure:"username"`
	Password    string `mapstructure:"password"`
	Zone        string `mapstructure:"zone"`
	StorageUUID string `mapstructure:"storage_uuid"`
	StorageName string `mapstructure:"storage_name"`

	// Optional configuration values
	TemplatePrefix string        `mapstructure:"template_prefix"`
	TemplateName   string        `mapstructure:"template_name"`
	StorageSize    int           `mapstructure:"storage_size"`
	Timeout        time.Duration `mapstructure:"state_timeout_duration"`
	CloneZones     []string      `mapstructure:"clone_zones"`

	NetworkInterfaces []NetworkInterface `mapstructure:"network_interfaces"`

	SSHPrivateKeyPath string `mapstructure:"ssh_private_key_path"`
	SSHPublicKeyPath  string `mapstructure:"ssh_public_key_path"`

	ctx interpolate.Context
}

func (c *Config) Prepare(raws ...interface{}) ([]string, error) {
	err := config.Decode(c, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &c.ctx,
	}, raws...)

	if err != nil {
		return nil, err
	}

	c.setEnv()

	// defaults
	if c.TemplatePrefix == "" && len(c.TemplateName) == 0 {
		c.TemplatePrefix = DefaultTemplatePrefix
	}

	if c.StorageSize == 0 {
		c.StorageSize = DefaultStorageSize
	}

	if c.Timeout == 0 {
		c.Timeout = DefaultTimeout
	}

	if c.Comm.SSHUsername == "" {
		c.Comm.SSHUsername = DefaultSSHUsername
	}

	// validate
	var errs *packer.MultiError
	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
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

	if c.Zone == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("'zone' must be specified"),
		)
	}

	if c.StorageUUID == "" && c.StorageName == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("'storage_uuid' or 'storage_name' must be specified"),
		)
	}

	if len(c.TemplatePrefix) > 40 {
		errs = packer.MultiErrorAppend(
			errs, errors.New("'template_prefix' must be 0-40 characters"),
		)
	}

	if len(c.TemplateName) > 40 {
		errs = packer.MultiErrorAppend(
			errs, errors.New("'template_name' is limited to 40 characters"),
		)
	}

	if len(c.TemplatePrefix) > 0 && len(c.TemplateName) > 0 {
		errs = packer.MultiErrorAppend(
			errs, errors.New("you can either use 'template_prefix' or 'template_name' in your configuration"),
		)
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	return nil, nil
}

// get params from environment
func (c *Config) setEnv() {
	username := os.Getenv("UPCLOUD_API_USER")
	if username != "" && c.Username == "" {
		c.Username = username
	}

	password := os.Getenv("UPCLOUD_API_PASSWORD")
	if password != "" && c.Password == "" {
		c.Password = password
	}
}
