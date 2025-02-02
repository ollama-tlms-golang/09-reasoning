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

	fmt.Println("🌍", ollamaUrl, "📕", model)


	url, _ := url.Parse(ollamaUrl)
	client := api.NewClient(url, http.DefaultClient)

	systemInstructions, err := os.ReadFile("system-instructions.md")
	if err != nil {
		log.Fatal("😡:", err)
	}

	irisInstructions, err := os.ReadFile("iris-instructions.md")
	if err != nil {
		log.Fatal("😡:", err)
	}

	irisDatabase, err := os.ReadFile("iris-database.xml")
	if err != nil {
		log.Fatal("😡:", err)
	}

	// Verginica
	/*
		userContent := `Using the above information and the below information,
		Given a specimen with:
		- Petal width: 2,5 cm
		- Petal length: 6 cm
		- Sepal width: 3,3 cm
		- Sepal length: 6,3 cm
		What is the species of the iris?
		`
	*/

	// Versicolor
	userContent := `Using the above information and the below information, 
	Given a specimen with:
	- Petal width: 1,5 cm
	- Petal length: 4,5 cm
	- Sepal width: 3,2 cm
	- Sepal length: 6,4 cm
	What is the species of the iris?
	`

	// Prompt construction
	messages := []api.Message{
		{Role: "system", Content: string(systemInstructions)},
		{Role: "system", Content: "# Iris Database\n" + string(irisDatabase)},
		{Role: "system", Content: string(irisInstructions)},
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
