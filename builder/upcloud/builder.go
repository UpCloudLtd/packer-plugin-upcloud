package upcloud

import (
	"context"
	"fmt"
	"time"

	"github.com/UpCloudLtd/packer-plugin-upcloud/internal/driver"
	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
)

const (
	BuilderId                    = "upcloud.builder"
	defaultTimeout time.Duration = 1 * time.Hour
)

type Builder struct {
	config Config
	runner multistep.Runner
	driver driver.Driver
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) (generatedVars []string, warnings []string, err error) {
	warnings, errs := b.config.Prepare(raws...)
	if errs != nil {
		return nil, warnings, errs
	}

	buildGeneratedData := []string{
		"ServerUUID",
		"ServerTitle",
		"ServerSize",
		"TemplateUUID",
		"TemplateTitle",
		"TemplateSize",
	}
	return buildGeneratedData, nil, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	// NOTE: context deadline is not set by default.

	// Setup the state bag and initial state for the steps
	b.driver = driver.NewDriver(&driver.DriverConfig{
		Username:    b.config.Username,
		Password:    b.config.Password,
		Timeout:     b.config.Timeout,
		SSHUsername: b.config.Comm.SSHUsername,
	})

	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("hook", hook)
	state.Put("ui", ui)
	state.Put("driver", b.driver)

	generatedData := &packerbuilderdata.GeneratedData{State: state}

	// Build the steps
	steps := []multistep.Step{
		&StepCreateSSHKey{
			Debug:        b.config.PackerDebug,
			DebugKeyPath: fmt.Sprintf("ssh_key-%s.pem", b.config.PackerBuildName),
		},
		&StepCreateServer{
			Config:        &b.config,
			GeneratedData: generatedData,
		},
		b.communicatorStep(),
		&commonsteps.StepProvision{},
		&commonsteps.StepCleanupTempKeys{
			Comm: &b.config.Comm,
		},
		&StepTeardownServer{},
		&StepCreateTemplate{
			Config:        &b.config,
			GeneratedData: generatedData,
		},
	}

	// Run
	b.runner = commonsteps.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	// If there was an error, return that
	if err, ok := state.GetOk("error"); ok {
		return nil, err.(error)
	}

	templates, ok := state.GetOk("templates")
	if !ok {
		return nil, fmt.Errorf("No template found in state, the build was probably cancelled")
	}

	artifact := &Artifact{
		Templates: templates.([]*upcloud.Storage),
		config:    &b.config,
		driver:    b.driver,
		StateData: map[string]interface{}{
			"generated_data":        state.Get("generated_data"),
			"template_prefix":       b.config.TemplatePrefix,
			"template_name":         b.config.TemplateName,
			"source_template_uuid":  state.Get("source_template_uuid"),
			"source_template_title": state.Get("source_template_title"),
		},
	}

	return artifact, nil
}

// CommunicatorStep returns step based on communicator type
// We currently support only SSH communicator but 'none' type
// can also be used for e.g. testing purposes
func (b *Builder) communicatorStep() multistep.Step {
	switch b.config.Comm.Type {
	case "none":
		return &communicator.StepConnect{
			Config: &b.config.Comm,
		}
	default:
		return &communicator.StepConnect{
			Config:    &b.config.Comm,
			Host:      sshHostCallback,
			SSHConfig: b.config.Comm.SSHConfigFunc(),
		}
	}
}
