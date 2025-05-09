package ec2

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type SnapshotManager struct {
	Client *ec2.Client
}

func NewSnapshotManagerFromClient(client *ec2.Client) *SnapshotManager {
	return &SnapshotManager{
		Client: client,
	}
}

func (s *SnapshotManager) CreateSnapshot(ctx context.Context, volumeID, instanceID string) (string, error) {
	input := &ec2.CreateSnapshotInput{
		VolumeId:    aws.String(volumeID),
		Description: aws.String(fmt.Sprintf("Snapshot before deleting instance %s", instanceID)),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeSnapshot,
				Tags: []types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(fmt.Sprintf("%s-snapshot", instanceID)),
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
