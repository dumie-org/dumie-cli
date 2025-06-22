package iam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

const (
	roleName       = "DumieInstanceManagerRole"
	profileName    = "DumieInstanceManagerProfile"
	policyName     = "DumieInstanceManagerPolicy"
	policyDocument = `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": [
					"ec2:CreateImage",
					"ec2:DeregisterImage",
					"ec2:DescribeImages",
					"ec2:TerminateInstances",
					"ec2:CreateTags",
					"ec2:DescribeInstances",
					"ec2:DescribeVolumes",
					"ec2:CreateSnapshot",
					"ec2:DeleteSnapshot",
					"ec2:DescribeSnapshots"
				],
				"Resource": "*"
			},
			{
				"Effect": "Allow",
				"Action": [
					"dynamodb:PutItem",
					"dynamodb:GetItem",
					"dynamodb:DeleteItem",
					"dynamodb:DescribeTable",
					"dynamodb:CreateTable"
				],
				"Resource": "arn:aws:dynamodb:*:*:table/dumie-lock-table"
			}
		]
	}`
	trustPolicyDocument = `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {
					"Service": "ec2.amazonaws.com"
				},
				"Action": "sts:AssumeRole"
			}
		]
	}`
)

// CreateInstanceManagerRole creates an IAM role and instance profile for EC2 instances to manage themselves
func CreateInstanceManagerRole(client *iam.Client) error {
	ctx := context.TODO()

	// Check if role already exists
	_, err := client.GetRole(ctx, &iam.GetRoleInput{
		RoleName: aws.String(roleName),
	})
	if err == nil {
		fmt.Printf("IAM role %s already exists\n", roleName)
	} else {
		// Create the role
		createRoleInput := &iam.CreateRoleInput{
			RoleName:                 aws.String(roleName),
			AssumeRolePolicyDocument: aws.String(trustPolicyDocument),
			Description:              aws.String("Role for Dumie instance management"),
		}

		_, err = client.CreateRole(ctx, createRoleInput)
		if err != nil {
			return fmt.Errorf("failed to create IAM role: %w", err)
		}

		// Create the policy
		createPolicyInput := &iam.CreatePolicyInput{
			PolicyName:     aws.String(policyName),
			PolicyDocument: aws.String(policyDocument),
			Description:    aws.String("Policy for Dumie instance management"),
		}

		policyOutput, err := client.CreatePolicy(ctx, createPolicyInput)
		if err != nil {
			return fmt.Errorf("failed to create IAM policy: %w", err)
		}

		// Attach the policy to the role
		attachPolicyInput := &iam.AttachRolePolicyInput{
			RoleName:  aws.String(roleName),
			PolicyArn: policyOutput.Policy.Arn,
		}

		_, err = client.AttachRolePolicy(ctx, attachPolicyInput)
		if err != nil {
			return fmt.Errorf("failed to attach policy to role: %w", err)
		}

		fmt.Printf("Successfully created IAM role %s and attached policy %s\n", roleName, policyName)
	}

	// Check if instance profile exists
	_, err = client.GetInstanceProfile(ctx, &iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String(profileName),
	})
	if err == nil {
		fmt.Printf("Instance profile %s already exists\n", profileName)
		return nil
	}

	// Create instance profile
	createProfileInput := &iam.CreateInstanceProfileInput{
		InstanceProfileName: aws.String(profileName),
		Path:                aws.String("/"),
	}

	_, err = client.CreateInstanceProfile(ctx, createProfileInput)
	if err != nil {
		return fmt.Errorf("failed to create instance profile: %w", err)
	}

	// Add role to instance profile
	addRoleInput := &iam.AddRoleToInstanceProfileInput{
		InstanceProfileName: aws.String(profileName),
		RoleName:            aws.String(roleName),
	}

	_, err = client.AddRoleToInstanceProfile(ctx, addRoleInput)
	if err != nil {
		return fmt.Errorf("failed to add role to instance profile: %w", err)
	}

	fmt.Printf("Successfully created instance profile %s and attached role %s\n", profileName, roleName)
	return nil
}

// GetInstanceManagerRoleARN returns the ARN of the instance manager role
func GetInstanceManagerRoleARN(client *iam.Client) (string, error) {
	ctx := context.TODO()

	// Get the instance profile
	profile, err := client.GetInstanceProfile(ctx, &iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String(profileName),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get instance profile: %w", err)
	}

	return *profile.InstanceProfile.Arn, nil
}
