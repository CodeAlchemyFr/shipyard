package manifests

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

const deploymentTemplate = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .App.Name }}
  labels:
    app: {{ .App.Name }}
    managed-by: shipyard
    {{- if .Version }}
    shipyard.version: "{{ .Version.Version }}"
    shipyard.image-tag: "{{ .Version.ImageTag }}"
    shipyard.image-hash: "{{ .Version.ImageHash }}"
    shipyard.deployed-at: "{{ .Version.Timestamp.Format "2006-01-02T15:04:05Z07:00" }}"
    {{- if .Version.RollbackTo }}
    shipyard.rollback-from: "{{ .Version.RollbackTo }}"
    {{- end }}
    {{- end }}
spec:
  replicas: {{ .Scaling.Min }}
  selector:
    matchLabels:
      app: {{ .App.Name }}
  template:
    metadata:
      labels:
        app: {{ .App.Name }}
    spec:
      {{- if .ImagePullSecrets }}
      imagePullSecrets:
      {{- range .ImagePullSecrets }}
      - name: {{ . }}
      {{- end }}
      {{- end }}
      containers:
      - name: {{ .App.Name }}
        image: {{ .App.Image }}
        ports:
        - containerPort: {{ .App.Port }}
        env:
        {{- range $key, $value := .Env }}
        - name: {{ $key }}
          value: "{{ $value }}"
        {{- end }}
        {{- if .Secrets }}
        envFrom:
        - secretRef:
            name: {{ .App.Name }}-secrets
        {{- end }}
        resources:
          requests:
            cpu: {{ .Resources.CPU }}
            memory: {{ .Resources.Memory }}
          limits:
            cpu: {{ .Resources.CPU }}
            memory: {{ .Resources.Memory }}
        livenessProbe:
          httpGet:
            path: /health
            port: {{ .App.Port }}
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: {{ .App.Port }}
          initialDelaySeconds: 5
          periodSeconds: 5
---
{{- if gt .Scaling.Max .Scaling.Min }}
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: {{ .App.Name }}-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: {{ .App.Name }}
  minReplicas: {{ .Scaling.Min }}
  maxReplicas: {{ .Scaling.Max }}
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: {{ .Scaling.TargetCPU }}
{{- end }}
`

// generateDeployment creates the deployment.yaml file for an application
func (g *Generator) generateDeployment(appDir string) error {
	tmpl, err := template.New("deployment").Parse(deploymentTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse deployment template: %w", err)
	}

	filePath := filepath.Join(appDir, "deployment.yaml")
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create deployment file %s: %w", filePath, err)
	}
	defer file.Close()

	// Create template data with version info
	templateData := struct {
		*Config
		Version          *DeploymentVersion
		ImagePullSecrets []string
	}{
		Config:           g.config,
		Version:          g.version,
		ImagePullSecrets: g.imagePullSecrets,
	}

	if err := tmpl.Execute(file, templateData); err != nil {
		return fmt.Errorf("failed to execute deployment template: %w", err)
	}

	fmt.Printf("📄 Generated: %s", filePath)
	if g.version != nil {
		fmt.Printf(" (version: %s)", g.version.Version)
	}
	fmt.Printf("\n")
	return nil
}