package setup

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"ask/config"
)

// Run starts the interactive setup process
func Run() error {
	fmt.Println("🤖 Welcome to Ask CLI Setup!")
	fmt.Println("This will help you configure your OpenAI API key and preferred model.")
	fmt.Println()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// Get API Key
	fmt.Println("📝 Step 1: OpenAI API Key")
	fmt.Println("You can get your API key from: https://platform.openai.com/account/api-keys")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	// Check if API key already exists
	if cfg.APIKey != "" {
		fmt.Print("API key already configured. Do you want to update it? (y/N): ")
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			fmt.Println("Keeping existing API key.")
		} else {
			cfg.APIKey = ""
		}
	}

	if cfg.APIKey == "" {
		fmt.Print("Enter your OpenAI API key: ")
		apiKey, _ := reader.ReadString('\n')
		apiKey = strings.TrimSpace(apiKey)

		if apiKey == "" {
			return fmt.Errorf("API key cannot be empty")
		}

		cfg.APIKey = apiKey
	}

	// Get Model Selection
	fmt.Println()
	fmt.Println("🤖 Step 2: Choose your preferred model")
	fmt.Println("Available models:")

	models := config.GetAvailableModels()
	for i, model := range models {
		fmt.Printf("  %d. %s\n", i+1, model)
	}
	fmt.Println()

	// Check if model already configured
	if cfg.Model != "" {
		fmt.Printf("Current model: %s\n", cfg.Model)
		fmt.Print("Do you want to change it? (y/N): ")
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			fmt.Println("Keeping existing model.")
		} else {
			cfg.Model = ""
		}
	}

	if cfg.Model == "" {
		fmt.Print("Enter the number of your preferred model (1-6): ")
		modelChoice, _ := reader.ReadString('\n')
		modelChoice = strings.TrimSpace(modelChoice)

		var choice int
		fmt.Sscanf(modelChoice, "%d", &choice)

		if choice < 1 || choice > len(models) {
			return fmt.Errorf("invalid choice. Please select a number between 1 and %d", len(models))
		}

		cfg.Model = models[choice-1]
	}

	// Save configuration
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %v", err)
	}

	fmt.Println()
	fmt.Println("✅ Setup complete!")
	fmt.Printf("Configuration saved to: %s\n", config.GetConfigPath())
	fmt.Println()
	fmt.Println("You can now use the ask CLI:")
	fmt.Println("  ask \"Your question here\"")
	fmt.Println()
	fmt.Println("To reconfigure, run: ask --setup")

	return nil
}

// GetConfigPath returns the path to the config file
func GetConfigPath() string {
	return config.GetConfigPath()
}
