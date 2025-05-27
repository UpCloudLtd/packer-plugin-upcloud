package upcloud //nolint:testpackage // not all fields can be exported in Artifact

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"

	"github.com/UpCloudLtd/packer-plugin-upcloud/internal/driver"

	_ "embed"
)

// Run tests: PACKER_ACC=1 go test -count 1 -v ./...  -timeout=120m

// json

//go:embed test-fixtures/json/basic.json
var testBuildBasic string

//go:embed test-fixtures/json/storage-uuid.json
var testBuilderStorageUuid string

//go:embed test-fixtures/json/storage-name.json
var testBuilderStorageName string

//go:embed test-fixtures/json/networking.json
var testBuilderNetworking string

//go:embed test-fixtures/json/basic_standard_tier.json
var testBuilderBasicStandardTier string

func TestBuilderAcc_default(t *testing.T) {
	t.Parallel()
	testAccPreCheck(t)

	testCase := &acctest.PluginTestCase{
		Name:     t.Name(),
		Template: testBuildBasic,
		Check:    checkTestResult(t),
		Teardown: teardown(t, t.Name()),
	}
	acctest.TestPlugin(t, testCase)
}

func TestBuilderAcc_storageUuid(t *testing.T) {
	t.Parallel()
	testAccPreCheck(t)
	testCase := &acctest.PluginTestCase{
		Name:     t.Name(),
		Template: testBuilderStorageUuid,
		Check:    checkTestResult(t),
		Teardown: teardown(t, t.Name()),
	}
	acctest.TestPlugin(t, testCase)
}

func TestBuilderAcc_storageName(t *testing.T) {
	t.Parallel()
	testAccPreCheck(t)
	testCase := &acctest.PluginTestCase{
		Name:     t.Name(),
		Template: testBuilderStorageName,
		Check:    checkTestResult(t),
		Teardown: teardown(t, t.Name()),
	}
	acctest.TestPlugin(t, testCase)
}

func TestBuilderAcc_standardTier(t *testing.T) {
	t.Parallel()
	testAccPreCheck(t)
	testCase := &acctest.PluginTestCase{
		Name:     t.Name(),
		Template: testBuilderBasicStandardTier,
		Check:    checkTestResult(t),
		Teardown: teardown(t, t.Name()),
	}
	acctest.TestPlugin(t, testCase)
}

func TestBuilderAcc_networking(t *testing.T) {
	t.Parallel()
	testAccPreCheck(t)
	testCase := &acctest.PluginTestCase{
		Name:     t.Name(),
		Template: testBuilderNetworking,
		Check:    checkTestResult(t),
		Teardown: teardown(t, t.Name()),
	}
	acctest.TestPlugin(t, testCase)
}

// pkr.hcl

//go:embed test-fixtures/hcl2/basic.pkr.hcl
var testBuildBasicHcl string

//go:embed test-fixtures/hcl2/storage-uuid.pkr.hcl
var testBuilderStorageUuidHcl string

//go:embed test-fixtures/hcl2/storage-name.pkr.hcl
var testBuilderStorageNameHcl string

//go:embed test-fixtures/hcl2/network_interfaces.pkr.hcl
var testBuilderNetworkInterfacesHcl string

func TestBuilderAcc_default_hcl(t *testing.T) {
	t.Parallel()
	testAccPreCheck(t)
	testCase := &acctest.PluginTestCase{
		Name:     t.Name(),
		Template: testBuildBasicHcl,
		Check:    checkTestResult(t),
		Teardown: teardown(t, t.Name()),
	}
	acctest.TestPlugin(t, testCase)
}

func TestBuilderAcc_storageUuid_hcl(t *testing.T) {
	t.Parallel()
	testAccPreCheck(t)
	testCase := &acctest.PluginTestCase{
		Name:     t.Name(),
		Template: testBuilderStorageUuidHcl,
		Check:    checkTestResult(t),
		Teardown: teardown(t, t.Name()),
	}
	acctest.TestPlugin(t, testCase)
}

func TestBuilderAcc_storageName_hcl(t *testing.T) {
	t.Parallel()
	testAccPreCheck(t)
	testCase := &acctest.PluginTestCase{
		Name:     t.Name(),
		Template: testBuilderStorageNameHcl,
		Check:    checkTestResult(t),
		Teardown: teardown(t, t.Name()),
	}
	acctest.TestPlugin(t, testCase)
}

func TestBuilderAcc_network_interfaces(t *testing.T) {
	t.Parallel()
	testAccPreCheck(t)
	testCase := &acctest.PluginTestCase{
		Name:     t.Name(),
		Template: testBuilderNetworkInterfacesHcl,
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			re := regexp.MustCompile(`upcloud.network_interfaces: Selecting default ip '10.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}' as Server IP`)
			log, err := readLog(t, logfile)
			if err != nil {
				return err
			}
			// Log content is checked via regex, no need to print it
			if !re.MatchString(log) {
				return fmt.Errorf("Unable find default utility network IP from the log %s", logfile)
			}
			return nil
		},
		Teardown: teardown(t, t.Name()),
	}
	acctest.TestPlugin(t, testCase)
}

func testAccPreCheck(t *testing.T) {
	t.Helper()
	if v := driver.UsernameFromEnv(); v == "" {
		t.Skipf("%s or %s must be set for acceptance tests", driver.EnvConfigUsernameLegacy, driver.EnvConfigUsername)
	}
	if v := driver.PasswordFromEnv(); v == "" {
		t.Skipf("%s or %s must be set for acceptance tests", driver.EnvConfigPasswordLegacy, driver.EnvConfigPassword)
	}
}

func readLog(t *testing.T, logfile string) (string, error) {
	t.Helper()
	logs, err := os.Open(logfile) // #nosec G304 -- logfile path is controlled by Packer SDK acctest framework
	if err != nil {
		return "", fmt.Errorf("Unable find %s", logfile)
	}
	defer func() {
		_ = logs.Close()
	}()

	logsBytes, err := io.ReadAll(logs)
	if err != nil {
		return "", fmt.Errorf("Unable to read %s", logfile)
	}
	return string(logsBytes), nil
}

func checkTestResult(t *testing.T) func(*exec.Cmd, string) error {
	t.Helper()
	return func(buildCommand *exec.Cmd, logfile string) error {
		log, err := readLog(t, logfile)
		if err != nil {
			return err
		}

		if buildCommand.ProcessState != nil {
			if buildCommand.ProcessState.ExitCode() != 0 {
				return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
			}
		}

		_, err = getUuidsFromLog(t, log)
		if err != nil {
			return err
		}
		return nil
	}
}

var re = regexp.MustCompile(`"Storage template created, UUID: (.*?)"`)

func getUuidsFromLog(t *testing.T, log string) ([]string, error) {
	t.Helper()
	var match string
	ms := re.FindAllStringSubmatch(log, -1)
	for _, m := range ms {
		match = m[1]
	}
	if match == "" {
		return nil, errors.New("Created template UUIDs not found in the log")
	}

	uuid := []string{}
	for _, item := range strings.Split(match, ",") {
		item = strings.TrimSpace(item)
		uuid = append(uuid, item)
	}
	return uuid, nil
}

func teardown(t *testing.T, testName string) func() error {
	t.Helper()
	logfile := fmt.Sprintf("packer_log_%s.txt", testName)
	return func() error {
		ctx, cancel := contextWithDefaultTimeout()
		defer cancel()
		log, err := readLog(t, logfile)
		if err != nil {
			return err
		}

		uuids, err := getUuidsFromLog(t, log)
		if err != nil {
			return err
		}

		drv := driver.NewDriver(&driver.DriverConfig{
			Username: driver.UsernameFromEnv(),
			Password: driver.PasswordFromEnv(),
			Timeout:  DefaultTimeout,
		})

		for _, u := range uuids {
			t.Logf("Cleaning up created templates: %s", u)
			if err := drv.DeleteTemplate(ctx, u); err != nil {
				return err
			}
		}

		return nil
	}
}
