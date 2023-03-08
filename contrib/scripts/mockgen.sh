#!/usr/bin/env sh

mockgen_cmd="go run github.com/golang/mock/mockgen"

$mockgen_cmd -source ./x/uibc/expected_keepers.go -package mocks -destination ./x/uibc/mocks/keepers.go
