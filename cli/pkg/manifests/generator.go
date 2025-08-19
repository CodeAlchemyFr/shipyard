package manifests

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shipyard/cli/pkg/config"
	"github.com/shipyard/cli/pkg/k8s"
)

// Generator handles the generation of Kubernetes manifests
type Generator struct {
	config           *Config
	outputDir        string
	version          *DeploymentVersion // Add version tracking
	imagePullSecrets []string           // Registry secrets for private images
	interactiveMode  bool               // Whether to use interactive registry selection
}

// NewGenerator creates a new manifest generator
func NewGenerator(cfg *Config) *Generator {
	// Get manifests directory from global config
	manifestsDir, err := config.GetManifestsDir()
	if err != nil {
		// Fallback to local directory on error
		manifestsDir = "manifests"
	}
	
	return &Generator{
		config:           cfg,
		outputDir:        manifestsDir,
		version:          nil,
		imagePullSecrets: []string{},
		interactiveMode:  true, // Default to interactive mode
	}
}

// NewGeneratorWithMode creates a new manifest generator with specified interaction mode
func NewGeneratorWithMode(cfg *Config, interactive bool) *Generator {
	// Get manifests directory from global config
	manifestsDir, err := config.GetManifestsDir()
	if err != nil {
		// Fallback to local directory on error
		manifestsDir = "manifests"
	}
	
	return &Generator{
		config:           cfg,
		outputDir:        manifestsDir,
		version:          nil,
		imagePullSecrets: []string{},
		interactiveMode:  interactive,
	}
}

// NewGeneratorWithVersion creates a new manifest generator with version tracking
func NewGeneratorWithVersion(cfg *Config, version *DeploymentVersion) *Generator {
	// Get manifests directory from global config
	manifestsDir, err := config.GetManifestsDir()
	if err != nil {
		// Fallback to local directory on error
		manifestsDir = "manifests"
	}
	
	return &Generator{
		config:           cfg,
		outputDir:        manifestsDir,
		version:          version,
		imagePullSecrets: []string{},
		interactiveMode:  true, // Default to interactive mode
	}
}

// NewGeneratorWithVersionAndMode creates a new manifest generator with version tracking and interaction mode
func NewGeneratorWithVersionAndMode(cfg *Config, version *DeploymentVersion, interactive bool) *Generator {
	// Get manifests directory from global config
	manifestsDir, err := config.GetManifestsDir()
	if err != nil {
		// Fallback to local directory on error
		manifestsDir = "manifests"
	}
	
	return &Generator{
		config:           cfg,
		outputDir:        manifestsDir,
		version:          version,
		imagePullSecrets: []string{},
		interactiveMode:  interactive,
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

// CreateK8sClient creates a new Kubernetes client
func CreateK8sClient() (*k8s.Client, error) {
	return k8s.NewClient()
}

// DeleteManifestsFromDirectory deletes all Kubernetes resources from manifest files in a directory
func DeleteManifestsFromDirectory(client *k8s.Client, directory string) error {
	if client == nil {
		return fmt.Errorf("client is nil")
	}
	
	// Extract app name from directory path
	appName := filepath.Base(directory)
	
	// Use the client's delete method
	return client.DeleteManifests(appName)
}

