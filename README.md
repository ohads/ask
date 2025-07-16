# Ask Project

A Go CLI tool to get ChatGPT answers from the command line with easy setup and configuration.

## Features

- 🤖 Interactive setup process for API key and model configuration
- 🔧 Configurable model selection (GPT-4, GPT-3.5-turbo, etc.)
- 💾 Local configuration storage
- 🚀 Simple command-line interface
- 🔒 Secure API key storage

## Getting Started

### Prerequisites

- Go 1.20 or later
- An OpenAI API key ([get one here](https://platform.openai.com/account/api-keys))

### Installation

#### Option 1: Homebrew (Recommended)

```bash
brew tap ohads/ask
brew install ask
```

#### Option 2: Manual Build

1. Clone the repository
2. Navigate to the project directory
3. Build the application:

```bash
go build -o ask main.go
```

### First Time Setup

Run the interactive setup process to configure your API key and preferred model:

```bash
# If installed via Homebrew
ask --setup

# If built manually
./ask --setup
```

This will:
- Prompt you for your OpenAI API key
- Let you choose your preferred model
- Save the configuration to `~/.ask/config.json`

### Usage

After setup, you can use the CLI:

```bash
# Basic usage
ask "What is the capital of France?"

# Use a specific model for this request
ask --model gpt-4 "Explain quantum computing"

# Show help
ask --help
```

### Available Models

- `gpt-4.1-nano` - Latest GPT-4.1 nano model
- `gpt-4o` - Latest GPT-4 model
- `gpt-4o-mini` - Faster, more efficient GPT-4
- `gpt-4-turbo` - GPT-4 with extended context
- `gpt-4` - Standard GPT-4
- `gpt-3.5-turbo` - Fast and cost-effective
- `gpt-3.5-turbo-16k` - GPT-3.5 with extended context

### Configuration

The configuration is stored in `~/.ask/config.json` and includes:
- Your OpenAI API key
- Your preferred model

To reconfigure, run:
```bash
ask --setup
```

### Examples

```bash
# Ask a simple question
$ ask "What is the weather like in Paris?"
<ChatGPT's answer here>

# Use a specific model
$ ask --model gpt-4 "Write a Python function to sort a list"
<ChatGPT's answer here>

# Get help
$ ask --help
🤖 Ask CLI - Get ChatGPT answers from the command line
...
```

## Project Structure

```
ask/
├── main.go              # Main application entry point
├── config/
│   └── config.go        # Configuration management
├── setup/
│   └── setup.go         # Interactive setup process
├── go.mod               # Go module definition
├── Makefile             # Build automation
├── Formula/             # Homebrew formula
├── .github/workflows/   # GitHub Actions
├── .gitignore           # Git ignore file
└── README.md            # This file
```

## Development

### Building for Multiple Platforms

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Calculate SHA256 hashes for Homebrew
make sha256
```

### Local Installation

```bash
# Install locally
make install

# Uninstall
make uninstall
```

## Security

- Your API key is stored locally in `~/.ask/config.json`
- The config file has restricted permissions (600)
- Never commit your API key to version control

## Troubleshooting

- **"No API key configured"**: Run `ask --setup` to configure your API key
- **"OpenAI API error"**: Check your API key and internet connection
- **"No prompt provided"**: Make sure to provide a question in quotes

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request 