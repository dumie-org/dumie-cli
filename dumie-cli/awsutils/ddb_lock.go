/*
Copyright © 2025 Chanhyeok Seo chanhyeok.seo2@gmail.com
*/
package awsutils

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoDBLock struct {
	Client    *dynamodb.Client
	TableName string
	TTL       time.Duration
}

const tableName = "dumie-lock-table"
const ttl = 5 * time.Minute

func NewDynamoDBLock(client *dynamodb.Client) *DynamoDBLock {
	return &DynamoDBLock{
		Client:    client,
		TableName: tableName,
		TTL:       ttl,
	}
}

func SearchDynamoDBLockTable(client *dynamodb.Client) (bool, error) {
	_, err := client.DescribeTable(context.Background(), &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		if errors.As(err, new(*types.ResourceNotFoundException)) {
			return false, nil
		}
		return false, fmt.Errorf("failed to describe table: %w", err)
	}
	return true, nil
}

func (lock *DynamoDBLock) CreateLockTable(ctx context.Context) error {
	fmt.Printf("Creating DynamoDB lock table: %s\n", lock.TableName)
	_, err := lock.Client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(lock.TableName),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("LockID"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("LockID"),
				KeyType:       types.KeyTypeHash,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	})

	if err != nil {
		return fmt.Errorf("failed to create lock table: %w", err)
	}

	fmt.Printf("DynamoDB lock table %s created successfully.\n", lock.TableName)
	return nil
}

func (lock *DynamoDBLock) AcquireLock(ctx context.Context, lockID string) error {
	now := time.Now().Unix()
	expiration := now + int64(lock.TTL.Seconds())

	_, err := lock.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(lock.TableName),
		Item: map[string]types.AttributeValue{
			"LockID":  &types.AttributeValueMemberS{Value: lockID},
			"Expires": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", expiration)},
		},
		ConditionExpression: aws.String("attribute_not_exists(LockID) OR Expires < :now"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":now": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", now)},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to acquire lock for lockID %s: %w", lockID, err)
	}

	return nil
}

func (lock *DynamoDBLock) ReleaseLock(ctx context.Context, lockID string) error {
	_, err := lock.Client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(lock.TableName),
		Key: map[string]types.AttributeValue{
			"LockID": &types.AttributeValueMemberS{Value: lockID},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to release lock for lockID %s: %w", lockID, err)
	}

	return nil
}

func (lock *DynamoDBLock) CheckLockStatus(ctx context.Context, lockID string) (bool, error) {
	output, err := lock.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(lock.TableName),
		Key: map[string]types.AttributeValue{
			"LockID": &types.AttributeValueMemberS{Value: lockID},
		},
	})
	if err != nil {
		return false, fmt.Errorf("failed to check lock status for lockID %s: %w", lockID, err)
	}

	if output.Item == nil {
		return false, nil
	}

	now := time.Now().Unix()
	expires, ok := output.Item["Expires"].(*types.AttributeValueMemberN)
	if !ok {
		return false, fmt.Errorf("invalid Expires attribute for lockID %s", lockID)
	}

	expirationTime, err := strconv.ParseInt(expires.Value, 10, 64)
	if err != nil {
		return false, fmt.Errorf("failed to parse Expires attribute for lockID %s: %w", lockID, err)
	}

	return expirationTime > now, nil
}
