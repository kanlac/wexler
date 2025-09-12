package models

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// MCPConfig represents MCP (Model Context Protocol) server configurations
// Configurations are stored as base64 encoded JSON strings for security
type MCPConfig struct {
	Servers map[string]string `json:"servers" yaml:"servers"` // serverName -> base64 encoded JSON config
}

// NewMCPConfig creates a new empty MCP configuration
func NewMCPConfig() *MCPConfig {
	return &MCPConfig{
		Servers: make(map[string]string),
	}
}

// AddServer adds a server configuration by encoding the config as base64
func (m *MCPConfig) AddServer(serverName string, config interface{}) error {
	if m.Servers == nil {
		m.Servers = make(map[string]string)
	}
	
	// Convert config to JSON
	jsonData, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal server config for %s: %w", serverName, err)
	}
	
	// Encode as base64
	encoded := base64.StdEncoding.EncodeToString(jsonData)
	m.Servers[serverName] = encoded
	
	return nil
}

// GetServer retrieves and decodes a server configuration
func (m *MCPConfig) GetServer(serverName string) (map[string]interface{}, error) {
	if m.Servers == nil {
		return nil, fmt.Errorf("server %s not found", serverName)
	}
	
	encoded, exists := m.Servers[serverName]
	if !exists {
		return nil, fmt.Errorf("server %s not found", serverName)
	}
	
	// Decode from base64
	jsonData, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode server config for %s: %w", serverName, err)
	}
	
	// Parse JSON
	var config map[string]interface{}
	err = json.Unmarshal(jsonData, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal server config for %s: %w", serverName, err)
	}
	
	return config, nil
}

// RemoveServer removes a server configuration
func (m *MCPConfig) RemoveServer(serverName string) {
	if m.Servers != nil {
		delete(m.Servers, serverName)
	}
}

// ListServers returns all server names
func (m *MCPConfig) ListServers() []string {
	if m.Servers == nil {
		return []string{}
	}
	
	servers := make([]string, 0, len(m.Servers))
	for serverName := range m.Servers {
		servers = append(servers, serverName)
	}
	return servers
}

// HasServer checks if a server configuration exists
func (m *MCPConfig) HasServer(serverName string) bool {
	if m.Servers == nil {
		return false
	}
	_, exists := m.Servers[serverName]
	return exists
}

// Validate checks if the MCP configuration is valid
func (m *MCPConfig) Validate() error {
	if m == nil {
		return fmt.Errorf("MCP config is nil")
	}
	
	// Validate each server configuration can be decoded
	for serverName, encoded := range m.Servers {
		jsonData, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return fmt.Errorf("invalid base64 encoding for server %s: %w", serverName, err)
		}
		
		// Validate it's valid JSON
		var config interface{}
		err = json.Unmarshal(jsonData, &config)
		if err != nil {
			return fmt.Errorf("invalid JSON config for server %s: %w", serverName, err)
		}
	}
	
	return nil
}

// ToMCPJSON converts the MCP configuration to the standard .mcp.json format
func (m *MCPConfig) ToMCPJSON() ([]byte, error) {
	mcpServers := make(map[string]interface{})
	
	for serverName := range m.Servers {
		config, err := m.GetServer(serverName)
		if err != nil {
			return nil, fmt.Errorf("failed to decode server %s: %w", serverName, err)
		}
		mcpServers[serverName] = config
	}
	
	mcpFile := map[string]interface{}{
		"mcpServers": mcpServers,
	}
	
	return json.MarshalIndent(mcpFile, "", "  ")
}

// FromMCPJSON parses a standard .mcp.json file and creates MCP configuration
func FromMCPJSON(data []byte) (*MCPConfig, error) {
	var mcpFile map[string]interface{}
	err := json.Unmarshal(data, &mcpFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse MCP JSON: %w", err)
	}
	
	config := NewMCPConfig()
	
	mcpServers, ok := mcpFile["mcpServers"].(map[string]interface{})
	if !ok {
		return config, nil // Empty or invalid mcpServers section
	}
	
	for serverName, serverConfig := range mcpServers {
		err := config.AddServer(serverName, serverConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to add server %s: %w", serverName, err)
		}
	}
	
	return config, nil
}

// Clone creates a deep copy of the MCP configuration
func (m *MCPConfig) Clone() *MCPConfig {
	if m == nil {
		return nil
	}
	
	clone := NewMCPConfig()
	for serverName, encoded := range m.Servers {
		clone.Servers[serverName] = encoded
	}
	
	return clone
}