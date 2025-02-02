package main

import (
	"03-star-trek-rag/rag"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/ollama/ollama/api"
)

func TestGenerateChunk(t *testing.T) {
	ctx := context.Background()

	ollamaEmbeddingsUrl := os.Getenv("OLLAMA_EMBEDDINGS_HOST")
	if ollamaEmbeddingsUrl == "" {
		ollamaEmbeddingsUrl = "http://0.0.0.0:11434"
	}

	embeddingsModel := os.Getenv("EMBEDDINGS_LLM")
	if embeddingsModel == "" {
		embeddingsModel = "snowflake-arctic-embed:33m"
	}

	fmt.Println("🌍", ollamaEmbeddingsUrl, "📦", embeddingsModel)

	url, _ := url.Parse(ollamaEmbeddingsUrl)
	embeddingsClient := api.NewClient(url, http.DefaultClient)

	content, err := os.ReadFile("./federation-medical-database.xml")
	if err != nil {
		log.Fatal("😡:", err)
	}

	vectorStore := []rag.VectorRecord{}

	chunks := rag.SplitText(string(content), "</disease>")

	// Create embeddings from documents and save them in the store
	for idx, chunk := range chunks {
		fmt.Println("📝 Creating embedding nb:", idx)
		fmt.Println("📝 Chunk:", chunk)

		embedding, _ := rag.GetEmbeddingFromChunk(ctx, embeddingsClient, embeddingsModel, chunk)

		// Save the embedding in the vector store
		record := rag.VectorRecord{
			Prompt:    chunk,
			Embedding: embedding,
		}
		vectorStore = append(vectorStore, record)
	}

	// Marshal the store to JSON
	storeJSON, err := json.MarshalIndent(vectorStore, "", "  ")
	if err != nil {
		log.Fatal("Failed to marshal store to JSON:", err)
	}

	// Write the JSON to a file
	storeFile := "store-federation-medical-database.json"
	err = os.WriteFile(storeFile, storeJSON, 0644)
	if err != nil {
		log.Fatal("Failed to write store to file:", err)
	}

	fmt.Println("✅ Store persisted to", storeFile)

}
