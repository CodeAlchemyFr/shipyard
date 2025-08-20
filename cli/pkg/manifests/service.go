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
  name: {{ .App.GetDNSName }}
  namespace: {{ .App.GetNamespace }}
  labels:
    app: {{ .App.GetDNSName }}
    managed-by: shipyard
spec:
  type: {{ if .Service.Type }}{{ .Service.Type }}{{ else }}ClusterIP{{ end }}
  ports:
  - port: {{ .App.Port }}
    targetPort: {{ .App.Port }}
    protocol: TCP
    name: http
    {{- if and .Service.ExternalPort (eq .Service.Type "NodePort") }}
    nodePort: {{ .Service.ExternalPort }}
    {{- end }}
  selector:
    app: {{ .App.GetDNSName }}
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