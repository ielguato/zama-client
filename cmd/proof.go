package cmd

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/spf13/cobra"
)

const (
	rootHashFileName = "rootHash" // Global constant for the root hash file name
)

type ProofPart struct {
	Hash    []byte
	IsRight bool
}

func merkleProof(filename, segmentId string) {
	// Create the URL for the request using the global server base URL
	url := fmt.Sprintf("%s/requestProof/%s/%s", serverBaseURL, filename, segmentId)

	// Send an HTTP GET request to the server
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error querying the requestProof endpoint for file '%s', segment '%s': %v\n", filename, segmentId, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Requesting Merkle proof...")

		// Request was successful, parse the response body
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading response body for file '%s', segment '%s': %v\n", filename, segmentId, err)
			return
		}

		// Deserialize the JSON response into a slice of ProofPart
		var proof []ProofPart
		if err = json.Unmarshal(bodyBytes, &proof); err != nil {
			fmt.Printf("Error deserializing proof for file '%s', segment '%s': %v\n", filename, segmentId, err)
			return
		}
		// Load the roothash from disk
		hashFilePath := filepath.Join(segmentsDir, filename, rootHashFileName)
		rootHash, err := ioutil.ReadFile(hashFilePath)
		if err != nil {
			fmt.Println("Error Opening rootHash file", err)
			return
		}
		// Output the received proof
		fmt.Printf("Received proof for file '%s', segment '%s'\n", filename, segmentId)
		// Verify the merkleProof
		verified, err := verifyMerkleProof(rootHash, proof)
		if !verified {
			fmt.Printf("Verification for Merkle proof failed for file '%s', segment '%s'", filename, segmentId)
			return
		}

		fmt.Printf("Verified proof for file '%s', segment '%s'", filename, segmentId)
	} else {
		// Handle non-OK status codes
		fmt.Printf("Request for Merkle proof failed for file '%s', segment '%s'. Server returned status: %s\n", filename, segmentId, resp.Status)
	}
}

func verifyMerkleProof(rootHash []byte, proof []ProofPart) (bool, error) {
	if len(proof) == 0 || len(proof[0].Hash) != sha256.Size {
		return false, fmt.Errorf("Invalid proof")
	}

	currentHash := proof[0].Hash

	for _, part := range proof[1:] {
		if len(part.Hash) != sha256.Size {
			return false, fmt.Errorf("Invalid hash size in proof part")
		}

		var combinedHash [sha256.Size]byte
		if part.IsRight {
			// If the proof part is the right node, append it to the right of the current hash.
			combinedHash = sha256.Sum256(append(currentHash, part.Hash...))
		} else {
			// If the proof part is the left node, append it to the left of the current hash.
			combinedHash = sha256.Sum256(append(part.Hash, currentHash...))
		}

		currentHash = combinedHash[:]
	}

	return bytes.Equal(currentHash, rootHash), nil
}

// ProofCmd creates and returns the proof command
func ProofCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "proof [filename] [segmentID]",
		Short: "Request a MerkleProof for a file segment",
		Args:  cobra.ExactArgs(2), // Requires two arguments
		Run: func(cmd *cobra.Command, args []string) {
			filename := args[0]
			segmentId := args[1]
			merkleProof(filename, segmentId)
		},
	}
}
