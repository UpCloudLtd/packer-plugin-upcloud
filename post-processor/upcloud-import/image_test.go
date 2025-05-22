package upcloudimport

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImage(t *testing.T) {
	c := time.Now().Format(time.ANSIC)
	file, err := os.CreateTemp(os.TempDir(), "packer-test-import-image-*.raw")
	require.NoError(t, err)
	defer func() {
		if err := os.Remove(file.Name()); err != nil {
			t.Logf("Failed to remove temporary file: %v", err)
		}
	}()
	if _, err = file.WriteString(c); err != nil {
		if closeErr := file.Close(); closeErr != nil {
			t.Logf("Failed to close file: %v", closeErr)
		}
		t.Fatalf("write to temp file %s failed", file.Name())
	}
	if err := file.Close(); err != nil {
		t.Fatalf("Failed to close file: %v", err)
	}
	t.Logf("created new temp file %s", file.Name())

	im, err := newImage(file.Name())
	require.NoError(t, err)
	assert.Equal(t, len(c), int(im.Size()))
	assert.Equal(t, 1, int(im.SizeGB()))
}
