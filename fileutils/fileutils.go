package fileutils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	rootHashFileName = "rootHash" // Global constant for the root hash file name
	segmentsDir      = "segments"
)

// It recursively constructs the tree from the bottom up.
func updateMerkleRoot(segmentHashes [][]byte) []byte {
	// If there is only one hash, it is also the root hash.
	if len(segmentHashes) == 1 {
		return segmentHashes[0]
	}

	// Make sure the number of hashes is even by duplicating the last hash if necessary.
	if len(segmentHashes)%2 != 0 {
		segmentHashes = append(segmentHashes, segmentHashes[len(segmentHashes)-1])
	}

	var newLevel [][]byte

	// Combine each pair of adjacent hashes and hash the result,
	// working our way up the tree.
	for i := 0; i < len(segmentHashes); i += 2 {
		combinedHash := append(segmentHashes[i], segmentHashes[i+1]...)
		newHash := sha256.Sum256(combinedHash)
		newLevel = append(newLevel, newHash[:])
	}

	// Recursively call updateMerkleRoot on the new level.
	return updateMerkleRoot(newLevel)
}

// SplitFile splits a file into chunks and calculates a Merkle root hash.
func SplitFile(filename string, chunkSize int) ([]string, error) {
	// Create a subdirectory for segments based on the original file's name
	originalFileName := filepath.Base(filename)
	segmentsDir := filepath.Join(segmentsDir, originalFileName)
	if err := os.MkdirAll(segmentsDir, os.ModePerm); err != nil {
		return nil, err
	}

	// Open the input file
	inputFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer inputFile.Close()

	// Initialize variables to keep track of segment paths
	var segmentPaths []string
	var leaves [][]byte // To store the hashes of the file segments

	// Initialize a segment counter
	segmentCounter := 0

	for {
		// Create a buffer to read a chunk of data
		buffer := make([]byte, chunkSize)

		// Read a chunk from the input file
		n, err := inputFile.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		// Compute the hash of the segment
		segmentHash := sha256.Sum256(buffer[:n])
		leaves = append(leaves, segmentHash[:]) // Append hash to leaves

		// Create a segment file name using the segment counter
		segmentName := fmt.Sprintf("%d", segmentCounter)
		segmentPath := filepath.Join(segmentsDir, segmentName)

		// Write the chunk to the segment file
		err = os.WriteFile(segmentPath, buffer[:n], os.ModePerm)
		if err != nil {
			return nil, err
		}

		// Append the segment path to the list
		segmentPaths = append(segmentPaths, segmentPath)

		// Increment the segment counter
		segmentCounter++
	}

	// Now we compute the root hash using the leaves
	rootHash, err := computeMerkleRoot(leaves)
	if err != nil {
		return nil, err
	}
	// Save the root hash in a file named 'rootHash' in the segments directory
	hashFilePath := filepath.Join(segmentsDir, "rootHash")
	err = os.WriteFile(hashFilePath, rootHash, os.ModePerm)
	if err != nil {
		return nil, err
	}

	// Return the paths to the segment files
	return segmentPaths, nil
}

// calculateMerkleRoot calculates the Merkle root from a slice of leaf hashes.
func computeMerkleRoot(leafHashes [][]byte) ([]byte, error) {
	if len(leafHashes) == 0 {
		return nil, fmt.Errorf("no leaf hashes provided")
	}

	// Duplicate the last leaf if there is an odd number of leaves
	if len(leafHashes)%2 != 0 {
		leafHashes = append(leafHashes, leafHashes[len(leafHashes)-1])
	}

	for len(leafHashes) > 1 {
		var newLevel [][]byte
		for i := 0; i < len(leafHashes); i += 2 {
			// Hash the left and right leaves together
			combinedHash := append(leafHashes[i], leafHashes[i+1]...)
			parentHash := sha256.Sum256(combinedHash)
			newLevel = append(newLevel, parentHash[:])
		}

		// Duplicate the last node if there is an odd number of nodes
		if len(newLevel)%2 != 0 && len(newLevel) > 1 {
			newLevel = append(newLevel, newLevel[len(newLevel)-1])
		}

		leafHashes = newLevel
	}

	// Return the last hash left, which is the root
	return leafHashes[0], nil
}
