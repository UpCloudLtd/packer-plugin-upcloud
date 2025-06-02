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
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
)

// StepCreateServer represents the step that creates a server.
type StepCreateServer struct {
	Config        *Config
	GeneratedData *packerbuilderdata.GeneratedData
}

// Run runs the actual step.
func (s *StepCreateServer) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui, drv, sshKeyPublic, err := s.validateState(state)
	if err != nil {
		return stepHaltWithError(state, err)
	}

	storage, err := s.getStorage(ctx, ui, drv)
	if err != nil {
		return stepHaltWithError(state, err)
	}

	response, err := s.createServer(ctx, ui, drv, storage, sshKeyPublic)
	if err != nil {
		return stepHaltWithError(state, err)
	}

	addr, err := s.selectIPAddress(ui, response)
	if err != nil {
		return stepHaltWithError(state, err)
	}

	s.populateState(state, storage, response, addr)
	s.handleBootWait(ui)

	return multistep.ActionContinue
}

// validateState validates and extracts required values from the state bag.
func (s *StepCreateServer) validateState(state multistep.StateBag) (packer.Ui, driver.Driver, string, error) {
	ui, ok := state.Get("ui").(packer.Ui)
	if !ok {
		return nil, nil, "", errors.New("UI is not of expected type")
	}

	drv, ok := state.Get("driver").(driver.Driver)
	if !ok {
		return nil, nil, "", errors.New("driver is not of expected type")
	}

	rawSSHKeyPublic, ok := state.GetOk("ssh_key_public")
	if !ok {
		return nil, nil, "", errors.New("SSH public key is missing")
	}

	sshKeyPublic, ok := rawSSHKeyPublic.(string)
	if !ok {
		return nil, nil, "", errors.New("SSH public key is not of expected type")
	}

	return ui, drv, sshKeyPublic, nil
}

// getStorage retrieves the storage template to use for server creation.
func (s *StepCreateServer) getStorage(ctx context.Context, ui packer.Ui, drv driver.Driver) (*upcloud.Storage, error) {
	ui.Say("Getting storage...")
	storage, err := drv.GetStorage(ctx, s.Config.StorageUUID, s.Config.StorageName)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage (UUID: %s, Name: %s): %w", s.Config.StorageUUID, s.Config.StorageName, err)
	}
	return storage, nil
}

// createServer creates the server with the specified configuration.
func (s *StepCreateServer) createServer(ctx context.Context, ui packer.Ui, drv driver.Driver, storage *upcloud.Storage, sshKeyPublic string) (*upcloud.ServerDetails, error) {
	ui.Say(fmt.Sprintf("Creating server based on storage %q...", storage.Title))

	networking := defaultNetworking()
	if len(s.Config.NetworkInterfaces) > 0 {
		networking = convertNetworkTypes(s.Config.NetworkInterfaces)
	}

	response, err := drv.CreateServer(ctx, &driver.ServerOpts{
		StorageUUID:  storage.UUID,
		StorageSize:  s.Config.StorageSize,
		Zone:         s.Config.Zone,
		SSHPublicKey: sshKeyPublic,
		Networking:   networking,
		StorageTier:  s.Config.StorageTier,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create server in zone %s: %w", s.Config.Zone, err)
	}

	ui.Say(fmt.Sprintf("Server %q created and in 'started' state", response.Title))
	return response, nil
}

// selectIPAddress selects the appropriate IP address for the server.
func (s *StepCreateServer) selectIPAddress(ui packer.Ui, response *upcloud.ServerDetails) (*IPAddress, error) {
	addr, infType := s.Config.DefaultIPaddress()

	// Handle case when no default IP address is configured
	if addr == nil {
		return s.selectAutoIPAddress(ui, response)
	}

	// Handle case when default IP address needs to be resolved
	if addr.Address == "" {
		return s.selectDefaultIPAddress(ui, response, infType)
	}

	// Use the explicitly configured IP address
	ui.Say(fmt.Sprintf("Selecting default ip '%s' as Server IP", addr.Address))
	return addr, nil
}

// selectAutoIPAddress automatically selects a public IP address.
func (s *StepCreateServer) selectAutoIPAddress(ui packer.Ui, response *upcloud.ServerDetails) (*IPAddress, error) {
	addr, err := findIPAddressByType(response.IPAddresses, InterfaceTypePublic)
	if err != nil {
		return nil, err
	}
	ui.Say(fmt.Sprintf("Auto-selecting ip '%s' as Server IP", addr.Address))
	return addr, nil
}

// selectDefaultIPAddress selects an IP address based on the configured interface type.
func (s *StepCreateServer) selectDefaultIPAddress(ui packer.Ui, response *upcloud.ServerDetails, infType InterfaceType) (*IPAddress, error) {
	addr, err := findIPAddressByType(response.IPAddresses, infType)
	if err != nil {
		return nil, err
	}
	ui.Say(fmt.Sprintf("Selecting default ip '%s' as Server IP", addr.Address))
	return addr, nil
}

// populateState populates the state bag with server and storage information.
func (s *StepCreateServer) populateState(state multistep.StateBag, storage *upcloud.Storage, response *upcloud.ServerDetails, addr *IPAddress) {
	state.Put("source_template_uuid", storage.UUID)
	state.Put("source_template_title", storage.Title)
	state.Put("server_ip_address", addr)
	state.Put("server_uuid", response.UUID)
	state.Put("server_title", response.Title)

	s.GeneratedData.Put("ServerUUID", response.UUID)
	s.GeneratedData.Put("ServerTitle", response.Title)
	s.GeneratedData.Put("ServerSize", response.Plan)
}

// handleBootWait handles the boot wait period if configured.
func (s *StepCreateServer) handleBootWait(ui packer.Ui) {
	if s.Config.BootWait > 0 {
		ui.Say(fmt.Sprintf("Waitig boot: %s", s.Config.BootWait.String()))
		time.Sleep(s.Config.BootWait)
	}
}

// Cleanup stops and destroys the server if server details are found in the state.
func (s *StepCreateServer) Cleanup(state multistep.StateBag) {
	ctx, cancel := contextWithDefaultTimeout()
	defer cancel()
	// Extract server uuid, return if no uuid has been stored
	rawServerUUID, ok := state.GetOk("server_uuid")

	if !ok {
		return
	}

	serverUUID, ok := rawServerUUID.(string)
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

	err := driver.StopServer(ctx, serverUUID)
	if err != nil {
		ui.Error(err.Error())
		return
	}

	// delete server
	ui.Say(fmt.Sprintf("Deleting server %q...", serverTitle))

	err = driver.DeleteServer(ctx, serverUUID)
	if err != nil {
		ui.Error(err.Error())
		return
	}
}
