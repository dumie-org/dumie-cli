/*
Copyright © 2025 Chanhyeok Seo chanhyeok.seo2@gmail.com
*/
package awsutils

import (
	"context"
	"fmt"
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

func SearchDynamoDBLockTable(client *dynamodb.Client) (bool, error) {
	_, err := client.DescribeTable(context.Background(), &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		if err.Error() == "ResourceNotFoundException" {
			return false, nil
		}
		return false, fmt.Errorf("failed to describe table: %w", err)
	}
	return true, nil
}

func InitializeDynamoDBLockTable(client *dynamodb.Client) error {
	lock := &DynamoDBLock{
		Client:    client,
		TableName: tableName,
		TTL:       ttl,
	}

	err := lock.createLockTable(context.Background())
	if err != nil {
		return fmt.Errorf("failed to initialize DynamoDB lock table: %w", err)
	}
	return nil
}

func (lock *DynamoDBLock) createLockTable(ctx context.Context) error {
	fmt.Printf("Creating DynamoDB lock table: %s\n", lock.TableName)
	_, err := lock.Client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(lock.TableName),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("AccountID"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("AccountID"),
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

func (lock *DynamoDBLock) AcquireLock(ctx context.Context, accountID string) error {
	now := time.Now().Unix()
	expiration := now + int64(lock.TTL.Seconds())

	_, err := lock.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(lock.TableName),
		Item: map[string]types.AttributeValue{
			"AccountID": &types.AttributeValueMemberS{Value: accountID},
			"Expires":   &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", expiration)},
		},
		ConditionExpression: aws.String("attribute_not_exists(AccountID) OR Expires < :now"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":now": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", now)},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to acquire lock for account %s: %w", accountID, err)
	}

	return nil
}

func (lock *DynamoDBLock) ReleaseLock(ctx context.Context, accountID string) error {
	_, err := lock.Client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(lock.TableName),
		Key: map[string]types.AttributeValue{
			"AccountID": &types.AttributeValueMemberS{Value: accountID},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to release lock for account %s: %w", accountID, err)
	}

	return nil
}
