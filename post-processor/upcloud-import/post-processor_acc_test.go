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
	username := os.Getenv("UPCLOUD_API_USER")
	require.NotEmpty(t, username, "UPCLOUD_API_USER must be set for acceptance tests")
	password := os.Getenv("UPCLOUD_API_PASSWORD")
	require.NotEmpty(t, password, "UPCLOUD_API_PASSWORD must be set for acceptance tests")

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
	t1, err := driver.GetTemplateByName(testName, "pl-waw1")
	require.NoError(t, err)
	assert.NoError(t, driver.DeleteStorage(t1.UUID))

	t1, err = driver.GetTemplateByName(testName, "fi-hel2")
	require.NoError(t, err)
	assert.NoError(t, driver.DeleteStorage(t1.UUID))
}
