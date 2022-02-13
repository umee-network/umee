package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/umee-network/umee/price-feeder/config"
)

func TestKeyring_Validate(t *testing.T) {
	os.Setenv(config.EnvVariableBackend, "test")
	os.Setenv(config.EnvVariableDir, "dir")
	os.Setenv(config.EnvVariablePass, "pass")

	envKeyring, err := config.InitKeyring()
	require.NoError(t, err)
	testCases := []struct {
		name      string
		keyring   config.Keyring
		expectErr bool
	}{
		{
			"valid config",
			config.Keyring{
				Backend: "test",
				Dir:     "/Users/username/.umee",
				Pass:    "keyringPass",
			},
			false,
		},
		{
			"empty",
			config.Keyring{},
			true,
		},
		{
			"partially empty",
			config.Keyring{
				Backend: "test",
				Dir:     "",
				Pass:    "keyringPass",
			},
			true,
		},
		{
			"using NewKeyring",
			config.NewKeyring(
				"backend",
				"dir",
				"pass",
			),
			false,
		},
		{
			"using environment variables",
			envKeyring,
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.keyring.Validate() != nil, tc.expectErr)
		})
	}
}
