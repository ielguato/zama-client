package cmd

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

func deleteFile(filename string) {
	// Use fmt.Sprintf to format the server URL string
	serverURL := fmt.Sprintf("%s/delete/%s", serverBaseURL, filename)

	// Create a DELETE request to delete the file
	req, err := http.NewRequest("DELETE", serverURL, nil)
	if err != nil {
		fmt.Printf("Error creating DELETE request: %v\n", err)
		return
	}

	// Send the DELETE request to the server
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending DELETE request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Check the response from the server
	if resp.StatusCode == http.StatusOK {
		fmt.Printf("File deleted successfully: %s\n", filename)
	} else {
		fmt.Printf("File deletion failed. Server returned: %s\n", resp.Status)
	}
}

// DeleteCmd creates and returns the delete command
func DeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [filename]",
		Short: "Delete a file from the server",
		Args:  cobra.ExactArgs(1), // Requires one argument
		Run: func(cmd *cobra.Command, args []string) {
			filename := args[0]
			deleteFile(filename)
		},
	}
}
