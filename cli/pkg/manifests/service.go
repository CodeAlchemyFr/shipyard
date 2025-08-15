package manifests

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

const serviceTemplate = `apiVersion: v1
kind: Service
metadata:
  name: {{ .App.Name }}
  labels:
    app: {{ .App.Name }}
    managed-by: shipyard
spec:
  type: ClusterIP
  ports:
  - port: 80
    targetPort: {{ .App.Port }}
    protocol: TCP
    name: http
  selector:
    app: {{ .App.Name }}
`

// generateService creates the service.yaml file for an application
func (g *Generator) generateService(appDir string) error {
	tmpl, err := template.New("service").Parse(serviceTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse service template: %w", err)
	}

	filePath := filepath.Join(appDir, "service.yaml")
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create service file %s: %w", filePath, err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, g.config); err != nil {
		return fmt.Errorf("failed to execute service template: %w", err)
	}

	fmt.Printf("🔗 Generated: %s\n", filePath)
	return nil
}