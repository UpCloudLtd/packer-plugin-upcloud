package upcloud

import (
	"fmt"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

// wraps error logic
func stepHaltWithError(state multistep.StateBag, err error) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	state.Put("error", err)
	ui.Error(err.Error())
	return multistep.ActionHalt
}

// parse public ip from server details
func getServerIp(details *upcloud.ServerDetails) (string, error) {
	for _, ipAddress := range details.IPAddresses {
		if ipAddress.Access == upcloud.IPAddressAccessPublic && ipAddress.Family == upcloud.IPAddressFamilyIPv4 {
			return ipAddress.Address, nil
		}
	}
	return "", fmt.Errorf("Unable to find the public IPv4 address of the server")
}

func getNowString() string {
	return time.Now().Format("20060102-150405")
}

// sshHostCallback retrieves the public IPv4 address of the server
func sshHostCallback(state multistep.StateBag) (string, error) {
	return state.Get("server_ip").(string), nil
}

func convertNetworkTypes(rawNetworking []NetworkInterface) []request.CreateServerInterface {
	networking := []request.CreateServerInterface{}
	for _, iface := range rawNetworking {
		ips := []request.CreateServerIPAddress{}
		for _, ip := range iface.IPAddresses {
			ips = append(ips, request.CreateServerIPAddress{Family: ip.Family, Address: ip.Address})
		}
		networking = append(networking, request.CreateServerInterface{
			IPAddresses: ips,
			Type:        iface.Type,
			Network:     iface.Network,
		})
	}
	return networking
}
