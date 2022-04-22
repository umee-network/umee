package app_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/umee-network/umee/v2/app"
)

func TestVerifyAddressFormat(t *testing.T) {
	testCases := []struct {
		address        string
		addressBz      []byte
		expectedErr    error
		expectedErrMsg string
	}{
		{
			address:     "umeevaloper1zypqa76je7pxsdwkfah6mu9a583sju6xjettez",
			addressBz:   []byte{17, 2, 14, 251, 82, 207, 130, 104, 53, 214, 79, 111, 173, 240, 189, 161, 227, 9, 115, 70},
			expectedErr: nil,
		},
		{
			address:     "umee1zypqa76je7pxsdwkfah6mu9a583sju6xjavygg",
			addressBz:   []byte{17, 2, 14, 251, 82, 207, 130, 104, 53, 214, 79, 111, 173, 240, 189, 161, 227, 9, 115, 70},
			expectedErr: nil,
		},
		{
			address:     "umee173upq2qjwrfxdvt8k8zftqljn6uyjd4zk4tw60",
			addressBz:   []byte{17, 2, 14, 251, 82, 207, 130, 104, 53, 214, 79, 111, 173, 240, 189, 161, 227, 9, 115, 70},
			expectedErr: nil,
		},
		{
			address:     "umee1xhcxq4fvxth2hn3msmkpftkfpw73um7s4et3lh4r8cfmumk3qsmsj30e5c",
			addressBz:   []byte{17, 2, 14, 251, 82, 207, 130, 104, 53, 214, 79, 111, 173, 240, 189, 161, 227, 9, 115, 70},
			expectedErr: nil,
		},
		{
			address:     "cosmos14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s4hmalr",
			addressBz:   []byte{173, 228, 165, 245, 128, 58, 67, 152, 53, 198, 54, 57, 90, 141, 100, 141, 238, 87, 178, 252, 144, 217, 141, 193, 127, 168, 135, 21, 155, 105, 99, 139},
			expectedErr: nil,
		},
		{
			address:        "umee1xhcxq4fvxth2hn3msmkpftkfpw7_invalidaddrs_dask3qsmsj30e5c",
			addressBz:      []byte{173, 228, 165, 245, 128, 58, 67, 152, 53, 198, 54, 57, 90, 141, 100, 141, 238, 87, 178, 252, 144, 217, 141, 193, 127, 168, 135, 21, 155, 105, 99, 139, 43},
			expectedErr:    errors.New("any"),
			expectedErrMsg: "invalid address length; got: 33, should be: 20 or 32 for cosmwasm contract addr",
		},
		{
			address:        "umee1xhcxq4fvxth2hn3msmkpftkfpw73_invalidaddrs_j30e5c123",
			addressBz:      []byte{17, 2, 14, 251, 82, 207, 130, 104, 53, 214, 79, 111, 173, 240, 189, 161, 227, 9, 115, 70, 84},
			expectedErr:    errors.New("any"),
			expectedErrMsg: "invalid address length; got: 21, should be: 20 or 32 for cosmwasm contract addr",
		},
	}

	for _, testCase := range testCases {
		actualErr := app.VerifyAddressFormat(testCase.addressBz)
		failedMsg := fmt.Sprintf("The address: %s failed the verification", testCase.address)

		if testCase.expectedErr == nil {
			require.Nil(t, actualErr, failedMsg)
			continue
		}
		require.ErrorContains(t, actualErr, testCase.expectedErrMsg, failedMsg)
	}
}
