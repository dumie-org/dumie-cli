package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/dumie-org/dumie-cli/internal/aws/common"
	"github.com/spf13/cobra"
)

var showAll bool

func printInstanceTable(instances []types.Instance) {
	fmt.Printf("\n%-20s %-25s %-15s %-18s %-20s %-10s\n",
		"NAME", "INSTANCE ID", "STATE", "PUBLIC IP", "LAUNCH TIME", "RESTORED")
	fmt.Println(strings.Repeat("-", 110))

	for _, instance := range instances {

		name := "-"
		publicIP := "-"
		launchTime := "-"
		var restored string

		for _, tag := range instance.Tags {
			if *tag.Key == "Name" {
				name = *tag.Value
			}
			if *tag.Key == "Restored" && *tag.Value == "true" {
				restored = "snapshot"
			}
		}

		if instance.PublicIpAddress != nil {
			publicIP = *instance.PublicIpAddress
		}

		if instance.LaunchTime != nil {
			launchTime = instance.LaunchTime.Format("2006-01-02 15:04:05")
		}

		fmt.Printf("%-20s %-25s %-15s %-18s %-20s %-10s\n",
			name,
			*instance.InstanceId,
			string(instance.State.Name),
			publicIP,
			launchTime,
			restored,
		)
	}
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all EC2 instances managed by Dumie",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.TODO()

		// EC2 client 생성
		client, err := common.GetEC2AWSClient()
		if err != nil {
			fmt.Println("Failed to create EC2 client:", err)
			return
		}

		input := &ec2.DescribeInstancesInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("tag:ManagedBy"),
					Values: []string{"Dumie"},
				},
			},
		}
		output, err := client.DescribeInstances(ctx, input)
		if err != nil {
			fmt.Println("Failed to describe instances:", err)
			return
		}

		var filteredInstances []types.Instance
		for _, r := range output.Reservations {
			for _, inst := range r.Instances {
				if showAll || inst.State.Name == types.InstanceStateNameRunning {
					filteredInstances = append(filteredInstances, inst)
				}
			}
		}

		if len(filteredInstances) == 0 {
			fmt.Println("No instances managed by Dumie found.")
			return
		}

		printInstanceTable(filteredInstances)
	},
}

func init() {
	listCmd.Flags().BoolVar(&showAll, "all", false, "Show all instances including stopped or terminated ones")
	rootCmd.AddCommand(listCmd)
}
