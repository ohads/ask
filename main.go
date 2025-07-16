package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"ask/config"
	"ask/setup"
)

type ChatRequest struct {
	Model    string               `json:"model"`
	Messages []config.ChatMessage `json:"messages"`
}

// ChatMessage is now defined in config package

type ChatResponse struct {
	Choices []struct {
		Message config.ChatMessage `json:"message"`
	} `json:"choices"`
}

func main() {
	var (
		setupFlag      = flag.Bool("setup", false, "Run the interactive setup process")
		modelFlag      = flag.String("model", "", "Override the configured model for this request")
		helpFlag       = flag.Bool("help", false, "Show help information")
		showConfigFlag = flag.Bool("show-config", false, "Show the current configuration and exit")
		editConfigFlag = flag.Bool("edit-config", false, "Edit the current configuration")
		clearFlag      = flag.Bool("clear", false, "Clear conversation history")
		noContextFlag  = flag.Bool("no-context", false, "Don't use conversation history for this request")
		newContextFlag = flag.String("new-context", "", "Create a new context with the given name")
		switchFlag     = flag.String("switch", "", "Switch to context by ID or name")
		listFlag       = flag.Bool("list-contexts", false, "List all contexts")
		deleteFlag     = flag.String("delete-context", "", "Delete context by ID or name")
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
		currentContext := cfg.GetCurrentContext()
		if currentContext != nil {
			contextType := ""
			if currentContext.Name == "default" && len(cfg.Contexts) == 1 {
				contextType = " (default)"
			}
			fmt.Printf("  Current context: %s%s (%s)\n", currentContext.Name, contextType, currentContext.ID)
			fmt.Printf("  Conversation history: %d messages\n", len(currentContext.History))
			if len(currentContext.History) > 0 {
				fmt.Println("  Recent conversation:")
				history := currentContext.History
				start := len(history) - 4 // Show last 2 exchanges (4 messages)
				if start < 0 {
					start = 0
				}
				for i := start; i < len(history); i++ {
					role := history[i].Role
					content := history[i].Content
					if len(content) > 50 {
						content = content[:47] + "..."
					}
					fmt.Printf("    %s: %s\n", role, content)
				}
			}
		} else {
			fmt.Printf("  Conversation history: %d messages (legacy)\n", cfg.GetHistoryLength())
			if cfg.GetHistoryLength() > 0 {
				fmt.Println("  Recent conversation:")
				history := cfg.GetHistory()
				start := len(history) - 4 // Show last 2 exchanges (4 messages)
				if start < 0 {
					start = 0
				}
				for i := start; i < len(history); i++ {
					role := history[i].Role
					content := history[i].Content
					if len(content) > 50 {
						content = content[:47] + "..."
					}
					fmt.Printf("    %s: %s\n", role, content)
				}
			}
		}
		return
	}

	if *editConfigFlag {
		if err := editConfig(); err != nil {
			log.Fatalf("Failed to edit configuration: %v", err)
		}
		return
	}

	if *clearFlag {
		cfg, err := config.Load()
		if err != nil {
			log.Fatalf("Failed to load configuration: %v", err)
		}
		cfg.ClearCurrentContext()
		if err := config.Save(cfg); err != nil {
			log.Fatalf("Failed to save configuration: %v", err)
		}
		fmt.Println("🗑️  Conversation history cleared.")
		return
	}

	if *newContextFlag != "" {
		cfg, err := config.Load()
		if err != nil {
			log.Fatalf("Failed to load configuration: %v", err)
		}
		id, err := cfg.CreateNewContext(*newContextFlag)
		if err != nil {
			log.Fatalf("Failed to create context: %v", err)
		}
		if err := config.Save(cfg); err != nil {
			log.Fatalf("Failed to save configuration: %v", err)
		}
		fmt.Printf("✅ Created new context '%s' with ID: %s\n", *newContextFlag, id)
		return
	}

	if *switchFlag != "" {
		cfg, err := config.Load()
		if err != nil {
			log.Fatalf("Failed to load configuration: %v", err)
		}
		if err := switchToContext(cfg, *switchFlag); err != nil {
			log.Fatalf("Failed to switch context: %v", err)
		}
		if err := config.Save(cfg); err != nil {
			log.Fatalf("Failed to save configuration: %v", err)
		}
		return
	}

	if *listFlag {
		cfg, err := config.Load()
		if err != nil {
			log.Fatalf("Failed to load configuration: %v", err)
		}
		listContexts(cfg)
		return
	}

	if *deleteFlag != "" {
		cfg, err := config.Load()
		if err != nil {
			log.Fatalf("Failed to load configuration: %v", err)
		}
		if err := deleteContext(cfg, *deleteFlag); err != nil {
			log.Fatalf("Failed to delete context: %v", err)
		}
		if err := config.Save(cfg); err != nil {
			log.Fatalf("Failed to save configuration: %v", err)
		}
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

	// Ensure we have a current context (creates default if needed)
	currentContext := cfg.GetCurrentContext()
	if currentContext != nil && cfg.CurrentContext != "" {
		// Save configuration if a default context was created
		if err := config.Save(cfg); err != nil {
			log.Printf("Warning: Failed to save configuration: %v", err)
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

	// Prepare messages for API request
	var messages []config.ChatMessage

	// Add conversation history if not disabled
	if !*noContextFlag {
		messages = append(messages, cfg.GetCurrentContextHistory()...)
	}

	// Add current user message
	messages = append(messages, config.ChatMessage{
		Role:    "user",
		Content: prompt,
	})

	// Make API request
	chatReq := ChatRequest{
		Model:    model,
		Messages: messages,
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
		response := chatResp.Choices[0].Message.Content
		fmt.Println(response)

		// Save conversation history if not disabled
		if !*noContextFlag {
			cfg.AddToCurrentContext("user", prompt)
			cfg.AddToCurrentContext("assistant", response)
			if err := config.Save(cfg); err != nil {
				log.Printf("Warning: Failed to save conversation history: %v", err)
			}
		}
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
	fmt.Println("  --setup         Run the interactive setup process")
	fmt.Println("  --model         Override the configured model for this request")
	fmt.Println("  --help          Show this help message")
	fmt.Println("  --show-config   Show the current configuration")
	fmt.Println("  --edit-config   Edit the current configuration")
	fmt.Println("  --clear         Clear conversation history")
	fmt.Println("  --no-context    Don't use conversation history for this request")
	fmt.Println()
	fmt.Println("Context Management:")
	fmt.Println("  --new-context   Create a new context with the given name")
	fmt.Println("  --switch        Switch to context by ID or name")
	fmt.Println("  --list-contexts List all contexts")
	fmt.Println("  --delete-context Delete context by ID or name")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  ask \"What is the capital of France?\"")
	fmt.Println("  ask --model gpt-4 \"Explain quantum computing\"")
	fmt.Println("  ask --setup")
	fmt.Println("  ask \"Continue from where we left off\"  # Uses conversation history")
	fmt.Println("  ask --clear  # Clear conversation history")
	fmt.Println()
	fmt.Println("Context Examples:")
	fmt.Println("  ask --new-context \"Python Project\"  # Create new context")
	fmt.Println("  ask --list-contexts                   # List all contexts")
	fmt.Println("  ask --switch \"Python Project\"       # Switch to context")
	fmt.Println("  ask \"What is a decorator?\"          # Use current context")
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

func editConfig() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	fmt.Println("🔧 Edit Ask CLI Configuration")
	fmt.Println("Leave blank to keep current value.")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	// Edit API Key
	fmt.Printf("Current API Key: %s\n", maskAPIKey(cfg.APIKey))
	fmt.Print("New API Key (or press Enter to keep current): ")
	apiKey, _ := reader.ReadString('\n')
	apiKey = strings.TrimSpace(apiKey)
	if apiKey != "" {
		cfg.APIKey = apiKey
	}

	// Edit Model
	fmt.Println()
	fmt.Printf("Current Model: %s\n", cfg.Model)
	fmt.Println("Available models:")
	models := config.GetAvailableModels()
	for i, model := range models {
		fmt.Printf("  %d. %s\n", i+1, model)
	}
	fmt.Print("New Model (number or name, or press Enter to keep current): ")
	modelChoice, _ := reader.ReadString('\n')
	modelChoice = strings.TrimSpace(modelChoice)

	if modelChoice != "" {
		// Try to parse as number first
		var choice int
		if _, err := fmt.Sscanf(modelChoice, "%d", &choice); err == nil {
			if choice >= 1 && choice <= len(models) {
				cfg.Model = models[choice-1]
			} else {
				return fmt.Errorf("invalid choice. Please select a number between 1 and %d", len(models))
			}
		} else {
			// Check if it's a valid model name
			validModel := false
			for _, model := range models {
				if model == modelChoice {
					cfg.Model = modelChoice
					validModel = true
					break
				}
			}
			if !validModel {
				return fmt.Errorf("invalid model name: %s", modelChoice)
			}
		}
	}

	// Save configuration
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %v", err)
	}

	fmt.Println()
	fmt.Println("✅ Configuration updated successfully!")
	return nil
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
    opts="--setup --model --help --show-config --edit-config --clear --no-context --new-context --switch --list-contexts --delete-context completion"
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

func switchToContext(cfg *config.Config, identifier string) error {
	// Try to find by ID first
	if err := cfg.SwitchContext(identifier); err == nil {
		context := cfg.GetCurrentContext()
		fmt.Printf("✅ Switched to context: %s (%s)\n", context.Name, context.ID)
		return nil
	}

	// Try to find by name
	contexts := cfg.ListContexts()
	for _, context := range contexts {
		if context.Name == identifier {
			if err := cfg.SwitchContext(context.ID); err != nil {
				return err
			}
			fmt.Printf("✅ Switched to context: %s (%s)\n", context.Name, context.ID)
			return nil
		}
	}

	return fmt.Errorf("context not found: %s", identifier)
}

func listContexts(cfg *config.Config) {
	contexts := cfg.ListContexts()
	currentContext := cfg.GetCurrentContext()

	if len(contexts) == 0 {
		fmt.Println("📝 No contexts found.")
		fmt.Println("Create a new context with: ask --new-context \"context name\"")
		return
	}

	fmt.Println("📝 Available contexts:")
	fmt.Println()

	for i, context := range contexts {
		marker := " "
		if currentContext != nil && currentContext.ID == context.ID {
			marker = "▶"
		}

		// Parse and format the updated time
		updated, _ := time.Parse(time.RFC3339, context.Updated)
		timeStr := updated.Format("Jan 02, 15:04")

		contextName := context.Name
		if context.Name == "default" && len(contexts) == 1 {
			contextName = "default (auto-created)"
		}

		fmt.Printf("%s %s (%s)\n", marker, contextName, context.ID)
		fmt.Printf("    Messages: %d | Updated: %s\n", len(context.History), timeStr)

		if i < len(contexts)-1 {
			fmt.Println()
		}
	}
}

func deleteContext(cfg *config.Config, identifier string) error {
	// Try to find by ID first
	if err := cfg.DeleteContext(identifier); err == nil {
		fmt.Printf("🗑️  Deleted context: %s\n", identifier)
		return nil
	}

	// Try to find by name
	contexts := cfg.ListContexts()
	for _, context := range contexts {
		if context.Name == identifier {
			if err := cfg.DeleteContext(context.ID); err != nil {
				return err
			}
			fmt.Printf("🗑️  Deleted context: %s (%s)\n", context.Name, context.ID)
			return nil
		}
	}

	return fmt.Errorf("context not found: %s", identifier)
}
