package util

import (
	"os"
)

func GetEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func BatchIndices(indices []string, batchSize int) [][]string {
	var batches [][]string
	for batchSize < len(indices) {
		indices, batches = indices[batchSize:], append(batches, indices[0:batchSize:batchSize])
	}
	batches = append(batches, indices)
	return batches
}
