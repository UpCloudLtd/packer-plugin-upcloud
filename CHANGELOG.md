# Changelog

All notable changes to this project will be documented in this file.
See updating [Changelog example here](https://keepachangelog.com/en/1.0.0/)

## [Unreleased]

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

[Unreleased]: https://github.com/UpCloudLtd/packer-plugin-upcloud/compare/v1.5.1...HEAD
[1.5.1]: https://github.com/UpCloudLtd/packer-plugin-upcloud/compare/v1.5.0...v1.5.1
[1.5.0]: https://github.com/UpCloudLtd/packer-plugin-upcloud/compare/v1.4.0...v1.5.0
[1.4.0]: https://github.com/UpCloudLtd/packer-plugin-upcloud/compare/v1.3.1...v1.4.0
[1.3.1]: https://github.com/UpCloudLtd/packer-plugin-upcloud/compare/v1.3.0...v1.3.1
[1.3.0]: https://github.com/UpCloudLtd/packer-plugin-upcloud/compare/v1.2.0...v1.3.0
[1.2.0]: https://github.com/UpCloudLtd/packer-plugin-upcloud/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/UpCloudLtd/packer-plugin-upcloud/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/UpCloudLtd/packer-plugin-upcloud/releases/tag/v1.0.0