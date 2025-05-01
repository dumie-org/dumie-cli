package cmd

import (
	"fmt"
	"os"

	"github.com/dumie-org/dumie-cli/awsutils"
	"github.com/dumie-org/dumie-cli/awsutils/common"
	"github.com/spf13/cobra"
)

var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "Manage EC2 key pairs",
	Long:  `Generate and manage EC2 key pairs for instance access`,
}

var generateKeyCmd = &cobra.Command{
	Use:   "generate [key-name]",
	Short: "Generate a new EC2 key pair",
	Long:  `Generate a new EC2 key pair and save the private key to a file`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		client, err := awsutils.GetEC2AWSClient()
		if err != nil {
			fmt.Printf("Error creating AWS client: %v\n", err)
			os.Exit(1)
		}

		keyPairName, err := common.GenerateKeyPair(client)
		if err != nil {
			fmt.Printf("Error generating key pair: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully generated key pair:\n")
		fmt.Printf("Key Pair Name: %s\n", keyPairName)
	},
}

func init() {
	rootCmd.AddCommand(keyCmd)
	keyCmd.AddCommand(generateKeyCmd)
}
