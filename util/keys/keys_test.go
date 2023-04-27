package keys

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestExtractAddressAndString(t *testing.T) {
	testCases := []struct {
		name       string
		startIndex int
		key        []byte
		expectErr  string
		addr       []byte
		s          string
		nextIndex  int
	}{
		{
			"typical case",
			2,
			[]byte{2, 2, 1, 4, 97, 98, 99, 0},
			"",
			[]byte{4},
			"abc",
			8,
		},
		{
			"trailing data",
			2,
			[]byte{2, 2, 1, 4, 97, 98, 99, 0, 1, 2, 3},
			"",
			[]byte{4},
			"abc",
			8,
		},
		{
			"no prefix",
			0,
			[]byte{1, 4, 97, 98, 99, 0},
			"",
			[]byte{4},
			"abc",
			6,
		},
		{
			"zero length address, correctly prefixed",
			2,
			[]byte{2, 2, 0, 97, 98, 99, 0},
			"",
			[]byte{},
			"abc",
			7,
		},
		{
			"empty string, properly terminated",
			2,
			[]byte{2, 2, 1, 4, 0},
			"key too short to contain non-empty null terminated string after startIndex",
			[]byte{4},
			"",
			0,
		},
		{
			"null terminator before string",
			2,
			[]byte{2, 2, 1, 4, 0, 0},
			"empty null-terminated string not allowed in key",
			[]byte{},
			"",
			0,
		},
		{
			"key too short (empty)",
			2,
			[]byte{},
			"key too short",
			[]byte{},
			"",
			0,
		},
		{
			"key too short (prefix only)",
			2,
			[]byte{2, 2},
			"key too short",
			[]byte{},
			"",
			0,
		},
		{
			"key too short (no addr)",
			2,
			[]byte{2, 2, 1},
			"key too short",
			[]byte{},
			"",
			0,
		},
		{
			"key too short (big addr len prefix)",
			2,
			[]byte{2, 2, 42, 4, 4, 4, 97, 98, 99, 0},
			"key too short",
			[]byte{},
			"",
			0,
		},
		{
			"non-terminated string",
			2,
			[]byte{2, 2, 1, 4, 97, 98, 99},
			"null terminator not found",
			[]byte{},
			"",
			0,
		},
	}

	for _, tc := range testCases {
		addr, denom, read, err := ExtractAddressAndString(tc.startIndex, tc.key)
		if tc.expectErr != "" {
			assert.ErrorContains(t, err, tc.expectErr)
		} else {
			assert.Equal(t, denom, tc.s, tc.name)
			assert.Equal(t, read, tc.nextIndex, tc.name)
			assert.DeepEqual(t, []byte(addr), tc.addr)
		}
	}
}

func TestExtractAddress(t *testing.T) {
	testCases := []struct {
		name       string
		startIndex int
		key        []byte
		expectErr  string
		addr       []byte
		nextIndex  int
	}{
		{
			"typical case",
			2,
			[]byte{2, 2, 1, 4},
			"",
			[]byte{4},
			4,
		},
		{
			"trailing data",
			2,
			[]byte{2, 2, 1, 4, 97, 98, 99, 0, 1, 2, 3},
			"",
			[]byte{4},
			4,
		},
		{
			"no prefix",
			0,
			[]byte{1, 4},
			"",
			[]byte{4},
			2,
		},
		{
			"zero length address, correctly prefixed",
			2,
			[]byte{2, 2, 0, 97, 98, 99, 0},
			"",
			[]byte{},
			3,
		},
		{
			"key too short (empty)",
			2,
			[]byte{},
			"key too short",
			[]byte{},
			0,
		},
		{
			"key too short (prefix only)",
			2,
			[]byte{2, 2},
			"key too short",
			[]byte{},
			0,
		},
		{
			"key too short (no addr)",
			2,
			[]byte{2, 2, 1},
			"key too short",
			[]byte{},
			0,
		},
		{
			"key too short (big addr len prefix)",
			2,
			[]byte{2, 2, 42, 4, 4, 4, 97, 98, 99, 0},
			"key too short",
			[]byte{},
			0,
		},
	}

	for _, tc := range testCases {
		addr, read, err := ExtractAddress(tc.startIndex, tc.key)
		if tc.expectErr != "" {
			assert.ErrorContains(t, err, tc.expectErr)
		} else {
			assert.Equal(t, read, tc.nextIndex, tc.name)
			assert.DeepEqual(t, []byte(addr), tc.addr)
		}
	}
}

func TestExtractString(t *testing.T) {
	testCases := []struct {
		name       string
		startIndex int
		key        []byte
		expectErr  string
		s          string
		nextIndex  int
	}{
		{
			"typical case",
			2,
			[]byte{2, 2, 97, 98, 99, 0},
			"",
			"abc",
			6,
		},
		{
			"trailing data",
			2,
			[]byte{2, 2, 97, 98, 99, 0, 1, 2, 3},
			"",
			"abc",
			6,
		},
		{
			"no prefix",
			0,
			[]byte{97, 98, 99, 0},
			"",
			"abc",
			4,
		},
		{
			"empty string, properly terminated",
			2,
			[]byte{2, 2, 0},
			"key too short to contain non-empty null terminated string after startIndex",
			"",
			0,
		},
		{
			"null terminator before string",
			2,
			[]byte{2, 2, 0, 9, 9},
			"empty null-terminated string not allowed in key",
			"",
			0,
		},
		{
			"key too short (empty)",
			2,
			[]byte{},
			"key too short",
			"",
			0,
		},
		{
			"key too short (prefix only)",
			2,
			[]byte{2, 2},
			"key too short",
			"",
			0,
		},
		{
			"non-terminated string",
			2,
			[]byte{2, 2, 97, 98, 99},
			"null terminator not found",
			"",
			0,
		},
	}

	for _, tc := range testCases {
		denom, read, err := ExtractString(tc.startIndex, tc.key)
		if tc.expectErr != "" {
			assert.ErrorContains(t, err, tc.expectErr)
		} else {
			assert.Equal(t, denom, tc.s, tc.name)
			assert.Equal(t, read, tc.nextIndex, tc.name)
		}
	}
}
