package upcloud

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
)

// wraps error logic.
func stepHaltWithError(state multistep.StateBag, err error) multistep.StepAction {
	uiRaw := state.Get("ui")
	if ui, ok := uiRaw.(packer.Ui); ok {
		ui.Error(err.Error())
	}
	state.Put("error", err)
	return multistep.ActionHalt
}

// Find IP address by type from list of IP addresses.
func findIPAddressByType(addrs upcloud.IPAddressSlice, infType InterfaceType) (*IPAddress, error) {
	var ipv6 *IPAddress
	for _, ipAddress := range addrs {
		if ipAddress.Access == string(infType) {
			switch ipAddress.Family {
			case upcloud.IPAddressFamilyIPv4:
				// prefer IPv4 over IPv6 - return first matching IPv4 interface if found
				return &IPAddress{Address: ipAddress.Address, Family: ipAddress.Family}, nil
			case upcloud.IPAddressFamilyIPv6:
				// not returning IPv6 because there might be IPv4 address coming up in the slice
				ipv6 = &IPAddress{Address: ipAddress.Address, Family: ipAddress.Family}
			}
		}
	}
	// return IPv6 if found
	if ipv6 != nil {
		return ipv6, nil
	}
	return nil, fmt.Errorf("unable to find '%s' IP address", infType)
}

func getNowString() string {
	return time.Now().Format("20060102-150405")
}

// sshHostCallback returns server's IP addresss.
// Note that IPv6 address needs to be enclosed in square brackets.
func sshHostCallback(state multistep.StateBag) (string, error) {
	addr, ok := state.Get("server_ip_address").(*IPAddress)
	if !ok || addr == nil {
		return "", errors.New("unable to get server_ip_address from state")
	}
	if addr.Family == upcloud.IPAddressFamilyIPv6 {
		return fmt.Sprintf("[%s]", addr.Address), nil
	}
	return addr.Address, nil
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
			Type:        string(iface.Type),
			Network:     iface.Network,
		})
	}
	return networking
}

func contextWithDefaultTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), defaultTimeout)
}

func defaultNetworking() []request.CreateServerInterface {
	return []request.CreateServerInterface{
		{
			IPAddresses: []request.CreateServerIPAddress{
				{
					Family: upcloud.IPAddressFamilyIPv4,
				},
			},
			Type: upcloud.IPAddressAccessPublic,
		},
	}
}
