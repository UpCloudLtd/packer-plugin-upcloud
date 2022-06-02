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
		img, err := a.buildHCPPackerRegistryMetadata()
		if err != nil {
			log.Printf("[DEBUG] error encountered when creating a registry image %v", err)
			return nil
		}
		return img
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

func (a *Artifact) buildHCPPackerRegistryMetadata() (*image.Image, error) {
	img, err := image.FromArtifact(a,
		image.WithID(a.Templates[0].UUID),
		image.WithRegion(a.config.Zone),
		image.WithProvider("upcloud"),
	)
	if err != nil {
		return img, err
	}

	if v, ok := a.StateData["source_template_uuid"].(string); ok {
		img.SourceImageID = v
		img.Labels["source_id"] = v
	}
	// Comma separated list of zones which can be used to per zone template UUIDs
	zones := []string{a.config.Zone}
	zones = append(zones, a.config.CloneZones...)
	img.Labels["zones"] = strings.Join(zones, ",")
	for _, t := range a.Templates {
		img.Labels[t.Zone] = t.UUID
	}
	if v, ok := a.StateData["source_template_title"].(string); ok {
		img.Labels["source"] = v
	}
	img.Labels["name"] = a.Templates[0].Title
	img.Labels["name_prefix"] = a.config.TemplatePrefix
	img.Labels["size"] = fmt.Sprint(a.Templates[0].Size)
	return img, nil
}
