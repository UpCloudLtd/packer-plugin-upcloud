package upcloudimport

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/UpCloudLtd/packer-plugin-upcloud/internal/driver"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud"
)

const artifactDestroyTimeout time.Duration = time.Minute * 30

type Artifact struct {
	postProcessor *PostProcessor
	templates     []*upcloud.Storage
	stateData     map[string]interface{}
	driver        driver.Driver
}

func (a *Artifact) BuilderId() string {
	return BuilderID
}

func (a *Artifact) Files() []string {
	return []string{}
}

func (a *Artifact) Id() string {
	return a.templates[0].UUID
}

func (a *Artifact) String() string {
	return fmt.Sprintf("%s [%s]", a.templates[0].Title, strings.Join(a.postProcessor.config.Zones, ", "))
}

func (a *Artifact) State(name string) interface{} {
	return a.stateData[name]
}

func (a *Artifact) Destroy() error {
	ctx, cancel := context.WithTimeout(context.Background(), artifactDestroyTimeout)
	defer cancel()
	for _, t := range a.templates {
		err := a.driver.DeleteTemplate(ctx, t.UUID)
		if err != nil {
			return err
		}
	}
	return nil
}
