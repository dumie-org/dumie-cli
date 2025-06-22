/*
Copyright © 2025 Chanhyeok Seo chanhyeok.seo2@gmail.com
*/
package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/dumie-org/dumie-cli/internal/aws/common"
)

func GetDefaultVPCID(client *ec2.Client) (*string, error) {
	describeVPCsInput := &ec2.DescribeVpcsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("isDefault"),
				Values: []string{"true"},
			},
		},
	}
	describeVPCsOutput, err := client.DescribeVpcs(context.TODO(), describeVPCsInput)
	if err != nil {
		return nil, fmt.Errorf("error describing VPCs: %w", err)
	}
	if len(describeVPCsOutput.Vpcs) == 0 {
		// TODO: In this case, Dumie should ask if the user wants to create a new default VPC and create if needed.
		return nil, fmt.Errorf("no default VPC found")
	}
	return describeVPCsOutput.Vpcs[0].VpcId, nil
}

func CreateOrGetSecurityGroup(client *ec2.Client, groupName string) (*string, error) {
	describeSGInput := &ec2.DescribeSecurityGroupsInput{
		GroupNames: []string{groupName},
	}
	describeSGOutput, err := client.DescribeSecurityGroups(context.TODO(), describeSGInput)
	if err == nil && len(describeSGOutput.SecurityGroups) > 0 {
		return describeSGOutput.SecurityGroups[0].GroupId, nil
	}

	vpcID, err := GetDefaultVPCID(client)
	if err != nil {
		return nil, fmt.Errorf("error getting default VPC ID: %w", err)
	}

	createSGInput := &ec2.CreateSecurityGroupInput{
		GroupName:   &groupName,
		Description: aws.String("Security Group managed by Dumie"),
		VpcId:       vpcID,
	}
	createSGOutput, err := client.CreateSecurityGroup(context.TODO(), createSGInput)
	if err != nil {
		return nil, fmt.Errorf("error creating Security Group: %w", err)
	}

	authorizeSGInput := &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: createSGOutput.GroupId,
		IpPermissions: []types.IpPermission{
			{
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int32(22),
				ToPort:     aws.Int32(22),
				IpRanges: []types.IpRange{
					{
						// TODO: This should be restricted to the user's IP address only *IMPORTANT
						CidrIp: aws.String("0.0.0.0/0"),
					},
				},
			},
		},
	}
	_, err = client.AuthorizeSecurityGroupIngress(context.TODO(), authorizeSGInput)
	if err != nil {
		return nil, fmt.Errorf("error authorizing Security Group ingress: %w", err)
	}

	return createSGOutput.GroupId, nil
}

func SearchEC2Instance(client *ec2.Client, profile string) (*string, error) {
	describeInstancesInput := &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:Name"),
				Values: []string{profile},
			},
			{
				Name:   aws.String("instance-state-name"),
				Values: []string{"running", "pending"},
			},
		},
	}
	describeInstancesOutput, err := client.DescribeInstances(context.TODO(), describeInstancesInput)
	if err != nil {
		return nil, fmt.Errorf("error describing instances: %w", err)
	}

	if len(describeInstancesOutput.Reservations) == 0 {
		return nil, nil
	}

	return describeInstancesOutput.Reservations[0].Instances[0].InstanceId, nil
}

func LaunchEC2Instance(client *ec2.Client, profile string, amiID string, instanceType types.InstanceType, sgID *string, keyName string) (*string, error) {
	tags := []types.TagSpecification{
		{
			ResourceType: types.ResourceTypeInstance,
			Tags: []types.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String(profile),
				},
				{
					Key:   aws.String("ManagedBy"),
					Value: aws.String("Dumie"),
				},
			},
		},
	}
	runInstancesInput := &ec2.RunInstancesInput{
		TagSpecifications: tags,
		ImageId:           &amiID,
		InstanceType:      instanceType,
		MinCount:          aws.Int32(1),
		MaxCount:          aws.Int32(1),
		SecurityGroupIds: []string{
			*sgID,
		},
		KeyName: aws.String(keyName),
	}
	runInstancesOutput, err := client.RunInstances(context.TODO(), runInstancesInput)
	if err != nil {
		return nil, fmt.Errorf("error running instances: %w", err)
	}

	instanceID := runInstancesOutput.Instances[0].InstanceId

	err = waitForInstanceRunning(context.TODO(), client, *instanceID)
	if err != nil {
		return nil, fmt.Errorf("error waiting for instance: %w", err)
	}

	return instanceID, nil
}

func waitForInstanceRunning(ctx context.Context, client *ec2.Client, instanceID string) error {
	checker := NewEC2StatusChecker(client, instanceID)
	return common.WaitForResourceStatus(ctx, checker)
}

func GetLatestAmazonLinuxAMI(client *ec2.Client) (string, error) {
	describeImagesInput := &ec2.DescribeImagesInput{
		Owners: []string{"amazon"},
		Filters: []types.Filter{
			{
				Name:   aws.String("name"),
				Values: []string{"amzn2-ami-hvm-*-x86_64-gp2"},
			},
			{
				Name:   aws.String("state"),
				Values: []string{"available"},
			},
		},
	}
	describeImagesOutput, err := client.DescribeImages(context.TODO(), describeImagesInput)
	if err != nil {
		return "", fmt.Errorf("error describing images: %w", err)
	}

	if len(describeImagesOutput.Images) == 0 {
		return "", fmt.Errorf("no Amazon Linux AMI found")
	}

	latestImage := describeImagesOutput.Images[0]
	for _, image := range describeImagesOutput.Images {
		if image.CreationDate != nil && *image.CreationDate > *latestImage.CreationDate {
			latestImage = image
		}
	}

	return *latestImage.ImageId, nil
}

func GetRootVolumeID(ctx context.Context, client *ec2.Client, instanceID string) (string, error) {
	output, err := client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil || len(output.Reservations) == 0 || len(output.Reservations[0].Instances) == 0 {
		return "", fmt.Errorf("could not find instance: %w", err)
	}

	for _, mapping := range output.Reservations[0].Instances[0].BlockDeviceMappings {
		if mapping.DeviceName != nil && *mapping.DeviceName == "/dev/xvda" {
			return *mapping.Ebs.VolumeId, nil
		}
	}

	return "", fmt.Errorf("root volume not found")
}

func TerminateInstance(ctx context.Context, client *ec2.Client, instanceID string) error {
	input := &ec2.TerminateInstancesInput{
		InstanceIds: []string{instanceID},
	}

	_, err := client.TerminateInstances(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to terminate instance %s: %w", instanceID, err)
	}

	return nil
}

func RegisterAMIFromSnapshot(ctx context.Context, client *ec2.Client, snapshotID string) (string, error) {
	name := fmt.Sprintf("dumie-ami-from-%s", snapshotID)

	// check existing ami
	describeInput := &ec2.DescribeImagesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("name"),
				Values: []string{name},
			},
		},
		Owners: []string{"self"},
	}
	describeOutput, err := client.DescribeImages(ctx, describeInput)
	if err == nil && len(describeOutput.Images) > 0 {
		return *describeOutput.Images[0].ImageId, nil
	}

	input := &ec2.RegisterImageInput{
		Name: aws.String(fmt.Sprintf("dumie-ami-from-%s", snapshotID)),
		BlockDeviceMappings: []types.BlockDeviceMapping{
			{
				DeviceName: aws.String("/dev/xvda"),
				Ebs: &types.EbsBlockDevice{
					SnapshotId:          aws.String(snapshotID),
					VolumeType:          types.VolumeTypeGp2,
					DeleteOnTermination: aws.Bool(true),
				},
			},
		},
		RootDeviceName:     aws.String("/dev/xvda"),
		VirtualizationType: aws.String("hvm"),
	}

	result, err := client.RegisterImage(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to register AMI from snapshot: %w", err)
	}

	return *result.ImageId, nil
}
