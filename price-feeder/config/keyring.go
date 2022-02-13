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

// NewKeyring returns a new Keyring object
func NewKeyring(backend string, dir string, pass string) Keyring {
	return Keyring{
		Backend: backend,
		Dir:     dir,
		Pass:    pass,
	}
}

// Validate makes sure no empty strings were entered
func (k Keyring) Validate() error {
	if k.Backend == "" || k.Pass == "" || k.Dir == "" {
		return fmt.Errorf("keyring is invalid")
	}
	return nil
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

// InitKeyring parses the environment variables and returns a keyring object.
// If the necessary environment variables aren't set, get them from console input.
func InitKeyring() (Keyring, error) {
	keyring := NewKeyring(
		os.Getenv(EnvVariableBackend),
		os.Getenv(EnvVariableDir),
		os.Getenv(EnvVariablePass),
	)

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
