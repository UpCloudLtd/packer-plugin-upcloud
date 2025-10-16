# Changelog

All notable changes to this project will be documented in this file.
See updating [Changelog example here](https://keepachangelog.com/en/1.0.0/)

## [Unreleased]

## [1.9.2] - 2025-10-16

### Changed

- Increase default timeout to `20m`.

## [1.9.1] - 2025-08-29

### Fixed

- Network interface validation in builder configuration
- Vulnerability regarding LZMA archives in ulikunitz/xz (CVE-2025-58058)

## [1.9.0] - 2025-08-26

### Added

- Support for using API token from system keyring.

### Changed

- If both basic authentication and API token are provided, the API token will be used instead of raising an error.

## [1.8.1] - 2025-08-18

### Fixed

- Use `zip` archive format for release assets
- Update Go version to 1.25.0

## [1.8.0] - 2025-08-11

### Added

- `storage_size` parameter to upcloud-import post-processor

### Fixed

- Update Go version to 1.24.6

### Changed

- Update `packer-plugin-sdk` to [v0.6.2](https://github.com/hashicorp/packer-plugin-sdk/releases/tag/v0.6.2)
- Update `upcloud-go-api` to [v8.23.0](https://github.com/UpCloudLtd/upcloud-go-api/releases/tag/v8.23.0)

## [1.7.0] - 2025-06-10

### Added

- Authentication token support through `UPCLOUD_TOKEN` environment variable and `token` config parameter

### Fixed

- Update Go version to 1.24.4

## [1.6.0] - 2025-05-23

### Added

- Enable storage tier customisation through `storage_tier` parameter

### Fixed

- Update Go version to 1.24 and fix security vulnerabilities

### Changed

- Update `packer-plugin-sdk` to [v0.6.1](https://github.com/hashicorp/packer-plugin-sdk/releases/tag/v0.6.1)
- Update `upcloud-go-api` to [v8.18.0](https://github.com/UpCloudLtd/upcloud-go-api/releases/tag/v8.18.0)

## [1.5.3] - 2024-01-02

### Fixed
- Update UpCloud Go SDK to v6 and fix security vulnerabilities

## [1.5.2] - 2022-12-15

### Fixed
- enable server metadata automatically when required by base image

### Changed
- update `packer-plugin-sdk` to [v0.3.0](https://github.com/hashicorp/packer-plugin-sdk/blob/main/CHANGELOG.md#030-june-09-2022)
- update upcloud-go-api to v5.1.0

## [1.5.1] - 2022-06-10

### Fixed
- client timeout when importing large images

## [1.5.0] - 2022-06-09

### Added
- HCP Packer image metadata support
- enviroment variables `UPCLOUD_USERNAME` and `UPCLOUD_PASSWORD` for authentication

### Changed
- new upcloud-go-api version v4.6.0

### Deprecated
- environment variables `UPCLOUD_API_USER` and `UPCLOUD_API_PASSWORD`

## [1.4.0] - 2022-05-30

### Added
- new `upcloud-import` post-processor

### Changed
- new upcloud-go-api version v4.5.2

## [1.3.1] - 2022-02-24

### Changed
- update documentation

### Fixed
- update docs.zip content to be compatible with Hashicorp website

## [1.3.0] - 2022-02-22

### Added
- add new template name param 
- new default IP address flag to select used interface/IP during build
- new wait_boot flag adds ability to wait N time for server to boot up and start all services
- add "none" communicator support
- support for IPv6 interfaces

### Changed
- update README file
- update acceptance test to embed HCL2 configs 
- drop public IPv4 interface requirement

### Fixed
- fix network interface config

## [1.2.0] - 2021-06-17

### Fixed
- fix template prefix usage

## [1.1.0] 2021-05-30

### Changed
- bump go version to 1.6 to enable darwin/arm build
- update dependencies
- update intergration tests

## [1.0.0] 2021-02-19

### Changed
- Upgrade to Packer 1.7.0
- Copy codebase from https://github.com/UpCloudLtd/upcloud-packer

[Unreleased]: https://github.com/UpCloudLtd/packer-plugin-upcloud/compare/v1.9.2...HEAD
[1.9.2]: https://github.com/UpCloudLtd/packer-plugin-upcloud/compare/v1.9.1...v1.9.2
[1.9.1]: https://github.com/UpCloudLtd/packer-plugin-upcloud/compare/v1.9.0...v1.9.1
[1.9.0]: https://github.com/UpCloudLtd/packer-plugin-upcloud/compare/v1.8.1...v1.9.0
[1.8.1]: https://github.com/UpCloudLtd/packer-plugin-upcloud/compare/v1.8.0...v1.8.1
[1.8.0]: https://github.com/UpCloudLtd/packer-plugin-upcloud/compare/v1.7.0...v1.8.0
[1.7.0]: https://github.com/UpCloudLtd/packer-plugin-upcloud/compare/v1.6.0...v1.7.0
[1.6.0]: https://github.com/UpCloudLtd/packer-plugin-upcloud/compare/v1.5.3...v1.6.0
[1.5.3]: https://github.com/UpCloudLtd/packer-plugin-upcloud/compare/v1.5.2...v1.5.3
[1.5.2]: https://github.com/UpCloudLtd/packer-plugin-upcloud/compare/v1.5.1...v1.5.2
[1.5.1]: https://github.com/UpCloudLtd/packer-plugin-upcloud/compare/v1.5.0...v1.5.1
[1.5.0]: https://github.com/UpCloudLtd/packer-plugin-upcloud/compare/v1.4.0...v1.5.0
[1.4.0]: https://github.com/UpCloudLtd/packer-plugin-upcloud/compare/v1.3.1...v1.4.0
[1.3.1]: https://github.com/UpCloudLtd/packer-plugin-upcloud/compare/v1.3.0...v1.3.1
[1.3.0]: https://github.com/UpCloudLtd/packer-plugin-upcloud/compare/v1.2.0...v1.3.0
[1.2.0]: https://github.com/UpCloudLtd/packer-plugin-upcloud/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/UpCloudLtd/packer-plugin-upcloud/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/UpCloudLtd/packer-plugin-upcloud/releases/tag/v1.0.0