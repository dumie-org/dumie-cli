package cmd

import (
	"context"
	"fmt"

	"github.com/dumie-org/dumie-cli/internal/aws/common"
	"github.com/dumie-org/dumie-cli/internal/aws/ec2"
	"github.com/spf13/cobra"
)

// deleteCmd deletes an EC2 instance by profile name and stores a snapshot of its root volume
var deleteCmd = &cobra.Command{
	Use:   "delete [profile]",
	Short: "Delete an instance by profile and create a snapshot of its root volume",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		profile := args[0]
		ctx := context.TODO()

		// Create EC2 Client
		ec2Client, err := common.GetEC2AWSClient()
		if err != nil {
			fmt.Println("Failed to create EC2 client:", err)
			return
		}

		// Find instance by tag:Name = profile
		instanceIDPtr, err := ec2.SearchEC2Instance(ec2Client, profile)
		if err != nil {
			fmt.Printf("Failed to find instance for profile [%s]: %v\n", profile, err)
			return
		}
		if instanceIDPtr == nil {
			fmt.Printf("No running instance found for profile [%s]\n", profile)
			return
		}
		instanceID := *instanceIDPtr

		// Get root volume ID
		volumeID, err := ec2.GetRootVolumeID(ctx, ec2Client, instanceID)
		if err != nil {
			fmt.Println("Failed to get root volume ID:", err)
			return
		}

		// Create snapshot with tag Name = profile
		snapshotMgr := ec2.NewSnapshotManagerFromClient(ec2Client)
		snapshotID, err := snapshotMgr.CreateSnapshot(ctx, volumeID, instanceID, profile)
		if err != nil {
			fmt.Println("Failed to create snapshot:", err)
			return
		}
		fmt.Printf("Snapshot [%s] successfully created for instance [%s] (profile: %s)\n", snapshotID, instanceID, profile)

		// Terminate instance
		err = ec2.TerminateInstance(ctx, ec2Client, instanceID)
		if err != nil {
			fmt.Println("Failed to terminate instance:", err)
			return
		}
		fmt.Printf("Instance [%s] (profile: %s) terminated successfully.\n", instanceID, profile)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
