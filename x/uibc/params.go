package uibc

import (
	fmt "fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultParams returns default genesis params
func DefaultParams() Params {
	return Params{
		IbcStatus:                   IBCTransferStatus_IBC_TRANSFER_STATUS_QUOTA_ENABLED,
		TotalQuota:                  sdk.NewDec(1_600_000),      // $1.6M
		TokenQuota:                  sdk.NewDec(1_200_000),      // $1.2M
		QuotaDuration:               time.Second * 60 * 60 * 24, // 24h
		InflowOutflowQuotaBase:      sdk.NewDec(1_000_000),      // 1M
		InflowOutflowQuotaRate:      sdk.MustNewDecFromStr("0.25"),
		InflowOutflowTokenQuotaBase: sdk.NewDec(900_000), // $0.9M
	}
}

func (p Params) Validate() error {
	if err := validateIBCTransferStatus(p.IbcStatus); err != nil {
		return err
	}
	if err := validateQuotaDuration(p.QuotaDuration); err != nil {
		return err
	}
	if err := validateQuota(p.TotalQuota, "total quota"); err != nil {
		return err
	}
	if err := validateQuota(p.TokenQuota, "quota per token"); err != nil {
		return err
	}
	if err := validateQuota(p.InflowOutflowQuotaBase, "total inflow outflow quota base"); err != nil {
		return err
	}
	if err := validateQuotaRate(p.InflowOutflowQuotaRate, "total inflow outflow quota rate"); err != nil {
		return err
	}
	if err := validateQuota(p.InflowOutflowTokenQuotaBase, "total inflow outflow quota token base"); err != nil {
		return err
	}
	if p.TotalQuota.LT(p.TokenQuota) {
		return fmt.Errorf("token quota shouldn't be less than quota per denom")
	}

	return nil
}

func validateIBCTransferStatus(status IBCTransferStatus) error {
	if status == IBCTransferStatus_IBC_TRANSFER_STATUS_QUOTA_DISABLED ||
		status == IBCTransferStatus_IBC_TRANSFER_STATUS_QUOTA_ENABLED ||
		status == IBCTransferStatus_IBC_TRANSFER_STATUS_QUOTA_IN_DISABLED ||
		status == IBCTransferStatus_IBC_TRANSFER_STATUS_QUOTA_OUT_DISABLED ||
		status == IBCTransferStatus_IBC_TRANSFER_STATUS_TRANSFERS_PAUSED {
		return nil
	}

	return fmt.Errorf("invalid ibc-transfer status: %s", status.String())
}

func validateQuotaDuration(d time.Duration) error {
	if d <= 0 {
		return fmt.Errorf("quota duration time must be positive: %d", d)
	}

	return nil
}

func validateQuota(q sdk.Dec, typ string) error {
	if q.IsNil() || q.IsNegative() {
		return fmt.Errorf("%s must be not negative: %s", typ, q)
	}
	return nil
}

func validateQuotaRate(q sdk.Dec, typ string) error {
	if q.LT(sdk.ZeroDec()) || q.GT(sdk.NewDec(2)) {
		return fmt.Errorf("%s must be between 0 and 2: %s", typ, q)
	}

	return validateQuota(q, typ)
}

// IBCTransferEnabled returns true if the ibc-transfer is enabled for both inflow and outflow."
func (status IBCTransferStatus) IBCTransferEnabled() bool {
	return status != IBCTransferStatus_IBC_TRANSFER_STATUS_TRANSFERS_PAUSED
}

// InflowQuotaEnabled returns true if inflow quota check is enabled.
func (status IBCTransferStatus) InflowQuotaEnabled() bool {
	// outflow disabled means inflow check enabled
	switch status {
	case IBCTransferStatus_IBC_TRANSFER_STATUS_QUOTA_ENABLED, IBCTransferStatus_IBC_TRANSFER_STATUS_QUOTA_OUT_DISABLED:
		return true
	default:
		return false
	}
}

// OutflowQuotaEnabled returns true if outflow quota check is enabled.
func (status IBCTransferStatus) OutflowQuotaEnabled() bool {
	// inflow disabled means outflow check enabled
	switch status {
	case IBCTransferStatus_IBC_TRANSFER_STATUS_QUOTA_ENABLED, IBCTransferStatus_IBC_TRANSFER_STATUS_QUOTA_IN_DISABLED:
		return true
	default:
		return false
	}
}
