package manifests

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/shipyard/cli/pkg/registry"
)

const registrySecretTemplate = `apiVersion: v1
kind: Secret
metadata:
  name: {{ .SecretName }}
  labels:
    app: {{ .AppName }}
    managed-by: shipyard
type: kubernetes.io/dockerconfigjson
data:
  .dockerconfigjson: {{ .DockerConfigJSON }}
`

// RegistrySecretData represents data for generating registry secrets
type RegistrySecretData struct {
	SecretName         string
	AppName           string
	DockerConfigJSON  string
}

// GenerateRegistrySecrets generates Kubernetes secrets for container registries
func (g *Generator) GenerateRegistrySecrets(appDir string) ([]string, error) {
	// Get registry manager
	manager, err := registry.NewManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize registry manager: %w", err)
	}
	defer manager.Close()

	// Get registry for the app image
	imageRegistry, err := manager.GetRegistryForImage(g.config.App.Image)
	if err != nil {
		// No registry needed, return empty list
		return []string{}, nil
	}

	// Create Docker config secret
	dockerConfig, err := manager.CreateDockerConfigSecret(imageRegistry)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker config: %w", err)
	}

	// Convert to JSON
	dockerConfigBytes, err := json.Marshal(dockerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal docker config: %w", err)
	}

	// Base64 encode for Kubernetes secret
	dockerConfigB64 := base64.StdEncoding.EncodeToString(dockerConfigBytes)

	// Generate secret name
	secretName := fmt.Sprintf("%s-registry-secret", g.config.App.Name)

	// Create secret data
	secretData := RegistrySecretData{
		SecretName:        secretName,
		AppName:          g.config.App.Name,
		DockerConfigJSON: dockerConfigB64,
	}

	// Parse template
	tmpl, err := template.New("registry-secret").Parse(registrySecretTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse registry secret template: %w", err)
	}

	// Create secret file
	filePath := filepath.Join(appDir, "registry-secret.yaml")
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create registry secret file %s: %w", filePath, err)
	}
	defer file.Close()

	// Execute template
	if err := tmpl.Execute(file, secretData); err != nil {
		return nil, fmt.Errorf("failed to execute registry secret template: %w", err)
	}

	fmt.Printf("🔐 Generated: %s (registry: %s)\n", filePath, imageRegistry.RegistryURL)
	
	return []string{secretName}, nil
}