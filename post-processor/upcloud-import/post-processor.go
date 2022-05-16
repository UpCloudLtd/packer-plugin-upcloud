package upcloudimport

import (
	"context"
	"errors"
	"fmt"

	"github.com/UpCloudLtd/packer-plugin-upcloud/internal/driver"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

const (
	BuilderID string = "packer.post-processor.upcloud-import"

	// https://www.packer.io/docs/provisioners/file
	fileBuilderID = "packer.file"

	// https://www.packer.io/docs/post-processors/artifice
	artificeBuilderID = "packer.post-processor.artifice"

	// https://www.packer.io/docs/post-processors/compress
	compressBuilderID string = "packer.post-processor.compress"

	// https://www.packer.io/plugins/builders/qemu
	qemuBuilderID string = "transcend.qemu"

	stateUI        string = "ui"
	stateArtifact  string = "artifact"
	stateStorages  string = "storages"
	stateTemplates string = "templates"
)

type PostProcessor struct {
	config *Config
	runner multistep.Runner
	driver driver.Driver
}

func (p *PostProcessor) ConfigSpec() hcldec.ObjectSpec {
	return p.config.FlatMapstructure().HCL2Spec()
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	var err error
	if p.config, err = NewConfig(raws...); err != nil {
		return err
	}
	p.driver = driver.NewDriver(&driver.DriverConfig{
		Username: p.config.Username,
		Password: p.config.Password,
		Timeout:  p.config.Timeout,
	})
	return p.validate()
}

// PostProcess takes a previously created Artifact and produces another
// Artifact. If an error occurs, it should return that error. If `keep` is
// true, then the previous artifact defaults to being kept if user has not
// given a value to keep_input_artifact. If forceOverride is true, then any
// user input for keep_input_artifact is ignored and the artifact is either
// kept or discarded according to the value set in `keep`.
// PostProcess is cancellable using context
func (p *PostProcessor) PostProcess(ctx context.Context, ui packer.Ui, a packer.Artifact) (packer.Artifact, bool, bool, error) {
	switch a.BuilderId() {
	case qemuBuilderID, fileBuilderID, compressBuilderID, artificeBuilderID:
		break
	default:
		return nil, false, false,
			fmt.Errorf("unsupported artifact type %s (suported types: %s, %s, %s, %s)",
				a.BuilderId(), qemuBuilderID, fileBuilderID, compressBuilderID, artificeBuilderID)
	}

	if len(a.Files()) < 1 {
		return nil, false, false, fmt.Errorf("%s didn't receive any files", BuilderID)
	}
	im, err := newImage(a.Files()[0])
	if err != nil {
		return nil, false, false, err
	}
	state := new(multistep.BasicStateBag)
	state.Put(stateUI, ui)
	state.Put(stateArtifact, a)
	state.Put(stateStorages, make([]*upcloud.Storage, 0))
	state.Put(stateTemplates, make([]*upcloud.Storage, 0))

	steps := []multistep.Step{
		&stepCreateStorage{postProcessor: p, image: im},
		&stepUploadImage{postProcessor: p, image: im},
		&stepCloneStorage{postProcessor: p},
		&stepCreateTemplate{postProcessor: p},
	}
	p.runner = commonsteps.NewRunnerWithPauseFn(steps, p.config.PackerConfig, ui, state)
	p.runner.Run(ctx, state)

	if e, ok := state.GetOk("error"); ok {
		return nil, false, false, e.(error)
	}

	if _, ok := state.GetOk(multistep.StateHalted); ok {
		return nil, false, false, errors.New("post-processing halted")
	}

	return &Artifact{
		postProcessor: p,
		templates:     state.Get(stateTemplates).([]*upcloud.Storage),
		stateData: map[string]interface{}{
			"generated_data": state.Get("generated_data"),
		},
		driver: p.driver,
	}, false, false, nil
}

func (p *PostProcessor) validate() error {
	if !p.config.ReplaceExisting {
		for _, zone := range p.config.Zones {
			s, err := p.driver.GetTemplateByName(p.config.TemplateName, zone)
			if err == nil && s.UUID != "" {
				return fmt.Errorf("template with the name '%s' already exists at %s zone. Change the name or set replace_existing to true", s.Title, zone)
			}
		}
	}
	availableZones := p.driver.GetAvailableZones()
	if len(availableZones) == 0 {
		return errors.New("unable to get available zones")
	}
	for _, zone := range p.config.Zones {
		if !zoneExists(zone, availableZones) {
			return fmt.Errorf("'%s' is not valid zone", zone)
		}
	}
	return nil
}

func zoneExists(zone string, availableZones []string) bool {
	for _, z := range availableZones {
		if zone == z {
			return true
		}
	}
	return false
}
