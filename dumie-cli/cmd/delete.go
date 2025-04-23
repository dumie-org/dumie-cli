package cmd

import (
	"context"
	"fmt"
	"github.com/dumie-org/dumie-cli/awsutils"
	"github.com/dumie-org/dumie-cli/awsutils/ec2"
	"github.com/spf13/cobra"
)

// deleteCmd deletes an EC2 instance and stores a snapshot of its root volume
var deleteCmd = &cobra.Command{
	Use:   "delete [instance-id]",
	Short: "Delete an instance and create a snapshot of its root volume",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		instanceID := args[0]
		ctx := context.TODO()

		// Create EC2 Client
		ec2Client, err := awsutils.GetEC2AWSClient()
		if err != nil {
			fmt.Println("Failed to create EC2 client:", err)
			return
		}

		// Get Root Volume ID
		volumeID, err := ec2.GetRootVolumeID(ctx, ec2Client, instanceID)
		if err != nil {
			fmt.Println("Failed to get root volume ID:", err)
			return
		}

		// Create Snapshot
		snapshotMgr := ec2.NewSnapshotManagerFromClient(ec2Client)
		snapshotID, err := snapshotMgr.CreateSnapshot(ctx, volumeID, instanceID)
		if err != nil {
			fmt.Println("Failed to create snapshot:", err)
			return
		}
		fmt.Printf("Snapshot [%s] successfully created for instance [%s]\n", snapshotID, instanceID)

		// Terminate Instance
		err = ec2.TerminateInstance(ctx, ec2Client, instanceID)
		if err != nil {
			fmt.Println("Failed to terminate instance:", err)
			return
		}
		fmt.Println("Instance terminated successfully.")
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
