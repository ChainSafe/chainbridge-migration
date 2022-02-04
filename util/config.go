package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// v1 Bridge configuration

type V1BridgeConfig struct {
	Chains       []RawChainConfig `json:"chains"`
	KeystorePath string           `json:"keystorePath,omitempty"`
}

type RawChainConfig struct {
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	Id       string            `json:"id"`       // ChainID
	Endpoint string            `json:"endpoint"` // url for rpc endpoint
	From     string            `json:"from"`     // address of key to use
	Opts     map[string]string `json:"opts"`
}

func (c *V1BridgeConfig) validate() error {
	for _, chain := range c.Chains {
		if chain.Type == "" {
			return fmt.Errorf("required field chain.Type empty for chain %s", chain.Id)
		}
		if chain.Endpoint == "" {
			return fmt.Errorf("required field chain.Endpoint empty for chain %s", chain.Id)
		}
		if chain.Name == "" {
			return fmt.Errorf("required field chain.Name empty for chain %s", chain.Id)
		}
		if chain.Id == "" {
			return fmt.Errorf("required field chain.Id empty for chain %s", chain.Id)
		}
		if chain.From == "" {
			return fmt.Errorf("required field chain.From empty for chain %s", chain.Id)
		}
	}
	return nil
}

func GetV1BridgeConfig(configPath string) (*V1BridgeConfig, error) {
	if configPath == "" {
		return nil, errors.New("")
	}

	var config V1BridgeConfig
	err := loadFile(configPath, &config)
	if err != nil {
		return nil, err
	}

	err = config.validate()
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// script configuration

type Config struct {
	ConfigurationPath string             `json:"configurationPath"`
	PrivateKeys       map[string]string  `json:"privateKeys"`
	StartingBlocks    map[string]string  `json:"startingBlocks"`
	Tokens            map[string][]Token `json:"tokens"`
	AutoPauseBridge   bool               `json:"autoPauseBridge"`
}

type Token struct {
	HandlerAddress  string `json:"handlerAddress"`
	TokenAddress    string `json:"tokenAddress"`
	Recipient       string `json:"recipient"`
	AmountOrTokenID string `json:"amountOrTokenID"`
	Type            string `json:"type"`
}

const DefaultConfigPath = "./configuration.json"

func GetConfig(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = DefaultConfigPath
	}

	var config Config
	err := loadFile(configPath, &config)
	if err != nil {
		return nil, err
	}

	if config.ConfigurationPath == "" {
		return nil, errors.New("require configuration path defined")
	}

	return &config, nil
}

func loadFile(file string, obj interface{}) error {
	ext := filepath.Ext(file)
	fp, err := filepath.Abs(file)
	if err != nil {
		return err
	}

	fmt.Printf("Loading configuration from path: %s\n", filepath.Clean(fp))

	f, err := os.Open(filepath.Clean(fp))
	if err != nil {
		return err
	}

	if ext == ".json" {
		if err = json.NewDecoder(f).Decode(&obj); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unrecognized extention: %s", ext)
	}

	return nil
}
