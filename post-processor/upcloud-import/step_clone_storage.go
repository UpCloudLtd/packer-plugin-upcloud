package upcloudimport

import (
	"context"
	"fmt"
	"sync"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepCloneStorage struct {
	postProcessor *PostProcessor
}

func (s *stepCloneStorage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if len(s.postProcessor.config.Zones) < 2 {
		return multistep.ActionContinue
	}
	ui := state.Get(stateUI).(packer.Ui)
	storages, err := getStorages(state)
	if err != nil {
		return haltOnError(ui, state, err)
	}
	if len(storages) < 1 {
		ui.Error("no storages to clone from")
		return multistep.ActionHalt
	}

	zones := s.postProcessor.config.Zones[1:]
	var halt bool
	var wg sync.WaitGroup
	wg.Add(len(zones))
	for _, z := range zones {
		go func(zone string) {
			defer wg.Done()
			ui.Say(fmt.Sprintf("Cloning storage '%s' from %s to %s", storages[0].Title, storages[0].Zone, zone))
			t, err := s.postProcessor.driver.CloneStorage(storages[0].UUID, zone, storages[0].Title)
			if err != nil {
				ui.Error(err.Error())
				halt = true
				return
			}
			storages = append(storages, t)
		}(z)
	}
	wg.Wait()
	if halt {
		return multistep.ActionHalt
	}
	state.Put(stateStorages, storages)
	return multistep.ActionContinue
}

func (s *stepCloneStorage) Cleanup(state multistep.StateBag) {
	ui := state.Get(stateUI).(packer.Ui)
	if err := cleanupDevices(ui, s.postProcessor.driver, state); err != nil {
		ui.Error(err.Error())
	}
}
