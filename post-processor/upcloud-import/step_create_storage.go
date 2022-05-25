package upcloudimport

import (
	"context"
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
	ui := state.Get(stateUI).(packer.Ui)
	storages, err := getStorages(state)
	if err != nil {
		return haltOnError(ui, state, err)
	}
	size := s.image.SizeGB()
	if size < storageMinSizeGB {
		size = storageMinSizeGB
	}
	ui.Say(fmt.Sprintf("Creating storage device (%dGB) for '%s' image", size, s.image.File()))
	storage, err := s.postProcessor.driver.CreateTemplateStorage(
		fmt.Sprintf("%s-%s", BuilderID, time.Now().Format(timestampSuffixLayout)), s.postProcessor.config.Zones[0], size)

	if err != nil {
		return haltOnError(ui, state, err)
	}
	storages = append(storages, storage)
	state.Put(stateStorages, storages)
	ui.Say(fmt.Sprintf("Storage '%s' (%s) created", storage.Title, storage.UUID))
	return multistep.ActionContinue
}

func (s *stepCreateStorage) Cleanup(state multistep.StateBag) {
	ui := state.Get(stateUI).(packer.Ui)
	if err := cleanupDevices(ui, s.postProcessor.driver, state); err != nil {
		ui.Error(err.Error())
	}
}
