package upcloudimport

import (
	"context"
	_ "embed"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/UpCloudLtd/packer-plugin-upcloud/internal/driver"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	if os.Getenv("PACKER_ACC") != "1" {
		t.Skip("skip acceptance test")
	}
	ctx := context.Background()
	username := driver.UsernameFromEnv()
	if username == "" {
		t.Skipf("%s or %s must be set for acceptance tests", driver.EnvConfigUsernameLegacy, driver.EnvConfigUsername)
	}

	password := driver.PasswordFromEnv()
	if password == "" {
		t.Skipf("%s or %s must be set for acceptance tests", driver.EnvConfigPasswordLegacy, driver.EnvConfigPassword)
	}

	testName := fmt.Sprintf("%s-acc-test-%s", BuilderID, time.Now().Format(timestampSuffixLayout))
	imageFile, err := ioutil.TempFile(os.TempDir(), fmt.Sprintf("%s-*.raw", testName))
	require.NoError(t, err)
	defer os.Remove(imageFile.Name())

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
		context.Background(),
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
