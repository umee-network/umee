module github.com/umee-network/umee

go 1.16

// TODO: Use next point release of github.com/peggyjv/gravity-bridge/module
// after v0.1.18 after it's released.
require (
	github.com/Microsoft/go-winio v0.5.0 // indirect
	github.com/cenkalti/backoff/v4 v4.1.1 // indirect
	github.com/containerd/continuity v0.1.0 // indirect
	github.com/cosmos/cosmos-sdk v0.43.0-rc3
	github.com/cosmos/go-bip39 v1.0.0
	github.com/cosmos/ibc-go v1.0.0-rc3
	github.com/ethereum/go-ethereum v1.9.25
	github.com/gogo/protobuf v1.3.3
	github.com/golangci/golangci-lint v1.41.1
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/opencontainers/runc v1.0.1 // indirect
	github.com/ory/dockertest/v3 v3.7.0
	github.com/peggyjv/gravity-bridge/module v0.1.19-0.20210806174644-f75251c7d6a0
	github.com/spf13/cast v1.4.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.8.0
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/tendermint v0.34.11
	github.com/tendermint/tm-db v0.6.4
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	golang.org/x/net v0.0.0-20210716203947-853a461950ff // indirect
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
	google.golang.org/genproto v0.0.0-20210617175327-b9e0b3197ced
	google.golang.org/grpc v1.38.0
)

replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

replace github.com/99designs/keyring => github.com/cosmos/keyring v1.1.7-0.20210622111912-ef00f8ac3d76
