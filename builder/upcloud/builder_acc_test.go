package upcloud

import (
	_ "embed"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	internal "github.com/UpCloudLtd/packer-plugin-upcloud/internal"
	"github.com/hashicorp/packer-plugin-sdk/acctest"
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

func TestBuilderAcc_default(t *testing.T) {
	testAccPreCheck(t)

	testCase := &acctest.PluginTestCase{
		Name:     t.Name(),
		Template: testBuildBasic,
		Check:    checkTestResult(),
		Teardown: teardown(t.Name()),
	}
	acctest.TestPlugin(t, testCase)
}

func TestBuilderAcc_storageUuid(t *testing.T) {
	testAccPreCheck(t)
	testCase := &acctest.PluginTestCase{
		Name:     t.Name(),
		Template: testBuilderStorageUuid,
		Check:    checkTestResult(),
		Teardown: teardown(t.Name()),
	}
	acctest.TestPlugin(t, testCase)
}

func TestBuilderAcc_storageName(t *testing.T) {
	testAccPreCheck(t)
	testCase := &acctest.PluginTestCase{
		Name:     t.Name(),
		Template: testBuilderStorageName,
		Check:    checkTestResult(),
		Teardown: teardown(t.Name()),
	}
	acctest.TestPlugin(t, testCase)
}

func TestBuilderAcc_networking(t *testing.T) {
	testAccPreCheck(t)
	testCase := &acctest.PluginTestCase{
		Name:     t.Name(),
		Template: testBuilderNetworking,
		Check:    checkTestResult(),
		Teardown: teardown(t.Name()),
	}
	acctest.TestPlugin(t, testCase)
}

// pkr.hcl

//go:embed test-fixtures/json/basic.json
var testBuildBasicHcl string

//go:embed test-fixtures/json/storage-uuid.json
var testBuilderStorageUuidHcl string

//go:embed test-fixtures/json/storage-name.json
var testBuilderStorageNameHcl string

func TestBuilderAcc_default_hcl(t *testing.T) {
	testAccPreCheck(t)
	testCase := &acctest.PluginTestCase{
		Name:     t.Name(),
		Template: testBuildBasicHcl,
		Check:    checkTestResult(),
		Teardown: teardown(t.Name()),
	}
	acctest.TestPlugin(t, testCase)
}

func TestBuilderAcc_storageUuid_hcl(t *testing.T) {
	testAccPreCheck(t)
	testCase := &acctest.PluginTestCase{
		Name:     t.Name(),
		Template: testBuilderStorageUuidHcl,
		Check:    checkTestResult(),
		Teardown: teardown(t.Name()),
	}
	acctest.TestPlugin(t, testCase)
}

func TestBuilderAcc_storageName_hcl(t *testing.T) {
	testAccPreCheck(t)
	testCase := &acctest.PluginTestCase{
		Name:     t.Name(),
		Template: testBuilderStorageNameHcl,
		Check:    checkTestResult(),
		Teardown: teardown(t.Name()),
	}
	acctest.TestPlugin(t, testCase)
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("UPCLOUD_API_USER"); v == "" {
		t.Skip("UPCLOUD_API_USER must be set for acceptance tests")
	}
	if v := os.Getenv("UPCLOUD_API_PASSWORD"); v == "" {
		t.Skip("UPCLOUD_API_PASSWORD must be set for acceptance tests")
	}
}

func readLog(logfile string) (string, error) {
	logs, err := os.Open(logfile)
	if err != nil {
		return "", fmt.Errorf("Unable find %s", logfile)
	}
	defer logs.Close()

	logsBytes, err := ioutil.ReadAll(logs)
	if err != nil {
		return "", fmt.Errorf("Unable to read %s", logfile)
	}
	return string(logsBytes), nil
}

func checkTestResult() func(*exec.Cmd, string) error {
	return func(buildCommand *exec.Cmd, logfile string) error {

		log, err := readLog(logfile)
		if err != nil {
			return err
		}

		fmt.Print(log)

		if buildCommand.ProcessState != nil {
			if buildCommand.ProcessState.ExitCode() != 0 {
				return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
			}
		}

		_, err = getUuidsFromLog(log)
		if err != nil {
			return err
		}
		return nil
	}
}

var re = regexp.MustCompile(`"Storage template created, UUID: (.*?)"`)

func getUuidsFromLog(log string) ([]string, error) {
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

func teardown(testName string) func() error {
	logfile := fmt.Sprintf("packer_log_%s.txt", testName)

	return func() error {

		log, err := readLog(logfile)
		if err != nil {
			return err
		}

		uuids, err := getUuidsFromLog(log)
		if err != nil {
			return err
		}

		driver := internal.NewDriver(&internal.DriverConfig{
			Username: os.Getenv("UPCLOUD_API_USER"),
			Password: os.Getenv("UPCLOUD_API_PASSWORD"),
			Timeout:  DefaultTimeout,
		})

		for _, u := range uuids {
			fmt.Printf("Cleaning up created templates: %s\n", u)
			if err := driver.DeleteTemplate(u); err != nil {
				return err
			}
		}

		return nil
	}
}
