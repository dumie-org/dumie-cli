/*
Copyright © 2025 Chanhyeok Seo chanhyeok.seo2@gmail.com
*/
package ddb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoDBStatusChecker struct {
	client    *dynamodb.Client
	tableName string
}

func NewDynamoDBStatusChecker(client *dynamodb.Client, tableName string) *DynamoDBStatusChecker {
	return &DynamoDBStatusChecker{client, tableName}
}

func (c *DynamoDBStatusChecker) CheckStatus(ctx context.Context) (string, error) {
	describeInput := &dynamodb.DescribeTableInput{
		TableName: aws.String(c.tableName),
	}

	describeOutput, err := c.client.DescribeTable(ctx, describeInput)
	if err != nil {
		return "", fmt.Errorf("error describing table: %w", err)
	}

	return string(describeOutput.Table.TableStatus), nil
}

func (c *DynamoDBStatusChecker) IsTargetStatus(currentStatus string) bool {
	return currentStatus == string(types.TableStatusActive)
}

func (c *DynamoDBStatusChecker) IsErrorStatus(currentStatus string) bool {
	return currentStatus == string(types.TableStatusDeleting)
}

func (c *DynamoDBStatusChecker) GetResourceID() string {
	return c.tableName
}

func (c *DynamoDBStatusChecker) GetResourceType() string {
	return "DynamoDB table"
}
