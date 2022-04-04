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
      , "set -x; exitcode=0; command=\"apt-get update && apt install -y --no-install-recommends ack apt-transport-https bsdmainutils ca-certificates curl debian-keyring debian-archive-keyring iputils-ping jq less lsof nano ncat net-tools nmap supervisor sysstat telnet traceroute vim \"; for i in 1 2 3 4 5; do eval \"$command\"; exitcode=$?; bash -c \"exit $exitcode\" && break || sleep 5; done; bash -c \"exit $exitcode\""
      , "curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | tee /etc/apt/trusted.gpg.d/caddy-stable.asc"
      , "curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list"
      , "apt-get update"
      , "apt install -y --no-install-recommends caddy"
      , "apt dist-upgrade -y"
      , "rm -rf /var/lib/{apt,dpkg,cache,log}/"
      , "curl -sLf https://github.com/informalsystems/ibc-rs/releases/download/v0.13.0/hermes-v0.13.0-x86_64-unknown-linux-gnu.tar.gz | tar -C /usr/local/bin -xz"
      , "curl -sLf https://github.com/cosmos/gaia/releases/download/v6.0.4/gaiad-v6.0.4-linux-amd64 -o /usr/local/bin/gaiad"
      , "chmod a+x /usr/local/bin/gaiad"
      , "curl -sLf https://github.com/CosmWasm/wasmvm/raw/v0.16.6/api/libwasmvm.so -o /usr/local/lib/libwasmvm.so"
      , "ldconfig"
      , "curl -sLf https://github.com/osmosis-labs/osmosis/releases/download/v7.0.4/osmosisd-7.0.4-linux-amd64 -o /usr/local/bin/osmosisd"
      , "chmod a+x /usr/local/bin/osmosisd"
      , "curl -sLf https://github.com/CosmosContracts/juno/releases/download/v2.1.0/junod -o /usr/local/bin/junod"
      , "chmod a+x /usr/local/bin/junod"
      , "cd /tmp && curl -sLqf https://github.com/umee-network/umee/releases/download/price-feeder/v0.1.4/price-feeder-v0.1.4-linux-amd64.tar.gz | tar --strip-components 1 -xz"
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
