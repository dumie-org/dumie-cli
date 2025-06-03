/*
Copyright © 2025 Chanhyeok Seo chanhyeok.seo2@gmail.com
*/
package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/dumie-org/dumie-cli/internal/aws/common"
)

func GetEC2AWSClient() (*ec2.Client, error) {
	cfgData, err := common.LoadAWSConfig()
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
	cfgData, err := common.LoadAWSConfig()
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
