#!/usr/bin/env bash

if [ "$(uname)" == "Darwin" ]; then
  # shellcheck disable=SC2038
  find . -type f  -path '*.go' ! -path '*.pb.go' ! -path '*.pb.gw.go' ! -path '*/mock/*' | xargs sed -i '' '/import/, /)/ {/^$/ d;}'
else
  # shellcheck disable=SC2038
  find . -type f  -path '*.go' ! -path '*.pb.go' ! -path '*.pb.gw.go' ! -path '*/mock/*' | xargs sed -i '/import/, /)/ {/^$/ d}'
fi
