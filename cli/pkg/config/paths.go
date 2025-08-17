package config

import (
	"os"
	"path/filepath"
)

// GetShipyardDir returns the global Shipyard directory (~/.shipyard)
func GetShipyardDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	
	shipyardDir := filepath.Join(homeDir, ".shipyard")
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(shipyardDir, 0755); err != nil {
		return "", err
	}
	
	return shipyardDir, nil
}

// GetManifestsDir returns the manifests directory (~/.shipyard/manifests)
func GetManifestsDir() (string, error) {
	shipyardDir, err := GetShipyardDir()
	if err != nil {
		return "", err
	}
	
	manifestsDir := filepath.Join(shipyardDir, "manifests")
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(manifestsDir, 0755); err != nil {
		return "", err
	}
	
	return manifestsDir, nil
}

// GetDatabasePath returns the database file path (~/.shipyard/manifests/shipyard.db)
func GetDatabasePath() (string, error) {
	manifestsDir, err := GetManifestsDir()
	if err != nil {
		return "", err
	}
	
	return filepath.Join(manifestsDir, "shipyard.db"), nil
}

// GetAppsDir returns the apps manifests directory (~/.shipyard/manifests/apps)
func GetAppsDir() (string, error) {
	manifestsDir, err := GetManifestsDir()
	if err != nil {
		return "", err
	}
	
	appsDir := filepath.Join(manifestsDir, "apps")
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(appsDir, 0755); err != nil {
		return "", err
	}
	
	return appsDir, nil
}

// GetSharedDir returns the shared manifests directory (~/.shipyard/manifests/shared)
func GetSharedDir() (string, error) {
	manifestsDir, err := GetManifestsDir()
	if err != nil {
		return "", err
	}
	
	sharedDir := filepath.Join(manifestsDir, "shared")
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(sharedDir, 0755); err != nil {
		return "", err
	}
	
	return sharedDir, nil
}