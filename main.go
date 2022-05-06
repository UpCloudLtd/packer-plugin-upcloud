package main

import (
	"fmt"
	"os"

	"github.com/UpCloudLtd/packer-plugin-upcloud/builder/upcloud"
	upcloudimport "github.com/UpCloudLtd/packer-plugin-upcloud/post-processor/upcloud-import"
	"github.com/UpCloudLtd/packer-plugin-upcloud/version"

	"github.com/hashicorp/packer-plugin-sdk/plugin"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterBuilder(plugin.DEFAULT_NAME, new(upcloud.Builder))
	pps.RegisterPostProcessor("import", new(upcloudimport.PostProcessor))
	pps.SetVersion(version.PluginVersion)
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
