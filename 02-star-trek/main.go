package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/ollama/ollama/api"
)

func main() {

	ctx := context.Background()

	ollamaUrl := os.Getenv("OLLAMA_HOST")
	if ollamaUrl == "" {
		ollamaUrl = "http://0.0.0.0:11434"
	}

	model := os.Getenv("LLM")
	if model == "" {
		model = "deepseek-r1:1.5b"
	}
	// deepseek-r1:1.5b 🔴
	// deepseek-r1:7b 🔴
	// deepseek-r1:8b 🔴
	// deepseek-r1:14b 🔴
	// deepseek-r1:32b 🔴

	fmt.Println("🌍", ollamaUrl, "📕", model)

	/*
		client, errCli := api.ClientFromEnvironment()
		if errCli != nil {
			log.Fatal("😡:", errCli)
		}
	*/

	url, _ := url.Parse(ollamaUrl)
	client := api.NewClient(url, http.DefaultClient)

	systemInstructions, err := os.ReadFile("system-instructions.md")
	if err != nil {
		log.Fatal("😡:", err)
	}

	medicalInstructions, err := os.ReadFile("medical-instructions.md")
	if err != nil {
		log.Fatal("😡:", err)
	}

	federationMedicalDatabase, err := os.ReadFile("federation-medical-database.xml")
	if err != nil {
		log.Fatal("😡:", err)
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

	// Prompt construction
	messages := []api.Message{
		{Role: "system", Content: string(systemInstructions)},
		{Role: "system", Content: string(medicalInstructions)},
		{Role: "system", Content: "# Federation Medical Database\n" + string(federationMedicalDatabase)},
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
	errChat := client.Chat(ctx, req, respFunc)
	if errChat != nil {
		log.Fatal("😡:", errChat)
	}

	fmt.Println("")
	fmt.Println("")
}
