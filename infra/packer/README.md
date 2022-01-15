
# [Packer](https://www.packer.io)

Our Cold-Image creation tool.

[Tutorials](https://learn.hashicorp.com/packer)

## Mac Installation

```bash
brew tap hashicorp/tap
brew install hashicorp/tap/packer
```

## Testing Packer hcl changes outside of CI
```bash
export GITHUB_SHA=foobar
bin/pack-and-distribute-only-docker
```
