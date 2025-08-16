package manifests

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/shipyard/cli/pkg/database"
)

// DeploymentVersion represents a deployment version (keeping same struct)
type DeploymentVersion struct {
	ID          int64             `json:"id"`
	Version     string            `json:"version"`
	Image       string            `json:"image"`
	ImageTag    string            `json:"image_tag"`
	ImageHash   string            `json:"image_hash"`
	Timestamp   time.Time         `json:"timestamp"`
	Config      *Config           `json:"config"`
	ConfigHash  string            `json:"config_hash"`
	Status      string            `json:"status"` // pending, success, failed
	RollbackTo  string            `json:"rollback_to,omitempty"`
	CompletedAt *time.Time        `json:"completed_at,omitempty"`
	ErrorMessage string           `json:"error_message,omitempty"`
}

// VersionManager handles deployment versioning with SQLite
type VersionManager struct {
	appName string
	db      *database.DB
}

// NewVersionManager creates a new version manager with SQLite backend
func NewVersionManager(appName string) *VersionManager {
	db, err := database.NewDB()
	if err != nil {
		// In a real implementation, we'd handle this error better
		panic(fmt.Sprintf("Failed to initialize database: %v", err))
	}

	return &VersionManager{
		appName: appName,
		db:      db,
	}
}

// GenerateVersion creates a new deployment version
func (vm *VersionManager) GenerateVersion(config *Config) (*DeploymentVersion, error) {
	// Generate version based on timestamp
	now := time.Now()
	version := fmt.Sprintf("v%d", now.Unix())
	
	// Extract image tag from image
	imageTag := extractImageTag(config.App.Image)
	
	// Generate image hash (simplified - in real use, would fetch from registry)
	imageHash := vm.generateImageHash(config.App.Image)
	
	// Generate config hash
	configHash, err := vm.generateConfigHash(config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate config hash: %w", err)
	}
	
	deployVersion := &DeploymentVersion{
		Version:    version,
		Image:      config.App.Image,
		ImageTag:   imageTag,
		ImageHash:  imageHash,
		Timestamp:  now,
		Config:     config,
		ConfigHash: configHash,
		Status:     "pending",
	}
	
	return deployVersion, nil
}

// SaveVersion saves a deployment version to the database
func (vm *VersionManager) SaveVersion(version *DeploymentVersion) error {
	// Get or create app
	appID, err := vm.db.GetOrCreateApp(vm.appName)
	if err != nil {
		return fmt.Errorf("failed to get/create app: %w", err)
	}

	// Serialize config to JSON
	configJSON, err := json.Marshal(version.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Insert deployment record
	query := `
		INSERT INTO deployments (
			app_id, version, image, image_tag, image_hash,
			config_json, config_hash, status, rollback_to_version,
			deployed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := vm.db.GetConnection().Exec(
		query,
		appID,
		version.Version,
		version.Image,
		version.ImageTag,
		version.ImageHash,
		string(configJSON),
		version.ConfigHash,
		version.Status,
		version.RollbackTo,
		version.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("failed to save deployment version: %w", err)
	}

	// Get the inserted ID
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get deployment ID: %w", err)
	}

	version.ID = id
	fmt.Printf("📝 Saved deployment version: %s (ID: %d)\n", version.Version, version.ID)
	return nil
}

// LoadVersionHistory loads deployment history from database
func (vm *VersionManager) LoadVersionHistory() ([]*DeploymentVersion, error) {
	query := `
		SELECT 
			id, version, image, image_tag, image_hash,
			config_json, config_hash, status, rollback_to_version,
			deployed_at, completed_at, error_message
		FROM deployment_history 
		WHERE app_name = ?
		ORDER BY deployed_at DESC
		LIMIT 50`

	rows, err := vm.db.GetConnection().Query(query, vm.appName)
	if err != nil {
		return nil, fmt.Errorf("failed to query deployment history: %w", err)
	}
	defer rows.Close()

	var versions []*DeploymentVersion
	for rows.Next() {
		var version DeploymentVersion
		var configJSON string
		var rollbackTo *string
		var completedAt *time.Time
		var errorMessage *string

		err := rows.Scan(
			&version.ID,
			&version.Version,
			&version.Image,
			&version.ImageTag,
			&version.ImageHash,
			&configJSON,
			&version.ConfigHash,
			&version.Status,
			&rollbackTo,
			&version.Timestamp,
			&completedAt,
			&errorMessage,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deployment row: %w", err)
		}

		// Deserialize config
		var config Config
		if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
		version.Config = &config

		// Handle nullable fields
		if rollbackTo != nil {
			version.RollbackTo = *rollbackTo
		}
		if completedAt != nil {
			version.CompletedAt = completedAt
		}
		if errorMessage != nil {
			version.ErrorMessage = *errorMessage
		}

		versions = append(versions, &version)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating deployment rows: %w", err)
	}

	return versions, nil
}

// UpdateVersionStatus updates the status of a deployment version
func (vm *VersionManager) UpdateVersionStatus(version, status string) error {
	now := time.Now()
	
	query := `
		UPDATE deployments 
		SET status = ?, completed_at = ?
		WHERE version = ? AND app_id = (SELECT id FROM apps WHERE name = ?)`

	result, err := vm.db.GetConnection().Exec(query, status, now, version, vm.appName)
	if err != nil {
		return fmt.Errorf("failed to update deployment status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("deployment version %s not found", version)
	}

	return nil
}

// UpdateVersionError updates the error message for a failed deployment
func (vm *VersionManager) UpdateVersionError(version, errorMessage string) error {
	query := `
		UPDATE deployments 
		SET status = 'failed', completed_at = ?, error_message = ?
		WHERE version = ? AND app_id = (SELECT id FROM apps WHERE name = ?)`

	_, err := vm.db.GetConnection().Exec(query, time.Now(), errorMessage, version, vm.appName)
	if err != nil {
		return fmt.Errorf("failed to update deployment error: %w", err)
	}

	return nil
}

// GetLatestSuccessfulVersion returns the latest successful deployment
func (vm *VersionManager) GetLatestSuccessfulVersion() (*DeploymentVersion, error) {
	query := `
		SELECT 
			id, version, image, image_tag, image_hash,
			config_json, config_hash, status, rollback_to_version,
			deployed_at, completed_at, error_message
		FROM deployment_history 
		WHERE app_name = ? AND status = 'success'
		ORDER BY deployed_at DESC
		LIMIT 1`

	var version DeploymentVersion
	var configJSON string
	var rollbackTo *string
	var completedAt *time.Time
	var errorMessage *string

	err := vm.db.GetConnection().QueryRow(query, vm.appName).Scan(
		&version.ID,
		&version.Version,
		&version.Image,
		&version.ImageTag,
		&version.ImageHash,
		&configJSON,
		&version.ConfigHash,
		&version.Status,
		&rollbackTo,
		&version.Timestamp,
		&completedAt,
		&errorMessage,
	)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, fmt.Errorf("no successful deployment found")
		}
		return nil, fmt.Errorf("failed to get latest successful version: %w", err)
	}

	// Deserialize config
	var config Config
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	version.Config = &config

	// Handle nullable fields
	if rollbackTo != nil {
		version.RollbackTo = *rollbackTo
	}
	if completedAt != nil {
		version.CompletedAt = completedAt
	}
	if errorMessage != nil {
		version.ErrorMessage = *errorMessage
	}

	return &version, nil
}

// GetVersionByIdentifier returns a version by version string or image tag
func (vm *VersionManager) GetVersionByIdentifier(identifier string) (*DeploymentVersion, error) {
	query := `
		SELECT 
			id, version, image, image_tag, image_hash,
			config_json, config_hash, status, rollback_to_version,
			deployed_at, completed_at, error_message
		FROM deployment_history 
		WHERE app_name = ? AND (version = ? OR image_tag = ?)
		ORDER BY deployed_at DESC
		LIMIT 1`

	var version DeploymentVersion
	var configJSON string
	var rollbackTo *string
	var completedAt *time.Time
	var errorMessage *string

	err := vm.db.GetConnection().QueryRow(query, vm.appName, identifier, identifier).Scan(
		&version.ID,
		&version.Version,
		&version.Image,
		&version.ImageTag,
		&version.ImageHash,
		&configJSON,
		&version.ConfigHash,
		&version.Status,
		&rollbackTo,
		&version.Timestamp,
		&completedAt,
		&errorMessage,
	)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, fmt.Errorf("version %s not found", identifier)
		}
		return nil, fmt.Errorf("failed to get version by identifier: %w", err)
	}

	// Deserialize config
	var config Config
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	version.Config = &config

	// Handle nullable fields
	if rollbackTo != nil {
		version.RollbackTo = *rollbackTo
	}
	if completedAt != nil {
		version.CompletedAt = completedAt
	}
	if errorMessage != nil {
		version.ErrorMessage = *errorMessage
	}

	return &version, nil
}

// ListVersions returns formatted version history
func (vm *VersionManager) ListVersions(limit int) ([]*DeploymentVersion, error) {
	query := `
		SELECT 
			id, version, image, image_tag, image_hash,
			config_json, config_hash, status, rollback_to_version,
			deployed_at, completed_at, error_message
		FROM deployment_history 
		WHERE app_name = ?
		ORDER BY deployed_at DESC`
	
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := vm.db.GetConnection().Query(query, vm.appName)
	if err != nil {
		return nil, fmt.Errorf("failed to query deployment versions: %w", err)
	}
	defer rows.Close()

	var versions []*DeploymentVersion
	for rows.Next() {
		var version DeploymentVersion
		var configJSON string
		var rollbackTo *string
		var completedAt *time.Time
		var errorMessage *string

		err := rows.Scan(
			&version.ID,
			&version.Version,
			&version.Image,
			&version.ImageTag,
			&version.ImageHash,
			&configJSON,
			&version.ConfigHash,
			&version.Status,
			&rollbackTo,
			&version.Timestamp,
			&completedAt,
			&errorMessage,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan version row: %w", err)
		}

		// We don't need to deserialize config for listing
		// Just set it to nil to save memory
		version.Config = nil

		// Handle nullable fields
		if rollbackTo != nil {
			version.RollbackTo = *rollbackTo
		}
		if completedAt != nil {
			version.CompletedAt = completedAt
		}
		if errorMessage != nil {
			version.ErrorMessage = *errorMessage
		}

		versions = append(versions, &version)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating version rows: %w", err)
	}

	return versions, nil
}

// Close closes the database connection
func (vm *VersionManager) Close() error {
	return vm.db.Close()
}

// generateImageHash creates a hash for the image (simplified)
func (vm *VersionManager) generateImageHash(image string) string {
	hash := sha256.Sum256([]byte(image))
	return fmt.Sprintf("%x", hash)[:12]
}

// generateConfigHash creates a hash for the configuration
func (vm *VersionManager) generateConfigHash(config *Config) (string, error) {
	// Create a copy without dynamic fields
	configCopy := *config
	
	// Convert to JSON for consistent hashing
	data, err := json.Marshal(configCopy)
	if err != nil {
		return "", err
	}
	
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)[:12], nil
}

// extractImageTag extracts tag from image name
func extractImageTag(image string) string {
	parts := strings.Split(image, ":")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return "latest"
}

// DeleteApp removes an application and all its data from the database
func (vm *VersionManager) DeleteApp() error {
	// Start a transaction
	tx, err := vm.db.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Get app ID
	var appID int64
	err = tx.QueryRow("SELECT id FROM apps WHERE name = ?", vm.appName).Scan(&appID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			// App doesn't exist, nothing to delete
			return nil
		}
		return fmt.Errorf("failed to get app ID: %w", err)
	}

	// Delete related records (foreign keys should handle this, but being explicit)
	tables := []string{
		"deployments", 
		"domains", 
		"metrics", 
		"health_checks", 
		"events",
	}

	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s WHERE app_id = ?", table)
		if _, err := tx.Exec(query, appID); err != nil {
			return fmt.Errorf("failed to delete from %s: %w", table, err)
		}
	}

	// Finally delete the app itself
	if _, err := tx.Exec("DELETE FROM apps WHERE id = ?", appID); err != nil {
		return fmt.Errorf("failed to delete app: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}