/*
Copyright © 2025 Chanhyeok Seo chanhyeok.seo2@gmail.com
*/
package cmd

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/chanhyeokseo/dumie/awsutils"
	"github.com/spf13/cobra"
)

var manualCmd = &cobra.Command{
	Use:   "manual",
	Short: "Dumie manual manager",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := awsutils.GetAWSClient()

		profile := args[0]

		if err != nil {
			fmt.Printf("Error getting AWS client: %v\n", err)
			return
		}

		defaultAMI, err := awsutils.GetLatestAmazonLinuxAMI(client)
		if err != nil {
			fmt.Printf("Error getting latest Amazon Linux AMI: %v\n", err)
			return
		}

		instanceType := types.InstanceTypeT2Micro
		securityGroupName := "dumie-default-sg"

		// Create or get the Security Group
		sgID, err := awsutils.CreateOrGetSecurityGroup(client, securityGroupName)
		if err != nil {
			fmt.Printf("Error creating or getting Security Group: %v\n", err)
			return
		}

		// Search if the EC2 instance already exists
		existingInstanceID, err := awsutils.SearchEC2Instance(client, profile)
		if err != nil {
			fmt.Printf("Error searching for existing EC2 instance: %v\n", err)
			return
		}

		if existingInstanceID != nil {
			fmt.Printf("EC2 instance already exists with ID: %s\n", *existingInstanceID)
			return
		}

		// Launch the EC2 instance
		instanceID, err := awsutils.LaunchEC2Instance(client, profile, defaultAMI, instanceType, sgID)
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
