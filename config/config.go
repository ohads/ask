package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type Config struct {
	APIKey         string             `json:"api_key"`
	Model          string             `json:"model"`
	History        []ChatMessage      `json:"history,omitempty"`
	Contexts       map[string]Context `json:"contexts,omitempty"`
	CurrentContext string             `json:"current_context,omitempty"`
}

type Context struct {
	ID      string        `json:"id"`
	Name    string        `json:"name"`
	History []ChatMessage `json:"history"`
	Created string        `json:"created"`
	Updated string        `json:"updated"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

var (
	configDir  string
	configFile string
)

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("Failed to get home directory: %v", err))
	}
	configDir = filepath.Join(homeDir, ".ask")
	configFile = filepath.Join(configDir, "config.json")
}

// Load loads the configuration from file
func Load() (*Config, error) {
	config := &Config{
		Model: "gpt-3.5-turbo", // default model
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return config, nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return config, nil
}

// Save saves the configuration to file
func Save(config *Config) error {
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// GetAvailableModels returns a list of available OpenAI models
func GetAvailableModels() []string {
	return []string{
		"gpt-4.1-nano",
		"gpt-4o",
		"gpt-4o-mini",
		"gpt-4-turbo",
		"gpt-4",
		"gpt-3.5-turbo",
		"gpt-3.5-turbo-16k",
	}
}

// GetConfigPath returns the path to the config file
func GetConfigPath() string {
	return configFile
}

// AddToHistory adds a message to the conversation history
func (c *Config) AddToHistory(role, content string) {
	c.History = append(c.History, ChatMessage{
		Role:    role,
		Content: content,
	})
}

// ClearHistory clears the conversation history
func (c *Config) ClearHistory() {
	c.History = nil
}

// GetHistory returns the conversation history
func (c *Config) GetHistory() []ChatMessage {
	return c.History
}

// GetHistoryLength returns the number of messages in history
func (c *Config) GetHistoryLength() int {
	return len(c.History)
}

// Initialize contexts if not exists
func (c *Config) InitContexts() {
	if c.Contexts == nil {
		c.Contexts = make(map[string]Context)
	}
}

// CreateNewContext creates a new conversation context
func (c *Config) CreateNewContext(name string) (string, error) {
	c.InitContexts()

	// Generate unique ID
	id := generateID()

	// Check if name already exists
	for _, ctx := range c.Contexts {
		if ctx.Name == name {
			return "", fmt.Errorf("context with name '%s' already exists", name)
		}
	}

	now := time.Now().Format(time.RFC3339)
	context := Context{
		ID:      id,
		Name:    name,
		History: []ChatMessage{},
		Created: now,
		Updated: now,
	}

	c.Contexts[id] = context
	c.CurrentContext = id

	return id, nil
}

// SwitchContext switches to a different context
func (c *Config) SwitchContext(contextID string) error {
	c.InitContexts()

	if _, exists := c.Contexts[contextID]; !exists {
		return fmt.Errorf("context with ID '%s' not found", contextID)
	}

	c.CurrentContext = contextID
	return nil
}

// GetCurrentContext returns the current context
func (c *Config) GetCurrentContext() *Context {
	c.InitContexts()

	if c.CurrentContext == "" {
		// Create default context if none exists
		if len(c.Contexts) == 0 {
			id, err := c.CreateNewContext("default")
			if err != nil {
				// This shouldn't happen with "default" name, but handle it gracefully
				return nil
			}
			c.CurrentContext = id
		} else {
			// If contexts exist but none is selected, select the first one
			for id := range c.Contexts {
				c.CurrentContext = id
				break
			}
		}
	}

	if context, exists := c.Contexts[c.CurrentContext]; exists {
		return &context
	}

	return nil
}

// GetCurrentContextHistory returns the history of the current context
func (c *Config) GetCurrentContextHistory() []ChatMessage {
	context := c.GetCurrentContext()
	if context == nil {
		return []ChatMessage{}
	}
	return context.History
}

// AddToCurrentContext adds a message to the current context
func (c *Config) AddToCurrentContext(role, content string) {
	context := c.GetCurrentContext()
	if context == nil {
		// Fall back to legacy history
		c.AddToHistory(role, content)
		return
	}

	context.History = append(context.History, ChatMessage{
		Role:    role,
		Content: content,
	})
	context.Updated = time.Now().Format(time.RFC3339)
	c.Contexts[c.CurrentContext] = *context
}

// ClearCurrentContext clears the history of the current context
func (c *Config) ClearCurrentContext() {
	context := c.GetCurrentContext()
	if context == nil {
		// Fall back to legacy history
		c.ClearHistory()
		return
	}

	context.History = []ChatMessage{}
	context.Updated = time.Now().Format(time.RFC3339)
	c.Contexts[c.CurrentContext] = *context
}

// DeleteContext deletes a context
func (c *Config) DeleteContext(contextID string) error {
	c.InitContexts()

	if _, exists := c.Contexts[contextID]; !exists {
		return fmt.Errorf("context with ID '%s' not found", contextID)
	}

	delete(c.Contexts, contextID)

	// If we deleted the current context, clear it
	if c.CurrentContext == contextID {
		c.CurrentContext = ""
	}

	return nil
}

// ListContexts returns all contexts
func (c *Config) ListContexts() []Context {
	c.InitContexts()

	var contexts []Context
	for _, context := range c.Contexts {
		contexts = append(contexts, context)
	}

	// Sort by updated time (newest first)
	sort.Slice(contexts, func(i, j int) bool {
		return contexts[i].Updated > contexts[j].Updated
	})

	return contexts
}

// generateID generates a unique ID for contexts
func generateID() string {
	return fmt.Sprintf("ctx_%d", time.Now().UnixNano())
}
