package rag

import (
	"context"
	"errors"
	"log"
	"math"
	"strings"

	"github.com/ollama/ollama/api"
)

func SplitText(text, delimiter string) []string {
	return strings.Split(text, delimiter)
}

func ChunkText(text string, chunkSize, overlap int) []string {
	chunks := []string{}
	for start := 0; start < len(text); start += chunkSize - overlap {
		end := start + chunkSize
		if end > len(text) {
			end = len(text)
		}
		chunks = append(chunks, text[start:end])
	}
	return chunks
}

type VectorRecord struct {
	Prompt    string    `json:"prompt"`
	Embedding []float64 `json:"embedding"`
}

type Similarity struct {
	Prompt           string
	CosineSimilarity float64
}

func GetEmbeddingFromChunk(ctx context.Context, client *api.Client, embeddingsModel, doc string) ([]float64, error) {

	req := &api.EmbeddingRequest{
		Model:  embeddingsModel,
		Prompt: doc,
	}
	// get embeddings
	resp, err := client.Embeddings(ctx, req)
	if err != nil {
		log.Println("😡:", err, "[from GetEmbeddingFromChunk]")
		return nil, err
	}
	return resp.Embedding, nil
}

// CosineSimilarity calculates the cosine similarity between two vectors
// Returns a value between -1 and 1, where:
// 1 means vectors are identical
// 0 means vectors are perpendicular
// -1 means vectors are opposite
func CosineSimilarity(vec1, vec2 []float64) (float64, error) {
	if len(vec1) != len(vec2) {
		return 0, errors.New("vectors must have the same length")
	}

	// Calculate dot product
	dotProduct := 0.0
	magnitude1 := 0.0
	magnitude2 := 0.0

	for i := 0; i < len(vec1); i++ {
		dotProduct += vec1[i] * vec2[i]
		magnitude1 += vec1[i] * vec1[i]
		magnitude2 += vec2[i] * vec2[i]
	}

	magnitude1 = math.Sqrt(magnitude1)
	magnitude2 = math.Sqrt(magnitude2)

	// Check for zero magnitudes to avoid division by zero
	if magnitude1 == 0 || magnitude2 == 0 {
		return 0, errors.New("vector magnitude cannot be zero")
	}

	return dotProduct / (magnitude1 * magnitude2), nil
}
