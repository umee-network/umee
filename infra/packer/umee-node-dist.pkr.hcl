variable "git_sha" {
  type = string
}

packer {
  required_plugins {
    docker = {
      version = ">= 1.0.3"
      source  = "github.com/hashicorp/docker"
    }
    googlecompute = {
      version = ">= 1.0.9"
      source  = "github.com/hashicorp/googlecompute"
    }
  }
}

source "docker" "final" {
  image  = "ubuntu:20.04"
  commit = true
  changes = [
    "ENTRYPOINT [\"/usr/local/bin/umeed\"]"
  ]
}

source "googlecompute" "final" {
  project_id   = "umeedefi"
  source_image_family = "ubuntu-minimal-2004-lts"
  ssh_username = "root"
  zone         = "us-central1-a"
  image_name   = "umee-node-${var.git_sha}"
  machine_type = "n2-standard-2"
}

build {
  name = "umee-node"

  sources = [
    "source.docker.final",
    "source.googlecompute.final"
  ]

  provisioner "shell" {
    inline = [ "/usr/bin/cloud-init status --wait" ]
    only = ["googlecompute.final"]
  }

  provisioner "shell" {
    inline = [ "sed -i 's/http:\\/\\/.\\+\\/ubuntu/http:\\/\\/mirrors.edge.kernel.org\\/ubuntu/g' /etc/apt/sources.list"
      , "apt update && apt install -y --no-install-recommends ack apt-transport-https bsdmainutils ca-certificates curl debian-keyring debian-archive-keyring gpg iputils-ping jq less lsof nano ncat net-tools nmap supervisor sysstat telnet traceroute vim"
      , "curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg"
      , "curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list"
      , "apt update"
      , "apt install -y --no-install-recommends caddy"
      , "apt dist-upgrade -y"
      , "rm -rf /var/lib/{apt,dpkg,cache,log}/"
      , "curl -sLf https://github.com/informalsystems/ibc-rs/releases/download/v0.14.0/hermes-v0.14.0-x86_64-unknown-linux-gnu.tar.gz | tar -C /usr/local/bin -xz"
      , "curl -sLf https://github.com/terra-money/core/releases/download/v2.0.0-rc.1/terra_2.0.0-rc.1_Linux_x86_64.tar.gz | tar -C /usr/local/bin -xz"
      , "curl -sLf https://github.com/cosmos/gaia/releases/download/v7.0.1/gaiad-v7.0.1-linux-amd64 -o /usr/local/bin/gaiad"
      , "chmod a+x /usr/local/bin/gaiad"
      , "curl -sLf https://github.com/CosmWasm/wasmvm/raw/v1.0.0-beta10/api/libwasmvm.so -o /usr/local/lib/libwasmvm.so"
      , "ldconfig"
      , "curl -sLf https://github.com/osmosis-labs/osmosis/releases/download/v7.2.0/osmosisd-7.2.0-linux-amd64 -o /usr/local/bin/osmosisd"
      , "chmod a+x /usr/local/bin/osmosisd"
      , "curl -sLf https://github.com/CosmosContracts/juno/releases/download/v3.1.1/junod -o /usr/local/bin/junod"
      , "chmod a+x /usr/local/bin/junod"
      , "cd /tmp && curl -sLqf https://github.com/umee-network/umee/releases/download/price-feeder/v0.2.1/price-feeder-v0.2.1-linux-amd64.tar.gz | tar --strip-components 1 -xz"
      , "cp /tmp/price-feeder /usr/local/bin/"
      , "cd /tmp && curl -sLqf https://github.com/umee-network/peggo/releases/download/v0.2.7/peggo-v0.2.7-linux-amd64.tar.gz | tar --strip-components 1 -xz"
      , "cp /tmp/peggo /usr/local/bin/"
    ]
  }

  provisioner "file" {
    source      = "dist/x86_64/"
    destination = "/usr/local/bin"
  }

  post-processors {
    post-processor "docker-tag" {
      repository = "us-docker.pkg.dev/umeedefi/stack/node"
      tags       = ["${var.git_sha}"]
      only       = ["docker.final"]
    }
    post-processor "docker-push" {
      only = ["docker.final"]
    }
  }
}