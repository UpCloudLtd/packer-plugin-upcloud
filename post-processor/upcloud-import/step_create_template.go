package upcloudimport

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

const (
	timestampLayout       string = "2006-01-02 15:04:05"
	timestampSuffixLayout string = "20060102-150405"
)

type stepCreateTemplate struct {
	postProcessor *PostProcessor
}

func (s *stepCreateTemplate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get(stateUI).(packer.Ui)

	storages, err := getStorages(state)
	if err != nil {
		return haltOnError(ui, state, err)
	}

	templates, err := getTemplates(state)
	if err != nil {
		return haltOnError(ui, state, err)
	}

	var halt bool
	var wg sync.WaitGroup
	wg.Add(len(storages))
	for _, storage := range storages {
		ui.Say(fmt.Sprintf("Creating template based on storage '%s' (%s) [%s]", storage.Title, storage.UUID, storage.Zone))
		go func(sto *upcloud.Storage) {
			defer wg.Done()
			template, err := s.createTemplateBasedOnStorage(ui, sto)
			if err != nil {
				halt = true
				return
			}
			templates = append(templates, template)
		}(storage)
	}

	wg.Wait()

	if err := cleanupDevices(ui, s.postProcessor.driver, state); err != nil {
		return haltOnError(ui, state, err)
	}

	state.Put(stateTemplates, templates)

	if halt {
		if err := cleanupTemplates(ui, s.postProcessor.driver, state); err != nil {
			ui.Error(err.Error())
		}
		return multistep.ActionHalt
	}
	return multistep.ActionContinue
}

func (s *stepCreateTemplate) Cleanup(state multistep.StateBag) {
	ui := state.Get(stateUI).(packer.Ui)
	if err := cleanupDevices(ui, s.postProcessor.driver, state); err != nil {
		ui.Error(err.Error())
	}
}

func (s *stepCreateTemplate) createTemplateBasedOnStorage(ui packer.Ui, storage *upcloud.Storage) (*upcloud.Storage, error) {
	var existingTemplate *upcloud.Storage
	var err error
	t1 := time.Now()
	name := s.postProcessor.config.TemplateName
	if s.postProcessor.config.ReplaceExisting {
		existingTemplate, err = s.postProcessor.driver.GetTemplateByName(s.postProcessor.config.TemplateName, storage.Zone)
		if err == nil && existingTemplate.UUID != "" {
			name = fmt.Sprintf("%s-%s-tmp", name, time.Now().Format(timestampSuffixLayout))
			ui.Say(fmt.Sprintf("Replacing previously created (%s) template '%s' [%s]",
				existingTemplate.Created.Format(timestampLayout), existingTemplate.Title, existingTemplate.Zone))
		} else {
			existingTemplate = nil
		}
	}
	template, err := s.postProcessor.driver.CreateTemplate(storage.UUID, name)
	if err != nil {
		ui.Error(err.Error())
		return nil, err
	}
	if existingTemplate != nil {
		ui.Say(fmt.Sprintf("Deleting existing template '%s' (%s) [%s]", existingTemplate.Title, existingTemplate.UUID, existingTemplate.Zone))
		if err := s.postProcessor.driver.DeleteStorage(existingTemplate.UUID); err != nil {
			ui.Error(err.Error())
			return nil, err
		}
		ui.Say(fmt.Sprintf("Renamimg temporary template '%s' to %s [%s]", template.Title, s.postProcessor.config.TemplateName, template.Zone))
		s.postProcessor.driver.RenameStorage(template.UUID, s.postProcessor.config.TemplateName)
	}

	ui.Say(fmt.Sprintf("Template '%s' created in %s [%s]", name, time.Since(t1), storage.Zone))
	return template, nil
}
