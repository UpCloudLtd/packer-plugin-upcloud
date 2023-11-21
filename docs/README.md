This is a plugin for Packer which can be used to generate storage templates on UpCloud. 
It utilises the [UpCloud Go API](https://github.com/UpCloudLtd/upcloud-go-api) to interface with the UpCloud API.

### Installation

To install this plugin, copy and paste this code into your Packer configuration, then run [`packer init`](https://www.packer.io/docs/commands/init).


```hcl
packer {
  required_plugins {
    upcloud = {
      version = ">=v1.0.0"
      source  = "github.com/UpCloudLtd/upcloud"
    }
  }
}

```

Alternatively, you can use `packer plugins install` to manage installation of this plugin.

```sh
$ packer plugins install github.com/UpCloudLtd/upcloud
```


### Components

#### Builders


- [upcloud](/packer/integrations/upcloudltd/latest/components/builder/upcloud) - The upcloud builder is used to generate storage templates on UpCloud.

#### Post-processors

- [upcloud-import](/packer/integrations/upcloudltd/latest/components/post-processor/import) - The upcloud import post-processors is used to import disk images to UpCloud.

### JSON Templates
From Packer version 1.7.0, template HCL2 becomes officially the preferred way to write Packer configuration. While the `json` format is still supported, but certain new features, such as `packer init` works only in newer HCL2 format.
If you are using `json` config templates, please consider upgrading them using the packer built-in command:

```sh
$ packer hcl2_upgrade example.json
Successfully created example.json.pkr.hcl
```
