#!/usr/bin/env sh

mockgen_cmd="go run github.com/golang/mock/mockgen"

$mockgen_cmd -source ./x/ugov/keeper_interfaces.go -package mocks -destination ./x/ugov/mocks/keepers.go
$mockgen_cmd -source ./x/uibc/expected_keepers.go -package mocks -destination ./x/uibc/mocks/keepers.go
$mockgen_cmd -source ./x/metoken/expected_keepers.go -package mocks -destination ./x/metoken/mocks/keepers.go
$mockgen_cmd -source ./app/inflation/expected_keepers.go -package mocks -destination ./app/inflation/mocks/keepers.go
