#!/usr/bin/env sh

#mockgen_cmd="mockgen"
mockgen_cmd="go run github.com/golang/mock/mockgen"

$mockgen_cmd -source ./x/uibc/expected_keepers.go -package fixtures -destination ./x/uibc/fixtures/keepers.go
