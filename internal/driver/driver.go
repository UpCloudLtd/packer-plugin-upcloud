package driver

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/credentials"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/client"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
)

const (
	DefaultPlan                      string = "1xCPU-2GB"
	DefaultHostname                  string = "custom"
	EnvConfigUsername                string = "UPCLOUD_USERNAME"
	EnvConfigPassword                string = "UPCLOUD_PASSWORD"
	EnvConfigAPIToken                string = "UPCLOUD_TOKEN"
	EnvConfigUsernameLegacy          string = "UPCLOUD_API_USER"
	EnvConfigPasswordLegacy          string = "UPCLOUD_API_PASSWORD"
	upcloudErrorCodeMetadataDisabled string = "METADATA_DISABLED_ON_CLOUD-INIT"
)

type (
	// ServerManager handles server lifecycle operations.
	ServerManager interface {
		CreateServer(ctx context.Context, opts *ServerOpts) (*upcloud.ServerDetails, error)
		DeleteServer(ctx context.Context, serverUUID string) error
		StopServer(ctx context.Context, serverUUID string) error
		// TODO: rename method or split into two separate method GetStorageByUUID and GetTemplateByName
		GetServerStorage(ctx context.Context, serverUUID string) (*upcloud.ServerStorageDevice, error)
	}

	// StorageManager handles storage operations.
	StorageManager interface {
		GetStorage(ctx context.Context, storageUUID, templateName string) (*upcloud.Storage, error)
		RenameStorage(ctx context.Context, storageUUID, name string) (*upcloud.Storage, error)
		CloneStorage(ctx context.Context, storageUUID, zone, title string) (*upcloud.Storage, error)
		CreateTemplateStorage(ctx context.Context, title, zone string, size int, tier string) (*upcloud.Storage, error)
		ImportStorage(ctx context.Context, storageUUID, contentType string, f io.Reader) (*upcloud.StorageImportDetails, error)
		WaitStorageOnline(ctx context.Context, storageUUID string) (*upcloud.Storage, error)
		DeleteStorage(ctx context.Context, storageUUID string) error
	}

	// TemplateManager handles template operations.
	TemplateManager interface {
		GetTemplateByName(ctx context.Context, name, zone string) (*upcloud.Storage, error)
		CreateTemplate(ctx context.Context, storageUUID, templateTitle string) (*upcloud.Storage, error)
		DeleteTemplate(ctx context.Context, templateUUID string) error
	}

	// ZoneManager handles zone operations.
	ZoneManager interface {
		GetAvailableZones(ctx context.Context) []string
	}

	// Driver combines all management interfaces.
	Driver interface {
		ServerManager
		StorageManager
		TemplateManager
		ZoneManager
	}

	driver struct {
		svc    *service.Service
		config *DriverConfig
	}

	DriverConfig struct {
		Username    string
		Password    string
		Token       string
		Timeout     time.Duration
		SSHUsername string
	}

	ServerOpts struct {
		StorageUUID  string
		StorageSize  int
		Zone         string
		SSHPublicKey string
		Networking   []request.CreateServerInterface
		StorageTier  string
	}
)

func NewDriver(c *DriverConfig) Driver {
	var cl *client.Client

	// Use API token if provided, otherwise fall back to username/password
	if c.Token != "" {
		// TODO: Update this with a proper token auth wrapper when upcloud-go-api supports it
		cl = client.New("", "", client.WithBearerAuth(c.Token))
	} else {
		cl = client.New(c.Username, c.Password)
	}

	svc := service.New(cl)
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
		var upcloudErr *upcloud.Problem
		if errors.As(err, &upcloudErr) && upcloudErr.ErrorCode() == upcloudErrorCodeMetadataDisabled {
			request.Metadata = upcloud.True
			if response, err = d.svc.CreateServer(ctx, request); err != nil {
				return nil, fmt.Errorf("error creating metadata enabled server: %w", err)
			}
		} else {
			return nil, fmt.Errorf("error creating server: %w", err)
		}
	}

	// Wait for server to start
	err = d.waitDesiredState(ctx, response.UUID, upcloud.ServerStateStarted)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (d *driver) DeleteServer(ctx context.Context, serverUUID string) error {
	err := d.svc.DeleteServerAndStorages(ctx, &request.DeleteServerAndStoragesRequest{
		UUID: serverUUID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete server %s and its storages: %w", serverUUID, err)
	}
	return nil
}

func (d *driver) StopServer(ctx context.Context, serverUUID string) error {
	// Ensure the instance is not in maintenance state
	err := d.waitUndesiredState(ctx, serverUUID, upcloud.ServerStateMaintenance)
	if err != nil {
		return err
	}

	// Check current server state and do nothing if already stopped
	response, err := d.getServerDetails(ctx, serverUUID)
	if err != nil {
		return err
	}

	if response.State == upcloud.ServerStateStopped {
		return nil
	}

	// Stop server
	_, err = d.svc.StopServer(ctx, &request.StopServerRequest{
		UUID: serverUUID,
	})
	if err != nil {
		return fmt.Errorf("failed to stop server: %w", err)
	}

	// Wait for server to stop
	err = d.waitDesiredState(ctx, serverUUID, upcloud.ServerStateStopped)
	if err != nil {
		return err
	}
	return nil
}

func (d *driver) CreateTemplate(ctx context.Context, serverStorageUUID, templateTitle string) (*upcloud.Storage, error) {
	// create image
	response, err := d.svc.TemplatizeStorage(ctx, &request.TemplatizeStorageRequest{
		UUID:  serverStorageUUID,
		Title: templateTitle,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating image: %w", err)
	}
	return d.WaitStorageOnline(ctx, response.UUID)
}

func (d *driver) WaitStorageOnline(ctx context.Context, storageUUID string) (*upcloud.Storage, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, d.config.Timeout)
	defer cancel()

	details, err := d.svc.WaitForStorageState(timeoutCtx, &request.WaitForStorageStateRequest{
		UUID:         storageUUID,
		DesiredState: upcloud.StorageStateOnline,
	})
	if err != nil {
		return nil, fmt.Errorf("error while waiting for storage to change state to 'online': %w", err)
	}
	return &details.Storage, nil
}

func (d *driver) GetTemplateByName(ctx context.Context, name, zone string) (*upcloud.Storage, error) {
	response, err := d.svc.GetStorages(ctx, &request.GetStoragesRequest{
		Type: upcloud.StorageTypeTemplate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get template storages: %w", err)
	}

	for _, s := range response.Storages {
		if strings.EqualFold(s.Title, name) && (zone != "" && zone == s.Zone) {
			return &s, nil
		}
	}

	return nil, fmt.Errorf("failed to find storage by name %q", name)
}

// fetch storage by uuid or name.
func (d *driver) GetStorage(ctx context.Context, storageUUID, storageName string) (*upcloud.Storage, error) {
	if storageUUID != "" {
		storage, err := d.getStorageByUUID(ctx, storageUUID)
		if err != nil {
			return nil, fmt.Errorf("error retrieving storage by uuid %q: %w", storageUUID, err)
		}
		return storage, nil
	}

	if storageName != "" {
		storage, err := d.getStorageByName(ctx, storageName)
		if err != nil {
			return nil, fmt.Errorf("error retrieving storage by name %q: %w", storageName, err)
		}
		return storage, nil
	}
	return nil, errors.New("error retrieving storage")
}

func (d *driver) RenameStorage(ctx context.Context, storageUUID, name string) (*upcloud.Storage, error) {
	details, err := d.svc.ModifyStorage(ctx, &request.ModifyStorageRequest{
		UUID:  storageUUID,
		Title: name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to rename storage %s to %s: %w", storageUUID, name, err)
	}

	return d.WaitStorageOnline(ctx, details.UUID)
}

func (d *driver) CreateTemplateStorage(ctx context.Context, title, zone string, size int, tier string) (*upcloud.Storage, error) {
	storage, err := d.svc.CreateStorage(ctx, &request.CreateStorageRequest{
		Size:  size,
		Tier:  tier,
		Title: title,
		Zone:  zone,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create template storage %s in zone %s: %w", title, zone, err)
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
		return nil, fmt.Errorf("failed to create storage import for %s: %w", storageUUID, err)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, d.config.Timeout)
	defer cancel()

	result, err := d.svc.WaitForStorageImportCompletion(timeoutCtx, &request.WaitForStorageImportCompletionRequest{
		StorageUUID: storageUUID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to wait for storage import completion for %s: %w", storageUUID, err)
	}
	return result, nil
}

func (d *driver) DeleteTemplate(ctx context.Context, templateUUID string) error {
	return d.DeleteStorage(ctx, templateUUID)
}

func (d *driver) DeleteStorage(ctx context.Context, storageUUID string) error {
	err := d.svc.DeleteStorage(ctx, &request.DeleteStorageRequest{
		UUID: storageUUID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete storage %s: %w", storageUUID, err)
	}
	return nil
}

func (d *driver) CloneStorage(ctx context.Context, storageUUID, zone, title string) (*upcloud.Storage, error) {
	response, err := d.svc.CloneStorage(ctx, &request.CloneStorageRequest{
		UUID:  storageUUID,
		Zone:  zone,
		Title: title,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to clone storage %s to zone %s with title %s: %w", storageUUID, zone, title, err)
	}
	return d.WaitStorageOnline(ctx, response.UUID)
}

func (d *driver) getStorageByUUID(ctx context.Context, storageUUID string) (*upcloud.Storage, error) {
	response, err := d.svc.GetStorageDetails(ctx, &request.GetStorageDetailsRequest{
		UUID: storageUUID,
	})
	if err != nil {
		return nil, fmt.Errorf("error fetching storages: %w", err)
	}
	return &response.Storage, nil
}

func (d *driver) getStorageByName(ctx context.Context, storageName string) (*upcloud.Storage, error) {
	response, err := d.svc.GetStorages(ctx, &request.GetStoragesRequest{
		Type: upcloud.StorageTypeTemplate,
	})
	if err != nil {
		return nil, fmt.Errorf("error fetching storages: %w", err)
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
		return nil, fmt.Errorf("failed to find storage by name %q", storageName)
	}
	return &storage, nil
}

func (d *driver) waitDesiredState(ctx context.Context, serverUUID, state string) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, d.config.Timeout)
	defer cancel()

	request := &request.WaitForServerStateRequest{
		UUID:         serverUUID,
		DesiredState: state,
	}
	if _, err := d.svc.WaitForServerState(timeoutCtx, request); err != nil {
		return fmt.Errorf("error while waiting for server to change state to %q: %w", state, err)
	}
	return nil
}

func (d *driver) waitUndesiredState(ctx context.Context, serverUUID, state string) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, d.config.Timeout)
	defer cancel()

	request := &request.WaitForServerStateRequest{
		UUID:           serverUUID,
		UndesiredState: state,
	}
	if _, err := d.svc.WaitForServerState(timeoutCtx, request); err != nil {
		return fmt.Errorf("error while waiting for server to change state from %q: %w", state, err)
	}
	return nil
}

func (d *driver) getServerDetails(ctx context.Context, serverUUID string) (*upcloud.ServerDetails, error) {
	response, err := d.svc.GetServerDetails(ctx, &request.GetServerDetailsRequest{
		UUID: serverUUID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get details for server: %w", err)
	}
	return response, nil
}

func (d *driver) GetServerStorage(ctx context.Context, serverUUID string) (*upcloud.ServerStorageDevice, error) {
	details, err := d.getServerDetails(ctx, serverUUID)
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
		return nil, fmt.Errorf("failed to find storage type disk for server %q", serverUUID)
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
				Storage: opts.StorageUUID,
				Title:   titleDisk,
				Size:    opts.StorageSize,
				Tier:    opts.StorageTier,
			},
		},
		Networking: &request.CreateServerNetworking{
			Interfaces: opts.Networking,
		},
		LoginUser: &request.LoginUser{
			CreatePassword: "no",
			Username:       d.config.SSHUsername,
			SSHKeys:        []string{opts.SSHPublicKey},
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

func CredentialsFromEnv(username, password, token string) (credentials.Credentials, error) {
	config := credentials.Credentials{
		Username: username,
		Password: password,
		Token:    token,
	}

	if config.Username == "" {
		config.Username = os.Getenv(EnvConfigUsernameLegacy)
	}
	if config.Password == "" {
		config.Password = os.Getenv(EnvConfigPasswordLegacy)
	}

	return credentials.Parse(config) //nolint:wrapcheck // Use the original error from shared credentials package
}
