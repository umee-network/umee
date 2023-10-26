package query

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
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	"github.com/umee-network/umee/v6/x/incentive"
	ltypes "github.com/umee-network/umee/v6/x/leverage/types"
	"github.com/umee-network/umee/v6/x/metoken"
	otypes "github.com/umee-network/umee/v6/x/oracle/types"
	ugovtypes "github.com/umee-network/umee/v6/x/ugov"
	uibctypes "github.com/umee-network/umee/v6/x/uibc"
)

// stargateWhitelist keeps whitelist and its deterministic
// response binding for stargate queries.
//
// The query can be multi-thread, so we have to use
// thread safe sync.Map.
var stargateWhitelist sync.Map

const (
	ibcBaseQueryPath      = "/ibc.applications.transfer.v1.Query/"
	authBaseQueryPath     = "/cosmos.auth.v1beta1.Query/"
	bankBaseQueryPath     = "/cosmos.bank.v1beta1.Query/"
	distrBaseQueryPath    = "/cosmos.distribution.v1beta1.Query/"
	govBaseQueryPath      = "/cosmos.gov.v1beta1.Query/"
	slashingBaseQueryPath = "/cosmos.slashing.v1beta1.Query/"
	stakingQueryPath      = "/cosmos.staking.v1beta1.Query/"

	// umee
	ugovBaseQueryPath      = "/umee.ugov.v1.Query/M"
	leverageBaseQueryPath  = "/umee.leverage.v1.Query/"
	oracleBaseQueryPath    = "/umee.oracle.v1.Query/"
	uibcBaseQueryPath      = "/umee.uibc.v1.Query/"
	incentiveBaseQueryPath = "/umee.incentive.v1.Query/"
	metokenBaseQueryPath   = "/umee.metoken.v1.Query/" // #nosec G101
)

func init() {
	// ibc queries
	setWhitelistedQuery(ibcBaseQueryPath+"DenomTrace", &ibctransfertypes.QueryDenomTraceResponse{})

	// cosmos-sdk queries

	// auth
	setWhitelistedQuery(authBaseQueryPath+"Account", &authtypes.QueryAccountResponse{})
	setWhitelistedQuery(authBaseQueryPath+"Params", &authtypes.QueryParamsResponse{})

	// bank
	setWhitelistedQuery(bankBaseQueryPath+"Balance", &banktypes.QueryBalanceResponse{})
	setWhitelistedQuery(bankBaseQueryPath+"DenomMetadata", &banktypes.QueryDenomsMetadataResponse{})
	setWhitelistedQuery(bankBaseQueryPath+"Params", &banktypes.QueryParamsResponse{})
	setWhitelistedQuery(bankBaseQueryPath+"SupplyOf", &banktypes.QuerySupplyOfResponse{})

	// distribution
	setWhitelistedQuery(distrBaseQueryPath+"Params", &distributiontypes.QueryParamsResponse{})
	setWhitelistedQuery(distrBaseQueryPath+"DelegatorWithdrawAddress",
		&distributiontypes.QueryDelegatorWithdrawAddressResponse{})
	setWhitelistedQuery(distrBaseQueryPath+"ValidatorCommission",
		&distributiontypes.QueryValidatorCommissionResponse{})

	// gov
	setWhitelistedQuery(govBaseQueryPath+"Deposit", &govtypes.QueryDepositResponse{})
	setWhitelistedQuery(govBaseQueryPath+"Params", &govtypes.QueryParamsResponse{})
	setWhitelistedQuery(govBaseQueryPath+"Vote", &govtypes.QueryVoteResponse{})

	// slashing
	setWhitelistedQuery(slashingBaseQueryPath+"Params", &slashingtypes.QueryParamsResponse{})
	setWhitelistedQuery(slashingBaseQueryPath+"SigningInfo", &slashingtypes.QuerySigningInfoResponse{})

	// staking
	setWhitelistedQuery(stakingQueryPath+"Delegation", &stakingtypes.QueryDelegationResponse{})
	setWhitelistedQuery(stakingQueryPath+"Params", &stakingtypes.QueryParamsResponse{})
	setWhitelistedQuery(stakingQueryPath+"Validator", &stakingtypes.QueryValidatorResponse{})

	// umee native module queries

	// ugov
	setWhitelistedQuery(ugovBaseQueryPath+"MinGasPrice", &ugovtypes.QueryMinGasPriceResponse{})

	// leverage
	setWhitelistedQuery(leverageBaseQueryPath+"Params", &ltypes.QueryParamsResponse{})
	setWhitelistedQuery(leverageBaseQueryPath+"RegisteredTokens", &ltypes.QueryRegisteredTokensResponse{})
	setWhitelistedQuery(leverageBaseQueryPath+"MarketSummary", &ltypes.QueryMarketSummaryResponse{})
	setWhitelistedQuery(leverageBaseQueryPath+"AccountBalances", &ltypes.QueryAccountBalancesResponse{})
	setWhitelistedQuery(leverageBaseQueryPath+"AccountSummary", &ltypes.QueryAccountSummaryResponse{})
	setWhitelistedQuery(leverageBaseQueryPath+"LiquidationTargets", &ltypes.QueryLiquidationTargetsResponse{})
	setWhitelistedQuery(leverageBaseQueryPath+"BadDebts", &ltypes.QueryBadDebtsResponse{})
	setWhitelistedQuery(leverageBaseQueryPath+"MaxWithdraw", &ltypes.QueryMaxWithdrawResponse{})
	setWhitelistedQuery(leverageBaseQueryPath+"MaxBorrow", &ltypes.QueryMaxBorrowResponse{})

	// oracle
	setWhitelistedQuery(oracleBaseQueryPath+"ExchangeRates", &otypes.QueryExchangeRatesResponse{})
	setWhitelistedQuery(oracleBaseQueryPath+"ActiveExchangeRates", &otypes.QueryActiveExchangeRatesResponse{})
	setWhitelistedQuery(oracleBaseQueryPath+"FeederDelegation", &otypes.QueryFeederDelegationResponse{})
	setWhitelistedQuery(oracleBaseQueryPath+"MissCounter", &otypes.QueryMissCounterResponse{})
	setWhitelistedQuery(oracleBaseQueryPath+"SlashWindow", &otypes.QuerySlashWindowResponse{})
	setWhitelistedQuery(oracleBaseQueryPath+"AggregatePrevote", &otypes.QueryAggregatePrevoteResponse{})
	setWhitelistedQuery(oracleBaseQueryPath+"AggregatePrevotes", &otypes.QueryAggregatePrevotesResponse{})
	setWhitelistedQuery(oracleBaseQueryPath+"AggregateVote", &otypes.QueryAggregateVoteResponse{})
	setWhitelistedQuery(oracleBaseQueryPath+"AggregateVotes", &otypes.QueryAggregateVotesResponse{})
	setWhitelistedQuery(oracleBaseQueryPath+"Params", &otypes.QueryParamsResponse{})
	setWhitelistedQuery(oracleBaseQueryPath+"Medians", &otypes.QueryMediansResponse{})
	setWhitelistedQuery(oracleBaseQueryPath+"MedianDeviations", &otypes.QueryMedianDeviationsResponse{})
	setWhitelistedQuery(oracleBaseQueryPath+"AvgPrice", &otypes.QueryAvgPriceResponse{})

	// uibc
	setWhitelistedQuery(uibcBaseQueryPath+"Params", &uibctypes.QueryParamsResponse{})
	setWhitelistedQuery(uibcBaseQueryPath+"Outflows", &uibctypes.QueryOutflowsResponse{})
	setWhitelistedQuery(uibcBaseQueryPath+"AllOutflows", &uibctypes.QueryAllOutflowsResponse{})

	// incentive
	setWhitelistedQuery(incentiveBaseQueryPath+"Params", &incentive.QueryParamsResponse{})
	setWhitelistedQuery(incentiveBaseQueryPath+"TotalBonded", &incentive.QueryTotalBondedResponse{})
	setWhitelistedQuery(incentiveBaseQueryPath+"TotalUnbonding", &incentive.QueryTotalUnbondingResponse{})
	setWhitelistedQuery(incentiveBaseQueryPath+"AccountBonds", &incentive.QueryAccountBondsResponse{})
	setWhitelistedQuery(incentiveBaseQueryPath+"PendingRewards", &incentive.QueryPendingRewardsResponse{})
	setWhitelistedQuery(incentiveBaseQueryPath+"CompletedIncentivePrograms",
		&incentive.QueryCompletedIncentiveProgramsResponse{})
	setWhitelistedQuery(incentiveBaseQueryPath+"OngoingIncentivePrograms",
		&incentive.QueryOngoingIncentiveProgramsResponse{})
	setWhitelistedQuery(incentiveBaseQueryPath+"UpcomingIncentivePrograms",
		&incentive.QueryUpcomingIncentiveProgramsResponse{})
	setWhitelistedQuery(incentiveBaseQueryPath+"IncentiveProgram", &incentive.QueryIncentiveProgramResponse{})
	setWhitelistedQuery(incentiveBaseQueryPath+"CurrentRates", &incentive.QueryCurrentRatesResponse{})
	setWhitelistedQuery(incentiveBaseQueryPath+"ActualRates", &incentive.QueryActualRates{})
	setWhitelistedQuery(incentiveBaseQueryPath+"LastRewardTime", &incentive.QueryLastRewardTimeResponse{})

	// metoken
	setWhitelistedQuery(metokenBaseQueryPath+"Params", &metoken.QueryParamsResponse{})
	setWhitelistedQuery(metokenBaseQueryPath+"Indexes", &metoken.QueryIndexesResponse{})
	setWhitelistedQuery(metokenBaseQueryPath+"SwapFee", &metoken.QuerySwapFeeResponse{})
	setWhitelistedQuery(metokenBaseQueryPath+"RedeemFee", &metoken.QueryRedeemFeeResponse{})
	setWhitelistedQuery(metokenBaseQueryPath+"IndexBalances", &metoken.QueryIndexBalancesResponse{})
	setWhitelistedQuery(metokenBaseQueryPath+"IndexPrices", &metoken.QueryIndexPricesResponse{})
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
