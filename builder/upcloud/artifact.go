package upcloud

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/packer/registry/image"

	"github.com/UpCloudLtd/packer-plugin-upcloud/internal/driver"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
)

// packersdk.Artifact implementation.
type Artifact struct {
	config    *Config
	driver    driver.Driver
	Templates []*upcloud.Storage

	// StateData should store data such as GeneratedData
	// to be shared with post-processors
	StateData map[string]interface{}
}

func (*Artifact) BuilderId() string { //nolint:revive // method is required by packer-plugin-sdk
	return BuilderID
}

func (a *Artifact) Files() []string {
	return []string{}
}

func (a *Artifact) Id() string { //nolint:revive // method is required by packer-plugin-sdk
	result := []string{}
	for _, t := range a.Templates {
		result = append(result, t.UUID)
	}
	return strings.Join(result, ",")
}

func (a *Artifact) String() string {
	return fmt.Sprintf("Storage template created, UUID: %s", a.Id())
}

func (a *Artifact) State(name string) interface{} {
	if name == image.ArtifactStateURI {
		images, err := a.buildHCPPackerRegistryMetadata()
		if err != nil {
			log.Printf("[DEBUG] error encountered when creating a registry image %v", err)
			return nil
		}
		return images
	}
	return a.StateData[name]
}

func (a *Artifact) Destroy() error {
	ctx, cancel := contextWithDefaultTimeout()
	defer cancel()
	for _, t := range a.Templates {
		err := a.driver.DeleteTemplate(ctx, t.UUID)
		if err != nil {
			return fmt.Errorf("failed to delete template %s: %w", t.UUID, err)
		}
	}
	return nil
}

func (a *Artifact) buildHCPPackerRegistryMetadata() ([]*image.Image, error) {
	var sourceTemplateUUID, sourceTemplateTitle string
	if v, ok := a.StateData["source_template_uuid"]; ok {
		if uuid, ok := v.(string); ok {
			sourceTemplateUUID = uuid
		}
	}

	if v, ok := a.StateData["source_template_title"]; ok {
		if title, ok := v.(string); ok {
			sourceTemplateTitle = title
		}
	}

	images := make([]*image.Image, 0)
	for _, template := range a.Templates {
		img, err := image.FromArtifact(a,
			image.WithID(template.UUID),
			image.WithRegion(template.Zone),
			image.WithProvider("upcloud"),
		)
		if err != nil {
			return images, fmt.Errorf("failed to create registry image for template %s: %w", template.UUID, err)
		}

		img.SourceImageID = sourceTemplateUUID
		img.Labels["source_id"] = sourceTemplateUUID
		img.Labels["source"] = sourceTemplateTitle
		img.Labels["name"] = template.Title
		img.Labels["name_prefix"] = a.config.TemplatePrefix
		img.Labels["size"] = strconv.Itoa(template.Size)
		images = append(images, img)
	}
	return images, nil
}
