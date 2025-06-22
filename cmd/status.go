package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/dumie-org/dumie-cli/internal/aws/common"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status <profile>",
	Short: "Show the status of a Dumie-managed instance",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		profile := args[0]
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client, err := common.GetEC2AWSClient()
		if err != nil {
			fmt.Println("Failed to create EC2 client:", err)
			return
		}

		// Describe EC2 instance
		input := &ec2.DescribeInstancesInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("tag:Name"),
					Values: []string{profile},
				},
				{
					Name:   aws.String("tag:ManagedBy"),
					Values: []string{"Dumie"},
				},
			},
		}

		output, err := client.DescribeInstances(ctx, input)
		if err != nil {
			fmt.Printf("Error retrieving instance: %v\n", err)
			return
		}

		// Prefer a running instance if multiple instances exist
		var selected *types.Instance
		for _, r := range output.Reservations {
			for _, inst := range r.Instances {
				if selected == nil || inst.State.Name == types.InstanceStateNameRunning {
					selected = &inst
				}
			}
		}

		if selected != nil {
			var source = "base AMI"
			for _, tag := range selected.Tags {
				if *tag.Key == "Restored" && *tag.Value == "true" {
					source = "snapshot"
					break
				}
			}

			publicIP := "-"
			if selected.PublicIpAddress != nil {
				publicIP = *selected.PublicIpAddress
			}

			launchTime := "-"
			if selected.LaunchTime != nil {
				launchTime = selected.LaunchTime.Local().Format("2006-01-02 15:04:05")
			}

			fmt.Println()
			fmt.Printf("Profile:     %s\n", profile)
			fmt.Printf("Instance ID: %s\n", *selected.InstanceId)
			fmt.Printf("State:       %s\n", selected.State.Name)
			fmt.Printf("Public IP:   %s\n", publicIP)
			fmt.Printf("Launch Time: %s\n", launchTime)
			fmt.Printf("Source:      %s\n", source)
		} else {
			fmt.Println("No active instance found for this profile.")
			checkSnapshot(ctx, client, profile)
		}
	},
}

func checkSnapshot(ctx context.Context, client *ec2.Client, profile string) {
	snapInput := &ec2.DescribeSnapshotsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:ManagedBy"),
				Values: []string{"Dumie"},
			},
			{
				Name:   aws.String("tag:Name"),
				Values: []string{profile},
			},
		},
		OwnerIds: []string{"self"},
	}
	output, err := client.DescribeSnapshots(ctx, snapInput)
	if err != nil {
		fmt.Println("Error checking snapshots:", err)
		return
	}

	if len(output.Snapshots) > 0 {
		fmt.Printf("\nFound %d snapshot(s) for profile [%s]:\n", len(output.Snapshots), profile)
		for _, snap := range output.Snapshots {
			createdAt := "-"
			if snap.StartTime != nil {
				createdAt = snap.StartTime.Local().Format("2006-01-02 15:04:05")
			}
			fmt.Printf("- Snapshot ID: %s\n", *snap.SnapshotId)
			fmt.Printf("  Created At:  %s\n", createdAt)
			fmt.Printf("  Size (GiB):  %d\n", snap.VolumeSize)
		}
	} else {
		fmt.Printf("No instance or snapshot found for profile [%s].\n", profile)
	}
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
