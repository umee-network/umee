module github.com/umee-network/umee

go 1.16

require (
	github.com/cosmos/cosmos-sdk v0.42.4-0.20210623214207-eb0fc466c99b
	github.com/gogo/protobuf v1.3.3
	github.com/golangci/golangci-lint v1.41.1
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/peggyjv/gravity-bridge/module v0.0.0-20210629194415-c9d37207deff
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/tendermint/tendermint v0.34.11
	github.com/tendermint/tm-db v0.6.4
	google.golang.org/genproto v0.0.0-20210114201628-6edceaf6022f
	google.golang.org/grpc v1.37.0
)

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.2-alpha.regen.4

replace github.com/gogo/grpc => google.golang.org/grpc v1.33.2
