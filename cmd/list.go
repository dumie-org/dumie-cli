package cmd

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/dumie-org/dumie-cli/internal/aws/common"
	"github.com/spf13/cobra"
)

type ProfileInfo struct {
	Name       string
	InstanceID string
	Status     string
	PublicIP   string
	LaunchTime string
}

func printInstanceTable(profiles []ProfileInfo) {
	fmt.Printf("\n%-20s %-25s %-15s %-18s %-20s\n",
		"NAME", "INSTANCE ID", "STATE", "PUBLIC IP", "LAUNCH TIME")
	fmt.Println(strings.Repeat("-", 105))

	for _, p := range profiles {
		fmt.Printf("%-20s %-25s %-20s %-18s %-20s\n",
			p.Name,
			p.InstanceID,
			p.Status,
			p.PublicIP,
			p.LaunchTime,
		)
	}
}

var showAll bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all EC2 instances managed by Dumie",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client, err := common.GetEC2AWSClient()
		if err != nil {
			fmt.Println("Failed to create EC2 client:", err)
			return
		}

		profileMap := map[string]ProfileInfo{}

		// collect instances with ManagedBy=Dumie
		ec2Input := &ec2.DescribeInstancesInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("tag:ManagedBy"),
					Values: []string{"Dumie"},
				},
			},
		}
		ec2Output, err := client.DescribeInstances(ctx, ec2Input)
		if err != nil {
			fmt.Println("Failed to describe instances:", err)
			return
		}

		for _, r := range ec2Output.Reservations {
			for _, inst := range r.Instances {
				var name, publicIP, launchTime string
				name = "-"
				publicIP = "-"

				for _, tag := range inst.Tags {
					if *tag.Key == "Name" {
						name = *tag.Value
					}
				}

				if inst.PublicIpAddress != nil {
					publicIP = *inst.PublicIpAddress
				}
				if inst.LaunchTime != nil {
					launchTime = inst.LaunchTime.Local().Format("2006-01-02 15:04:05")
				}

				profileMap[name] = ProfileInfo{
					Name:       name,
					InstanceID: *inst.InstanceId,
					Status:     string(inst.State.Name),
					PublicIP:   publicIP,
					LaunchTime: launchTime,
				}
			}
		}

		// collect snapshots to find archived profiles
		if showAll {
			snapInput := &ec2.DescribeSnapshotsInput{
				Filters: []types.Filter{
					{
						Name:   aws.String("tag:ManagedBy"),
						Values: []string{"Dumie"},
					},
				},
				OwnerIds: []string{"self"},
			}
			snapOutput, err := client.DescribeSnapshots(ctx, snapInput)
			if err == nil {
				for _, snap := range snapOutput.Snapshots {
					var profile string
					for _, tag := range snap.Tags {
						if *tag.Key == "Name" {
							profile = *tag.Value
							break
						}
					}
					if profile != "" {
						if _, exists := profileMap[profile]; !exists {
							profileMap[profile] = ProfileInfo{
								Name:       profile,
								InstanceID: "-",
								Status:     "archived",
								PublicIP:   "-",
								LaunchTime: "-",
							}
						}
					}
				}
			}
		}

		if len(profileMap) == 0 {
			fmt.Println("No profiles managed by Dumie found.")
			return
		}

		var profiles []ProfileInfo
		for _, p := range profileMap {
			if !showAll && !strings.HasPrefix(p.Status, "running") {
				continue
			}
			profiles = append(profiles, p)
		}

		printInstanceTable(profiles)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVar(&showAll, "all", false, "Show all instances and snapshot-only profiles")
	flag.CommandLine.Parse([]string{}) // for compatibility with cobra+flag
}
