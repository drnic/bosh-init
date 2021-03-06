# BOSH Micro CLI [![Build Status](https://travis-ci.org/cloudfoundry/bosh-micro-cli.svg?branch=master)](https://travis-ci.org/cloudfoundry/bosh-micro-cli)

This is the BOSH Micro CLI rewritten to support external CPIs.

* Documentation: [docs.cloudfoundry.org/bosh](http://docs.cloudfoundry.org/bosh)
* IRC: `#bosh` on freenode
* Google groups:
  [bosh-users](https://groups.google.com/a/cloudfoundry.org/group/bosh-users/topics) &
  [bosh-dev](https://groups.google.com/a/cloudfoundry.org/group/bosh-dev/topics) &
  [vcap-dev](https://groups.google.com/a/cloudfoundry.org/group/vcap-dev/topics) (for CF)

## Usage

1. Build micro:

  ```
  bin/build
  ```

2. Set up deployment manifest:

  ```
  out/bosh-micro deployment manifest.yml
  ```

3. Deploy

  ```
  out/bosh-micro deploy stemcell.tgz cpi-release.tgz
  ```

where `cpi-release.tgz` is a BOSH CPI release and `stemcell.tgz` is a BOSH stemcell appropriate for the CPI release.

Please see the [CLI workflow](docs/cli_workflow.md) for more information on creating a manifest.

## Logging

To output logs during bosh-micro commands set the `BOSH_MICRO_LOG_LEVEL` environment variable to one of the following values: 

DEBUG, INFO, WARN, ERROR, NONE (default)

To output logs to a file set the `BOSH_MICRO_LOG_PATH` environment variable to the path of the file to create and/or append to. 

By default (when `BOSH_MICRO_LOG_LEVEL` is not NONE) logs write to STDOUT (debug & info) & STDERR (warn & error).


## Contributing

### Set up a workstation for development

Note: This guide assumes a few things:

- You have gcc (or an equivalent)
- You can install packages (brew, apt-get, or equivalent)

Get Golang and its dependencies (Mac example, replace with your package manager of choice):

- `brew update`
- `brew install go`
- `brew install git` (Go needs git for the `go get` command)
- `brew install hg` (Go needs mercurial for the `go get` command)

Clone and set up the BOSH Micro CLI repository:

- `go get -d github.com/cloudfoundry/bosh-micro-cli`
- `cd $GOPATH/src/github.com/cloudfoundry/bosh-micro-cli`

From here on out we assume you're working in `$GOPATH/src/github.com/cloudfoundry/bosh-micro-cli`

To build the micro cli:

- `bin/build` # The `bosh-micro` binary will be located in `out/`

Install tools used by the BOSH Micro CLI test suite:

- `bin/go get code.google.com/p/go.tools/cmd/vet`
- `bin/go get github.com/golang/lint/golint`

### Running unit tests

Each package in the agent has its own unit tests and there are integration tests in the `integration` package.
You can also run all tests with `bin/test`.

### Running acceptance tests

Install sshpass:

- `brew install https://raw.github.com/eugeneoden/homebrew/eca9de1/Library/Formula/sshpass.rb`

Acceptance tests can be run in a VM with the following vagrant providers:

* virtualbox
* vmware_fusion
* vmware_workstation
* aws

#### Running tests using local provider

  When using a local provider, you may choose to download both the
  bosh-warden-cpi-release and stemcell to a local directory to then be copied
  into the VM. To take advantage of this feature, export the following variables
  prior to running the tests (absolute paths are required):

      $ export BOSH_MICRO_CPI_RELEASE_PATH=/tmp/bosh-warden-cpi-9.tgz
      $ export BOSH_MICRO_STEMCELL_PATH=/tmp/bosh-stemcell-348-warden-boshlite-ubuntu-trusty-go_agent.tgz
      $ ./bin/test-acceptance-with-vm --provider=virtualbox

#### Running tests using AWS provider

  When using the AWS provider, you will need to provide the following:

      $ export BOSH_MICRO_PRIVATE_KEY=/tmp/id_rsa

      # The following variables are required by Bosh Lite
      $ export BOSH_AWS_ACCESS_KEY_ID=access_key
      $ export BOSH_AWS_SECRET_ACCESS_KEY=secret
      $ export BOSH_LITE_KEYPAIR=keypair
      $ export BOSH_LITE_SECURITY_GROUP=sg-1234
      $ export BOSH_LITE_PRIVATE_KEY=$BOSH_MICRO_PRIVATE_KEY

#### Running tests against existing VM

Acceptance tests use configuration file specified via `BOSH_MICRO_CONFIG_PATH`. The format of the configuration file is basic JSON.

```
{
  "vm_username": "TEST_VM_USERNAME",
  "vm_ip": "TEST_VM_IP",
  "private_key_path": "TEST_VM_PRIVATE_KEY_PATH",
  "cpi_release_path": "CPI_RELEASE_PATH",
  "cpi_release_url": "CPI_RELEASE_URL",
  "stemcell_path": "STEMCELL_PATH",
  "stemcell_url": "STEMCELL_URL"
}
```

Run tests:

```
BOSH_MICRO_CONFIG_PATH=config.json bin/test-acceptance
```
