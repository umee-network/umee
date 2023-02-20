package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v3"
)

var (
	oneDec           = sdk.OneDec()
	minVoteThreshold = sdk.NewDecWithPrec(33, 2) // 0.33
)

// maxium number of decimals allowed for VoteThreshold
const (
	MaxVoteThresholdPrecision  = 2
	MaxVoteThresholdMultiplier = 100 // must be 10^MaxVoteThresholdPrecision
)

// Parameter keys
var (
	KeyVotePeriod               = []byte("VotePeriod")
	KeyVoteThreshold            = []byte("VoteThreshold")
	KeyRewardBand               = []byte("RewardBand")
	KeyRewardDistributionWindow = []byte("RewardDistributionWindow")
	KeyAcceptList               = []byte("AcceptList")
	KeySlashFraction            = []byte("SlashFraction")
	KeySlashWindow              = []byte("SlashWindow")
	KeyMinValidPerWindow        = []byte("MinValidPerWindow")
	KeyHistoricStampPeriod      = []byte("HistoricStampPeriod")
	KeyMedianStampPeriod        = []byte("MedianStampPeriod")
	KeyMaximumPriceStamps       = []byte("MaximumPriceStamps")
	KeyMaximumMedianStamps      = []byte("MedianStampAmount")
)

// Default parameter values
const (
	DefaultVotePeriod               = BlocksPerMinute / 2 // 30 seconds
	DefaultSlashWindow              = BlocksPerWeek       // window for a week
	DefaultRewardDistributionWindow = BlocksPerYear       // window for a year
	DefaultHistoricStampPeriod      = BlocksPerMinute * 5 // window for 5 minutes
	DefaultMaximumPriceStamps       = 36                  // retain for 3 hours
	DefaultMedianStampPeriod        = BlocksPerHour * 3   // window for 3 hours
	DefaultMaximumMedianStamps      = 24                  // retain for 3 days
)

// Default parameter values
var (
	DefaultVoteThreshold = sdk.NewDecWithPrec(50, 2) // 50%
	DefaultRewardBand    = sdk.NewDecWithPrec(2, 2)  // 2% (-1, 1)
	DefaultAcceptList    = DenomList{
		{
			BaseDenom:   UmeeDenom,
			SymbolDenom: UmeeSymbol,
			Exponent:    UmeeExponent,
		},
		{
			BaseDenom:   AtomDenom,
			SymbolDenom: AtomSymbol,
			Exponent:    AtomExponent,
		},
	}
	DefaultSlashFraction     = sdk.NewDecWithPrec(1, 4) // 0.01%
	DefaultMinValidPerWindow = sdk.NewDecWithPrec(5, 2) // 5%
)

var _ paramstypes.ParamSet = &Params{}

// DefaultParams creates default oracle module parameters
func DefaultParams() Params {
	return Params{
		VotePeriod:               DefaultVotePeriod,
		VoteThreshold:            DefaultVoteThreshold,
		RewardBand:               DefaultRewardBand,
		RewardDistributionWindow: DefaultRewardDistributionWindow,
		AcceptList:               DefaultAcceptList,
		SlashFraction:            DefaultSlashFraction,
		SlashWindow:              DefaultSlashWindow,
		MinValidPerWindow:        DefaultMinValidPerWindow,
		HistoricStampPeriod:      DefaultHistoricStampPeriod,
		MedianStampPeriod:        DefaultMedianStampPeriod,
		MaximumPriceStamps:       DefaultMaximumPriceStamps,
		MaximumMedianStamps:      DefaultMaximumMedianStamps,
	}
}

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramstypes.KeyTable {
	return paramstypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of oracle module's parameters.
func (p *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(
			KeyVotePeriod,
			&p.VotePeriod,
			validateVotePeriod,
		),
		paramstypes.NewParamSetPair(
			KeyVoteThreshold,
			&p.VoteThreshold,
			validateVoteThreshold,
		),
		paramstypes.NewParamSetPair(
			KeyRewardBand,
			&p.RewardBand,
			validateRewardBand,
		),
		paramstypes.NewParamSetPair(
			KeyRewardDistributionWindow,
			&p.RewardDistributionWindow,
			validateRewardDistributionWindow,
		),
		paramstypes.NewParamSetPair(
			KeyAcceptList,
			&p.AcceptList,
			validateAcceptList,
		),
		paramstypes.NewParamSetPair(
			KeySlashFraction,
			&p.SlashFraction,
			validateSlashFraction,
		),
		paramstypes.NewParamSetPair(
			KeySlashWindow,
			&p.SlashWindow,
			validateSlashWindow,
		),
		paramstypes.NewParamSetPair(
			KeyMinValidPerWindow,
			&p.MinValidPerWindow,
			validateMinValidPerWindow,
		),
		paramstypes.NewParamSetPair(
			KeyHistoricStampPeriod,
			&p.HistoricStampPeriod,
			validateHistoricStampPeriod,
		),
		paramstypes.NewParamSetPair(
			KeyMedianStampPeriod,
			&p.MedianStampPeriod,
			validateMedianStampPeriod,
		),
		paramstypes.NewParamSetPair(
			KeyMaximumPriceStamps,
			&p.MaximumPriceStamps,
			validateMaximumPriceStamps,
		),
		paramstypes.NewParamSetPair(
			KeyMaximumMedianStamps,
			&p.MaximumMedianStamps,
			validateMaximumMedianStamps,
		),
	}
}

// String implements fmt.Stringer interface
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// Validate performs basic validation on oracle parameters.
func (p Params) Validate() error {
	if p.VotePeriod == 0 {
		return fmt.Errorf("oracle parameter VotePeriod must be > 0, is %d", p.VotePeriod)
	}
	if p.VoteThreshold.LTE(sdk.NewDecWithPrec(33, 2)) {
		return fmt.Errorf("oracle parameter VoteThreshold must be greater than 33 percent")
	}

	if p.RewardBand.GT(sdk.OneDec()) || p.RewardBand.IsNegative() {
		return fmt.Errorf("oracle parameter RewardBand must be between [0, 1]")
	}

	if p.RewardDistributionWindow < p.VotePeriod {
		return fmt.Errorf("oracle parameter RewardDistributionWindow must be greater than or equal with VotePeriod")
	}

	if p.SlashFraction.GT(sdk.OneDec()) || p.SlashFraction.IsNegative() {
		return fmt.Errorf("oracle parameter SlashFraction must be between [0, 1]")
	}

	if p.SlashWindow < p.VotePeriod {
		return fmt.Errorf("oracle parameter SlashWindow must be greater than or equal with VotePeriod")
	}

	if p.MinValidPerWindow.GT(sdk.OneDec()) || p.MinValidPerWindow.IsNegative() {
		return fmt.Errorf("oracle parameter MinValidPerWindow must be between [0, 1]")
	}

	if p.HistoricStampPeriod > p.MedianStampPeriod {
		return fmt.Errorf("oracle parameter MedianStampPeriod must be greater than or equal with HistoricStampPeriod")
	}

	if p.HistoricStampPeriod%p.VotePeriod != 0 || p.MedianStampPeriod%p.VotePeriod != 0 {
		return fmt.Errorf("oracle parameters HistoricStampPeriod and MedianStampPeriod must be exact multiples of VotePeriod")
	}

	for _, denom := range p.AcceptList {
		if len(denom.BaseDenom) == 0 {
			return fmt.Errorf("oracle parameter AcceptList Denom must have BaseDenom")
		}
		if len(denom.SymbolDenom) == 0 {
			return fmt.Errorf("oracle parameter AcceptList Denom must have SymbolDenom")
		}
	}

	return nil
}

func validateVotePeriod(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("vote period must be positive: %d", v)
	}

	return nil
}

func validateVoteThreshold(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return ValidateVoteThreshold(v)
}

func validateRewardBand(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("reward band must be positive: %s", v)
	}

	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("reward band is too large: %s", v)
	}

	return nil
}

func validateRewardDistributionWindow(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("reward distribution window must be positive: %d", v)
	}

	return nil
}

func validateAcceptList(i interface{}) error {
	v, ok := i.(DenomList)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	for _, d := range v {
		if len(d.BaseDenom) == 0 {
			return fmt.Errorf("oracle parameter AcceptList Denom must have BaseDenom")
		}
		if len(d.SymbolDenom) == 0 {
			return fmt.Errorf("oracle parameter AcceptList Denom must have SymbolDenom")
		}
	}

	return nil
}

func validateSlashFraction(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("slash fraction must be positive: %s", v)
	}

	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("slash fraction is too large: %s", v)
	}

	return nil
}

func validateSlashWindow(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("slash window must be positive: %d", v)
	}

	return nil
}

func validateMinValidPerWindow(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("min valid per window must be positive: %s", v)
	}

	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("min valid per window is too large: %s", v)
	}

	return nil
}

func validateHistoricStampPeriod(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v < 1 {
		return fmt.Errorf("historic stamp period must be positive: %d", v)
	}

	return nil
}

func validateMedianStampPeriod(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v < 1 {
		return fmt.Errorf("median stamp period must be positive: %d", v)
	}

	return nil
}

func validateMaximumPriceStamps(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v < 1 {
		return fmt.Errorf("maximum price stamps must be positive: %d", v)
	}

	return nil
}

func validateMaximumMedianStamps(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v < 1 {
		return fmt.Errorf("maximum median stamps must be positive: %d", v)
	}

	return nil
}

// ValidateVoteThreshold validates oracle exchange rates power vote threshold.
// Must be
// * a decimal value > 0.33 and <= 1.
// * max precision is 2 (so 0.501 is not allowed)
func ValidateVoteThreshold(x sdk.Dec) error {
	if x.LTE(minVoteThreshold) || x.GT(oneDec) {
		return sdkerrors.ErrInvalidRequest.Wrapf("threshold must be bigger than %s and <= 1", minVoteThreshold)
	}
	i := x.MulInt64(100).TruncateInt64()
	x2 := sdk.NewDecWithPrec(i, MaxVoteThresholdPrecision)
	if !x2.Equal(x) {
		return sdkerrors.ErrInvalidRequest.Wrap("threshold precision must be maximum 2 decimals")
	}
	return nil
}
