package types

//nolint
const (
	MicroLunaDenom  = "uluna"
	MicroUSDDenom   = "uusd"
	MicroKRWDenom   = "ukrw"
	MicroSDRDenom   = "usdr"
	MicroCNYDenom   = "ucny"
	MicroJPYDenom   = "ujpy"
	MicroEURDenom   = "ueur"
	MicroGBPDenom   = "ugbp"
	MicroMNTDenom   = "umnt"
	BlocksPerMinute = uint64(10)
	BlocksPerHour   = BlocksPerMinute * 60
	BlocksPerDay    = BlocksPerHour * 24
	BlocksPerWeek   = BlocksPerDay * 7
	BlocksPerMonth  = BlocksPerDay * 30
	BlocksPerYear   = BlocksPerDay * 365

	MicroUnit = int64(1e6)
)
