package types

// Oracle module event types
const (
	EventTypeExchangeRateUpdate = "exchange_rate_update"
	EventTypePrevote            = "prevote"
	EventTypeVote               = "vote"
	EventTypeFeedDelegate       = "feed_delegate"
	EventTypeAggregatePrevote   = "aggregate_prevote"
	EventTypeAggregateVote      = "aggregate_vote"

	EventAttrKeyDenom         = "denom"
	EventAttrKeyVoter         = "voter"
	EventAttrKeyExchangeRate  = "exchange_rate"
	EventAttrKeyExchangeRates = "exchange_rates"
	EventAttrKeyOperator      = "operator"
	EventAttrKeyFeeder        = "feeder"
	EventAttrValueCategory    = ModuleName
)
