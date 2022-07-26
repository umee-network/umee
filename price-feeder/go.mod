module github.com/umee-network/umee/price-feeder

go 1.18

require (
	github.com/cosmos/cosmos-sdk v0.46.0-rc3
	github.com/go-playground/validator/v10 v10.11.0
	github.com/golangci/golangci-lint v1.47.2
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.5.0
	github.com/justinas/alice v1.2.0
	github.com/mitchellh/mapstructure v1.5.0
	github.com/rs/cors v1.8.2
	github.com/rs/zerolog v1.27.0
	github.com/sirkon/goproxy v1.4.8
	github.com/spf13/cobra v1.5.0
	github.com/stretchr/testify v1.8.0
	github.com/tendermint/tendermint v0.34.20
	github.com/umee-network/umee/v2 v2.0.1
	golang.org/x/sync v0.0.0-20220513210516-0976fa681c29
	google.golang.org/grpc v1.48.0
	gopkg.in/yaml.v2 v2.4.0
)


replace (
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	github.com/umee-network/umee/v2 => ../
)
