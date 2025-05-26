package upcloudimport

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
)

const (
	timestampLayout       string = "2006-01-02 15:04:05"
	timestampSuffixLayout string = "20060102-150405"
)

type stepCreateTemplate struct {
	postProcessor *PostProcessor
}

func (s *stepCreateTemplate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	uiRaw := state.Get(stateUI)
	ui, ok := uiRaw.(packer.Ui)
	if !ok {
		return haltOnError(nil, state, errors.New("UI is not of expected type"))
	}

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
			template, err := s.createTemplateBasedOnStorage(ctx, ui, sto)
			if err != nil {
				halt = true
				return
			}
			templates = append(templates, template)
		}(storage)
	}

	wg.Wait()

	if err := cleanupDevices(ctx, ui, s.postProcessor.driver, state); err != nil {
		return haltOnError(ui, state, err)
	}

	state.Put(stateTemplates, templates)

	if halt {
		if err := cleanupTemplates(ctx, ui, s.postProcessor.driver, state); err != nil {
			ui.Error(err.Error())
		}
		return multistep.ActionHalt
	}
	return multistep.ActionContinue
}

func (s *stepCreateTemplate) Cleanup(state multistep.StateBag) {
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

func (s *stepCreateTemplate) createTemplateBasedOnStorage(ctx context.Context, ui packer.Ui, storage *upcloud.Storage) (*upcloud.Storage, error) {
	var existingTemplate *upcloud.Storage
	var err error
	t1 := time.Now()
	name := s.postProcessor.config.TemplateName
	if s.postProcessor.config.ReplaceExisting {
		existingTemplate, err = s.postProcessor.driver.GetTemplateByName(ctx, s.postProcessor.config.TemplateName, storage.Zone)
		if err == nil && existingTemplate.UUID != "" {
			name = fmt.Sprintf("%s-%s-tmp", name, time.Now().Format(timestampSuffixLayout))
			ui.Say(fmt.Sprintf("Replacing previously created (%s) template '%s' [%s]",
				existingTemplate.Created.Format(timestampLayout), existingTemplate.Title, existingTemplate.Zone))
		} else {
			existingTemplate = nil
		}
	}
	template, err := s.postProcessor.driver.CreateTemplate(ctx, storage.UUID, name)
	if err != nil {
		ui.Error(err.Error())
		return nil, err
	}
	if existingTemplate != nil {
		ui.Say(fmt.Sprintf("Deleting existing template '%s' (%s) [%s]", existingTemplate.Title, existingTemplate.UUID, existingTemplate.Zone))
		if err := s.postProcessor.driver.DeleteStorage(ctx, existingTemplate.UUID); err != nil {
			ui.Error(err.Error())
			return nil, err
		}
		ui.Say(fmt.Sprintf("Renamimg temporary template '%s' to %s [%s]", template.Title, s.postProcessor.config.TemplateName, template.Zone))
		template, err = s.postProcessor.driver.RenameStorage(ctx, template.UUID, s.postProcessor.config.TemplateName)
		if err != nil {
			ui.Error(err.Error())
			return nil, err
		}
	}

	ui.Say(fmt.Sprintf("Template '%s' created in %s [%s]", name, time.Since(t1), storage.Zone))
	return template, nil
}
