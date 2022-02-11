package config

import (
	"fmt"
	"os"
)

const (
	EnvVariableBackend = "PRICE_FEEDER_BACKEND"
	EnvVariableDir     = "PRICE_FEEDER_DIR"
	EnvVariablePass    = "PRICE_FEEDER_PASS"
)

// Keyring defines the required Umee keyring configuration.
type Keyring struct {
	Backend string
	Pass    string
	Dir     string
}

// NewKeyring parses the environment variables and returns a keyring object.
// If the necessary environment variables aren't set, return an error.
func NewKeyring() (Keyring, error) {
	keyring := Keyring{
		Backend: os.Getenv(EnvVariableBackend),
		Pass:    os.Getenv(EnvVariablePass),
		Dir:     os.Getenv(EnvVariableDir),
	}
	if keyring.Validate() != nil {
		keyring, err := keyring.GetStdInput()
		if err != nil {
			return Keyring{}, err
		}
		if keyring.Validate() != nil {
			return Keyring{}, fmt.Errorf("invalid values set for keyring")
		}
		return keyring, nil
	}
	return keyring, nil
}

// GetStdInput gets the keyring from console input
func (Keyring) GetStdInput() (Keyring, error) {
	keyring := Keyring{}
	fmt.Printf("Enter keyring backend: ")
	fmt.Scanf("%s", &keyring.Backend)
	fmt.Printf("Enter keyring directory: ")
	fmt.Scanf("%s", &keyring.Dir)
	fmt.Printf("Enter keyring password: ")
	fmt.Scanf("%s", &keyring.Pass)
	return keyring, nil
}

// Validate makes sure no empty strings were entered
func (k Keyring) Validate() error {
	if k.Backend == "" || k.Pass == "" || k.Dir == "" {
		return fmt.Errorf("keyring is invalid")
	}
	return nil
}
