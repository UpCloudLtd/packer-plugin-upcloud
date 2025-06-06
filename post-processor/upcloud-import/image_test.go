//go:build !integration

package upcloudimport_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	upcloudimport "github.com/UpCloudLtd/packer-plugin-upcloud/post-processor/upcloud-import"
)

func TestImage(t *testing.T) {
	t.Parallel()
	c := time.Now().Format(time.ANSIC)
	file, err := os.CreateTemp(os.TempDir(), "packer-test-import-image-*.raw")
	require.NoError(t, err)
	defer func() {
		if err := os.Remove(file.Name()); err != nil {
			t.Logf("warning: failed to remove temp file: %v", err)
		}
	}()
	if _, err = file.WriteString(c); err != nil {
		if closeErr := file.Close(); closeErr != nil {
			t.Logf("warning: failed to close file: %v", closeErr)
		}
		t.Fatalf("write to temp file %s failed", file.Name())
	}
	if err := file.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}
	t.Logf("created new temp file %s", file.Name())

	im, err := upcloudimport.NewImage(file.Name())
	require.NoError(t, err)
	assert.Equal(t, len(c), int(im.Size()))
	assert.Equal(t, 1, im.SizeGB())
}
