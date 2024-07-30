package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Address            string `yaml:"address"`
	InsecureConnection bool   `yaml:"insecureConnection"`
	Token              string `yaml:"token"`

	Subscriptions struct {
		Transaction struct {
			Enable                      bool     `yaml:"enable"`
			TransactionsVote            bool     `yaml:"transactionsVote"`
			TransactionsFailed          bool     `yaml:"transactionsFailed"`
			TransactionsAccountsInclude []string `yaml:"transactionsAccountsInclude"`
		} `yaml:"transaction"`
	} `yaml:"subscriptions"`
}

func ReadConfig(filename string) Config {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Failed to unmarshal config file: %v", err)
	}

	return config
}
