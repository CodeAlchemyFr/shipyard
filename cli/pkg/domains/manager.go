package domains

import (
	"fmt"
	"strings"
	"time"

	"github.com/shipyard/cli/pkg/database"
)

// Domain represents a domain configuration
type Domain struct {
	ID         int64     `json:"id"`
	AppID      int64     `json:"app_id"`
	AppName    string    `json:"app_name"`
	Hostname   string    `json:"hostname"`
	BaseDomain string    `json:"base_domain"`
	Path       string    `json:"path"`
	SSLEnabled bool      `json:"ssl_enabled"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// DomainGroup represents domains grouped by base domain
type DomainGroup struct {
	BaseDomain string    `json:"base_domain"`
	Domains    []Domain  `json:"domains"`
}

// Manager handles domain operations
type Manager struct {
	db *database.DB
}

// NewManager creates a new domain manager
func NewManager() (*Manager, error) {
	db, err := database.NewDB()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return &Manager{
		db: db,
	}, nil
}

// AddDomain adds a new domain for an app
func (m *Manager) AddDomain(appName, hostname string) error {
	// Get or create app
	appID, err := m.db.GetOrCreateApp(appName)
	if err != nil {
		return fmt.Errorf("failed to get/create app: %w", err)
	}

	// Extract base domain
	baseDomain := extractBaseDomain(hostname)

	// Check if hostname already exists
	var existingApp string
	err = m.db.GetConnection().QueryRow(`
		SELECT a.name 
		FROM domains d 
		JOIN apps a ON d.app_id = a.id 
		WHERE d.hostname = ?`, hostname).Scan(&existingApp)
	
	if err == nil {
		if existingApp == appName {
			return fmt.Errorf("domain %s already exists for app %s", hostname, appName)
		}
		return fmt.Errorf("domain %s is already used by app %s", hostname, existingApp)
	}

	// Insert new domain
	query := `
		INSERT INTO domains (app_id, hostname, base_domain, path, ssl_enabled)
		VALUES (?, ?, ?, ?, ?)`

	_, err = m.db.GetConnection().Exec(query, appID, hostname, baseDomain, "/", true)
	if err != nil {
		return fmt.Errorf("failed to add domain: %w", err)
	}

	return nil
}

// RemoveDomain removes a domain from an app
func (m *Manager) RemoveDomain(appName, hostname string) error {
	query := `
		DELETE FROM domains 
		WHERE hostname = ? AND app_id = (SELECT id FROM apps WHERE name = ?)`

	result, err := m.db.GetConnection().Exec(query, hostname, appName)
	if err != nil {
		return fmt.Errorf("failed to remove domain: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("domain %s not found for app %s", hostname, appName)
	}

	return nil
}

// GetDomainsForApp returns all domains for a specific app
func (m *Manager) GetDomainsForApp(appName string) ([]Domain, error) {
	query := `
		SELECT id, hostname, base_domain, path, ssl_enabled, created_at, updated_at
		FROM domain_overview
		WHERE app_name = ?
		ORDER BY hostname`

	rows, err := m.db.GetConnection().Query(query, appName)
	if err != nil {
		return nil, fmt.Errorf("failed to query domains: %w", err)
	}
	defer rows.Close()

	var domains []Domain
	for rows.Next() {
		var domain Domain
		err := rows.Scan(
			&domain.ID,
			&domain.Hostname,
			&domain.BaseDomain,
			&domain.Path,
			&domain.SSLEnabled,
			&domain.CreatedAt,
			&domain.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan domain row: %w", err)
		}
		domain.AppName = appName
		// We need to get app_id for the domain struct - let's get it from the apps table
		err = m.db.GetConnection().QueryRow("SELECT id FROM apps WHERE name = ?", appName).Scan(&domain.AppID)
		if err != nil {
			return nil, fmt.Errorf("failed to get app_id for app %s: %w", appName, err)
		}
		domains = append(domains, domain)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating domain rows: %w", err)
	}

	return domains, nil
}

// GetAllDomains returns all domains grouped by base domain
func (m *Manager) GetAllDomains() ([]DomainGroup, error) {
	query := `
		SELECT d.id, a.name as app_name, d.hostname, d.base_domain, d.path, d.ssl_enabled, d.created_at, d.updated_at, a.id as app_id
		FROM domains d
		JOIN apps a ON d.app_id = a.id
		ORDER BY d.base_domain, d.hostname`

	rows, err := m.db.GetConnection().Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all domains: %w", err)
	}
	defer rows.Close()

	domainMap := make(map[string][]Domain)
	for rows.Next() {
		var domain Domain
		err := rows.Scan(
			&domain.ID,
			&domain.AppName,
			&domain.Hostname,
			&domain.BaseDomain,
			&domain.Path,
			&domain.SSLEnabled,
			&domain.CreatedAt,
			&domain.UpdatedAt,
			&domain.AppID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan domain row: %w", err)
		}

		domainMap[domain.BaseDomain] = append(domainMap[domain.BaseDomain], domain)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating domain rows: %w", err)
	}

	// Convert map to slice
	var groups []DomainGroup
	for baseDomain, domains := range domainMap {
		groups = append(groups, DomainGroup{
			BaseDomain: baseDomain,
			Domains:    domains,
		})
	}

	return groups, nil
}

// GetDomainsByBaseDomain returns all domains for a specific base domain
func (m *Manager) GetDomainsByBaseDomain(baseDomain string) ([]Domain, error) {
	query := `
		SELECT d.id, a.name as app_name, d.hostname, d.base_domain, d.path, d.ssl_enabled, d.created_at, d.updated_at, a.id as app_id
		FROM domains d
		JOIN apps a ON d.app_id = a.id
		WHERE d.base_domain = ?
		ORDER BY d.hostname`

	rows, err := m.db.GetConnection().Query(query, baseDomain)
	if err != nil {
		return nil, fmt.Errorf("failed to query domains by base domain: %w", err)
	}
	defer rows.Close()

	var domains []Domain
	for rows.Next() {
		var domain Domain
		err := rows.Scan(
			&domain.ID,
			&domain.AppName,
			&domain.Hostname,
			&domain.BaseDomain,
			&domain.Path,
			&domain.SSLEnabled,
			&domain.CreatedAt,
			&domain.UpdatedAt,
			&domain.AppID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan domain row: %w", err)
		}

		domains = append(domains, domain)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating domain rows: %w", err)
	}

	return domains, nil
}

// SyncDomainsFromConfig syncs domains from paas.yaml config to database
func (m *Manager) SyncDomainsFromConfig(appName string, configDomains []string) error {
	// Get current domains from database
	currentDomains, err := m.GetDomainsForApp(appName)
	if err != nil {
		return fmt.Errorf("failed to get current domains: %w", err)
	}

	// Create maps for comparison
	currentMap := make(map[string]bool)
	for _, domain := range currentDomains {
		currentMap[domain.Hostname] = true
	}

	configMap := make(map[string]bool)
	for _, domain := range configDomains {
		configMap[domain] = true
	}

	// Add new domains from config
	for _, domain := range configDomains {
		if !currentMap[domain] {
			if err := m.AddDomain(appName, domain); err != nil {
				return fmt.Errorf("failed to add domain %s: %w", domain, err)
			}
			fmt.Printf("➕ Added domain: %s\n", domain)
		}
	}

	// Remove domains not in config
	for _, domain := range currentDomains {
		if !configMap[domain.Hostname] {
			if err := m.RemoveDomain(appName, domain.Hostname); err != nil {
				return fmt.Errorf("failed to remove domain %s: %w", domain.Hostname, err)
			}
			fmt.Printf("➖ Removed domain: %s\n", domain.Hostname)
		}
	}

	return nil
}

// GetBaseDomains returns all unique base domains
func (m *Manager) GetBaseDomains() ([]string, error) {
	query := `SELECT DISTINCT base_domain FROM domains ORDER BY base_domain`

	rows, err := m.db.GetConnection().Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query base domains: %w", err)
	}
	defer rows.Close()

	var baseDomains []string
	for rows.Next() {
		var baseDomain string
		if err := rows.Scan(&baseDomain); err != nil {
			return nil, fmt.Errorf("failed to scan base domain: %w", err)
		}
		baseDomains = append(baseDomains, baseDomain)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating base domain rows: %w", err)
	}

	return baseDomains, nil
}

// Close closes the database connection
func (m *Manager) Close() error {
	return m.db.Close()
}

// extractBaseDomain extracts the base domain from a hostname
func extractBaseDomain(hostname string) string {
	parts := strings.Split(hostname, ".")
	if len(parts) >= 2 {
		return strings.Join(parts[len(parts)-2:], ".")
	}
	return hostname
}