# UpCloud Packer builder

![Build Status](https://github.com/UpCloudLtd/packer-plugin-upcloud/workflows/test/badge.svg)
![Release Status](https://github.com/UpCloudLtd/packer-plugin-upcloud/workflows/release/badge.svg)

This is a builder plugin for Packer which can be used to generate storage templates on UpCloud. It utilises the [UpCloud Go API](https://github.com/UpCloudLtd/upcloud-go-api) to interface with the UpCloud API.

## Installation

### Installing using Packer Packet Manager

In order to use `packer init`:

 1. You need to have Packer version ">=1.7.0" installed. 
 2. The config template should be in `hcl` format.
 3. The config template file, when in `hcl` format, must have the extension `.pkr.hcl`.
 4. The config template must contain `required_plugins` block. For example:

```hcl
...
packer {
    required_plugins {
        upcloud = {
            version = ">=v1.0.0"
            source = "github.com/UpCloudLtd/upcloud"
        }
    }
}
...
```

Run following command and check the output (NOTE: The file extension must be `*.pkr.hcl`):
```sh
$ packer init examples/basic_example.pkr.hcl

Installed plugin github.com/upcloudltd/upcloud v1.0.0 in "/Users/johndoe/.packer.d/plugins/github.com/upcloudltd/upcloud/packer-plugin-upcloud_v1.0.0_x5.0_darwin_amd64"
```

After successful plugin installation, you can run the build command:
```sh
$ packer build examples/basic_example.pkr.hcl

...
```

From Packer version 1.7.0, template HCL2 becomes officially the preferred way to write Packer configuration. While the `json` format is still supported, but certain new features, such as `packer init` works only in newer HCL2 format.
If you are using `json` config templates, please consider upgrading them using the packer built-in command:

```sh
$ packer hcl2_upgrade example.json
Successfully created example.json.pkr.hcl
```




### Pre-built binaries

You can download the pre-built binaries of the plugin from the [GitHub releases page](https://github.com/UpCloudLtd/packer-plugin-upcloud/releases). Just download the archive for your operating system and architecture, unpack it, and place the binary in the appropriate location, e.g. on Linux `~/.packer.d/plugins`. Make sure the file is executable, then install [Packer](https://www.packer.io/).

Please note that use of symlinks with plugin binary can break Packer functionality ([see for details and workaround](https://github.com/hashicorp/packer/issues/10749#issuecomment-800120263)).


### Installing from source

#### Prerequisites

You will need to have the [Go](https://golang.org/) programming language and the [Packer](https://www.packer.io/) itself installed. You can find instructions to install each of the prerequisites at their documentation.

#### Building and installing

Run the following commands to download and install the plugin from the source.

```sh
git clone https://github.com/UpCloudLtd/packer-plugin-upcloud
cd packer-plugin-upcloud
go build
cp packer-plugin-upcloud ~/.packer.d/plugins/
```

## Usage

The builder will automatically generate a temporary SSH key pair for the `root` user which is used for provisioning. This means that if you do not provision a user during the process you will not be able to gain access to your server.

If you want to login to a server deployed with the template, you might want to include an SSH key to your `root` user by replacing the `<ssh-rsa_key>` in the below example with your public key.

Here is a sample template, which you can also find in the `examples/` directory. It reads your UpCloud API credentials from the environment variables and creates an Ubuntu 20.04 LTS server in the `nl-ams1` region.

```json
{
  "variables": {
    "username": "{{ env `UPCLOUD_API_USER` }}",
    "password": "{{ env `UPCLOUD_API_PASSWORD` }}"
  },
  "builders": [
    {
      "type": "upcloud",
      "username": "{{ user `username` }}",
      "password": "{{ user `password` }}",
      "zone": "nl-ams1",
      "storage_uuid": "01000000-0000-4000-8000-000030200200"
    }
  ],
  "provisioners": [
    {
      "type": "shell",
      "inline": [
        "apt-get update",
        "apt-get upgrade -y",
        "echo '<ssh-rsa_key>' | tee /root/.ssh/authorized_keys"
      ]
    }
  ]
}
```

You will need to provide a username and a password with the access rights to the API functions to authenticate. We recommend setting up a subaccount with only the API privileges for security purposes. You can do this at your [UpCloud Control Panel](https://my.upcloud.com/account) in the My Account menu under the User Accounts tab.

Enter the API user credentials in your terminal with the following commands. Replace the `<API_username>` and `<API_password>` with your user details.

```sh
export UPCLOUD_API_USER=<API_username>
export UPCLOUD_API_PASSWORD=<API_password>
```
Then run Packer using the example template with the command underneath.
```
packer build examples/basic_example.json
```
If everything goes according to plan, you should see something like the example output below.

```sh
upcloud: output will be in this color.

==> upcloud: Creating temporary ssh key...
==> upcloud: Getting storage...
==> upcloud: Creating server based on storage "Ubuntu Server 20.04 LTS (Focal Fossa)"...
==> upcloud: Server "packer-custom-image-20210206-213858" created and in 'started' state
==> upcloud: Using ssh communicator to connect: 94.237.109.25
==> upcloud: Waiting for SSH to become available...
==> upcloud: Connected to SSH!
==> upcloud: Provisioning with shell script: /var/folders/pt/x34s6zq90qxb78q8fcwx6jx80000gn/T/packer-shell198358867
    upcloud: Hit:1 http://archive.ubuntu.com/ubuntu focal InRelease
    upcloud: Get:2 http://archive.ubuntu.com/ubuntu focal-updates InRelease [114 kB]
...
==> upcloud: Stopping server "packer-custom-image-20210206-213858"...
==> upcloud: Server "packer-custom-image-20210206-213858" is now in 'stopped' state
==> upcloud: Creating storage template for server "packer-custom-image-20210206-213858"...
==> upcloud: Storage template for server "packer-custom-image-20210206-213858" created
==> upcloud: Stopping server "packer-custom-image-20210206-213858"...
==> upcloud: Deleting server "packer-custom-image-20210206-213858"...
Build 'upcloud' finished after 3 minutes 19 seconds.
```

## Configuration reference

This section describes the available configuration options for the builder. Please note that the purpose of the builder is to create a storage template that can be used as a source for deploying new servers, therefore the temporary server used for building the template is not configurable.

### Required values

* `username` (string) The username to use when interfacing with the UpCloud API.
* `password` (string) The password to use when interfacing with the UpCloud API.
* `zone` (string) The zone in which the server and template should be created (e.g. `nl-ams1`).
* `storage_uuid` (string) The UUID of the storage you want to use as a template when creating the server.


### Optional values

* `storage_name` (string) The name of the storage that will be used to find the first matching storage in the list of existing templates. Note that `storage_uuid` parameter has higher priority. You should use either `storage_uuid` or `storage_name` for not strict matching (e.g "ubuntu server 20.04").
* `storage_size` (int) The storage size in gigabytes. Defaults to `25`. Changing this value is useful if you aim to build a template for larger server configurations where the preconfigured server disk is larger than 25 GB. The operating system disk can also be later extended if needed. Note that Windows templates require large storage size, than default 25 Gb.
* `state_timeout_duration` (string) The amount of time to wait for resource state changes. Defaults to `5m`.
* `template_prefix` (string) The prefix to use for the generated template title. Defaults to an empty string, meaning the prefix will be the storage title. You can use this option to easily differentiate between different templates.
* `template_name` (string) Similarly to `template_prefix`, but this will allow you to set the full template name and not just the prefix. Defaults to an empty string, meaning the name will be the storage title. You can use this option to easily differentiate between different templates. It cannot be used in conjunction with the prefix setting.
* `clone_zones` ([]string) The array of extra zones (locations) where created templates should be cloned. Note that default `state_timeout_duration` is not enough for cloning, better to increase a value depending on storage size.
* `ssh_private_key_path` (string) Path to SSH Private Key that will be used for provisioning and stored in the template.
* `ssh_public_key_path` (string) Path to SSH Public Key that will be used for provisioning.
* `network_interfaces` (array) The array of network interfaces to request during the creation of the server for building the packer image. Example:

```json
    ...
    "network_interfaces": [
        {
            "type": "public",
            "ip_addresses": [
                {
                    "family": "IPv4"
                }
            ]
        },
        {
            "type": "private",
            "network": "{{ private_network_uuid }}",
            "ip_addresses": [
                {
                    "family": "IPv4",
                    "address": "192.168.3.123"
                }
            ]
        },
        {
            "type": "utility",
            "ip_addresses": [
                {
                    "family": "IPv4"
                }
            ]
        }
    ]
    ...
```


## License

This project is distributed under the [MIT License](https://opensource.org/licenses/MIT), see LICENSE.txt for more information.
