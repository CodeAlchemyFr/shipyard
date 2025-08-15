package registry

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/shipyard/cli/pkg/database"
)

// Registry represents a container registry configuration
type Registry struct {
	ID           int64     `json:"id"`
	RegistryURL  string    `json:"registry_url"`
	Username     string    `json:"username"`
	Password     string    `json:"password"` // Encrypted
	Email        string    `json:"email"`
	RegistryType string    `json:"registry_type"`
	IsDefault    bool      `json:"is_default"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Manager handles registry credentials
type Manager struct {
	db  *database.DB
	key []byte // AES encryption key
}

// NewManager creates a new registry manager
func NewManager() (*Manager, error) {
	db, err := database.NewDB()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Simple encryption key (in production, use proper key management)
	key := make([]byte, 32)
	copy(key, []byte("shipyard-secret-key-32-bytes!!!"))

	return &Manager{
		db:  db,
		key: key,
	}, nil
}

// AddRegistry adds a new registry credential
func (m *Manager) AddRegistry(registryURL, username, password, email, registryType string, isDefault bool) error {
	// Encrypt password
	encryptedPassword, err := m.encrypt(password)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}

	// Normalize registry URL
	registryURL = normalizeRegistryURL(registryURL)

	// Insert registry
	query := `
		INSERT INTO registry_credentials (registry_url, username, password, email, registry_type, is_default)
		VALUES (?, ?, ?, ?, ?, ?)`

	_, err = m.db.GetConnection().Exec(query, registryURL, username, encryptedPassword, email, registryType, isDefault)
	if err != nil {
		return fmt.Errorf("failed to add registry: %w", err)
	}

	return nil
}

// GetRegistry gets a registry by URL
func (m *Manager) GetRegistry(registryURL string) (*Registry, error) {
	registryURL = normalizeRegistryURL(registryURL)

	query := `
		SELECT id, registry_url, username, password, email, registry_type, is_default, created_at, updated_at
		FROM registry_credentials
		WHERE registry_url = ?`

	var registry Registry
	var encryptedPassword string

	err := m.db.GetConnection().QueryRow(query, registryURL).Scan(
		&registry.ID,
		&registry.RegistryURL,
		&registry.Username,
		&encryptedPassword,
		&registry.Email,
		&registry.RegistryType,
		&registry.IsDefault,
		&registry.CreatedAt,
		&registry.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("registry not found: %w", err)
	}

	// Decrypt password
	registry.Password, err = m.decrypt(encryptedPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt password: %w", err)
	}

	return &registry, nil
}

// GetRegistryForImage determines which registry to use for an image
func (m *Manager) GetRegistryForImage(image string) (*Registry, error) {
	// Extract registry from image
	registryURL := extractRegistryFromImage(image)
	
	// Try to find specific registry
	registry, err := m.GetRegistry(registryURL)
	if err == nil {
		return registry, nil
	}

	// If not found, try default registry
	return m.GetDefaultRegistry()
}

// GetDefaultRegistry gets the default registry
func (m *Manager) GetDefaultRegistry() (*Registry, error) {
	query := `
		SELECT id, registry_url, username, password, email, registry_type, is_default, created_at, updated_at
		FROM registry_credentials
		WHERE is_default = 1
		LIMIT 1`

	var registry Registry
	var encryptedPassword string

	err := m.db.GetConnection().QueryRow(query).Scan(
		&registry.ID,
		&registry.RegistryURL,
		&registry.Username,
		&encryptedPassword,
		&registry.Email,
		&registry.RegistryType,
		&registry.IsDefault,
		&registry.CreatedAt,
		&registry.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("no default registry found: %w", err)
	}

	// Decrypt password
	registry.Password, err = m.decrypt(encryptedPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt password: %w", err)
	}

	return &registry, nil
}

// ListRegistries lists all registries
func (m *Manager) ListRegistries() ([]Registry, error) {
	query := `
		SELECT id, registry_url, username, password, email, registry_type, is_default, created_at, updated_at
		FROM registry_credentials
		ORDER BY is_default DESC, registry_url`

	rows, err := m.db.GetConnection().Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query registries: %w", err)
	}
	defer rows.Close()

	var registries []Registry
	for rows.Next() {
		var registry Registry
		var encryptedPassword string

		err := rows.Scan(
			&registry.ID,
			&registry.RegistryURL,
			&registry.Username,
			&encryptedPassword,
			&registry.Email,
			&registry.RegistryType,
			&registry.IsDefault,
			&registry.CreatedAt,
			&registry.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan registry row: %w", err)
		}

		// Don't decrypt password for listing (security)
		registry.Password = "***"
		registries = append(registries, registry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating registry rows: %w", err)
	}

	return registries, nil
}

// RemoveRegistry removes a registry
func (m *Manager) RemoveRegistry(registryURL string) error {
	registryURL = normalizeRegistryURL(registryURL)

	query := `DELETE FROM registry_credentials WHERE registry_url = ?`

	result, err := m.db.GetConnection().Exec(query, registryURL)
	if err != nil {
		return fmt.Errorf("failed to remove registry: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("registry %s not found", registryURL)
	}

	return nil
}

// SetDefaultRegistry sets a registry as default
func (m *Manager) SetDefaultRegistry(registryURL string) error {
	registryURL = normalizeRegistryURL(registryURL)

	query := `UPDATE registry_credentials SET is_default = 1 WHERE registry_url = ?`

	result, err := m.db.GetConnection().Exec(query, registryURL)
	if err != nil {
		return fmt.Errorf("failed to set default registry: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("registry %s not found", registryURL)
	}

	return nil
}

// Close closes the database connection
func (m *Manager) Close() error {
	return m.db.Close()
}

// encrypt encrypts a string using AES
func (m *Manager) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(m.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decrypts a string using AES
func (m *Manager) decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(m.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// normalizeRegistryURL normalizes a registry URL
func normalizeRegistryURL(registryURL string) string {
	// Remove protocol
	registryURL = strings.TrimPrefix(registryURL, "https://")
	registryURL = strings.TrimPrefix(registryURL, "http://")
	
	// Handle Docker Hub special case
	if registryURL == "docker.io" || registryURL == "index.docker.io" || registryURL == "" {
		return "docker.io"
	}
	
	return strings.TrimSuffix(registryURL, "/")
}

// extractRegistryFromImage extracts registry URL from image name
func extractRegistryFromImage(image string) string {
	parts := strings.Split(image, "/")
	
	// If no slash, it's Docker Hub
	if len(parts) == 1 {
		return "docker.io"
	}
	
	// If first part contains dot, it's likely a registry URL
	if strings.Contains(parts[0], ".") || strings.Contains(parts[0], ":") {
		return parts[0]
	}
	
	// Default to Docker Hub
	return "docker.io"
}

// CreateDockerConfigSecret creates Docker config for imagePullSecrets
func (m *Manager) CreateDockerConfigSecret(registry *Registry) (map[string]interface{}, error) {
	// Create Docker config format
	auth := base64.StdEncoding.EncodeToString([]byte(registry.Username + ":" + registry.Password))
	
	config := map[string]interface{}{
		"auths": map[string]interface{}{
			registry.RegistryURL: map[string]interface{}{
				"username": registry.Username,
				"password": registry.Password,
				"email":    registry.Email,
				"auth":     auth,
			},
		},
	}
	
	return config, nil
}