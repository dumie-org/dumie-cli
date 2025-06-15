package iam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

const (
	roleName       = "DumieInstanceManagerRole"
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

// CreateInstanceManagerRole creates an IAM role for EC2 instances to manage themselves
func CreateInstanceManagerRole(client *iam.Client) error {
	ctx := context.TODO()

	// Check if role already exists
	_, err := client.GetRole(ctx, &iam.GetRoleInput{
		RoleName: aws.String(roleName),
	})
	if err == nil {
		fmt.Printf("IAM role %s already exists\n", roleName)
		return nil
	}

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
	return nil
}

// GetInstanceManagerRoleARN returns the ARN of the instance manager role
func GetInstanceManagerRoleARN(client *iam.Client) (string, error) {
	ctx := context.TODO()

	role, err := client.GetRole(ctx, &iam.GetRoleInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get role: %w", err)
	}

	return *role.Role.Arn, nil
}
