package unit

import (
	"encoding/base64"
	"encoding/json"
	"mindful/src/models"
	"testing"
)

func TestMCPConfigEncoding(t *testing.T) {
	tests := []struct {
		name   string
		config interface{}
		want   string
	}{
		{
			name: "simple server configuration",
			config: map[string]interface{}{
				"command": "python",
				"args":    []string{"-m", "context7"},
			},
			want: "eyJhcmdzIjpbIi1tIiwiY29udGV4dDciXSwiY29tbWFuZCI6InB5dGhvbiJ9",
		},
		{
			name: "complex server configuration",
			config: map[string]interface{}{
				"command": "node",
				"args":    []string{"server.js"},
				"env": map[string]string{
					"NODE_ENV": "production",
					"PORT":     "3000",
				},
				"cwd": "/path/to/server",
			},
		},
		{
			name:   "empty configuration",
			config: map[string]interface{}{},
		},
		{
			name: "configuration with special characters",
			config: map[string]interface{}{
				"command": "python",
				"args":    []string{"-m", "mcp-server"},
				"env": map[string]string{
					"API_KEY":  "secret-key-123!@#$%^&*()",
					"DATABASE": "mysql://user:pass@localhost/db",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create MCP config and add server
			mcpConfig := models.NewMCPConfig()
			err := mcpConfig.AddServer("test-server", tt.config)
			if err != nil {
				t.Fatalf("AddServer() error = %v", err)
			}

			// Verify server was stored
			if !mcpConfig.HasServer("test-server") {
				t.Error("HasServer() should return true after adding server")
			}

			// Retrieve and decode server config
			retrieved, err := mcpConfig.GetServer("test-server")
			if err != nil {
				t.Fatalf("GetServer() error = %v", err)
			}

			// Compare original and retrieved configurations
			if !equalConfigs(tt.config.(map[string]interface{}), retrieved) {
				t.Errorf("Retrieved config doesn't match original.\nOriginal: %+v\nRetrieved: %+v", tt.config, retrieved)
			}

			// Test specific encoding if provided
			if tt.want != "" {
				encoded := mcpConfig.Servers["test-server"]
				if encoded != tt.want {
					t.Errorf("Encoding mismatch.\nGot: %s\nWant: %s", encoded, tt.want)
				}
			}
		})
	}
}

func TestBase64EncodingDecoding(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "simple JSON",
			input:   `{"command": "python", "args": ["-m", "server"]}`,
			wantErr: false,
		},
		{
			name:    "empty JSON",
			input:   `{}`,
			wantErr: false,
		},
		{
			name:    "JSON with special characters",
			input:   `{"key": "value with spaces and symbols!@#$%^&*()"}`,
			wantErr: false,
		},
		{
			name:    "JSON with unicode",
			input:   `{"message": "Hello 世界 🌍"}`,
			wantErr: false,
		},
		{
			name:    "complex nested JSON",
			input:   `{"server": {"host": "localhost", "port": 8080}, "config": {"debug": true, "timeout": 30}}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode to base64
			encoded := base64.StdEncoding.EncodeToString([]byte(tt.input))

			// Decode from base64
			decoded, err := base64.StdEncoding.DecodeString(encoded)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify round-trip encoding
				if string(decoded) != tt.input {
					t.Errorf("Round-trip encoding failed.\nOriginal: %s\nDecoded: %s", tt.input, string(decoded))
				}

				// Verify it's valid JSON
				var jsonData interface{}
				if err := json.Unmarshal(decoded, &jsonData); err != nil {
					t.Errorf("Decoded data is not valid JSON: %v", err)
				}
			}
		})
	}
}

func TestMCPConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		servers map[string]string
		wantErr bool
	}{
		{
			name: "valid MCP configuration",
			servers: map[string]string{
				"context7":   "eyJjb21tYW5kIjoicHl0aG9uIiwiYXJncyI6WyItbSIsImNvbnRleHQ3Il19",
				"filesystem": "eyJjb21tYW5kIjoibm9kZSIsImFyZ3MiOlsic2VydmVyLmpzIl19",
			},
			wantErr: false,
		},
		{
			name:    "empty MCP configuration",
			servers: map[string]string{},
			wantErr: false,
		},
		{
			name: "invalid base64 encoding",
			servers: map[string]string{
				"invalid": "not-valid-base64!@#$",
			},
			wantErr: true,
		},
		{
			name: "valid base64 but invalid JSON",
			servers: map[string]string{
				"invalid-json": base64.StdEncoding.EncodeToString([]byte("not valid json")),
			},
			wantErr: true,
		},
		{
			name: "mixed valid and invalid servers",
			servers: map[string]string{
				"valid":   "eyJjb21tYW5kIjoicHl0aG9uIn0=",
				"invalid": "invalid-base64",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mcpConfig := &models.MCPConfig{
				Servers: tt.servers,
			}

			err := mcpConfig.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMCPJSONConversion(t *testing.T) {
	tests := []struct {
		name    string
		servers map[string]interface{}
		wantErr bool
	}{
		{
			name: "simple server configuration",
			servers: map[string]interface{}{
				"context7": map[string]interface{}{
					"command": "python",
					"args":    []interface{}{"-m", "context7"},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple servers",
			servers: map[string]interface{}{
				"context7": map[string]interface{}{
					"command": "python",
					"args":    []interface{}{"-m", "context7"},
				},
				"filesystem": map[string]interface{}{
					"command": "node",
					"args":    []interface{}{"server.js"},
				},
			},
			wantErr: false,
		},
		{
			name:    "empty servers",
			servers: map[string]interface{}{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create MCP config and add servers
			mcpConfig := models.NewMCPConfig()
			for serverName, serverConfig := range tt.servers {
				err := mcpConfig.AddServer(serverName, serverConfig)
				if err != nil {
					t.Fatalf("AddServer() error = %v", err)
				}
			}

			// Convert to MCP JSON
			jsonData, err := mcpConfig.ToMCPJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("ToMCPJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Parse the JSON back
				parsedConfig, err := models.FromMCPJSON(jsonData)
				if err != nil {
					t.Fatalf("FromMCPJSON() error = %v", err)
				}

				// Verify all servers are present
				for serverName := range tt.servers {
					if !parsedConfig.HasServer(serverName) {
						t.Errorf("Parsed config missing server: %s", serverName)
					}
				}

				// Verify server count matches
				if len(parsedConfig.ListServers()) != len(tt.servers) {
					t.Errorf("Server count mismatch. Got %d, want %d",
						len(parsedConfig.ListServers()), len(tt.servers))
				}
			}
		})
	}
}

func TestMCPConfigOperations(t *testing.T) {
	mcpConfig := models.NewMCPConfig()

	// Test adding servers
	serverConfigs := map[string]interface{}{
		"context7": map[string]interface{}{
			"command": "python",
			"args":    []interface{}{"-m", "context7"},
		},
		"filesystem": map[string]interface{}{
			"command": "node",
			"args":    []interface{}{"fs-server.js"},
		},
	}

	for name, config := range serverConfigs {
		if err := mcpConfig.AddServer(name, config); err != nil {
			t.Fatalf("AddServer(%s) error = %v", name, err)
		}
	}

	// Test listing servers
	servers := mcpConfig.ListServers()
	if len(servers) != len(serverConfigs) {
		t.Errorf("ListServers() returned %d servers, want %d", len(servers), len(serverConfigs))
	}

	// Test HasServer
	for name := range serverConfigs {
		if !mcpConfig.HasServer(name) {
			t.Errorf("HasServer(%s) should return true", name)
		}
	}

	if mcpConfig.HasServer("nonexistent") {
		t.Error("HasServer(nonexistent) should return false")
	}

	// Test removing server
	mcpConfig.RemoveServer("context7")
	if mcpConfig.HasServer("context7") {
		t.Error("HasServer(context7) should return false after removal")
	}

	// Test cloning
	clone := mcpConfig.Clone()
	if clone == nil {
		t.Error("Clone() returned nil")
	}

	if len(clone.ListServers()) != len(mcpConfig.ListServers()) {
		t.Error("Clone() server count doesn't match original")
	}
}

// Helper function to compare configurations (order-independent)
func equalConfigs(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for key, valueA := range a {
		valueB, exists := b[key]
		if !exists {
			return false
		}

		// Handle different types
		switch vA := valueA.(type) {
		case string:
			if vB, ok := valueB.(string); !ok || vA != vB {
				return false
			}
		case []interface{}:
			vB, ok := valueB.([]interface{})
			if !ok || len(vA) != len(vB) {
				return false
			}
			for i, itemA := range vA {
				if itemA != vB[i] {
					return false
				}
			}
		case []string:
			// Handle []string vs []interface{} conversion after JSON marshalling
			if vB, ok := valueB.([]interface{}); ok {
				if len(vA) != len(vB) {
					return false
				}
				for i, itemA := range vA {
					if itemB, ok := vB[i].(string); !ok || itemA != itemB {
						return false
					}
				}
			} else if vB, ok := valueB.([]string); ok {
				if len(vA) != len(vB) {
					return false
				}
				for i, itemA := range vA {
					if itemA != vB[i] {
						return false
					}
				}
			} else {
				return false
			}
		case map[string]interface{}:
			if vB, ok := valueB.(map[string]interface{}); !ok || !equalConfigs(vA, vB) {
				return false
			}
		case map[string]string:
			// Handle map[string]string vs map[string]interface{} conversion after JSON marshalling
			if vB, ok := valueB.(map[string]interface{}); ok {
				if len(vA) != len(vB) {
					return false
				}
				for k, vA_val := range vA {
					if vB_val, exists := vB[k]; !exists {
						return false
					} else if vB_str, ok := vB_val.(string); !ok || vA_val != vB_str {
						return false
					}
				}
			} else if vB, ok := valueB.(map[string]string); ok {
				if len(vA) != len(vB) {
					return false
				}
				for k, vA_val := range vA {
					if vB_val, exists := vB[k]; !exists || vA_val != vB_val {
						return false
					}
				}
			} else {
				return false
			}
		default:
			if valueA != valueB {
				return false
			}
		}
	}

	return true
}
