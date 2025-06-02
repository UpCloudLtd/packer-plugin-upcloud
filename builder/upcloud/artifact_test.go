package upcloud //nolint:testpackage // not all fields can be exported in Artifact

import (
	"fmt"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packer/registry/image"
	"github.com/stretchr/testify/assert"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
)

func TestArtifact_impl(t *testing.T) {
	t.Parallel()
	var _ packersdk.Artifact = new(Artifact)
}

func TestArtifact_Id(t *testing.T) {
	t.Parallel()
	uuid1 := "some-uuid-1"
	uuid2 := "some-uuid-2"
	expected := fmt.Sprintf("%s,%s", uuid1, uuid2)

	templates := []*upcloud.Storage{}
	templates = append(templates, &upcloud.Storage{UUID: uuid1}, &upcloud.Storage{UUID: uuid2})

	a := &Artifact{Templates: templates}
	result := a.Id()

	if result != expected {
		t.Errorf("Expected: %q, got: %q", expected, result)
	}
}

func TestArtifact_String(t *testing.T) {
	t.Parallel()
	expected := `Storage template created, UUID: some-uuid`

	templates := []*upcloud.Storage{}
	templates = append(templates, &upcloud.Storage{UUID: "some-uuid"})

	a := &Artifact{Templates: templates}
	result := a.String()

	if result != expected {
		t.Errorf("Expected: %q, got: %q", expected, result)
	}
}

func TestArtifact_Metadata(t *testing.T) {
	t.Parallel()
	templates := []*upcloud.Storage{}
	templates = append(templates,
		&upcloud.Storage{
			UUID:  "some-uuid",
			Size:  10,
			Title: "some-title",
			Zone:  "fi-hel1",
		},
		&upcloud.Storage{
			UUID:  "some-other-uuid",
			Size:  10,
			Title: "some-title",
			Zone:  "fi-hel2",
		},
	)

	a := &Artifact{
		Templates: templates,
		config: &Config{
			Zone:           "fi-hel1",
			CloneZones:     []string{"fi-hel2"},
			TemplatePrefix: "prefix",
		},
		StateData: map[string]interface{}{
			"source_template_title": "source-title",
			"source_template_uuid":  "source-uuid",
		},
	}
	result := a.State(image.ArtifactStateURI)
	got, ok := result.([]*image.Image)
	if !ok {
		t.Fatalf("Expected []*image.Image, got %T", result)
	}
	want := &image.Image{
		ImageID:        "some-uuid",
		ProviderName:   "upcloud",
		ProviderRegion: "fi-hel1",
		Labels: map[string]string{
			"source":      "source-title",
			"source_id":   "source-uuid",
			"name":        "some-title",
			"name_prefix": "prefix",
			"size":        "10",
		},
		SourceImageID: "source-uuid",
	}
	assert.Equal(t, want, got[0])
}
