package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"ask/config"
	"ask/setup"
)

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
}

func main() {
	var (
		setupFlag = flag.Bool("setup", false, "Run the interactive setup process")
		modelFlag = flag.String("model", "", "Override the configured model for this request")
		helpFlag  = flag.Bool("help", false, "Show help information")
	)
	flag.Parse()

	if *helpFlag {
		showHelp()
		return
	}

	if *setupFlag {
		if err := setup.Run(); err != nil {
			log.Fatalf("Setup failed: %v", err)
		}
		return
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Check if API key is configured
	if cfg.APIKey == "" {
		fmt.Println("❌ No API key configured.")
		fmt.Println("Please run the setup process:")
		fmt.Println("  ask --setup")
		os.Exit(1)
	}

	// Get prompt from command line arguments
	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("❌ No prompt provided.")
		fmt.Println("Usage: ask \"your question here\"")
		fmt.Println("For help: ask --help")
		os.Exit(1)
	}
	prompt := strings.Join(args, " ")

	// Determine which model to use
	model := cfg.Model
	if *modelFlag != "" {
		model = *modelFlag
	}

	// Make API request
	chatReq := ChatRequest{
		Model: model,
		Messages: []ChatMessage{{
			Role:    "user",
			Content: prompt,
		}},
	}
	body, err := json.Marshal(chatReq)
	if err != nil {
		log.Fatalf("Failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		log.Fatalf("OpenAI API error: %s", string(b))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		log.Fatalf("Failed to decode response: %v", err)
	}

	if len(chatResp.Choices) > 0 {
		fmt.Println(chatResp.Choices[0].Message.Content)
	} else {
		fmt.Println("No response from ChatGPT.")
	}
}

func showHelp() {
	fmt.Println("🤖 Ask CLI - Get ChatGPT answers from the command line")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  ask \"your question here\"")
	fmt.Println("  ask --model gpt-4 \"your question here\"")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --setup    Run the interactive setup process")
	fmt.Println("  --model    Override the configured model for this request")
	fmt.Println("  --help     Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  ask \"What is the capital of France?\"")
	fmt.Println("  ask --model gpt-4 \"Explain quantum computing\"")
	fmt.Println("  ask --setup")
	fmt.Println()
	fmt.Println("Configuration:")
	fmt.Printf("  Config file: %s\n", config.GetConfigPath())
	fmt.Println("  Run 'ask --setup' to configure your API key and preferred model")
}
