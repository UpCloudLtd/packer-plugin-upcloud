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
	defer os.Remove(file.Name())
	if _, err = file.WriteString(c); err != nil {
		file.Close()
		t.Fatalf("write to temp file %s failed", file.Name())
	}
	file.Close()
	t.Logf("created new temp file %s", file.Name())

	im, err := newImage(file.Name())
	require.NoError(t, err)
	assert.Equal(t, len(c), int(im.Size()))
	assert.Equal(t, 1, int(im.SizeGB()))
}
