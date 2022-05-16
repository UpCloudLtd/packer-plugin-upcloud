package upcloudimport

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
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

	cs := sha256.New()
	if _, err := io.Copy(cs, fd); err != nil {
		return haltOnError(ui, state, err)
	}

	// reset Reader after io.Copy
	if _, err := fd.Seek(0, 0); err != nil {
		return haltOnError(ui, state, err)
	}

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
	csString := hex.EncodeToString(cs.Sum(nil)[:])
	if importDetails.SHA256Sum != csString {
		return haltOnError(ui, state, fmt.Errorf("uploaded image checksum mismatch want '%s' got '%s'", csString, importDetails.SHA256Sum))
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
