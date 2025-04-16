/*
Copyright © 2025 Chanhyeok Seo chanhyeok.seo2@gmail.com
*/
package cmd

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/dumie-org/dumie-cli/awsutils"
	"github.com/dumie-org/dumie-cli/awsutils/ddb"
	"github.com/dumie-org/dumie-cli/awsutils/ec2"
	"github.com/spf13/cobra"
)

var manualCmd = &cobra.Command{
	Use:   "manual",
	Short: "Dumie manual manager",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {

		profile := args[0]

		lockClient, err := awsutils.GetDynamoDBClient()
		if err != nil {
			fmt.Printf("Error getting DynamoDB client: %v\n", err)
			return
		}

		lock := ddb.NewDynamoDBLock(lockClient)

		fmt.Println("Acquiring deployment lock...")
		err = lock.AcquireLock(context.TODO(), profile)
		if err != nil {
			fmt.Printf("Error acquiring deployment lock: %v\n", err)
			return
		}
		fmt.Println("Deployment lock acquired.")
		defer func() {
			fmt.Println("Releasing deployment lock...")
			err := lock.ReleaseLock(context.TODO(), profile)
			if err != nil {
				fmt.Printf("Error releasing deployment lock: %v\n", err)
			}
			fmt.Println("Deployment lock released.")
		}()

		client, err := awsutils.GetEC2AWSClient()
		if err != nil {
			fmt.Printf("Error getting AWS client: %v\n", err)
			return
		}

		defaultAMI, err := ec2.GetLatestAmazonLinuxAMI(client)
		if err != nil {
			fmt.Printf("Error getting latest Amazon Linux AMI: %v\n", err)
			return
		}

		instanceType := types.InstanceTypeT2Micro
		securityGroupName := "dumie-default-sg"

		sgID, err := ec2.CreateOrGetSecurityGroup(client, securityGroupName)
		if err != nil {
			fmt.Printf("Error creating or getting Security Group: %v\n", err)
			return
		}

		existingInstanceID, err := ec2.SearchEC2Instance(client, profile)
		if err != nil {
			fmt.Printf("Error searching for existing EC2 instance: %v\n", err)
			return
		}

		if existingInstanceID != nil {
			fmt.Printf("EC2 instance already exists with ID: %s\n", *existingInstanceID)
			return
		}

		instanceID, err := ec2.LaunchEC2Instance(client, profile, defaultAMI, instanceType, sgID)
		if err != nil {
			fmt.Printf("Error launching EC2 instance: %v\n", err)
			return
		}

		fmt.Printf("EC2 instance launched successfully with ID: %s\n", *instanceID)
	},
}

func init() {
	deployCmd.AddCommand(manualCmd)
}
