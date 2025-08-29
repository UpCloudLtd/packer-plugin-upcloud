//go:generate packer-sdc mapstructure-to-hcl2 -type Config,NetworkInterface,IPAddress
//go:generate packer-sdc struct-markdown
package upcloud

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"

	"github.com/UpCloudLtd/packer-plugin-upcloud/internal/driver"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
)

type InterfaceType string

const (
	DefaultTemplatePrefix                 = "custom-image"
	DefaultSSHUsername                    = "root"
	DefaultCommunicator                   = "ssh"
	DefaultStorageSize                    = 25
	DefaultTimeout                        = 5 * time.Minute
	DefaultStorageTier                    = upcloud.StorageTierMaxIOPS
	InterfaceTypePublic     InterfaceType = upcloud.IPAddressAccessPublic
	InterfaceTypeUtility    InterfaceType = upcloud.IPAddressAccessUtility
	InterfaceTypePrivate    InterfaceType = upcloud.IPAddressAccessPrivate
	MaxTemplateNameLength                 = 40
	MaxTemplatePrefixLength               = 40
)

// for config type conversion.
type NetworkInterface struct {
	// List of IP Addresses
	IPAddresses []IPAddress `mapstructure:"ip_addresses"`

	// Network type (e.g. public, utility, private)
	Type InterfaceType `mapstructure:"type"`

	// Network UUID when connecting private network
	Network string `mapstructure:"network,omitempty"`
}

type IPAddress struct {
	// Default IP address. When set to `true` SSH communicator will connect to this IP after boot.
	Default bool `mapstructure:"default"`

	// IP address family (IPv4 or IPv6)
	Family string `mapstructure:"family"`

	// IP address. Note that at the moment using floating IPs is not supported.
	Address string `mapstructure:"address,omitempty"`
}

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`

	// The username to use when interfacing with the UpCloud API.
	Username string `mapstructure:"username"`

	// The password to use when interfacing with the UpCloud API.
	Password string `mapstructure:"password"`

	// The API token to use when interfacing with the UpCloud API. This is mutually exclusive with username and password.
	Token string `mapstructure:"token"`

	// The zone in which the server and template should be created (e.g. nl-ams1).
	Zone string `mapstructure:"zone" required:"true"`

	// The UUID of the storage you want to use as a template when creating the server.
	//
	// Optionally use `storage_name` parameter to find matching storage
	StorageUUID string `mapstructure:"storage_uuid" required:"true"`

	// The name of the storage that will be used to find the first matching storage in the list of existing templates.
	//
	// Note that `storage_uuid` parameter has higher priority. You should use either `storage_uuid` or `storage_name` for not strict matching (e.g "ubuntu server 20.04").
	StorageName string `mapstructure:"storage_name"`

	// The prefix to use for the generated template title. Defaults to `custom-image`.
	// You can use this option to easily differentiate between different templates.
	TemplatePrefix string `mapstructure:"template_prefix"`

	// Similarly to `template_prefix`, but this will allow you to set the full template name and not just the prefix.
	// Defaults to an empty string, meaning the name will be the storage title.
	// You can use this option to easily differentiate between different templates.
	// It cannot be used in conjunction with the prefix setting.
	TemplateName string `mapstructure:"template_name"`

	// The storage size in gigabytes. Defaults to `25`.
	// Changing this value is useful if you aim to build a template for larger server configurations where the preconfigured server disk is larger than 25 GB.
	// The operating system disk can also be later extended if needed. Note that Windows templates require large storage size, than default 25 Gb.
	StorageSize int `mapstructure:"storage_size"`

	// The storage tier to use. Available options are `maxiops`, `archive`, and `standard`. Defaults to `maxiops`.
	// For most production workloads, MaxIOPS is recommended for best performance.
	StorageTier string `mapstructure:"storage_tier"`

	// The amount of time to wait for resource state changes. Defaults to `5m`.
	Timeout time.Duration `mapstructure:"state_timeout_duration"`

	// The amount of time to wait after booting the server. Defaults to '0s'
	BootWait time.Duration `mapstructure:"boot_wait"`

	// The array of extra zones (locations) where created templates should be cloned.
	// Note that default `state_timeout_duration` is not enough for cloning, better to increase a value depending on storage size.
	CloneZones []string `mapstructure:"clone_zones"`

	// The array of network interfaces to request during the creation of the server for building the packer image.
	NetworkInterfaces []NetworkInterface `mapstructure:"network_interfaces"`

	// Path to SSH Private Key that will be used for provisioning and stored in the template.
	SSHPrivateKeyPath string `mapstructure:"ssh_private_key_path"`

	// Path to SSH Public Key that will be used for provisioning.
	SSHPublicKeyPath string `mapstructure:"ssh_public_key_path"`

	ctx interpolate.Context
}

// DefaultIPaddress returns default IP address and its type (public,private,utility).
func (c *Config) DefaultIPaddress() (*IPAddress, InterfaceType) {
	for _, iface := range c.NetworkInterfaces {
		for _, addr := range iface.IPAddresses {
			if addr.Default {
				return &addr, iface.Type
			}
		}
	}
	return nil, ""
}

func (c *Config) Prepare(raws ...interface{}) ([]string, error) {
	err := config.Decode(c, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &c.ctx,
	}, raws...)
	if err != nil {
		return nil, fmt.Errorf("failed to decode configuration: %w", err)
	}

	err = c.SetEnv()
	if err != nil {
		return nil, err
	}

	c.SetDefaults()

	if errs := c.validate(); errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	return nil, nil
}

// setDefaults sets default values for configuration fields.
func (c *Config) SetDefaults() {
	if c.TemplatePrefix == "" && len(c.TemplateName) == 0 {
		c.TemplatePrefix = DefaultTemplatePrefix
	}

	if c.StorageSize == 0 {
		c.StorageSize = DefaultStorageSize
	}

	if c.StorageTier == "" {
		c.StorageTier = DefaultStorageTier
	}

	if c.Timeout == 0 {
		c.Timeout = DefaultTimeout
	}

	if c.Comm.Type == "" {
		c.Comm.Type = DefaultCommunicator
	}

	if c.Comm.Type == "ssh" && c.Comm.SSHUsername == "" {
		c.Comm.SSHUsername = DefaultSSHUsername
	}
}

// validate validates the configuration and returns any errors.
func (c *Config) validate() *packer.MultiError {
	var errs *packer.MultiError
	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}

	// Validate required fields
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

	// Validate template configuration
	if templateErrs := c.validateTemplate(); templateErrs != nil {
		errs = packer.MultiErrorAppend(errs, templateErrs.Errors...)
	}

	// Validate network interfaces
	if networkErrs := c.validateNetworkInterfaces(); networkErrs != nil {
		errs = packer.MultiErrorAppend(errs, networkErrs.Errors...)
	}

	return errs
}

// validateTemplate checks template configuration.
func (c *Config) validateTemplate() *packer.MultiError {
	var errs *packer.MultiError

	if len(c.TemplatePrefix) > MaxTemplatePrefixLength {
		errs = packer.MultiErrorAppend(
			errs, errors.New("'template_prefix' must be 0-40 characters"),
		)
	}

	if len(c.TemplateName) > MaxTemplateNameLength {
		errs = packer.MultiErrorAppend(
			errs, errors.New("'template_name' is limited to 40 characters"),
		)
	}

	if len(c.TemplatePrefix) > 0 && len(c.TemplateName) > 0 {
		errs = packer.MultiErrorAppend(
			errs, errors.New("you can either use 'template_prefix' or 'template_name' in your configuration"),
		)
	}

	return errs
}

// get params from environment.
func (c *Config) SetEnv() error {
	creds, err := driver.CredentialsFromEnv(c.Username, c.Password, c.Token)
	if err != nil {
		return err //nolint:wrapcheck // Use the original error from the shared credentials package
	}

	c.Username = creds.Username
	c.Password = creds.Password
	c.Token = creds.Token
	return nil
}

// validateNetworkInterfaces checks network interface configuration.
func (c *Config) validateNetworkInterfaces() *packer.MultiError {
	var errs *packer.MultiError

	for i, iface := range c.NetworkInterfaces {
		// Validate interface type
		switch iface.Type {
		case InterfaceTypePublic, InterfaceTypePrivate, InterfaceTypeUtility:
			// valid
		default:
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("network interface %d has invalid type: %s", i, iface.Type))
		}

		// Validate network UUID for private networks
		if iface.Type == InterfaceTypePrivate {
			if iface.Network == "" {
				errs = packer.MultiErrorAppend(errs, fmt.Errorf("network interface %d: private network requires network UUID", i))
			} else if _, err := uuid.Parse(iface.Network); err != nil {
				errs = packer.MultiErrorAppend(errs, fmt.Errorf("network interface %d: invalid network UUID '%s'", i, iface.Network))
			}
		}

		// Validate IP addresses
		for j, ip := range iface.IPAddresses {
			if ip.Family != upcloud.IPAddressFamilyIPv4 && ip.Family != upcloud.IPAddressFamilyIPv6 {
				errs = packer.MultiErrorAppend(errs, fmt.Errorf("network interface %d, IP %d: invalid IP family '%s'", i, j, ip.Family))
			}

			if ip.Address == "" {
				continue
			}

			parsedIP := net.ParseIP(ip.Address)
			if parsedIP == nil {
				errs = packer.MultiErrorAppend(errs, fmt.Errorf("network interface %d, IP %d: invalid IP address '%s'", i, j, ip.Address))
				continue
			}

			isIPv4 := parsedIP.To4() != nil
			switch ip.Family {
			case upcloud.IPAddressFamilyIPv4:
				if !isIPv4 {
					errs = packer.MultiErrorAppend(errs, fmt.Errorf("network interface %d, IP %d: IP family is IPv4 but address is IPv6", i, j))
				}
			case upcloud.IPAddressFamilyIPv6:
				if isIPv4 {
					errs = packer.MultiErrorAppend(errs, fmt.Errorf("network interface %d, IP %d: IP family is IPv6 but address is IPv4", i, j))
				}
			}
		}
	}

	return errs
}
