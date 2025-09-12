package storage

import (
	"fmt"
	"os"
	"path/filepath"

	bolt "go.etcd.io/bbolt"
)

// Manager implements StorageManager interface for BoltDB storage
type Manager struct {
	db   *bolt.DB
	path string
}

// NewManager creates a new StorageManager instance
func NewManager(storagePath string) (*Manager, error) {
	if storagePath == "" {
		return nil, fmt.Errorf("storage path cannot be empty")
	}

	// Ensure storage directory exists
	if err := os.MkdirAll(filepath.Dir(storagePath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Open BoltDB database
	db, err := bolt.Open(storagePath, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create MCP bucket if it doesn't exist
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("mcp"))
		return err
	})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create MCP bucket: %w", err)
	}

	return &Manager{
		db:   db,
		path: storagePath,
	}, nil
}

// StoreMCP stores an MCP server configuration
func (m *Manager) StoreMCP(serverName string, config string) error {
	if serverName == "" {
		return fmt.Errorf("server name cannot be empty")
	}
	
	if config == "" {
		return fmt.Errorf("config cannot be empty")
	}

	return m.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("mcp"))
		if bucket == nil {
			return fmt.Errorf("mcp bucket not found")
		}
		
		return bucket.Put([]byte(serverName), []byte(config))
	})
}

// RetrieveMCP retrieves an MCP server configuration
func (m *Manager) RetrieveMCP(serverName string) (string, error) {
	if serverName == "" {
		return "", fmt.Errorf("server name cannot be empty")
	}

	var config string
	err := m.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("mcp"))
		if bucket == nil {
			return fmt.Errorf("mcp bucket not found")
		}
		
		data := bucket.Get([]byte(serverName))
		if data == nil {
			return fmt.Errorf("server %s not found", serverName)
		}
		
		config = string(data)
		return nil
	})

	return config, err
}

// ListMCP lists all stored MCP server configurations
func (m *Manager) ListMCP() (map[string]string, error) {
	configs := make(map[string]string)

	err := m.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("mcp"))
		if bucket == nil {
			return fmt.Errorf("mcp bucket not found")
		}

		return bucket.ForEach(func(key, value []byte) error {
			configs[string(key)] = string(value)
			return nil
		})
	})

	return configs, err
}

// DeleteMCP deletes an MCP server configuration
func (m *Manager) DeleteMCP(serverName string) error {
	if serverName == "" {
		return fmt.Errorf("server name cannot be empty")
	}

	return m.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("mcp"))
		if bucket == nil {
			return fmt.Errorf("mcp bucket not found")
		}

		// Check if key exists before deleting
		if bucket.Get([]byte(serverName)) == nil {
			return fmt.Errorf("server %s not found", serverName)
		}

		return bucket.Delete([]byte(serverName))
	})
}

// Close closes the database connection
func (m *Manager) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}