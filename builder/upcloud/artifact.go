package upcloud

import (
	"fmt"
	"log"
	"strings"

	"github.com/UpCloudLtd/packer-plugin-upcloud/internal/driver"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/hashicorp/packer-plugin-sdk/packer/registry/image"
)

// packersdk.Artifact implementation
type Artifact struct {
	config    *Config
	driver    driver.Driver
	Templates []*upcloud.Storage

	// StateData should store data such as GeneratedData
	// to be shared with post-processors
	StateData map[string]interface{}
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Files() []string {
	return []string{}
}

func (a *Artifact) Id() string {
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
	for _, t := range a.Templates {
		err := a.driver.DeleteTemplate(t.UUID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Artifact) buildHCPPackerRegistryMetadata() ([]*image.Image, error) {
	var sourceTemplateUUID, sourceTemplateTitle string
	if v, ok := a.StateData["source_template_uuid"]; ok {
		sourceTemplateUUID = v.(string)
	}

	if v, ok := a.StateData["source_template_title"]; ok {
		sourceTemplateTitle = v.(string)
	}

	images := make([]*image.Image, 0)
	for _, template := range a.Templates {
		img, err := image.FromArtifact(a,
			image.WithID(template.UUID),
			image.WithRegion(template.Zone),
			image.WithProvider("upcloud"),
		)
		if err != nil {
			return images, err
		}

		img.SourceImageID = sourceTemplateUUID
		img.Labels["source_id"] = sourceTemplateUUID
		img.Labels["source"] = sourceTemplateTitle
		img.Labels["name"] = template.Title
		img.Labels["name_prefix"] = a.config.TemplatePrefix
		img.Labels["size"] = fmt.Sprint(template.Size)
		images = append(images, img)
	}
	return images, nil
}
