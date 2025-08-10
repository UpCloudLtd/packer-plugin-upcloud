package upcloudimport

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

const (
	storageMinSizeGB int = 10
	storageMaxSizeGB int = 4096
)

type stepCreateStorage struct {
	postProcessor *PostProcessor
	image         *image
}

func (s *stepCreateStorage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	uiRaw := state.Get(stateUI)
	ui, ok := uiRaw.(packer.Ui)
	if !ok {
		return haltOnError(nil, state, errors.New("UI is not of expected type"))
	}
	storages, err := getStorages(state)
	if err != nil {
		return haltOnError(ui, state, err)
	}
	var size int
	if s.postProcessor.config.StorageSize > 0 {
		size = s.postProcessor.config.StorageSize
		ui.Say(fmt.Sprintf("Creating storage device (%dGB) for '%s' image using manually specified size", size, s.image.File()))
	} else {
		size = s.image.SizeGB()
		if size < storageMinSizeGB {
			size = storageMinSizeGB
		}
		ui.Say(fmt.Sprintf("Creating storage device (%dGB) for '%s' image", size, s.image.File()))
	}
	storage, err := s.postProcessor.driver.CreateTemplateStorage(ctx,
		fmt.Sprintf("%s-%s", BuilderID, time.Now().Format(timestampSuffixLayout)),
		s.postProcessor.config.Zones[0],
		size,
		s.postProcessor.config.StorageTier)
	if err != nil {
		return haltOnError(ui, state, err)
	}
	storages = append(storages, storage)
	state.Put(stateStorages, storages)
	ui.Say(fmt.Sprintf("Storage '%s' (%s) created", storage.Title, storage.UUID))
	return multistep.ActionContinue
}

func (s *stepCreateStorage) Cleanup(state multistep.StateBag) {
	ctx, cancel := contextWithDefaultTimeout()
	defer cancel()
	uiRaw := state.Get(stateUI)
	ui, ok := uiRaw.(packer.Ui)
	if !ok {
		return
	}
	if err := cleanupDevices(ctx, ui, s.postProcessor.driver, state); err != nil {
		ui.Error(err.Error())
	}
}
