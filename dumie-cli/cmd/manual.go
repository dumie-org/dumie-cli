/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// manualCmd represents the manual command
var manualCmd = &cobra.Command{
	Use:   "manual",
	Short: "Dumie manual manager",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("manual called")
	},
}

func init() {
	deployCmd.AddCommand(manualCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// manualCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// manualCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
