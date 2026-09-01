package upcloud

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"

	"github.com/UpCloudLtd/packer-plugin-upcloud/internal/driver"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
)

// StepCreateTemplate represents the step that creates a storage template from the newly created server.
type StepCreateTemplate struct {
	Config        *Config
	GeneratedData *packerbuilderdata.GeneratedData
}

// templateSource is a storage to templatize, together with the title to give the template.
type templateSource struct {
	storageUUID string
	title       string
}

// Run runs the actual step.
func (s *StepCreateTemplate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	serverUUIDRaw := state.Get("server_uuid")
	serverUUID, ok := serverUUIDRaw.(string)
	if !ok {
		return stepHaltWithError(state, errors.New("server_uuid is not of expected type"))
	}

	ui, ok := state.Get("ui").(packer.Ui)
	if !ok {
		return stepHaltWithError(state, errors.New("UI is not of expected type"))
	}
	drv, ok := state.Get("driver").(driver.Driver)
	if !ok {
		return stepHaltWithError(state, errors.New("driver is not of expected type"))
	}

	// get storage details for every disk of the server
	storages, err := drv.GetServerStorages(ctx, serverUUID)
	if err != nil {
		return stepHaltWithError(state, err)
	}

	baseTitle := s.baseTemplateTitle()

	var (
		artifacts          []*upcloud.Storage
		cleanupStorageUUID []string
	)
	if s.Config.ArtifactType == ArtifactTypeStorage {
		artifacts, err = s.createStorages(ctx, ui, drv, storages, baseTitle)
	} else {
		artifacts, cleanupStorageUUID, err = s.createTemplates(ctx, ui, drv, storages, baseTitle)
	}
	if err != nil {
		return stepHaltWithError(state, err)
	}

	state.Put("cleanup_storage_uuids", cleanupStorageUUID)
	state.Put("templates", artifacts)

	return multistep.ActionContinue
}

// createTemplates templatizes every disk of the builder server, plus a per-zone clone of each disk
// for every zone in 'clone_zones'. The intermediate clones are only there to be templatized, so
// they are returned for cleanup.
func (s *StepCreateTemplate) createTemplates(
	ctx context.Context, ui packer.Ui, drv driver.Driver, storages []upcloud.ServerStorageDevice, baseTitle string,
) ([]*upcloud.Storage, []string, error) {
	// cloning to zones
	cleanupStorageUUID := []string{}
	sources := make([]templateSource, 0, len(storages)*(len(s.Config.CloneZones)+1))

	for i, storage := range storages {
		title := templateTitleForDisk(baseTitle, i)
		sources = append(sources, templateSource{storageUUID: storage.UUID, title: title})

		for _, zone := range s.Config.CloneZones {
			ui.Say(fmt.Sprintf("Cloning storage %q to zone %q...", storage.UUID, zone))
			clonedTitle := fmt.Sprintf("packer-%s-cloned-disk%d", getNowString(), i+1)
			clonedStorage, err := drv.CloneStorage(ctx, storage.UUID, zone, clonedTitle, storage.Tier)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to clone disk %d: %w", i+1, err)
			}
			sources = append(sources, templateSource{storageUUID: clonedStorage.UUID, title: title})
			cleanupStorageUUID = append(cleanupStorageUUID, clonedStorage.UUID)
		}
	}
	ui.Say("Cloning completed...")

	// creating template
	templates := []*upcloud.Storage{}

	for _, source := range sources {
		ui.Say(fmt.Sprintf("Creating template for storage %q...", source.storageUUID))
		t, err := drv.CreateTemplate(ctx, source.storageUUID, source.title, s.Config.TemplateLabels)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create template %q: %w", source.title, err)
		}

		templates = append(templates, t)
		ui.Say(fmt.Sprintf("Template for storage %q created...", source.storageUUID))
	}

	return templates, cleanupStorageUUID, nil
}

// createStorages clones every disk of the builder server into a detached regular storage. One set
// serves every zone, since such a storage can be cloned into any zone, so 'clone_zones' does not
// apply here and is rejected by the configuration. The clones are the artifact itself, so there is
// nothing to clean up afterwards.
func (s *StepCreateTemplate) createStorages(
	ctx context.Context, ui packer.Ui, drv driver.Driver, storages []upcloud.ServerStorageDevice, baseTitle string,
) ([]*upcloud.Storage, error) {
	artifacts := make([]*upcloud.Storage, 0, len(storages))
	for i, storage := range storages {
		title := templateTitleForDisk(baseTitle, i)
		ui.Say(fmt.Sprintf("Cloning storage %q...", storage.UUID))
		clonedStorage, err := drv.CloneStorage(ctx, storage.UUID, s.Config.Zone, title, storage.Tier)
		if err != nil {
			return nil, fmt.Errorf("failed to clone disk %d: %w", i+1, err)
		}
		artifacts = append(artifacts, clonedStorage)
		ui.Say(fmt.Sprintf("Storage %q created...", clonedStorage.UUID))
	}
	ui.Say("Cloning completed...")

	return artifacts, nil
}

// baseTemplateTitle returns the title to give the created templates. It is evaluated once per
// build, so that every template of the same build shares the same timestamp.
func (s *StepCreateTemplate) baseTemplateTitle() string {
	// we either use template name or prefix.
	if len(s.Config.TemplatePrefix) > 0 {
		return fmt.Sprintf("%s-%s", s.Config.TemplatePrefix, getNowString())
	}
	return s.Config.TemplateName
}

// templateTitleForDisk returns the template title for the disk at the given index. The first disk
// keeps the base title as-is; the additional ones get a "-diskN" suffix, since templates created
// from the same server would otherwise be indistinguishable.
func templateTitleForDisk(baseTitle string, diskIndex int) string {
	if diskIndex == 0 {
		return baseTitle
	}
	return fmt.Sprintf("%s-disk%d", baseTitle, diskIndex+1)
}

// Cleanup cleans up after the step.
func (s *StepCreateTemplate) Cleanup(state multistep.StateBag) {
	rawStorageUuids, ok := state.GetOk("cleanup_storage_uuids")

	if !ok {
		return
	}
	ctx, cancel := contextWithDefaultTimeout()
	defer cancel()
	storageUuids, ok := rawStorageUuids.([]string)
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

	for _, uuid := range storageUuids {
		ui.Say(fmt.Sprintf("Delete storage %q...", uuid))

		err := driver.DeleteTemplate(ctx, uuid)
		if err != nil {
			ui.Error(err.Error())
		}
	}
}
