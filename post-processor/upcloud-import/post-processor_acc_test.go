package upcloudimport //nolint:testpackage // not all fields can be exported

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/UpCloudLtd/packer-plugin-upcloud/internal/driver"

	_ "embed"
)

type testArtifact struct {
	files []string
}

func (a *testArtifact) BuilderId() string             { return artificeBuilderID }
func (a *testArtifact) Files() []string               { return a.files }
func (a *testArtifact) Id() string                    { return "" }
func (a *testArtifact) String() string                { return "" }
func (a *testArtifact) State(name string) interface{} { return nil }
func (a *testArtifact) Destroy() error                { return nil }

func TestPostProcessorAcc_raw(t *testing.T) {
	t.Parallel()
	if os.Getenv("PACKER_ACC") != "1" {
		t.Skip("skip acceptance test")
	}
	ctx := t.Context()
	username := driver.UsernameFromEnv()
	if username == "" {
		t.Skipf("%s or %s must be set for acceptance tests", driver.EnvConfigUsernameLegacy, driver.EnvConfigUsername)
	}

	password := driver.PasswordFromEnv()
	if password == "" {
		t.Skipf("%s or %s must be set for acceptance tests", driver.EnvConfigPasswordLegacy, driver.EnvConfigPassword)
	}

	testName := fmt.Sprintf("%s-acc-test-%s", BuilderID, time.Now().Format(timestampSuffixLayout))
	imageFile, err := os.CreateTemp(os.TempDir(), fmt.Sprintf("%s-*.raw", testName))
	require.NoError(t, err)
	defer func() {
		if err := os.Remove(imageFile.Name()); err != nil {
			t.Logf("Warning: failed to remove temp file: %v", err)
		}
	}()

	var p PostProcessor
	err = p.Configure([]interface{}{map[string]interface{}{
		"username":         username,
		"password":         password,
		"zones":            []string{"pl-waw1", "fi-hel2"},
		"template_name":    testName,
		"replace_existing": true,
	}}...)
	require.NoError(t, err)

	a, _, _, err := p.PostProcess(
		t.Context(),
		packer.TestUi(t),
		&testArtifact{files: []string{imageFile.Name()}})

	require.NoError(t, err)
	require.NotNil(t, a)

	driver := driver.NewDriver(&driver.DriverConfig{
		Username: username,
		Password: password,
		Timeout:  time.Minute * 30,
	})
	t1, err := driver.GetTemplateByName(ctx, testName, "pl-waw1")
	require.NoError(t, err)
	assert.NoError(t, driver.DeleteStorage(ctx, t1.UUID))

	t1, err = driver.GetTemplateByName(ctx, testName, "fi-hel2")
	require.NoError(t, err)
	assert.NoError(t, driver.DeleteStorage(ctx, t1.UUID))
}
