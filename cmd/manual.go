/*
Copyright © 2025 Chanhyeok Seo chanhyeok.seo2@gmail.com
*/
package cmd

import (
	"context"
	"fmt"

	"github.com/dumie-org/dumie-cli/internal/aws/ec2"
	"github.com/spf13/cobra"
)

var manualCmd = &cobra.Command{
	Use:   "manual [profile]",
	Short: "Dumie manual manager",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		profile := args[0]
		ctx := context.TODO()

		instanceID, err := ec2.RestoreOrCreateInstance(ctx, profile)
		if err != nil {
			fmt.Println("Failed to deploy instance:", err)
			return
		}

		fmt.Printf("Instance [%s] launched successfully for profile [%s]\n", instanceID, profile)

		err = ec2.DeleteOldSnapshotsByProfile(ctx, profile)
		if err != nil {
			fmt.Println("Warning: failed to delete old snapshots:", err)
		}
	},
}

func init() {
	deployCmd.AddCommand(manualCmd)
}
