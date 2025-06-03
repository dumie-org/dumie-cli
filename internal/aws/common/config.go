/*
Copyright © 2025 Chanhyeok Seo chanhyeok.seo2@gmail.com
*/
package common

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type AWSConfig struct {
	AccessKeyID     string `json:"aws_access_key_id"`
	SecretAccessKey string `json:"aws_secret_access_key"`
	Region          string `json:"aws_region"`
	KeyPairName     string `json:"key_pair_name"`
}

const (
	// MaxRetries is the maximum number of retries for AWS operations
	MaxRetries = 300

	// RetryDelay is the delay between retries
	RetryDelay = 1 * time.Second

	// StatusUpdateInterval is the interval for showing status updates
	StatusUpdateInterval = 5

	// ConfigFilePath is the path to the config file
	ConfigFilePath = "aws_config.json"
)

// LoadAWSConfig loads the AWS configuration from the config file
func LoadAWSConfig() (*AWSConfig, error) {
	file, err := os.Open(ConfigFilePath)
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %w", err)
	}
	defer file.Close()

	var config AWSConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("error decoding config file: %w", err)
	}

	return &config, nil
}
