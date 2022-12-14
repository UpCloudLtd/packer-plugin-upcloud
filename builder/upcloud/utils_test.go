package upcloud

import (
	"reflect"
	"testing"

	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud/request"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

func TestFindIPAddressByType(t *testing.T) {
	want := "127.0.0.1"
	got, err := findIPAddressByType(upcloud.IPAddressSlice{
		upcloud.IPAddress{
			Access:  upcloud.IPAddressAccessPrivate,
			Address: want,
			Family:  "IPv4",
		},
		upcloud.IPAddress{
			Access:  upcloud.IPAddressAccessPrivate,
			Address: "127.0.0.2",
			Family:  "IPv4",
		},
	}, InterfaceTypePrivate)

	if err != nil {
		t.Fatal(err)
	}
	if got.Address != want {
		t.Errorf("findIPAddressByType failed want %s got %s", want, got.Address)
	}

	got, err = findIPAddressByType(upcloud.IPAddressSlice{
		upcloud.IPAddress{
			Access:  upcloud.IPAddressAccessPublic,
			Address: want,
			Family:  "IPv4",
		},
		upcloud.IPAddress{
			Access:  upcloud.IPAddressAccessPublic,
			Address: "::1/128",
			Family:  "IPv6",
		},
	}, InterfaceTypePublic)

	if err != nil {
		t.Fatal(err)
	}
	if got.Address != want {
		t.Errorf("findIPAddressByType failed want %s got %s", want, got.Address)
	}

	got, err = findIPAddressByType(upcloud.IPAddressSlice{
		upcloud.IPAddress{
			Access:  upcloud.IPAddressAccessPublic,
			Address: want,
			Family:  "IPv4",
		},
		upcloud.IPAddress{
			Access:  upcloud.IPAddressAccessPublic,
			Address: "::1/128",
			Family:  "IPv6",
		},
	}, InterfaceTypePrivate)

	if err == nil {
		t.Errorf("findIPAddressByType failed got %s instead of error", got.Address)
	}
}

func TestSSHHostCallback(t *testing.T) {
	stateIPv6 := multistep.BasicStateBag{}
	stateIPv6.Put("server_ip_address", &IPAddress{
		Default: false,
		Family:  "IPv6",
		Address: "IPv6_address",
	})
	want := "[IPv6_address]"
	got, err := sshHostCallback(&stateIPv6)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("IPv6 sshHostCallback failed want %s got %s", want, got)
	}
	stateIPv4 := multistep.BasicStateBag{}
	stateIPv4.Put("server_ip_address", &IPAddress{
		Default: false,
		Family:  "IPv4",
		Address: "IPv4_address",
	})
	want = "IPv4_address"
	got, err = sshHostCallback(&stateIPv4)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("IPv4 sshHostCallback failed want %s got %s", want, got)
	}
}

func TestConvertNetworkTypes(t *testing.T) {
	want := []request.CreateServerInterface{
		{
			IPAddresses: []request.CreateServerIPAddress{
				{
					Family:  "IPv4",
					Address: "127.0.0.3",
				},
			},
			Type:    upcloud.NetworkTypeUtility,
			Network: "",
		},
		{
			IPAddresses: []request.CreateServerIPAddress{
				{
					Family:  "IPv4",
					Address: "127.0.0.2",
				},
			},
			Type:    upcloud.NetworkTypePrivate,
			Network: "",
		},
		{
			IPAddresses: []request.CreateServerIPAddress{
				{
					Family:  "IPv6",
					Address: "127.0.0.1",
				},
			},
			Type:    upcloud.NetworkTypePublic,
			Network: "",
		},
	}
	got := convertNetworkTypes([]NetworkInterface{
		{
			IPAddresses: []IPAddress{{
				Family:  "IPv4",
				Address: "127.0.0.3",
			}},
			Type:    InterfaceTypeUtility,
			Network: "",
		},
		{
			IPAddresses: []IPAddress{{
				Family:  "IPv4",
				Address: "127.0.0.2",
			}},
			Type:    InterfaceTypePrivate,
			Network: "",
		},
		{
			IPAddresses: []IPAddress{{
				Family:  "IPv6",
				Address: "127.0.0.1",
			}},
			Type:    InterfaceTypePublic,
			Network: "",
		},
	})
	for i := range got {
		if !reflect.DeepEqual(want[i], got[i]) {
			t.Fatalf("convertNetworkTypes failed IP want %+v got %+v", want, got)
		}
	}
}
