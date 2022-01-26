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
}

build {
  name = "umee-node"
  sources = [
    "source.docker.final",
    "source.googlecompute.final"
  ]

  provisioner "shell" {
    inline = ["apt update"
      , "sed -i 's/http:\/\/.\+\/ubuntu/http:\/\/mirrors.edge.kernel.org\/ubuntu/g' /etc/apt/sources.list"
      , "apt update"
      , "apt install -y --no-install-recommends ack apt-transport-https bsdmainutils ca-certificates curl debian-keyring debian-archive-keyring iputils-ping jq less lsof nano ncat net-tools nmap supervisor sysstat telnet traceroute vim"
      , "curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | tee /etc/apt/trusted.gpg.d/caddy-stable.asc"
      , "curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list"
      , "apt update"
      , "apt install -y --no-install-recommends caddy"
      , "apt dist-upgrade -y"
      , "rm -rf /var/lib/{apt,dpkg,cache,log}/"
      , "curl -sLf https://github.com/informalsystems/ibc-rs/releases/download/v0.10.0/hermes-v0.10.0-x86_64-unknown-linux-gnu.tar.gz | tar -C /usr/local/bin -xz"
      , "curl -sLf https://github.com/cosmos/gaia/releases/download/v6.0.0/gaiad-v6.0.0-linux-amd64 -o /usr/local/bin/gaiad"
      , "chmod a+x /usr/local/bin/gaiad"
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
