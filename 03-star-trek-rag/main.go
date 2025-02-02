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
	"sort"
	"time"

	"github.com/ollama/ollama/api"
)

func main() {

	ctx := context.Background()

	ollamaUrl := os.Getenv("OLLAMA_HOST")
	if ollamaUrl == "" {
		ollamaUrl = "http://0.0.0.0:11434"
	}

	ollamaEmbeddingsUrl := os.Getenv("OLLAMA_EMBEDDINGS_HOST")
	if ollamaEmbeddingsUrl == "" {
		ollamaEmbeddingsUrl = "http://0.0.0.0:11434"
	}

	model := os.Getenv("LLM")
	if model == "" {
		model = "deepseek-r1:1.5b"
	}
	// deepseek-r1:1.5b 🔴🔥 with a base of 5 similarities 🟢 with a base of 4 similarities
	// deepseek-r1:7b 🟢 with a base of 5 similarities
	// deepseek-r1:8b 🟢 with a base of 5 similarities
	// deepseek-r1:14b 🟢 with a base of 5 similarities
	// deepseek-r1:32b 🟢 with a base of 5 similarities

	embeddingsModel := os.Getenv("EMBEDDINGS_LLM")
	if embeddingsModel == "" {
		embeddingsModel = "snowflake-arctic-embed:33m"
	}

	fmt.Println("🌍", ollamaUrl, "📕", model, "🌍", ollamaEmbeddingsUrl, "🌐", embeddingsModel)

	url, _ := url.Parse(ollamaUrl)
	chatClient := api.NewClient(url, http.DefaultClient)

	/*
		client, errCli := api.ClientFromEnvironment()
		if errCli != nil {
			log.Fatal("😡:", errCli)
		}
	*/

	url, _ = url.Parse(ollamaEmbeddingsUrl)
	embeddingsClient := api.NewClient(url, http.DefaultClient)

	systemInstructions, err := os.ReadFile("system-instructions.md")
	if err != nil {
		log.Fatal("😡:", err)
	}

	medicalInstructions, err := os.ReadFile("medical-instructions.md")
	if err != nil {
		log.Fatal("😡:", err)
	}

	/*
		federationMedicalDatabase, err := os.ReadFile("federation-medical-database.md")
		if err != nil {
			log.Fatal("😡:", err)
		}
	*/

	vectorStore := []rag.VectorRecord{}
	storeFile := "store-federation-medical-database.json"
	file, err := os.ReadFile(storeFile)
	if err != nil {
		log.Fatal("😡 Failed to read store file:", err)
	}
	if err := json.Unmarshal(file, &vectorStore); err != nil {
		log.Fatal("😡 Failed to unmarshal store:", err)
	}

	// Andorian Ice Plague
	userContent := `Using the above information and the below information, 

	The symptoms of the patient are:
        - Progressive skin crystallization
        - Internal organ freezing
        - Antennae necrosis
        - Hypothermic shock

	Make a diagnosis of the disease, provide the name of the disease and propose a treatment
	`

	/*
	   - Progressive skin crystallization
	   - Internal organ freezing
	   - Antennae necrosis
	   - Hypothermic shock

	*/

	// 🧠 Get the context from the similarities
	embeddingFromQuestion, _ := rag.GetEmbeddingFromChunk(ctx, embeddingsClient, embeddingsModel, userContent)

	// Search similarites between the question and the vectors of the store
	// 1- calculate the cosine similarity between the question and each vector in the store
	similarities := []rag.Similarity{}

	for _, vector := range vectorStore {
		cosineSimilarity, err := rag.CosineSimilarity(embeddingFromQuestion, vector.Embedding)
		if err != nil {
			log.Fatalln("😡", err)
		}

		// append to similarities
		similarities = append(similarities, rag.Similarity{
			Prompt:           vector.Prompt,
			CosineSimilarity: cosineSimilarity,
		})
	}

	// Select the N most similar chunks
	// retrieve in similarities the 5 records with the highest cosine similarity
	// sort the similarities
	sort.Slice(similarities, func(i, j int) bool {
		return similarities[i].CosineSimilarity > similarities[j].CosineSimilarity
	})

	// get the first N records
	topNSimilarities := similarities[:4]

	fmt.Println("🔍 Top similarities:")
	for _, similarity := range topNSimilarities {
		fmt.Println("🔍 Prompt:", similarity.Prompt)
		fmt.Println("🔍 Cosine similarity:", similarity.CosineSimilarity)
		fmt.Println("--------------------------------------------------")
	}

	// Create a new context with the top 5 chunks
	extractFromFederationMedicalDatabase := ""
	for _, similarity := range topNSimilarities {
		extractFromFederationMedicalDatabase += similarity.Prompt
	}

	// Prompt construction
	messages := []api.Message{
		{Role: "system", Content: string(systemInstructions)},
		{Role: "system", Content: extractFromFederationMedicalDatabase},
		{Role: "system", Content: string(medicalInstructions)},
		{Role: "user", Content: userContent},
	}

	stream := true
	//noStream  := false

	req := &api.ChatRequest{
		Model:    model,
		Messages: messages,
		Options: map[string]interface{}{
			"temperature":    0.0,
			"repeat_last_n":  2,
			"repeat_penalty": 2.2,
			"top_k":          10,
			"top_p":          0.5,
		},
		KeepAlive: &api.Duration{Duration: 1 * time.Minute},
		Stream:    &stream,
	}

	respFunc := func(resp api.ChatResponse) error {
		fmt.Print(resp.Message.Content)
		return nil
	}

	// Start the chat completion
	errChat := chatClient.Chat(ctx, req, respFunc)
	if errChat != nil {
		log.Fatal("😡:", errChat)
	}

	fmt.Println("")
	fmt.Println("")
}
