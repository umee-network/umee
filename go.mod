module github.com/umee-network/umee

go 1.17

require (
	github.com/cosmos/cosmos-sdk v0.44.1
	github.com/cosmos/go-bip39 v1.0.0
	github.com/cosmos/ibc-go v1.2.1
	github.com/ethereum/go-ethereum v1.10.9
	github.com/gogo/protobuf v1.3.3
	github.com/golangci/golangci-lint v1.42.1
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/ory/dockertest/v3 v3.8.0
	github.com/peggyjv/gravity-bridge/module v0.2.17
	github.com/spf13/cast v1.4.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.9.0
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/starport v0.18.0
	github.com/tendermint/tendermint v0.34.13
	github.com/tendermint/tm-db v0.6.4
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	google.golang.org/genproto v0.0.0-20210903162649-d08c68adba83
	google.golang.org/grpc v1.41.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace (
	github.com/99designs/keyring => github.com/cosmos/keyring v1.1.7-0.20210622111912-ef00f8ac3d76
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.2-alpha.regen.4
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
)
