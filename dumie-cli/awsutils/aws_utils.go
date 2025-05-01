/*
Copyright © 2025 Chanhyeok Seo chanhyeok.seo2@gmail.com
*/
package awsutils

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type AWSConfig struct {
	AccessKeyID     string `json:"aws_access_key_id"`
	SecretAccessKey string `json:"aws_secret_access_key"`
	Region          string `json:"aws_region"`
	KeyPairName     string `json:"key_pair_name"`
}

const configFilePath = "aws_config.json"

// LoadAWSConfig loads the AWS configuration from the config file
func LoadAWSConfig() (*AWSConfig, error) {
	file, err := os.Open(configFilePath)
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

func GetEC2AWSClient() (*ec2.Client, error) {
	cfgData, err := LoadAWSConfig()
	if err != nil {
		return nil, err
	}

	awsCfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfgData.AccessKeyID,
			cfgData.SecretAccessKey,
			"",
		)),
		config.WithRegion(cfgData.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	client := ec2.NewFromConfig(awsCfg)
	fmt.Println("Retrieved client:", client)
	return client, nil
}

func GetDynamoDBClient() (*dynamodb.Client, error) {
	cfgData, err := LoadAWSConfig()
	if err != nil {
		return nil, err
	}

	awsCfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfgData.AccessKeyID,
			cfgData.SecretAccessKey,
			"",
		)),
		config.WithRegion(cfgData.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	client := dynamodb.NewFromConfig(awsCfg)
	fmt.Println("Retrieved client:", client)
	return client, nil
}
