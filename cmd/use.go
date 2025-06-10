package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/dumie-org/dumie-cli/internal/aws/common"
	ec2utils "github.com/dumie-org/dumie-cli/internal/aws/ec2"
	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:   "use [profile]",
	Short: "Create or connect to an instance with SSH monitoring",
	Long: `Create or connect to an instance with SSH monitoring.
If no instance exists for the profile, it will create one with SSH monitoring enabled.
If an instance exists, it will connect to it.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		profile := args[0]

		// Get AWS config and clients
		cfgData, err := common.LoadAWSConfig()
		if err != nil {
			fmt.Printf("Failed to get AWS config: %v\n", err)
			return
		}

		awsCfg, err := config.LoadDefaultConfig(
			context.TODO(),
			config.WithRegion(cfgData.Region),
		)
		if err != nil {
			fmt.Printf("Failed to load AWS config: %v\n", err)
			return
		}

		ec2Client := ec2.NewFromConfig(awsCfg)

		// Check if instance exists
		instanceIDPtr, err := ec2utils.SearchEC2Instance(ec2Client, profile)
		if err != nil {
			fmt.Printf("Failed to find instance: %v\n", err)
			return
		}

		var instanceID string
		if instanceIDPtr == nil {
			// No instance exists, create one
			fmt.Printf("No instance found for profile [%s]. Creating new instance...\n", profile)

			// Get AMI ID
			amiID, err := ec2utils.GetLatestAmazonLinuxAMI(ec2Client)
			if err != nil {
				fmt.Printf("Failed to get AMI ID: %v\n", err)
				return
			}

			// Create security group
			sgID, err := ec2utils.CreateOrGetSecurityGroup(ec2Client, profile)
			if err != nil {
				fmt.Printf("Failed to create security group: %v\n", err)
				return
			}

			// Get key pair name
			keyName, err := common.GetKeyPairName()
			if err != nil {
				fmt.Printf("Failed to get key pair name: %v\n", err)
				return
			}

			// Launch instance with SSH monitoring script
			userDataPath := "scripts/user_data/ssh_monitor.sh"
			instanceIDPtr, err = ec2utils.LaunchEC2Instance(ec2Client, profile, amiID, types.InstanceTypeT2Micro, sgID, keyName, &userDataPath)
			if err != nil {
				fmt.Printf("Failed to launch instance: %v\n", err)
				return
			}
			instanceID = *instanceIDPtr
		} else {
			instanceID = *instanceIDPtr
		}

		// Get instance details
		instanceDetails, err := ec2Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{
			InstanceIds: []string{instanceID},
		})
		if err != nil {
			fmt.Printf("Failed to get instance details: %v\n", err)
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
	rootCmd.AddCommand(useCmd)
}
