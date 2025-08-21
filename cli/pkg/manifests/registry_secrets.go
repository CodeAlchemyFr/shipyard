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

// GenerateRegistrySecrets generates Kubernetes secrets for container registries with interactive or auto selection
func (g *Generator) GenerateRegistrySecrets(appDir string) ([]string, error) {
	// Get registry manager
	manager, err := registry.NewManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize registry manager: %w", err)
	}
	defer manager.Close()

	var selectedRegistries []*registry.Registry
	
	// Always use interactive selection
	selectedRegistries, err = manager.SelectRegistriesInteractive(g.config.App.Image)
	if err != nil {
		return nil, fmt.Errorf("failed to select registries interactively: %w", err)
	}

	if len(selectedRegistries) == 0 {
		// No registries selected, return empty list
		return []string{}, nil
	}

	var secretNames []string

	// Generate secrets for each selected registry
	for i, selectedRegistry := range selectedRegistries {
		secretName, err := g.generateSingleRegistrySecret(appDir, selectedRegistry, i)
		if err != nil {
			return nil, fmt.Errorf("failed to generate secret for registry %s: %w", selectedRegistry.RegistryURL, err)
		}
		secretNames = append(secretNames, secretName)
	}

	return secretNames, nil
}

// generateSingleRegistrySecret generates a secret for a single registry
func (g *Generator) generateSingleRegistrySecret(appDir string, selectedRegistry *registry.Registry, index int) (string, error) {
	// Get registry manager
	manager, err := registry.NewManager()
	if err != nil {
		return "", fmt.Errorf("failed to initialize registry manager: %w", err)
	}
	defer manager.Close()

	// Create Docker config secret
	dockerConfig, err := manager.CreateDockerConfigSecret(selectedRegistry)
	if err != nil {
		return "", fmt.Errorf("failed to create docker config: %w", err)
	}

	// Convert to JSON
	dockerConfigBytes, err := json.Marshal(dockerConfig)
	if err != nil {
		return "", fmt.Errorf("failed to marshal docker config: %w", err)
	}

	// Base64 encode for Kubernetes secret
	dockerConfigB64 := base64.StdEncoding.EncodeToString(dockerConfigBytes)

	// Generate unique secret name
	var secretName string
	if index == 0 {
		secretName = fmt.Sprintf("%s-registry-secret", g.config.App.Name)
	} else {
		secretName = fmt.Sprintf("%s-registry-secret-%d", g.config.App.Name, index+1)
	}

	// Create secret data
	secretData := RegistrySecretData{
		SecretName:        secretName,
		AppName:          g.config.App.Name,
		DockerConfigJSON: dockerConfigB64,
	}

	// Parse template
	tmpl, err := template.New("registry-secret").Parse(registrySecretTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse registry secret template: %w", err)
	}

	// Create secret file with unique name
	var fileName string
	if index == 0 {
		fileName = "registry-secret.yaml"
	} else {
		fileName = fmt.Sprintf("registry-secret-%d.yaml", index+1)
	}
	
	filePath := filepath.Join(appDir, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create registry secret file %s: %w", filePath, err)
	}
	defer file.Close()

	// Execute template
	if err := tmpl.Execute(file, secretData); err != nil {
		return "", fmt.Errorf("failed to execute registry secret template: %w", err)
	}

	
	return secretName, nil
}