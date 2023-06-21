package cw

import (
	"os"

	"gopkg.in/yaml.v3"
)

type ChainConfig struct {
	RPC       string            `yaml:"rpc"`
	GRPC      string            `yaml:"grpc"`
	API       string            `yaml:"api"`
	ChainID   string            `yaml:"chain_id"`
	Mnemonics map[string]string `yaml:"mnemonics"`
}

func ReadConfig(configFile string) (*ChainConfig, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	cc := ChainConfig{}
	err = yaml.Unmarshal(data, &cc)
	if err != nil {
		return nil, err
	}
	return &cc, nil
}
