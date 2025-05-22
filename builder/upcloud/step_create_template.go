package upcloud

import (
	"context"
	"fmt"

	"github.com/UpCloudLtd/packer-plugin-upcloud/internal/driver"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
)

// StepCreateTemplate represents the step that creates a storage template from the newly created server
type StepCreateTemplate struct {
	Config        *Config
	GeneratedData *packerbuilderdata.GeneratedData
}

// Run runs the actual step
func (s *StepCreateTemplate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	serverUuid := state.Get("server_uuid").(string)

	ui := state.Get("ui").(packer.Ui)
	drv := state.Get("driver").(driver.Driver)

	// get storage details
	storage, err := drv.GetServerStorage(ctx, serverUuid)
	if err != nil {
		return stepHaltWithError(state, err)
	}

	// cloning to zones
	cleanupStorageUuid := []string{}
	storageUuids := []string{}
	storageUuids = append(storageUuids, storage.UUID)

	for _, zone := range s.Config.CloneZones {
		ui.Say(fmt.Sprintf("Cloning storage %q to zone %q...", storage.UUID, zone))
		title := fmt.Sprintf("packer-%s-cloned-disk1", getNowString())
		clonedStorage, err := drv.CloneStorage(ctx, storage.UUID, zone, title)
		if err != nil {
			return stepHaltWithError(state, err)
		}
		storageUuids = append(storageUuids, clonedStorage.UUID)
		cleanupStorageUuid = append(cleanupStorageUuid, clonedStorage.UUID)
	}
	ui.Say("Cloning completed...")

	// creating template
	templates := []*upcloud.Storage{}

	// we either use template name or prefix.
	var templateTitle string
	if len(s.Config.TemplatePrefix) > 0 {
		templateTitle = fmt.Sprintf("%s-%s", s.Config.TemplatePrefix, getNowString())
	} else {
		templateTitle = s.Config.TemplateName
	}

	for _, uuid := range storageUuids {
		ui.Say(fmt.Sprintf("Creating template for storage %q...", uuid))
		t, err := drv.CreateTemplate(ctx, uuid, templateTitle)
		if err != nil {
			return stepHaltWithError(state, err)
		}

		templates = append(templates, t)
		ui.Say(fmt.Sprintf("Template for storage %q created...", uuid))
	}

	state.Put("cleanup_storage_uuids", cleanupStorageUuid)
	state.Put("templates", templates)

	return multistep.ActionContinue
}

// Cleanup cleans up after the step
func (s *StepCreateTemplate) Cleanup(state multistep.StateBag) {
	rawStorageUuids, ok := state.GetOk("cleanup_storage_uuids")

	if !ok {
		return
	}
	ctx, cancel := contextWithDefaultTimeout()
	defer cancel()
	storageUuids := rawStorageUuids.([]string)

	ui := state.Get("ui").(packer.Ui)
	driver := state.Get("driver").(driver.Driver)

	for _, uuid := range storageUuids {
		ui.Say(fmt.Sprintf("Delete storage %q...", uuid))

		err := driver.DeleteTemplate(ctx, uuid)
		if err != nil {
			ui.Error(err.Error())
		}
	}
}
