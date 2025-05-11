package ec2

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/dumie-org/dumie-cli/awsutils/common"
)

type SnapshotManager struct {
	Client *ec2.Client
}

func NewSnapshotManagerFromClient(client *ec2.Client) *SnapshotManager {
	return &SnapshotManager{
		Client: client,
	}
}

func (s *SnapshotManager) CreateSnapshot(ctx context.Context, volumeID, instanceID, profile string) (string, error) {
	input := &ec2.CreateSnapshotInput{
		VolumeId:    aws.String(volumeID),
		Description: aws.String(fmt.Sprintf("Snapshot before deleting instance %s", instanceID)),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeSnapshot,
				Tags: []types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(profile),
					},
					{
						Key:   aws.String("InstanceID"),
						Value: aws.String(instanceID),
					},
					{
						Key:   aws.String("ManagedBy"),
						Value: aws.String("Dumie"),
					},
				},
			},
		},
	}

	result, err := s.Client.CreateSnapshot(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to create snapshot: %w", err)
	}

	return *result.SnapshotId, nil
}

func TryRestoreFromSnapshot(ctx context.Context, client *ec2.Client, profile string) (string, error) {
	// Find Snapshot (tag:Name = profile)
	describeSnapshotsInput := &ec2.DescribeSnapshotsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:Name"),
				Values: []string{profile},
			},
		},
		OwnerIds: []string{"self"},
	}

	result, err := client.DescribeSnapshots(ctx, describeSnapshotsInput)
	if err != nil {
		return "", fmt.Errorf("failed to search snapshot: %w", err)
	}

	if len(result.Snapshots) == 0 {
		return "", nil // snapshot 없음 → 새 인스턴스 생성으로 fallback
	}

	snapshotID := *result.Snapshots[0].SnapshotId
	fmt.Println("🪄 Found snapshot for profile. Registering AMI from snapshot:", snapshotID)

	// Register AMI
	amiID, err := RegisterAMIFromSnapshot(ctx, client, snapshotID)
	if err != nil {
		return "", fmt.Errorf("failed to register AMI from snapshot: %w", err)
	}

	// Get SecurityGroup
	sgID, err := CreateOrGetSecurityGroup(client, "dumie-default-sg")
	if err != nil {
		return "", fmt.Errorf("failed to get SG: %w", err)
	}

	// Get Pem Key
	keyName, err := common.GenerateKeyPair(client)
	if err != nil {
		return "", fmt.Errorf("Error getting key pair name: %v\n", err)
	}

	// Launch EC2 Instance
	instanceIDPtr, err := LaunchEC2Instance(client, profile, amiID, types.InstanceTypeT2Micro, sgID, keyName)
	if err != nil {
		return "", fmt.Errorf("failed to launch instance from snapshot: %w", err)
	}

	return *instanceIDPtr, nil
}
