package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"
)

func downloadFile(filename string) {
	directoryName := path.Base(filename)
	segmentsDir := path.Join(segmentsDir, directoryName) // Assuming segmentsDirName is a global constant for the segments directory

	// Ensure the segments directory exists
	if _, err := os.Stat(segmentsDir); os.IsNotExist(err) {
		fmt.Printf("Segments directory not found: %s\n", segmentsDir)
		return
	}

	// Loop through segments and download them
	index := 0
	for {
		segmentName := fmt.Sprintf("%d", index)
		segmentPath := path.Join(segmentsDir, segmentName)

		if _, err := os.Stat(segmentPath); os.IsNotExist(err) {
			// No more segments to download
			break
		}

		downloadSegment(filename, segmentName)
		index++
	}

	// Merge the downloaded segments into the final file
	if err := mergeSegments(filename); err != nil {
		fmt.Printf("Error merging segments: %v\n", err)
	}
}

func downloadSegment(filename string, segmentName string) {
	// Use fmt.Sprintf to format the server URL string
	serverURL := fmt.Sprintf("%s/download/%s/%s", serverBaseURL, filename, segmentName)

	// Ensure the downloads directory exists
	os.MkdirAll(downloadDir, os.ModePerm)

	// Set the directory and path for the segment
	segmentDir := path.Join(downloadDir, filename)
	segmentPath := path.Join(segmentDir, segmentName)

	// Send a GET request to download the file
	resp, err := http.Get(serverURL)
	if err != nil {
		fmt.Printf("Error sending GET request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		// Ensure the segment's directory exists
		os.MkdirAll(segmentDir, os.ModePerm)

		// Create a new file to save the downloaded segment
		outFile, err := os.Create(segmentPath)
		if err != nil {
			fmt.Printf("Error creating the output file: %v\n", err)
			return
		}
		defer outFile.Close()

		// Copy the server's response to the local file
		_, err = io.Copy(outFile, resp.Body)
		if err != nil {
			fmt.Printf("Error copying response to file: %v\n", err)
			return
		}

		fmt.Printf("Segment downloaded successfully: %s\n", segmentPath)
	} else {
		fmt.Printf("Segment download failed: %s Server returned: %s\n", segmentPath, resp.Status)
	}
}

func mergeSegments(filename string) error {
	segmentsDir := filepath.Join(downloadDir, filename)          // Use the global constant for 'downloads' directory
	outputPath := filepath.Join(downloadDir, filename, filename) // Specify output file, use '.merged' to avoid name conflict

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	counter := 0
	for {
		segmentPath := filepath.Join(segmentsDir, fmt.Sprintf("%d", counter))
		if _, err := os.Stat(segmentPath); os.IsNotExist(err) {
			// No more segments to merge
			break
		}

		segmentFile, err := os.Open(segmentPath)
		if err != nil {
			return fmt.Errorf("failed to open segment file '%s': %w", segmentPath, err)
		}

		_, err = io.Copy(outputFile, segmentFile)
		segmentFile.Close() // Make sure to close the file after each iteration

		if err != nil {
			return fmt.Errorf("failed to copy segment '%s' into '%s': %w", segmentPath, outputPath, err)
		}

		// Remove segment as it's no longer necessary
		if err := os.Remove(segmentPath); err != nil {
			return fmt.Errorf("failed to remove segment '%s': %w", segmentPath, err)
		}
		counter++
	}
	fmt.Printf("Merged and deleted segments. File: %s\n", outputPath)

	return nil
}

func DownloadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "download [filename]",
		Short: "Download a file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			filename := args[0]
			// Call your download logic here
			downloadFile(filename)
		},
	}
}
