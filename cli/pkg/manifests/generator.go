package manifests

import (
	"fmt"
	"os"
	"path/filepath"
)

// Generator handles the generation of Kubernetes manifests
type Generator struct {
	config           *Config
	outputDir        string
	version          *DeploymentVersion // Add version tracking
	imagePullSecrets []string           // Registry secrets for private images
}

// NewGenerator creates a new manifest generator
func NewGenerator(config *Config) *Generator {
	return &Generator{
		config:           config,
		outputDir:        "manifests", // Base directory for all manifests
		version:          nil,
		imagePullSecrets: []string{},
	}
}

// NewGeneratorWithVersion creates a new manifest generator with version tracking
func NewGeneratorWithVersion(config *Config, version *DeploymentVersion) *Generator {
	return &Generator{
		config:           config,
		outputDir:        "manifests", // Base directory for all manifests
		version:          version,
		imagePullSecrets: []string{},
	}
}

// GenerateAppManifests creates all manifests for an application
func (g *Generator) GenerateAppManifests() error {
	appDir := filepath.Join(g.outputDir, "apps", g.config.App.Name)
	
	// Create app directory
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return fmt.Errorf("failed to create app directory %s: %w", appDir, err)
	}

	fmt.Printf("📁 Created directory: %s\n", appDir)

	// Generate registry secrets if needed
	imagePullSecrets, err := g.GenerateRegistrySecrets(appDir)
	if err != nil {
		return fmt.Errorf("failed to generate registry secrets: %w", err)
	}
	g.imagePullSecrets = imagePullSecrets

	// Generate deployment.yaml
	if err := g.generateDeployment(appDir); err != nil {
		return fmt.Errorf("failed to generate deployment: %w", err)
	}

	// Generate secrets.yaml (with base64 encoded values)
	if err := g.generateSecrets(appDir); err != nil {
		return fmt.Errorf("failed to generate secrets: %w", err)
	}

	// Generate service.yaml
	if err := g.generateService(appDir); err != nil {
		return fmt.Errorf("failed to generate service: %w", err)
	}

	return nil
}

// UpdateIngressManifests updates shared ingress files by domain (now uses database)
func (g *Generator) UpdateIngressManifests() error {
	// Use new database-based ingress generation
	return g.UpdateIngressFromDatabase(g.config.App.Name)
}

