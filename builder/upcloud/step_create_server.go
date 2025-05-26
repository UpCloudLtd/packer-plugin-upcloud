package upcloud

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"

	"github.com/UpCloudLtd/packer-plugin-upcloud/internal/driver"
)

// StepCreateServer represents the step that creates a server.
type StepCreateServer struct {
	Config        *Config
	GeneratedData *packerbuilderdata.GeneratedData
}

// Run runs the actual step.
func (s *StepCreateServer) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui, ok := state.Get("ui").(packer.Ui)
	if !ok {
		return stepHaltWithError(state, errors.New("UI is not of expected type"))
	}
	drv, ok := state.Get("driver").(driver.Driver)
	if !ok {
		return stepHaltWithError(state, errors.New("driver is not of expected type"))
	}

	rawSshKeyPublic, ok := state.GetOk("ssh_key_public")
	if !ok {
		return stepHaltWithError(state, errors.New("SSH public key is missing"))
	}
	sshKeyPublic, ok := rawSshKeyPublic.(string)
	if !ok {
		return stepHaltWithError(state, errors.New("SSH public key is not of expected type"))
	}

	ui.Say("Getting storage...")

	storage, err := drv.GetStorage(ctx, s.Config.StorageUUID, s.Config.StorageName)
	if err != nil {
		return stepHaltWithError(state, err)
	}

	ui.Say(fmt.Sprintf("Creating server based on storage %q...", storage.Title))

	networking := defaultNetworking()
	if len(s.Config.NetworkInterfaces) > 0 {
		networking = convertNetworkTypes(s.Config.NetworkInterfaces)
	}
	response, err := drv.CreateServer(ctx, &driver.ServerOpts{
		StorageUuid:  storage.UUID,
		StorageSize:  s.Config.StorageSize,
		Zone:         s.Config.Zone,
		SshPublicKey: sshKeyPublic,
		Networking:   networking,
		StorageTier:  s.Config.StorageTier,
	})
	if err != nil {
		return stepHaltWithError(state, err)
	}

	ui.Say(fmt.Sprintf("Server %q created and in 'started' state", response.Title))

	addr, infType := s.Config.DefaultIPaddress()
	if addr != nil {
		if addr.Address == "" {
			addr, err = findIPAddressByType(response.IPAddresses, infType)
			if err != nil {
				return stepHaltWithError(state, err)
			}
		}
		ui.Say(fmt.Sprintf("Selecting default ip '%s' as Server IP", addr.Address))
	} else {
		addr, err = findIPAddressByType(response.IPAddresses, InterfaceTypePublic)
		if err != nil {
			return stepHaltWithError(state, err)
		}
		ui.Say(fmt.Sprintf("Auto-selecting ip '%s' as Server IP", addr.Address))
	}

	state.Put("source_template_uuid", storage.UUID)
	state.Put("source_template_title", storage.Title)
	state.Put("server_ip_address", addr)
	state.Put("server_uuid", response.UUID)
	state.Put("server_title", response.Title)

	s.GeneratedData.Put("ServerUUID", response.UUID)
	s.GeneratedData.Put("ServerTitle", response.Title)
	s.GeneratedData.Put("ServerSize", response.Plan)
	if s.Config.BootWait > 0 {
		ui.Say(fmt.Sprintf("Waitig boot: %s", s.Config.BootWait.String()))
		time.Sleep(s.Config.BootWait)
	}
	return multistep.ActionContinue
}

// Cleanup stops and destroys the server if server details are found in the state.
func (s *StepCreateServer) Cleanup(state multistep.StateBag) {
	ctx, cancel := contextWithDefaultTimeout()
	defer cancel()
	// Extract server uuid, return if no uuid has been stored
	rawServerUuid, ok := state.GetOk("server_uuid")

	if !ok {
		return
	}

	serverUuid, ok := rawServerUuid.(string)
	if !ok {
		return
	}
	serverTitleRaw := state.Get("server_title")
	serverTitle, ok := serverTitleRaw.(string)
	if !ok {
		return
	}

	uiRaw := state.Get("ui")
	ui, ok := uiRaw.(packer.Ui)
	if !ok {
		return
	}
	driverRaw := state.Get("driver")
	driver, ok := driverRaw.(driver.Driver)
	if !ok {
		return
	}

	// stop server
	ui.Say(fmt.Sprintf("Stopping server %q...", serverTitle))

	err := driver.StopServer(ctx, serverUuid)
	if err != nil {
		ui.Error(err.Error())
		return
	}

	// delete server
	ui.Say(fmt.Sprintf("Deleting server %q...", serverTitle))

	err = driver.DeleteServer(ctx, serverUuid)
	if err != nil {
		ui.Error(err.Error())
		return
	}
}
