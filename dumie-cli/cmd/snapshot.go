package cmd

import (
	"context"
	"fmt"
	"github.com/dumie-org/dumie-cli/awsutils/common"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/spf13/cobra"

	"github.com/dumie-org/dumie-cli/awsutils"
	"github.com/dumie-org/dumie-cli/awsutils/ec2"
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Deploy an instance from a snapshot by ID",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.TODO()

		snapshotID, _ := cmd.Flags().GetString("id")
		if snapshotID == "" {
			fmt.Println("Please provide --id (snapshot ID) to deploy from.")
			return
		}

		ec2Client, err := awsutils.GetEC2AWSClient()
		if err != nil {
			fmt.Println("Failed to create EC2 client:", err)
			return
		}

		amiID, err := ec2.RegisterAMIFromSnapshot(ctx, ec2Client, snapshotID)
		if err != nil {
			fmt.Println("Failed to register AMI from snapshot:", err)
			return
		}
		fmt.Println("AMI registered:", amiID)

		sgID, err := ec2.CreateOrGetSecurityGroup(ec2Client, "dumie-default-sg")
		if err != nil {
			fmt.Println("Failed to create or get security group:", err)
			return
		}

		keyName, err := common.GenerateKeyPair(ec2Client)
		if err != nil {
			fmt.Printf("Error getting key pair name: %v\n", err)
			return
		}

		instanceType := types.InstanceTypeT2Micro
		instanceName := fmt.Sprintf("%s-instance", snapshotID)

		instanceIDPtr, err := ec2.LaunchEC2Instance(ec2Client, instanceName, amiID, instanceType, sgID, keyName)
		if err != nil {
			fmt.Println("Failed to launch instance from AMI:", err)
			return
		}

		fmt.Println("Instance launched successfully! InstanceID:", *instanceIDPtr)
	},
}

func init() {
	deployCmd.AddCommand(snapshotCmd)
	snapshotCmd.Flags().String("id", "", "Snapshot ID to deploy from (required)")
}
