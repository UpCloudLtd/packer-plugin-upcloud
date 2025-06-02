package upcloud

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"

	"github.com/UpCloudLtd/packer-plugin-upcloud/internal/driver"
)

// StepTeardownServer represents the step that stops the server before creating the image.
type StepTeardownServer struct{}

// Run runs the actual step.
func (s *StepTeardownServer) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	// Extract server details
	serverUUIDRaw := state.Get("server_uuid")
	serverUUID, ok := serverUUIDRaw.(string)
	if !ok {
		return stepHaltWithError(state, errors.New("server_uuid is not of expected type"))
	}
	serverTitleRaw := state.Get("server_title")
	serverTitle, ok := serverTitleRaw.(string)
	if !ok {
		return stepHaltWithError(state, errors.New("server_title is not of expected type"))
	}

	ui, ok := state.Get("ui").(packer.Ui)
	if !ok {
		return stepHaltWithError(state, errors.New("UI is not of expected type"))
	}
	driver, ok := state.Get("driver").(driver.Driver)
	if !ok {
		return stepHaltWithError(state, errors.New("driver is not of expected type"))
	}

	ui.Say(fmt.Sprintf("Stopping server %q...", serverTitle))

	err := driver.StopServer(ctx, serverUUID)
	if err != nil {
		return stepHaltWithError(state, err)
	}

	ui.Say(fmt.Sprintf("Server %q is now in 'stopped' state", serverTitle))

	return multistep.ActionContinue
}

func (s *StepTeardownServer) Cleanup(state multistep.StateBag) {}
