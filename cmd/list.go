package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/spf13/cobra"
)

// listFilesLocal lists the names of the directories in segmentsDir.
func listFilesLocal() {
	files, err := ioutil.ReadDir(segmentsDir)
	if err != nil {
		fmt.Printf("Error reading directory: %v\n", err)
		return
	}

	fmt.Println("Local directories:")
	for _, f := range files {
		if f.IsDir() {
			fmt.Println(f.Name())
		}
	}
}

// listFilesServer makes an API call to the serverBaseURL endpoint to list files.
func listFilesServer() {
	serverURL := fmt.Sprintf("%s/list", serverBaseURL)

	resp, err := http.Get(serverURL)
	if err != nil {
		fmt.Printf("Error making GET request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error listing files. Server returned: %s\n", resp.Status)
		return
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return
	}

	var fileList []string
	if err := json.Unmarshal(bodyBytes, &fileList); err != nil {
		fmt.Printf("Error unmarshalling response: %v\n", err)
		return
	}

	fmt.Println("Server files:")
	for _, file := range fileList {
		fmt.Println(file)
	}
}

// ListCmd creates and returns the list command
func ListCmd() *cobra.Command {
	var location string

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List files either locally or on the server",
		Long: `List command allows you to list directories or files depending on the location flag provided.
Use --location=local to list local directories, or --location=server to list files that are uploaded on the server.`,
		Run: func(cmd *cobra.Command, args []string) {
			switch location {
			case "local":
				listFilesLocal()
			case "server":
				listFilesServer()
			default:
				fmt.Println("Invalid location. Use 'local' or 'server'.")
			}
		},
	}

	listCmd.Flags().StringVarP(&location, "location", "l", "", "location to list files from (required)")
	listCmd.MarkFlagRequired("location")

	return listCmd
}
