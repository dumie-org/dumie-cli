/*
Copyright © 2025 Chanhyeok Seo chanhyeok.seo2@gmail.com
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/dumie-org/dumie-cli/internal/aws"
	"github.com/dumie-org/dumie-cli/internal/aws/common"
	"github.com/dumie-org/dumie-cli/internal/aws/ddb"
	"github.com/spf13/cobra"
)

type AWSConfig struct {
	AccessKeyID     string `json:"aws_access_key_id"`
	SecretAccessKey string `json:"aws_secret_access_key"`
	Region          string `json:"aws_region"`
	KeyPairName     string `json:"key_pair_name"`
}

const (
	lockTableName    = "dumie-lock-table"
	defaultAWSRegion = "us-east-1"
)

func loadConfig() (*AWSConfig, error) {
	config := &AWSConfig{}

	if _, err := os.Stat(common.ConfigFilePath); os.IsNotExist(err) {
		return config, nil
	}

	data, err := os.ReadFile(common.ConfigFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	return config, nil
}

func promptForInput(prompt, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s (%s): ", prompt, defaultValue)
	} else if prompt == "Enter AWS_REGION" {
		fmt.Printf("%s (us-east-1): ", prompt)
	} else {
		fmt.Printf("%s: ", prompt)
	}
	var input string
	fmt.Scanln(&input)
	if input == "" {
		return defaultValue
	}
	return input
}

func configureDynamoDBLockTable() error {
	fmt.Println("Now initializing DynamoDB lock table...")

	client, err := aws.GetDynamoDBClient()
	if err != nil {
		fmt.Printf("Error getting AWS client: %v\n", err)
		return err
	}

	isTableExists, err := ddb.SearchDynamoDBLockTable(client)
	if err != nil {
		fmt.Printf("Error searching for DynamoDB lock table: %v\n", err)
		return err
	}

	if isTableExists {
		fmt.Println("DynamoDB lock table already exists. Skipping creation.")
		return nil
	}

	lock := ddb.NewDynamoDBLock(client)

	err = lock.CreateLockTable(context.Background())
	if err != nil {
		fmt.Printf("Error creating DynamoDB lock table: %v\n", err)
		return err
	}

	fmt.Println("DynamoDB lock table initialized successfully.")
	return nil
}

func configureEC2KeyPair() error {
	fmt.Println("Configuring EC2 key pair...")

	client, err := aws.GetEC2AWSClient()
	if err != nil {
		return fmt.Errorf("error creating AWS client: %v", err)
	}

	keyPairName, err := common.GenerateKeyPair(client)
	if err != nil {
		return fmt.Errorf("error generating key pair: %v", err)
	}

	fmt.Printf("Successfully configured key pair.\n")
	fmt.Printf("Key Pair Name: %s\n", keyPairName)

	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}

	config.KeyPairName = keyPairName

	// Save the updated config
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling config: %v", err)
	}

	if err := os.WriteFile(common.ConfigFilePath, data, 0644); err != nil {
		return fmt.Errorf("error writing config file: %v", err)
	}

	return nil
}

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure Dumie manager to integrate with AWS",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := loadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		awsAccessKeyID := promptForInput("Enter AWS_ACCESS_KEY_ID", config.AccessKeyID)
		awsSecretAccessKey := promptForInput("Enter AWS_SECRET_ACCESS_KEY", config.SecretAccessKey)
		awsRegion := promptForInput("Enter AWS_REGION", config.Region)
		if awsRegion == "" {
			awsRegion = defaultAWSRegion
		}

		newConfig := AWSConfig{
			AccessKeyID:     awsAccessKeyID,
			SecretAccessKey: awsSecretAccessKey,
			Region:          awsRegion,
			KeyPairName:     config.KeyPairName,
		}

		file, err := os.Create(common.ConfigFilePath)
		if err != nil {
			fmt.Printf("Error creating config file: %v\n", err)
			return
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		if err := encoder.Encode(newConfig); err != nil {
			fmt.Printf("Error encoding config to file: %v\n", err)
			return
		}

		fmt.Println("Configuration saved successfully.")

		err = configureDynamoDBLockTable()
		if err != nil {
			fmt.Printf("Error configuring DynamoDB lock table: %v\n", err)
			return
		}

		err = configureEC2KeyPair()
		if err != nil {
			fmt.Printf("Error configuring EC2 key pair: %v\n", err)
			return
		}

		fmt.Println("Configuration completed successfully.")
	},
}

func init() {
	rootCmd.AddCommand(configureCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// configureCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configureCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
