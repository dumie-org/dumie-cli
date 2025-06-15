package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/dumie-org/dumie-cli/internal/aws/common"
	"github.com/dumie-org/dumie-cli/internal/aws/ddb"
	ec2utils "github.com/dumie-org/dumie-cli/internal/aws/ec2"
	"github.com/dumie-org/dumie-cli/internal/aws/iam"
	"github.com/spf13/cobra"
)

func connectToInstance(instanceID, publicDNS string) error {
	keyPairName, err := common.GetKeyPairName()
	if err != nil {
		return fmt.Errorf("failed to get key pair name: %v", err)
	}

	keyFilePath := filepath.Join(".", fmt.Sprintf("%s.pem", keyPairName))
	if _, err := os.Stat(keyFilePath); os.IsNotExist(err) {
		return fmt.Errorf("private key file not found: %s", keyFilePath)
	}

	sshArgs := []string{
		"-i", keyFilePath,
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		fmt.Sprintf("ec2-user@%s", publicDNS),
	}

	fmt.Printf("Connecting to instance [%s] at %s...\n", instanceID, publicDNS)
	sshCmd := exec.Command("ssh", sshArgs...)
	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr

	return sshCmd.Run()
}

func createNewInstance(profile string) (string, error) {
	iamClient, err := common.GetIAMClient()
	if err != nil {
		return "", fmt.Errorf("failed to get IAM client: %v", err)
	}

	roleARN, err := iam.GetInstanceManagerRoleARN(iamClient)
	if err != nil {
		return "", fmt.Errorf("failed to get IAM role ARN: %v", err)
	}

	userDataPath := filepath.Join("scripts", "user_data", "ssh_monitor.sh")
	return ec2utils.RestoreOrCreateInstance(context.TODO(), profile, &userDataPath, &roleARN)
}

var useCmd = &cobra.Command{
	Use:   "use [profile]",
	Short: "Create or connect to an instance with SSH monitoring",
	Long: `Create or connect to an instance with SSH monitoring.
If no instance exists for the profile, it will create one with SSH monitoring enabled.
If an instance exists, it will connect to it.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		profile := args[0]
		ctx := context.TODO()

		// Initialize DynamoDB lock
		ddbClient, err := common.GetDynamoDBAWSClient()
		if err != nil {
			fmt.Printf("Failed to get DynamoDB client: %v\n", err)
			return
		}

		lock := ddb.NewDynamoDBLock(ddbClient)

		// Check if table exists, create if it doesn't
		exists, err := ddb.SearchDynamoDBLockTable(ddbClient)
		if err != nil {
			fmt.Printf("Failed to check lock table: %v\n", err)
			return
		}
		if !exists {
			if err := lock.CreateLockTable(ctx); err != nil {
				fmt.Printf("Failed to create lock table: %v\n", err)
				return
			}
		}

		// Try to acquire lock for this profile with retry
		lockID := fmt.Sprintf("profile-%s", profile)
		startTime := time.Now()
		maxRetryTime := 10 * time.Minute
		retryInterval := 5 * time.Second

		for {
			err := lock.AcquireLock(ctx, lockID)
			if err == nil {
				break
			}

			elapsed := time.Since(startTime)
			if elapsed >= maxRetryTime {
				fmt.Printf("Failed to acquire lock after %v: %v\n", maxRetryTime, err)
				return
			}

			fmt.Printf("The instance is being created or in the termination process. Retrying to connect to the instance... (elapsed: %v)\n", elapsed.Round(time.Second))
			time.Sleep(retryInterval)
		}
		defer lock.ReleaseLock(ctx, lockID)

		ec2Client, err := common.GetEC2AWSClient()
		if err != nil {
			fmt.Printf("Failed to get EC2 client: %v\n", err)
			return
		}

		instanceIDPtr, err := ec2utils.SearchEC2Instance(ec2Client, profile)
		if err != nil {
			fmt.Printf("Failed to find instance: %v\n", err)
			return
		}

		var instanceID string
		if instanceIDPtr == nil {
			fmt.Printf("No instance found for profile [%s]. Creating new instance...\n", profile)
			instanceID, err = createNewInstance(profile)
			if err != nil {
				fmt.Printf("Failed to launch instance: %v\n", err)
				return
			}
		} else {
			instanceID = *instanceIDPtr
		}

		publicDNS, err := ec2utils.GetInstancePublicDNS(ec2Client, instanceID)
		if err != nil {
			fmt.Printf("Failed to get public DNS for instance [%s]: %v\n", instanceID, err)
			return
		}
		if publicDNS == "" {
			fmt.Printf("Instance [%s] has no public DNS name\n", instanceID)
			return
		}

		err = ec2utils.DeleteOldSnapshotsByProfile(ctx, profile)
		if err != nil {
			fmt.Println("Warning: failed to delete old snapshots:", err)
		}

		if err := connectToInstance(instanceID, publicDNS); err != nil {
			fmt.Printf("SSH connection failed: %v\n", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(useCmd)
}
