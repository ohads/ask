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
		setupFlag      = flag.Bool("setup", false, "Run the interactive setup process")
		modelFlag      = flag.String("model", "", "Override the configured model for this request")
		helpFlag       = flag.Bool("help", false, "Show help information")
		showConfigFlag = flag.Bool("show-config", false, "Show the current configuration and exit")
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

	if *showConfigFlag {
		cfg, err := config.Load()
		if err != nil {
			log.Fatalf("Failed to load configuration: %v", err)
		}
		fmt.Println("Current Ask CLI configuration:")
		fmt.Printf("  Config file: %s\n", config.GetConfigPath())
		fmt.Printf("  API Key: %s\n", maskAPIKey(cfg.APIKey))
		fmt.Printf("  Model: %s\n", cfg.Model)
		return
	}

	// Autocomplete support
	if len(os.Args) > 1 && os.Args[1] == "completion" {
		printCompletionScript()
		return
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Check if API key is configured
	if cfg.APIKey == "" {
		fmt.Println("🤖 No configuration found. Starting setup process...")
		fmt.Println()
		if err := setup.Run(); err != nil {
			log.Fatalf("Setup failed: %v", err)
		}
		// Reload configuration after setup
		cfg, err = config.Load()
		if err != nil {
			log.Fatalf("Failed to load configuration after setup: %v", err)
		}
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

func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "********"
	}
	return key[:4] + strings.Repeat("*", len(key)-8) + key[len(key)-4:]
}

func printCompletionScript() {
	models := config.GetAvailableModels()
	modelList := strings.Join(models, " ")
	fmt.Println(`# bash/zsh completion for ask
_ask_completions() {
    local cur prev opts models
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    opts="--setup --model --help --show-config completion"
    models="` + modelList + `"

    if [[ $prev == --model ]]; then
        COMPREPLY=( $(compgen -W "$models" -- $cur) )
        return 0
    fi

    if [[ $cur == -* ]]; then
        COMPREPLY=( $(compgen -W "$opts" -- $cur) )
        return 0
    fi
}

complete -F _ask_completions ask
# To enable: source <(ask completion)
`)
}
