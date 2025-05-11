package ec2

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/ec2"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/dumie-org/dumie-cli/awsutils"
	"github.com/dumie-org/dumie-cli/awsutils/common"
	"github.com/dumie-org/dumie-cli/awsutils/ddb"
)

func RestoreOrCreateInstance(ctx context.Context, profile string) (string, error) {
	lockClient, err := awsutils.GetDynamoDBClient()
	if err != nil {
		return "", fmt.Errorf("failed to get DDB client: %w", err)
	}
	lock := ddb.NewDynamoDBLock(lockClient)
	fmt.Println("Acquiring deployment lock for profile:", profile)
	if err := lock.AcquireLock(ctx, profile); err != nil {
		return "", fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer func() {
		fmt.Println("Releasing deployment lock for profile:", profile)
		if err := lock.ReleaseLock(ctx, profile); err != nil {
			fmt.Println("Failed to release lock:", err)
		}
	}()

	client, err := awsutils.GetEC2AWSClient()
	if err != nil {
		return "", fmt.Errorf("failed to get EC2 client: %w", err)
	}

	existing, err := SearchEC2Instance(client, profile)
	if err != nil {
		return "", fmt.Errorf("error checking existing instance: %w", err)
	}
	if existing != nil {
		return "", fmt.Errorf("instance already exists with ID: %s", *existing)
	}

	// Try restore from snapshot
	instanceID, err := TryRestoreFromSnapshot(ctx, client, profile)
	if err != nil {
		return "", err
	}
	if instanceID != "" {
		fmt.Println("Restored instance from snapshot:", instanceID)
		return instanceID, nil
	}

	// Launch new instance
	return launchNewInstance(ctx, client, profile)
}

func launchNewInstance(ctx context.Context, client *ec2.Client, profile string) (string, error) {
	fmt.Println("No snapshot found. Launching fresh instance.")

	amiID, err := GetLatestAmazonLinuxAMI(client)
	if err != nil {
		return "", fmt.Errorf("failed to get AMI: %w", err)
	}

	sgID, err := CreateOrGetSecurityGroup(client, "dumie-default-sg")
	if err != nil {
		return "", fmt.Errorf("failed to get security group: %w", err)
	}

	keyName, err := common.GenerateKeyPair(client)
	if err != nil {
		return "", fmt.Errorf("failed to get key pair: %w", err)
	}

	instanceIDPtr, err := LaunchEC2Instance(client, profile, amiID, types.InstanceTypeT2Micro, sgID, keyName)
	if err != nil {
		return "", fmt.Errorf("failed to launch instance: %w", err)
	}

	return *instanceIDPtr, nil
}
