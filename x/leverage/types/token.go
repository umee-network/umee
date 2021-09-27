package types

const (
	// UTokenPrefix defines the uToken denomination prefix for all uToken types.
	UTokenPrefix = "u/"
)

// UTokenFromTokenDenom returns the uToken denom given a token denom.
func UTokenFromTokenDenom(tokenDenom string) string {
	return UTokenPrefix + tokenDenom
}
