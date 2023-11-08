package main

import (
	"fmt"
	"os"
	"zama-client/cmd" // Replace with the module path in your go.mod

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{Use: "client"}

	// Add command functions from the cmd package
	rootCmd.AddCommand(cmd.UploadCmd())
	rootCmd.AddCommand(cmd.DownloadCmd())
	rootCmd.AddCommand(cmd.DeleteCmd())
	rootCmd.AddCommand(cmd.ProofCmd())
	rootCmd.AddCommand(cmd.ListCmd())
	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
