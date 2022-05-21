#!/usr/bin/env bash

# Because Mac OS's version of sed has a few kinks and bugs, this command cannot be run directly from inside the Makefile
# shellcheck disable=SC1003
sed -i '' '4i\'$'\noption go_package = "github.com/confio/ics23/go";' "$1"