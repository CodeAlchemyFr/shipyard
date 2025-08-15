package manifests

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

const secretsTemplate = `{{- if .Secrets }}
apiVersion: v1
kind: Secret
metadata:
  name: {{ .App.Name }}-secrets
  labels:
    app: {{ .App.Name }}
    managed-by: shipyard
type: Opaque
data:
{{- range $key, $value := .SecretsBase64 }}
  {{ $key }}: {{ $value }}
{{- end }}
{{- end }}
`

// generateSecrets creates the secrets.yaml file with base64 encoded values
func (g *Generator) generateSecrets(appDir string) error {
	// Skip if no secrets defined
	if len(g.config.Secrets) == 0 {
		fmt.Printf("ℹ️  No secrets defined for %s, skipping secrets.yaml\n", g.config.App.Name)
		return nil
	}

	// Convert secrets to base64
	secretsBase64 := make(map[string]string)
	for key, value := range g.config.Secrets {
		encoded := base64.StdEncoding.EncodeToString([]byte(value))
		secretsBase64[key] = encoded
	}

	// Create template data with base64 encoded secrets
	templateData := struct {
		*Config
		SecretsBase64 map[string]string
	}{
		Config:        g.config,
		SecretsBase64: secretsBase64,
	}

	tmpl, err := template.New("secrets").Parse(secretsTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse secrets template: %w", err)
	}

	filePath := filepath.Join(appDir, "secrets.yaml")
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create secrets file %s: %w", filePath, err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, templateData); err != nil {
		return fmt.Errorf("failed to execute secrets template: %w", err)
	}

	fmt.Printf("🔐 Generated: %s (with %d secrets base64 encoded)\n", filePath, len(g.config.Secrets))
	return nil
}