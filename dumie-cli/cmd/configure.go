/*
Copyright © 2025 Chanhyeok Seo chanhyeok.seo2@gmail.com
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/chanhyeokseo/dumie/awsutils"
	"github.com/spf13/cobra"
)

type AWSConfig struct {
	AccessKeyID     string `json:"aws_access_key_id"`
	SecretAccessKey string `json:"aws_secret_access_key"`
	Region          string `json:"aws_region"`
}

const (
	configFilePath   = "../../aws_config.json"
	lockTableName    = "dumie-lock-table"
	defaultAWSRegion = "us-east-1"
)

// loadConfig loads the AWS configuration from the config file
func loadConfig() (*AWSConfig, error) {
	file, err := os.Open(configFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &AWSConfig{}, nil // Return an empty config if the file does not exist
		}
		return nil, fmt.Errorf("error opening config file: %w", err)
	}
	defer file.Close()

	var config AWSConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("error decoding config file: %w", err)
	}

	return &config, nil
}

// promptForInput prompts the user for input, showing the default value if available
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

// configureCmd represents the configure command
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
		}

		file, err := os.Create(configFilePath)
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

		fmt.Println("Configuration saved successfully. Now initializing DynamoDB lock table...")

		client, err := awsutils.GetDynamoDBClient()
		if err != nil {
			fmt.Printf("Error getting AWS client: %v\n", err)
			return
		}

		isTableExists, err := awsutils.SearchDynamoDBLockTable(client)
		if err != nil {
			fmt.Printf("Error searching for DynamoDB lock table: %v\n", err)
			return
		}

		if isTableExists {
			fmt.Println("DynamoDB lock table already exists. Skipping creation.")
			return
		}

		lock := awsutils.NewDynamoDBLock(client)

		err = lock.CreateLockTable(context.Background())
		if err != nil {
			fmt.Printf("Error creating DynamoDB lock table: %v\n", err)
			return
		}

		fmt.Println("DynamoDB lock table initialized successfully.")
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
