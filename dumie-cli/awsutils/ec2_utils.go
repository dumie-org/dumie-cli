/*
Copyright © 2025 Chanhyeok Seo chanhyeok.seo2@gmail.com
*/
package awsutils

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// GetDefaultVPCID retrieves the default VPC ID
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

// CreateOrGetSecurityGroup creates or retrieves a Security Group
func CreateOrGetSecurityGroup(client *ec2.Client, groupName string) (*string, error) {
	// Check if the Security Group already exists
	describeSGInput := &ec2.DescribeSecurityGroupsInput{
		GroupNames: []string{groupName},
	}
	describeSGOutput, err := client.DescribeSecurityGroups(context.TODO(), describeSGInput)
	if err == nil && len(describeSGOutput.SecurityGroups) > 0 {
		return describeSGOutput.SecurityGroups[0].GroupId, nil
	}

	// Create a new Security Group
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

	// Authorize inbound traffic
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

// LaunchEC2Instance launches an EC2 instance
func LaunchEC2Instance(client *ec2.Client, amiID string, instanceType types.InstanceType, sgID *string) (*string, error) {
	runInstancesInput := &ec2.RunInstancesInput{
		ImageId:      &amiID,
		InstanceType: instanceType,
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
		SecurityGroupIds: []string{
			*sgID,
		},
	}

	runInstancesOutput, err := client.RunInstances(context.TODO(), runInstancesInput)
	if err != nil {
		return nil, fmt.Errorf("error running instances: %w", err)
	}

	instanceID := runInstancesOutput.Instances[0].InstanceId
	return instanceID, nil
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

	// Find the latest image by creation date
	latestImage := describeImagesOutput.Images[0]
	for _, image := range describeImagesOutput.Images {
		if image.CreationDate != nil && *image.CreationDate > *latestImage.CreationDate {
			latestImage = image
		}
	}

	return *latestImage.ImageId, nil
}
