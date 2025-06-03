/*
Copyright © 2025 Chanhyeok Seo chanhyeok.seo2@gmail.com
*/
package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type EC2StatusChecker struct {
	client     *ec2.Client
	instanceID string
}

func NewEC2StatusChecker(client *ec2.Client, instanceID string) *EC2StatusChecker {
	return &EC2StatusChecker{client, instanceID}
}

func (c *EC2StatusChecker) CheckStatus(ctx context.Context) (string, error) {
	describeInput := &ec2.DescribeInstancesInput{
		InstanceIds: []string{c.instanceID},
	}

	describeOutput, err := c.client.DescribeInstances(ctx, describeInput)
	if err != nil {
		return "", fmt.Errorf("error describing instance: %w", err)
	}

	if len(describeOutput.Reservations) == 0 || len(describeOutput.Reservations[0].Instances) == 0 {
		return "", fmt.Errorf("instance not found")
	}

	return string(describeOutput.Reservations[0].Instances[0].State.Name), nil
}

func (c *EC2StatusChecker) IsTargetStatus(currentStatus string) bool {
	return currentStatus == string(ec2types.InstanceStateNameRunning)
}

func (c *EC2StatusChecker) IsErrorStatus(currentStatus string) bool {
	return currentStatus == string(ec2types.InstanceStateNameTerminated) ||
		currentStatus == string(ec2types.InstanceStateNameShuttingDown)
}

func (c *EC2StatusChecker) GetResourceID() string {
	return c.instanceID
}

func (c *EC2StatusChecker) GetResourceType() string {
	return "EC2 instance"
}
