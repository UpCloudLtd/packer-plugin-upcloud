package upcloudimport

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepUploadImage struct {
	postProcessor *PostProcessor
	image         *image
}

func (s *stepUploadImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get(stateUI).(packer.Ui)
	storages, err := getStorages(state)
	if err != nil {
		return haltOnError(ui, state, err)
	}
	ui.Say(fmt.Sprintf("Starting to upload image '%s' (%s) into storage '%s'", s.image.File(), s.image.ContentType, storages[0].Title))

	fd, err := os.Open(s.image.Path)
	if err != nil {
		return haltOnError(ui, state, err)
	}
	defer fd.Close()

	t1 := time.Now()
	importDetails, err := s.postProcessor.driver.ImportStorage(storages[0].UUID, s.image.ContentType, fd)
	if err != nil {
		return haltOnError(ui, state, err)
	}

	ui.Say(fmt.Sprintf("Image '%s' uploaded to storage '%s' (%s) in %s", s.image.File(), storages[0].Title, storages[0].UUID, time.Since(t1)))
	ui.Say(fmt.Sprintf("Waiting storage '%s' to become online", storages[0].Title))

	if _, err := s.postProcessor.driver.WaitStorageOnline(storages[0].UUID); err != nil {
		return haltOnError(ui, state, err)
	}

	// do checksum check after storage is online so that cleanup works if there is a problem
	if err := s.image.CheckSHA256(importDetails.SHA256Sum); err != nil {
		return haltOnError(ui, state, err)
	}

	state.Put(stateStorages, storages)

	return multistep.ActionContinue
}

func (s *stepUploadImage) Cleanup(state multistep.StateBag) {
	ui := state.Get(stateUI).(packer.Ui)
	if err := cleanupDevices(ui, s.postProcessor.driver, state); err != nil {
		ui.Error(err.Error())
	}
}
