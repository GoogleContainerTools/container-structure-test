# v1.11.0 Release - 11/09/2021

Highlights:
* Use os.Lstat over os.Stat (#292)
* Add support for the "user" metadata. Related to #80. (#274)
* Move to Go 1.17 to support newer versions of macOS

Big thanks to everyone who contributed to this release:
* charlyx
* dduportal
* midnightconman

## Distribution

container-structure-test is distributed in binary form for Linux (arm64, amd64, s390x, ppc64le), OS X, and Windows systems for the v1.11.0 release, as well as a container image for running tests in Google Cloud Builder.

Binaries are available on Google Cloud Storage. The direct GCS links are:
[Linux/amd64](https://storage.googleapis.com/container-structure-test/v1.11.0/container-structure-test-linux-amd64)
[Linux/arm64](https://storage.googleapis.com/container-structure-test/v1.11.0/container-structure-test-linux-arm64)
[Linux/s390x](https://storage.googleapis.com/container-structure-test/v1.11.0/container-structure-test-linux-s390x)
[Linux/ppc64le](https://storage.googleapis.com/container-structure-test/v1.11.0/container-structure-test-linux-ppc64le)
[Darwin/amd64](https://storage.googleapis.com/container-structure-test/v1.11.0/container-structure-test-darwin-amd64)
[Windows](https://storage.googleapis.com/container-structure-test/v1.11.0/container-structure-test-windows-amd64.exe)

The container image can be found at `gcr.io/gcp-runtimes/container-structure-test:v1.11.0`.

## Installation

### OSX
```shell
curl -LO https://storage.googleapis.com/container-structure-test/v1.11.0/container-structure-test-darwin-amd64 && mv container-structure-test-darwin-amd64 container-structure-test && chmod +x container-structure-test && sudo mv container-structure-test /usr/local/bin/
```
Feel free to leave off the `sudo mv container-structure-test /usr/local/bin` if you would like to add container-structure-test to your path manually.

### Windows
 https://storage.googleapis.com/container-structure-test/v1.11.0/container-structure-test-windows-amd64.exe

### Linux
amd64: 
```shell
curl -LO https://storage.googleapis.com/container-structure-test/v1.11.0/container-structure-test-linux-amd64 && mv container-structure-test-linux-amd64 container-structure-test && chmod +x container-structure-test && sudo mv container-structure-test /usr/local/bin/
```
arm64: 
```shell
curl -LO https://storage.googleapis.com/container-structure-test/v1.11.0/container-structure-test-linux-arm64 && mv container-structure-test-linux-arm64 container-structure-test && chmod +x container-structure-test && sudo mv container-structure-test /usr/local/bin/
```
s390x: 
```shell
curl -LO https://storage.googleapis.com/container-structure-test/v1.11.0/container-structure-test-linux-s390x && mv container-structure-test-linux-s390x container-structure-test && chmod +x container-structure-test && sudo mv container-structure-test /usr/local/bin/
```
ppc64le: 
```shell
curl -LO https://storage.googleapis.com/container-structure-test/v1.11.0/container-structure-test-linux-ppc64le && mv container-structure-test-linux-ppc64le container-structure-test && chmod +x container-structure-test && sudo mv container-structure-test /usr/local/bin/
```
Feel free to leave off the `sudo mv container-structure-test /usr/local/bin` if you would like to add container-structure-test to your path manually.

## Usage
Documentation is available [here](https://github.com/GoogleCloudPlatform/container-structure-test/blob/master/README.md)

v1.10.0 Release - 01/07/2021

Highlights:
* :sparkles: Add new output format JUnit [#254](https://github.com/GoogleContainerTools/container-structure-test/pull/254)
* Produce linux/s390x and linux/ppc64le binaries to use in container_test [#269](https://github.com/GoogleContainerTools/container-structure-test/pull/269)

Big thanks to everyone who contributed to this release:
* barthy1
* charlyx

## Distribution

container-structure-test is distributed in binary form for Linux (arm64 and amd64) and OS X systems for the v1.10.0 release, as well as a container image for running tests in Google Cloud Builder.

Binaries are available on Google Cloud Storage. The direct GCS links are:
[Darwin/amd64](https://storage.googleapis.com/container-structure-test/v1.10.0/container-structure-test-darwin-amd64)
[Linux/amd64](https://storage.googleapis.com/container-structure-test/v1.10.0/container-structure-test-linux-amd64)
[Linux/arm64](https://storage.googleapis.com/container-structure-test/v1.10.0/container-structure-test-linux-arm64)

The container image can be found at `gcr.io/gcp-runtimes/container-structure-test:v1.10.0`.

## Installation

### OSX
```shell
curl -LO https://storage.googleapis.com/container-structure-test/v1.10.0/container-structure-test-darwin-amd64 && mv container-structure-test-darwin-amd64 container-structure-test && chmod +x container-structure-test && sudo mv container-structure-test /usr/local/bin/
```
Feel free to leave off the `sudo mv container-structure-test /usr/local/bin` if you would like to add container-structure-test to your path manually.

### Linux
amd64: 
```shell
curl -LO https://storage.googleapis.com/container-structure-test/v1.10.0/container-structure-test-linux-amd64 && mv container-structure-test-linux-amd64 container-structure-test && chmod +x container-structure-test && sudo mv container-structure-test /usr/local/bin/
```
arm64: 
```shell
curl -LO https://storage.googleapis.com/container-structure-test/v1.10.0/container-structure-test-linux-arm64 && mv container-structure-test-linux-arm64 container-structure-test && chmod +x container-structure-test && sudo mv container-structure-test /usr/local/bin/
```
Feel free to leave off the `sudo mv container-structure-test /usr/local/bin` if you would like to add container-structure-test to your path manually.

## Usage
Documentation is available [here](https://github.com/GoogleCloudPlatform/container-structure-test/blob/master/README.md)