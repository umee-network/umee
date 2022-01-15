# Umee Images

## Public Docker Image Usage

The latest (full) commit hash from the main branch can be used for the latest version of Umee.

For example:
```bash
docker run -it us-docker.pkg.dev/umeedefi/stack/node:<full-git-sha>

->

A daemon and client for interacting with the Umee network. Umee is a
Universal Capital Facility that can collateralize assets on one blockchain
towards borrowing assets on another blockchain.

Usage:
  umeed [command]

...
```


## [Packer](https://www.packer.io)

Our Cold-Image creation tool.

### Links
* [Tutorials](https://learn.hashicorp.com/packer)
* [Docker Builder Docs](https://www.packer.io/plugins/builders/docker)
* [GCP Builder Docs](https://www.packer.io/plugins/builders/googlecompute)

### Mac Installation

```bash
brew tap hashicorp/tap
brew install hashicorp/tap/packer
```

### Testing Packer hcl changes outside of CI
```bash
export GITHUB_SHA=foobar
bin/pack-and-distribute-only-docker
```
