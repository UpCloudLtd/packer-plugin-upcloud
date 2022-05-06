package upcloudimport

import (
	"fmt"

	"github.com/UpCloudLtd/packer-plugin-upcloud/internal/driver"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

func cleanupDevices(ui packer.Ui, driver driver.Driver, state multistep.StateBag) error {
	storages, err := getStorages(state)
	if err != nil {
		return err
	}
	for _, s := range storages {
		if err = deleteStorageIfExists(ui, driver, s); err != nil {
			return err
		}
	}
	state.Put(stateStorages, make([]*upcloud.Storage, 0))
	return nil
}

func cleanupTemplates(ui packer.Ui, driver driver.Driver, state multistep.StateBag) error {
	storages, err := getTemplates(state)
	if err != nil {
		return err
	}
	for _, s := range storages {
		if err = deleteStorageIfExists(ui, driver, s); err != nil {
			return err
		}
	}
	state.Put(stateTemplates, make([]*upcloud.Storage, 0))
	return nil
}

func deleteStorageIfExists(ui packer.Ui, driver driver.Driver, storage *upcloud.Storage) error {
	if _, err := driver.GetStorage(storage.UUID, ""); err == nil {
		ui.Say(fmt.Sprintf("Cleanup storage '%s' (%s)", storage.Title, storage.UUID))
		if err := driver.DeleteStorage(storage.UUID); err != nil {
			ui.Error(err.Error())
			return err
		}
	}
	return nil
}

func getStorages(state multistep.StateBag) ([]*upcloud.Storage, error) {
	storages, ok := state.Get(stateStorages).([]*upcloud.Storage)
	if !ok {
		return nil, fmt.Errorf("Unable to get '%s' from state", stateStorages)
	}
	return storages, nil
}

func getTemplates(state multistep.StateBag) ([]*upcloud.Storage, error) {
	storages, ok := state.Get(stateTemplates).([]*upcloud.Storage)
	if !ok {
		return nil, fmt.Errorf("Unable to get '%s' from state", stateTemplates)
	}
	return storages, nil
}

func haltOnError(ui packer.Ui, state multistep.StateBag, err error) multistep.StepAction {
	ui.Error(err.Error())
	return multistep.ActionHalt
}
