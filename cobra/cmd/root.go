/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dumie",
	Short: "Dumie, a smart on-demand instance manager",
	Long: `Dumie is a CLI tool designed to help you easily manage dummy instances used for testing purposes. 
It provides an automated and simple command-line interface that helps reduce cloud costs in testing environments. 
By tracking the active status of instances, it automatically terminates them to save costs and automatically saves the work state of testing environments.

Dumie has four types of managers:
1. Active Manager: This manager automatically terminates instances when they are not in use.
2. Schedule Manager: This manager automatically terminates instances based on a schedule.
3. TTL Manager: This manager automatically terminates instances after a certain period of time.
4. Manual Manager: This manager allows you to manually manage instances.
`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cobra.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
