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
	// deepseek-r1:1.5b
	// deepseek-r1:7b
	// deepseek-r1:8b
	// deepseek-r1:14b
	// deepseek-r1:32b

	fmt.Println("🌍", ollamaUrl, "📕", model)

	url, _ := url.Parse(ollamaUrl)
	chatClient := api.NewClient(url, http.DefaultClient)

	systemInstructions := `You are an expert in mathematics and you are tutoring a student. 
	All the output are in markdown format.
	Instructions:
	1. Use the EXACT numbers provided
	2. Do not round or modify the number
	3. Keep ALL decimal places as given
	4. Use plain numbers with units
	5. Explain the concept without using mathematical notation
	6. Describe the solution in simple terms without formulas

	`

	/*
		knowledge := `# Knowledge base
		<known-formulas>
			<formula>
				Area of a rectangle:
				- Formula: Area = length × width
				- Similar example: 3m × 2m = 6m²
			</formula>
		</known-formulas>
		`
	*/

	/*
		statement := `Given: Rectangle with length 8m, width 4m`
		userContent := "Question: Calculate the rectangle's area"

		statement := `Given: Circle with radius 5 meters`
		userContent := `Problem Calculate area and perimeter of circle
		Expected accuracy: 2 decimal places
		`
	*/

	statement := `Given: Rectangle with length 8m, width 4.28m`
	userContent := "Question: Calculate the rectangle's area"

	// Prompt construction
	messages := []api.Message{
		{Role: "system", Content: systemInstructions},
		//{Role: "system", Content: knowledge},
		{Role: "system", Content: statement},
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
