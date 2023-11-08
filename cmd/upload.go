package cmd

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"zama-client/fileutils"

	"github.com/spf13/cobra"
)

func uploadFile(filePath string) {
	// Split the file into segments and get their paths
	segmentPaths, err := fileutils.SplitFile(filePath, chunkSize)
	if err != nil {
		fmt.Printf("Error splitting the file: %v\n", err)
		return
	}

	// Upload each segment
	for _, segmentPath := range segmentPaths {
		err := uploadSegment(segmentPath, path.Base(filePath))
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

// uploadSegment uploads a file segment to a server and returns an error if it fails
func uploadSegment(filePath string, parentFileName string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening the segment: %w", err)
	}
	defer file.Close()

	// Construct the server URL where you want to upload the file
	serverURL := fmt.Sprintf("%s/upload/%s", serverBaseURL, parentFileName)

	// Create a new multipart form request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create a form field for the file
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("error creating form file for segment: %w", err)
	}

	// Copy the file content into the form field
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("error copying segment file: %w", err)
	}

	// Close the multipart writer to complete the form data
	if err := writer.Close(); err != nil {
		return fmt.Errorf("error closing writer: %w", err)
	}

	// Create a POST request to upload the file
	req, err := http.NewRequest("POST", serverURL, body)
	if err != nil {
		return fmt.Errorf("error creating POST request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the POST request to the server
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending POST request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response from the server
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("segment upload failed. Server returned: %s", resp.Status)
	}

	fmt.Printf("Segment uploaded successfully: %s\n", filePath)
	return nil
}

func UploadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "upload [file_path]",
		Short: "Upload a file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			uploadFile(args[0])
		},
	}
}
