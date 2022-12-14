package driver

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud/client"
	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud/service"
)

const (
	DefaultPlan                      string = "1xCPU-2GB"
	DefaultHostname                  string = "custom"
	EnvConfigUsername                string = "UPCLOUD_USERNAME"
	EnvConfigPassword                string = "UPCLOUD_PASSWORD"
	EnvConfigUsernameLegacy          string = "UPCLOUD_API_USER"
	EnvConfigPasswordLegacy          string = "UPCLOUD_API_PASSWORD"
	upcloudErrorCodeMetadataDisabled string = "METADATA_DISABLED_ON_CLOUD-INIT"
)

type (
	Driver interface {
		CreateServer(context.Context, *ServerOpts) (*upcloud.ServerDetails, error)
		DeleteServer(context.Context, string) error
		StopServer(context.Context, string) error
		// TODO: rename method or split into two separate method GetStorageByUUID and GetTemplateByName
		GetStorage(ctx context.Context, storageUUID, templateName string) (*upcloud.Storage, error)
		RenameStorage(ctx context.Context, storageUUID, name string) (*upcloud.Storage, error)
		GetServerStorage(context.Context, string) (*upcloud.ServerStorageDevice, error)
		CloneStorage(ctx context.Context, storageUUID, zone, title string) (*upcloud.Storage, error)
		GetTemplateByName(ctx context.Context, name, zone string) (*upcloud.Storage, error)
		CreateTemplate(ctx context.Context, storageUUID, templateTitle string) (*upcloud.Storage, error)
		CreateTemplateStorage(ctx context.Context, title, zone string, size int) (*upcloud.Storage, error)
		ImportStorage(ctx context.Context, storageUUID, contentType string, f io.Reader) (*upcloud.StorageImportDetails, error)
		WaitStorageOnline(ctx context.Context, storageUUID string) (*upcloud.Storage, error)
		DeleteTemplate(context.Context, string) error
		DeleteStorage(ctx context.Context, storageUUID string) error
		GetAvailableZones(context.Context) []string
	}

	driver struct {
		svc    *service.Service
		config *DriverConfig
	}

	DriverConfig struct {
		Username    string
		Password    string
		Timeout     time.Duration
		SSHUsername string
	}

	ServerOpts struct {
		StorageUuid  string
		StorageSize  int
		Zone         string
		SshPublicKey string
		Networking   []request.CreateServerInterface
	}
)

func NewDriver(c *DriverConfig) Driver {
	client := client.New(c.Username, c.Password)
	svc := service.New(client)
	return &driver{
		svc:    svc,
		config: c,
	}
}

func (d *driver) CreateServer(ctx context.Context, opts *ServerOpts) (*upcloud.ServerDetails, error) {
	// Create server
	request := d.prepareCreateRequest(opts)
	response, err := d.svc.CreateServer(ctx, request)
	if err != nil {
		var upcloudErr *upcloud.Error
		if errors.As(err, &upcloudErr) && upcloudErr.ErrorCode == upcloudErrorCodeMetadataDisabled {
			request.Metadata = upcloud.True
			if response, err = d.svc.CreateServer(ctx, request); err != nil {
				return nil, fmt.Errorf("Error creating metadata enabled server: %s", err)
			}
		} else {
			return nil, fmt.Errorf("Error creating server: %s", err)
		}
	}

	// Wait for server to start
	err = d.waitDesiredState(ctx, response.UUID, upcloud.ServerStateStarted)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (d *driver) DeleteServer(ctx context.Context, serverUuid string) error {
	return d.svc.DeleteServerAndStorages(ctx, &request.DeleteServerAndStoragesRequest{
		UUID: serverUuid,
	})
}

func (d *driver) StopServer(ctx context.Context, serverUuid string) error {
	// Ensure the instance is not in maintenance state
	err := d.waitUndesiredState(ctx, serverUuid, upcloud.ServerStateMaintenance)
	if err != nil {
		return err
	}

	// Check current server state and do nothing if already stopped
	response, err := d.getServerDetails(ctx, serverUuid)
	if err != nil {
		return err
	}

	if response.State == upcloud.ServerStateStopped {
		return nil
	}

	// Stop server
	_, err = d.svc.StopServer(ctx, &request.StopServerRequest{
		UUID: serverUuid,
	})
	if err != nil {
		return fmt.Errorf("Failed to stop server: %s", err)
	}

	// Wait for server to stop
	err = d.waitDesiredState(ctx, serverUuid, upcloud.ServerStateStopped)
	if err != nil {
		return err
	}
	return nil
}

func (d *driver) CreateTemplate(ctx context.Context, serverStorageUuid, templateTitle string) (*upcloud.Storage, error) {
	// create image
	response, err := d.svc.TemplatizeStorage(ctx, &request.TemplatizeStorageRequest{
		UUID:  serverStorageUuid,
		Title: templateTitle,
	})
	if err != nil {
		return nil, fmt.Errorf("Error creating image: %s", err)
	}
	return d.WaitStorageOnline(ctx, response.UUID)
}

func (d *driver) WaitStorageOnline(ctx context.Context, storageUuid string) (*upcloud.Storage, error) {
	details, err := d.svc.WaitForStorageState(ctx, &request.WaitForStorageStateRequest{
		UUID:         storageUuid,
		DesiredState: upcloud.StorageStateOnline,
		Timeout:      d.config.Timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("Error while waiting for storage to change state to 'online': %s", err)
	}
	return &details.Storage, nil
}

func (d *driver) GetTemplateByName(ctx context.Context, name, zone string) (*upcloud.Storage, error) {
	response, err := d.svc.GetStorages(ctx, &request.GetStoragesRequest{
		Type: upcloud.StorageTypeTemplate,
	})

	if err != nil {
		return nil, err
	}

	for _, s := range response.Storages {
		if strings.EqualFold(s.Title, name) && (zone != "" && zone == s.Zone) {
			return &s, nil
		}
	}

	return nil, fmt.Errorf("Failed to find storage by name %q", name)
}

// fetch storage by uuid or name
func (d *driver) GetStorage(ctx context.Context, storageUuid, storageName string) (*upcloud.Storage, error) {
	if storageUuid != "" {
		storage, err := d.getStorageByUuid(ctx, storageUuid)
		if err != nil {
			return nil, fmt.Errorf("Error retrieving storage by uuid %q: %s", storageUuid, err)
		}
		return storage, nil
	}

	if storageName != "" {
		storage, err := d.getStorageByName(ctx, storageName)
		if err != nil {
			return nil, fmt.Errorf("Error retrieving storage by name %q: %s", storageName, err)
		}
		return storage, nil

	}
	return nil, fmt.Errorf("Error retrieving storage")
}

func (d *driver) RenameStorage(ctx context.Context, storageUUID, name string) (*upcloud.Storage, error) {
	details, err := d.svc.ModifyStorage(ctx, &request.ModifyStorageRequest{
		UUID:  storageUUID,
		Title: name,
	})

	if err != nil {
		return nil, err
	}

	return d.WaitStorageOnline(ctx, details.Storage.UUID)
}

func (d *driver) CreateTemplateStorage(ctx context.Context, title, zone string, size int) (*upcloud.Storage, error) {
	storage, err := d.svc.CreateStorage(ctx, &request.CreateStorageRequest{
		Size:  size,
		Tier:  upcloud.StorageTierMaxIOPS,
		Title: title,
		Zone:  zone,
	})
	if err != nil {
		return nil, err
	}
	return d.WaitStorageOnline(ctx, storage.UUID)
}

func (d *driver) ImportStorage(ctx context.Context, storageUUID, contentType string, f io.Reader) (*upcloud.StorageImportDetails, error) {
	if _, err := d.svc.CreateStorageImport(ctx, &request.CreateStorageImportRequest{
		StorageUUID:    storageUUID,
		ContentType:    contentType,
		Source:         "direct_upload",
		SourceLocation: f,
	}); err != nil {
		return nil, err
	}
	return d.svc.WaitForStorageImportCompletion(ctx, &request.WaitForStorageImportCompletionRequest{
		StorageUUID: storageUUID,
		Timeout:     d.config.Timeout,
	})
}

func (d *driver) DeleteTemplate(ctx context.Context, templateUuid string) error {
	return d.DeleteStorage(ctx, templateUuid)
}

func (d *driver) DeleteStorage(ctx context.Context, storageUUID string) error {
	return d.svc.DeleteStorage(ctx, &request.DeleteStorageRequest{
		UUID: storageUUID,
	})
}

func (d *driver) CloneStorage(ctx context.Context, storageUuid string, zone string, title string) (*upcloud.Storage, error) {
	response, err := d.svc.CloneStorage(ctx, &request.CloneStorageRequest{
		UUID:  storageUuid,
		Zone:  zone,
		Title: title,
	})
	if err != nil {
		return nil, err
	}
	return d.WaitStorageOnline(ctx, response.UUID)
}

func (d *driver) getStorageByUuid(ctx context.Context, storageUuid string) (*upcloud.Storage, error) {
	response, err := d.svc.GetStorageDetails(ctx, &request.GetStorageDetailsRequest{
		UUID: storageUuid,
	})

	if err != nil {
		return nil, fmt.Errorf("Error fetching storages: %s", err)
	}
	return &response.Storage, nil
}

func (d *driver) getStorageByName(ctx context.Context, storageName string) (*upcloud.Storage, error) {
	response, err := d.svc.GetStorages(ctx, &request.GetStoragesRequest{
		Type: upcloud.StorageTypeTemplate,
	})

	if err != nil {
		return nil, fmt.Errorf("Error fetching storages: %s", err)
	}

	var found bool
	var storage upcloud.Storage
	for _, s := range response.Storages {
		// TODO: should we compare are these strings equal instead ?
		if strings.Contains(strings.ToLower(s.Title), strings.ToLower(storageName)) {
			found = true
			storage = s
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("Failed to find storage by name %q", storageName)
	}
	return &storage, nil
}

func (d *driver) waitDesiredState(ctx context.Context, serverUuid string, state string) error {
	request := &request.WaitForServerStateRequest{
		UUID:         serverUuid,
		DesiredState: state,
		Timeout:      d.config.Timeout,
	}
	if _, err := d.svc.WaitForServerState(ctx, request); err != nil {
		return fmt.Errorf("Error while waiting for server to change state to %q: %s", state, err)
	}
	return nil
}

func (d *driver) waitUndesiredState(ctx context.Context, serverUuid string, state string) error {
	request := &request.WaitForServerStateRequest{
		UUID:           serverUuid,
		UndesiredState: state,
		Timeout:        d.config.Timeout,
	}
	if _, err := d.svc.WaitForServerState(ctx, request); err != nil {
		return fmt.Errorf("Error while waiting for server to change state from %q: %s", state, err)
	}
	return nil
}

func (d *driver) getServerDetails(ctx context.Context, serverUuid string) (*upcloud.ServerDetails, error) {
	response, err := d.svc.GetServerDetails(ctx, &request.GetServerDetailsRequest{
		UUID: serverUuid,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to get details for server: %s", err)
	}
	return response, nil
}

func (d *driver) GetServerStorage(ctx context.Context, serverUuid string) (*upcloud.ServerStorageDevice, error) {
	details, err := d.getServerDetails(ctx, serverUuid)
	if err != nil {
		return nil, err
	}

	var found bool
	var storage upcloud.ServerStorageDevice
	for _, s := range details.StorageDevices {
		if s.Type == upcloud.StorageTypeDisk {
			found = true
			storage = s
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("Failed to find storage type disk for server %q", serverUuid)
	}
	return &storage, nil
}

func (d *driver) prepareCreateRequest(opts *ServerOpts) *request.CreateServerRequest {
	title := fmt.Sprintf("packer-%s-%s", DefaultHostname, getNowString())
	titleDisk := fmt.Sprintf("%s-disk1", DefaultHostname)

	request := request.CreateServerRequest{
		Title:            title,
		Hostname:         DefaultHostname,
		Zone:             opts.Zone,
		PasswordDelivery: request.PasswordDeliveryNone,
		Plan:             DefaultPlan,
		StorageDevices: []request.CreateServerStorageDevice{
			{
				Action:  request.CreateServerStorageDeviceActionClone,
				Storage: opts.StorageUuid,
				Title:   titleDisk,
				Size:    opts.StorageSize,
				Tier:    upcloud.StorageTierMaxIOPS,
			},
		},
		Networking: &request.CreateServerNetworking{
			Interfaces: opts.Networking,
		},
		LoginUser: &request.LoginUser{
			CreatePassword: "no",
			Username:       d.config.SSHUsername,
			SSHKeys:        []string{opts.SshPublicKey},
		},
	}
	return &request
}

func (d *driver) GetAvailableZones(ctx context.Context) []string {
	zones := make([]string, 0)
	if z, err := d.svc.GetZones(ctx); err == nil {
		for _, zone := range z.Zones {
			zones = append(zones, zone.ID)
		}
	}
	return zones
}

func getNowString() string {
	return time.Now().Format("20060102-150405")
}

func UsernameFromEnv() string {
	username := os.Getenv(EnvConfigUsernameLegacy)
	if username == "" {
		username = os.Getenv(EnvConfigUsername)
	}
	return username
}

func PasswordFromEnv() string {
	passwd := os.Getenv(EnvConfigPasswordLegacy)
	if passwd == "" {
		passwd = os.Getenv(EnvConfigPassword)
	}
	return passwd
}
