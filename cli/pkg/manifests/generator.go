package manifests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	// Handle CI/CD mode - check if this is initial deployment or update
	if g.config.CICD.Enabled {
		return g.handleCICDDeployment()
	}
	
	return g.generateStandardManifests()
}

// handleCICDDeployment manages CI/CD enabled deployments
func (g *Generator) handleCICDDeployment() error {
	appDir := filepath.Join(g.outputDir, "apps", g.config.App.Name)
	deploymentFile := filepath.Join(appDir, "deployment.yaml")
	
	// Check if this is the first deployment
	if _, err := os.Stat(deploymentFile); os.IsNotExist(err) {
		fmt.Printf("🚀 First deployment - creating manifests with real image, then switching to ${IMAGE_TAG}\n")
		
		// First: Create manifests with real image
		if err := g.generateStandardManifests(); err != nil {
			return err
		}
		
		// Note: Image will be replaced with ${IMAGE_TAG} after successful deployment
		return nil
	} else {
		fmt.Printf("🔄 CI/CD enabled - manifests already exist with ${IMAGE_TAG} placeholder\n")
		return nil
	}
}

// generateStandardManifests generates manifests for normal Shipyard deployment
func (g *Generator) generateStandardManifests() error {
	// Generate namespace if specified
	if err := g.generateNamespace(); err != nil {
		return err
	}
	
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

// updateDeploymentForCICD replaces the real image with ${IMAGE_TAG} placeholder
func (g *Generator) updateDeploymentForCICD() error {
	appDir := filepath.Join(g.outputDir, "apps", g.config.App.Name)
	deploymentFile := filepath.Join(appDir, "deployment.yaml")
	
	// Read current deployment file
	content, err := os.ReadFile(deploymentFile)
	if err != nil {
		return fmt.Errorf("failed to read deployment file: %w", err)
	}
	
	// Replace the image line with ${IMAGE_TAG}
	updatedContent := strings.ReplaceAll(string(content), 
		fmt.Sprintf("image: %s", g.config.App.Image),
		"image: ${IMAGE_TAG}")
	
	// Write updated content back
	err = os.WriteFile(deploymentFile, []byte(updatedContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write updated deployment file: %w", err)
	}
	
	fmt.Printf("🔄 Updated deployment.yaml: %s → ${IMAGE_TAG}\n", g.config.App.Image)
	return nil
}

// generateNamespace creates a namespace manifest if needed
func (g *Generator) generateNamespace() error {
	targetNamespace := g.config.App.GetNamespace()
	
	// Skip if using default namespace
	if targetNamespace == "default" {
		return nil
	}
	
	sharedDir := filepath.Join(g.outputDir, "shared")
	if err := os.MkdirAll(sharedDir, 0755); err != nil {
		return fmt.Errorf("failed to create shared directory %s: %w", sharedDir, err)
	}
	
	namespacePath := filepath.Join(sharedDir, fmt.Sprintf("namespace-%s.yaml", targetNamespace))
	
	// Check if namespace file already exists
	if _, err := os.Stat(namespacePath); err == nil {
		return nil // Already exists
	}
	
	namespaceContent := fmt.Sprintf(`apiVersion: v1
kind: Namespace
metadata:
  name: %s
  labels:
    managed-by: shipyard
    app: %s
`, targetNamespace, g.config.App.Name)
	
	err := os.WriteFile(namespacePath, []byte(namespaceContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to create namespace file: %w", err)
	}
	
	fmt.Printf("🏗️  Generated namespace: %s (for app: %s)\n", namespacePath, g.config.App.Name)
	return nil
}

// UpdateDeploymentForCICD replaces the real image with ${IMAGE_TAG} placeholder (public method)
func (g *Generator) UpdateDeploymentForCICD() error {
	return g.updateDeploymentForCICD()
}

