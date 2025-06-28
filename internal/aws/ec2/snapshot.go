package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/dumie-org/dumie-cli/internal/aws/common"
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

func TryRestoreFromSnapshot(ctx context.Context, client *ec2.Client, profile string, iamRoleARN *string) (string, error) {
	// Find Snapshot (tag:Name = profile)
	describeSnapshotsInput := &ec2.DescribeSnapshotsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:Name"),
				Values: []string{profile},
			},
			{
				Name:   aws.String("tag:ManagedBy"),
				Values: []string{"Dumie"},
			},
		},
		OwnerIds: []string{"self"},
	}

	result, err := client.DescribeSnapshots(ctx, describeSnapshotsInput)
	if err != nil {
		return "", fmt.Errorf("failed to search snapshot: %w", err)
	}

	if len(result.Snapshots) == 0 { // No matching snapshot found for profile
		return "", nil
	}

	snapshotID := *result.Snapshots[0].SnapshotId
	fmt.Println("Found snapshot for profile. Registering AMI from snapshot:", snapshotID)

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
	keyName, err := common.GetKeyPairName()
	if err != nil {
		return "", fmt.Errorf("error getting key pair name: %v", err)
	}

	// Launch EC2 Instance
	instanceIDPtr, err := LaunchEC2Instance(client, profile, amiID, types.InstanceTypeT2Micro, sgID, keyName, nil, iamRoleARN, true)
	if err != nil {
		return "", fmt.Errorf("failed to launch instance: %w", err)
	}

	return *instanceIDPtr, nil
}

func DeleteSnapshotAndAMIIfExists(ctx context.Context, client *ec2.Client, snapshotID string, profile string) error {
	// check AMI using the snapshot
	describeInput := &ec2.DescribeImagesInput{
		Owners: []string{"self"},
		Filters: []types.Filter{
			{
				Name:   aws.String("block-device-mapping.snapshot-id"),
				Values: []string{snapshotID},
			},
		},
	}

	images, err := client.DescribeImages(ctx, describeInput)
	if err != nil {
		return fmt.Errorf("failed to describe AMIs: %w", err)
	}

	// deregister ami
	for _, image := range images.Images {
		_, err := client.DeregisterImage(ctx, &ec2.DeregisterImageInput{
			ImageId: image.ImageId,
		})
		if err != nil {
			fmt.Printf("Warning: failed to deregister AMI [%s]: %v\n", *image.ImageId, err)
			return err
		}
		fmt.Printf("Deregistered AMI [%s] using snapshot [%s]\n", *image.ImageId, snapshotID)
	}

	// delete snapshot
	_, err = client.DeleteSnapshot(ctx, &ec2.DeleteSnapshotInput{
		SnapshotId: aws.String(snapshotID),
	})
	if err != nil {
		return fmt.Errorf("failed to delete snapshot [%s]: %w", snapshotID, err)
	}

	fmt.Printf("Deleted old snapshot [%s] for profile [%s]\n", snapshotID, profile)
	return nil
}

func DeleteOldSnapshotsByProfile(ctx context.Context, profile string) error {
	client, err := common.GetEC2AWSClient()
	if err != nil {
		return fmt.Errorf("failed to create EC2 client: %w", err)
	}

	input := &ec2.DescribeSnapshotsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:Name"),
				Values: []string{profile},
			},
			{
				Name:   aws.String("tag:ManagedBy"),
				Values: []string{"Dumie"},
			},
		},
		OwnerIds: []string{"self"},
	}

	result, err := client.DescribeSnapshots(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to describe snapshots: %w", err)
	}

	for _, snap := range result.Snapshots {
		err := DeleteSnapshotAndAMIIfExists(ctx, client, *snap.SnapshotId, profile)
		if err != nil {
			fmt.Printf("Warning: failed to delete snapshot [%s]: %v\n", *snap.SnapshotId, err)
		}
	}

	return nil
}
