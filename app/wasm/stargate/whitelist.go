package stargate

import (
	"fmt"
	"sync"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"

	ltypes "github.com/umee-network/umee/v5/x/leverage/types"
	otypes "github.com/umee-network/umee/v5/x/oracle/types"
	ugovtypes "github.com/umee-network/umee/v5/x/ugov"
	uibctypes "github.com/umee-network/umee/v5/x/uibc"
)

// stargateWhitelist keeps whitelist and its deterministic
// response binding for stargate queries.
//
// The query can be multi-thread, so we have to use
// thread safe sync.Map.
var stargateWhitelist sync.Map

// TODO: needs to finalize the which queries should allow
func init() {
	// ibc queries
	setWhitelistedQuery("/ibc.applications.transfer.v1.Query/DenomTrace", &ibctransfertypes.QueryDenomTraceResponse{})

	// cosmos-sdk queries

	// auth
	setWhitelistedQuery("/cosmos.auth.v1beta1.Query/Account", &authtypes.QueryAccountResponse{})
	setWhitelistedQuery("/cosmos.auth.v1beta1.Query/Params", &authtypes.QueryParamsResponse{})

	// bank
	setWhitelistedQuery("/cosmos.bank.v1beta1.Query/Balance", &banktypes.QueryBalanceResponse{})
	setWhitelistedQuery("/cosmos.bank.v1beta1.Query/DenomMetadata", &banktypes.QueryDenomsMetadataResponse{})
	setWhitelistedQuery("/cosmos.bank.v1beta1.Query/Params", &banktypes.QueryParamsResponse{})
	setWhitelistedQuery("/cosmos.bank.v1beta1.Query/SupplyOf", &banktypes.QuerySupplyOfResponse{})

	// distribution
	setWhitelistedQuery("/cosmos.distribution.v1beta1.Query/Params", &distributiontypes.QueryParamsResponse{})
	setWhitelistedQuery("/cosmos.distribution.v1beta1.Query/DelegatorWithdrawAddress",
		&distributiontypes.QueryDelegatorWithdrawAddressResponse{})
	setWhitelistedQuery("/cosmos.distribution.v1beta1.Query/ValidatorCommission",
		&distributiontypes.QueryValidatorCommissionResponse{})

	// gov
	setWhitelistedQuery("/cosmos.gov.v1beta1.Query/Deposit", &govtypes.QueryDepositResponse{})
	setWhitelistedQuery("/cosmos.gov.v1beta1.Query/Params", &govtypes.QueryParamsResponse{})
	setWhitelistedQuery("/cosmos.gov.v1beta1.Query/Vote", &govtypes.QueryVoteResponse{})

	// slashing
	setWhitelistedQuery("/cosmos.slashing.v1beta1.Query/Params", &slashingtypes.QueryParamsResponse{})
	setWhitelistedQuery("/cosmos.slashing.v1beta1.Query/SigningInfo", &slashingtypes.QuerySigningInfoResponse{})

	// staking
	setWhitelistedQuery("/cosmos.staking.v1beta1.Query/Delegation", &stakingtypes.QueryDelegationResponse{})
	setWhitelistedQuery("/cosmos.staking.v1beta1.Query/Params", &stakingtypes.QueryParamsResponse{})
	setWhitelistedQuery("/cosmos.staking.v1beta1.Query/Validator", &stakingtypes.QueryValidatorResponse{})

	// umee native module queries

	// ugov
	setWhitelistedQuery("/umee.ugov.v1.Query/MinGasPrice", &ugovtypes.QueryMinGasPriceResponse{})

	// leverage
	setWhitelistedQuery("/umee.leverage.v1.Query/Params", &ltypes.QueryParamsResponse{})
	setWhitelistedQuery("/umee.leverage.v1.Query/RegisteredTokens", &ltypes.QueryRegisteredTokensResponse{})
	setWhitelistedQuery("/umee.leverage.v1.Query/MarketSummary", &ltypes.QueryMarketSummaryResponse{})
	setWhitelistedQuery("/umee.leverage.v1.Query/AccountBalances", &ltypes.QueryAccountBalancesResponse{})
	setWhitelistedQuery("/umee.leverage.v1.Query/AccountSummary", &ltypes.QueryAccountSummaryResponse{})
	setWhitelistedQuery("/umee.leverage.v1.Query/LiquidationTargets", &ltypes.QueryLiquidationTargetsResponse{})
	setWhitelistedQuery("/umee.leverage.v1.Query/BadDebts", &ltypes.QueryBadDebtsResponse{})
	setWhitelistedQuery("/umee.leverage.v1.Query/MaxWithdraw", &ltypes.QueryMaxWithdrawResponse{})
	setWhitelistedQuery("/umee.leverage.v1.Query/MaxBorrow", &ltypes.QueryMaxBorrowResponse{})

	// oracle
	setWhitelistedQuery("/umee.oracle.v1.Query/ExchangeRates", &otypes.QueryExchangeRatesResponse{})
	setWhitelistedQuery("/umee.oracle.v1.Query/ActiveExchangeRates", &otypes.QueryActiveExchangeRatesResponse{})
	setWhitelistedQuery("/umee.oracle.v1.Query/FeederDelegation", &otypes.QueryFeederDelegationResponse{})
	setWhitelistedQuery("/umee.oracle.v1.Query/MissCounter", &otypes.QueryMissCounterResponse{})
	setWhitelistedQuery("/umee.oracle.v1.Query/SlashWindow", &otypes.QuerySlashWindowResponse{})
	setWhitelistedQuery("/umee.oracle.v1.Query/AggregatePrevote", &otypes.QueryAggregatePrevoteResponse{})
	setWhitelistedQuery("/umee.oracle.v1.Query/AggregatePrevotes", &otypes.QueryAggregatePrevotesResponse{})
	setWhitelistedQuery("/umee.oracle.v1.Query/AggregateVote", &otypes.QueryAggregateVoteResponse{})
	setWhitelistedQuery("/umee.oracle.v1.Query/AggregateVotes", &otypes.QueryAggregateVotesResponse{})
	setWhitelistedQuery("/umee.oracle.v1.Query/Params", &otypes.QueryParamsResponse{})
	setWhitelistedQuery("/umee.oracle.v1.Query/Medians", &otypes.QueryMediansResponse{})
	setWhitelistedQuery("/umee.oracle.v1.Query/MedianDeviations", &otypes.QueryMedianDeviationsResponse{})
	setWhitelistedQuery("/umee.oracle.v1.Query/AvgPrice", &otypes.QueryAvgPriceResponse{})

	// uibc
	setWhitelistedQuery("/umee.uibc.v1.Query/Params", &uibctypes.QueryParamsResponse{})
	setWhitelistedQuery("/umee.uibc.v1.Query/Outflows", &uibctypes.QueryOutflowsResponse{})
	setWhitelistedQuery("/umee.uibc.v1.Query/AllOutflows", &uibctypes.QueryAllOutflowsResponse{})
}

// GetWhitelistedQuery returns the whitelisted query at the provided path.
// If the query does not exist, or it was setup wrong by the chain, this returns an error.
func GetWhitelistedQuery(queryPath string) (codec.ProtoMarshaler, error) {
	protoResponseAny, isWhitelisted := stargateWhitelist.Load(queryPath)
	if !isWhitelisted {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("'%s' path is not allowed from the contract", queryPath)}
	}
	protoResponseType, ok := protoResponseAny.(codec.ProtoMarshaler)
	if !ok {
		return nil, wasmvmtypes.Unknown{}
	}
	return protoResponseType, nil
}

func setWhitelistedQuery(queryPath string, protoType codec.ProtoMarshaler) {
	stargateWhitelist.Store(queryPath, protoType)
}

func GetStargateWhitelistedPaths() (keys []string) {
	// Iterate over the map and collect the keys
	stargateWhitelist.Range(func(key, value interface{}) bool {
		keyStr, ok := key.(string)
		if !ok {
			panic("key is not a string")
		}
		keys = append(keys, keyStr)
		return true
	})

	return keys
}
