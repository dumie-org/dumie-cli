/*
Copyright © 2025 Chanhyeok Seo chanhyeok.seo2@gmail.com
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/dumie-org/dumie-cli/internal/aws/common"
	ec2utils "github.com/dumie-org/dumie-cli/internal/aws/ec2"
	"github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
	Use:   "connect [profile]",
	Short: "Connect to an EC2 instance via SSH",
	Long:  `Connect to a running EC2 instance using SSH with the configured key pair.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		profile := args[0]

		ec2Client, err := common.GetEC2AWSClient()
		if err != nil {
			fmt.Printf("Failed to create EC2 client: %v\n", err)
			return
		}

		instanceIDPtr, err := ec2utils.SearchEC2Instance(ec2Client, profile)
		if err != nil {
			fmt.Printf("Failed to find instance for profile [%s]: %v\n", profile, err)
			return
		}
		if instanceIDPtr == nil {
			fmt.Printf("No running instance found for profile [%s]\n", profile)
			return
		}
		instanceID := *instanceIDPtr

		instanceDetails, err := ec2Client.DescribeInstances(context.Background(), &ec2.DescribeInstancesInput{
			InstanceIds: []string{instanceID},
		})
		if err != nil {
			fmt.Printf("Failed to describe instance [%s]: %v\n", instanceID, err)
			return
		}

		instance := instanceDetails.Reservations[0].Instances[0]
		actualKeyName := ""
		if instance.KeyName != nil {
			actualKeyName = *instance.KeyName
		}

		fmt.Printf("Instance [%s] is using key pair: %s\n", instanceID, actualKeyName)

		publicDNS, err := ec2utils.GetInstancePublicDNS(ec2Client, instanceID)
		if err != nil {
			fmt.Printf("Failed to get public DNS for instance [%s]: %v\n", instanceID, err)
			return
		}
		if publicDNS == "" {
			fmt.Printf("Instance [%s] has no public DNS name\n", instanceID)
			return
		}

		keyPairName, err := common.GetKeyPairName()
		if err != nil {
			fmt.Printf("Failed to get key pair name: %v\n", err)
			return
		}

		keyFilePath := filepath.Join(".", fmt.Sprintf("%s.pem", keyPairName))
		if _, err := os.Stat(keyFilePath); os.IsNotExist(err) {
			fmt.Printf("Private key file not found: %s\n", keyFilePath)
			fmt.Println("Make sure the key file exists in the current directory.")
			return
		}

		sshArgs := []string{
			"-i", keyFilePath,
			"-o", "StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
			fmt.Sprintf("ec2-user@%s", publicDNS),
		}

		fmt.Printf("Connecting to instance [%s] at %s...\n", instanceID, publicDNS)
		fmt.Printf("SSH command: ssh %v\n", sshArgs)

		sshCmd := exec.Command("ssh", sshArgs...)
		sshCmd.Stdin = os.Stdin
		sshCmd.Stdout = os.Stdout
		sshCmd.Stderr = os.Stderr

		if err := sshCmd.Run(); err != nil {
			fmt.Printf("SSH connection failed: %v\n", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)
}
