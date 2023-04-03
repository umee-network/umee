package keys

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestLeadingAddressAndDenom(t *testing.T) {
	testCases := []struct {
		msg       string
		prefix    []byte
		key       []byte
		expectErr string
		addr      []byte
		denom     string
		read      int
	}{
		{
			"typical case",
			[]byte{2, 2},
			[]byte{2, 2, 1, 4, 97, 98, 99, 0},
			"",
			[]byte{4},
			"abc",
			8,
		},
		{
			"trailing data",
			[]byte{2, 2},
			[]byte{2, 2, 1, 4, 97, 98, 99, 0, 1, 2, 3},
			"",
			[]byte{4},
			"abc",
			8,
		},
		{
			"no prefix",
			[]byte{},
			[]byte{1, 4, 97, 98, 99, 0},
			"",
			[]byte{4},
			"abc",
			6,
		},
		{
			"zero length address, correctly prefixed",
			[]byte{2, 2},
			[]byte{2, 2, 0, 97, 98, 99, 0},
			"",
			[]byte{},
			"abc",
			7,
		},
		{
			"prefix variation", // not a problem as long as length is the same
			[]byte{1, 1},
			[]byte{2, 2, 1, 4, 97, 98, 99, 0},
			"",
			[]byte{4},
			"abc",
			8,
		},
		{
			"empty denom, properly terminated",
			[]byte{2, 2},
			[]byte{2, 2, 1, 4, 0},
			"",
			[]byte{4},
			"",
			5,
		},
		{
			"key too short (empty)",
			[]byte{2, 2},
			[]byte{},
			"key too short",
			[]byte{},
			"",
			0,
		},
		{
			"key too short (prefix only)",
			[]byte{2, 2},
			[]byte{2, 2},
			"key too short",
			[]byte{},
			"",
			0,
		},
		{
			"key too short (no addr)",
			[]byte{2, 2},
			[]byte{2, 2, 1},
			"key too short",
			[]byte{},
			"",
			0,
		},
		{
			"key too short 4 (too short for addr len prefix)",
			[]byte{2, 2},
			[]byte{2, 2, 42, 4, 4, 4, 97, 98, 99, 0},
			"key too short",
			[]byte{},
			"",
			0,
		},
		{
			"non-terminated denom",
			[]byte{2, 2},
			[]byte{2, 2, 1, 4, 97, 98, 99},
			"did not find expected null terminator",
			[]byte{},
			"",
			0,
		},
	}

	for _, tc := range testCases {
		addr, denom, read, err := LeadingAddressAndDenom(tc.prefix, tc.key)
		if tc.expectErr != "" {
			assert.ErrorContains(t, err, tc.expectErr)
		} else {
			assert.Equal(t, denom, tc.denom, tc.msg)
			assert.Equal(t, read, tc.read, tc.msg)
			assert.DeepEqual(t, []byte(addr), tc.addr)
		}
	}
}

func TestLeadingAddress(t *testing.T) {
	testCases := []struct {
		msg       string
		prefix    []byte
		key       []byte
		expectErr string
		addr      []byte
		read      int
	}{
		{
			"typical case",
			[]byte{2, 2},
			[]byte{2, 2, 1, 4},
			"",
			[]byte{4},
			4,
		},
		{
			"trailing data",
			[]byte{2, 2},
			[]byte{2, 2, 1, 4, 97, 98, 99, 0, 1, 2, 3},
			"",
			[]byte{4},
			4,
		},
		{
			"no prefix",
			[]byte{},
			[]byte{1, 4},
			"",
			[]byte{4},
			2,
		},
		{
			"zero length address, correctly prefixed",
			[]byte{2, 2},
			[]byte{2, 2, 0, 97, 98, 99, 0},
			"",
			[]byte{},
			3,
		},
		{
			"prefix variation", // not a problem as long as length is the same
			[]byte{1, 1},
			[]byte{2, 2, 1, 4, 97, 98, 99, 0},
			"",
			[]byte{4},
			4,
		},
		{
			"key too short (empty)",
			[]byte{2, 2},
			[]byte{},
			"key too short",
			[]byte{},
			0,
		},
		{
			"key too short (prefix only)",
			[]byte{2, 2},
			[]byte{2, 2},
			"key too short",
			[]byte{},
			0,
		},
		{
			"key too short (no addr)",
			[]byte{2, 2},
			[]byte{2, 2, 1},
			"key too short",
			[]byte{},
			0,
		},
		{
			"key too short 4 (too short for addr len prefix)",
			[]byte{2, 2},
			[]byte{2, 2, 42, 4, 4, 4, 97, 98, 99, 0},
			"key too short",
			[]byte{},
			0,
		},
	}

	for _, tc := range testCases {
		addr, read, err := LeadingAddress(tc.prefix, tc.key)
		if tc.expectErr != "" {
			assert.ErrorContains(t, err, tc.expectErr)
		} else {
			assert.Equal(t, read, tc.read, tc.msg)
			assert.DeepEqual(t, []byte(addr), tc.addr)
		}
	}
}

func TestLeadingDenom(t *testing.T) {
	testCases := []struct {
		msg       string
		prefix    []byte
		key       []byte
		expectErr string
		denom     string
		read      int
	}{
		{
			"typical case",
			[]byte{2, 2},
			[]byte{2, 2, 97, 98, 99, 0},
			"",
			"abc",
			6,
		},
		{
			"trailing data",
			[]byte{2, 2},
			[]byte{2, 2, 97, 98, 99, 0, 1, 2, 3},
			"",
			"abc",
			6,
		},
		{
			"no prefix",
			[]byte{},
			[]byte{97, 98, 99, 0},
			"",
			"abc",
			4,
		},
		{
			"prefix variation", // not a problem as long as length is the same
			[]byte{1, 1},
			[]byte{2, 2, 97, 98, 99, 0},
			"",
			"abc",
			6,
		},
		{
			"empty denom, properly terminated",
			[]byte{2, 2},
			[]byte{2, 2, 0},
			"",
			"",
			3,
		},
		{
			"key too short (empty)",
			[]byte{2, 2},
			[]byte{},
			"key too short",
			"",
			0,
		},
		{
			"key too short (prefix only)",
			[]byte{2, 2},
			[]byte{2, 2},
			"key too short",
			"",
			0,
		},
		{
			"non-terminated denom",
			[]byte{2, 2},
			[]byte{2, 2, 97, 98, 99},
			"did not find expected null terminator",
			"",
			0,
		},
	}

	for _, tc := range testCases {
		denom, read, err := LeadingDenom(tc.prefix, tc.key)
		if tc.expectErr != "" {
			assert.ErrorContains(t, err, tc.expectErr)
		} else {
			assert.Equal(t, denom, tc.denom, tc.msg)
			assert.Equal(t, read, tc.read, tc.msg)
		}
	}
}
