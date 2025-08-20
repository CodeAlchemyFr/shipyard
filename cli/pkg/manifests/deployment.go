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
  name: {{ .App.GetDNSName }}
  namespace: {{ .App.GetNamespace }}
  labels:
    app: {{ .App.GetDNSName }}
    managed-by: shipyard
    {{- if .Version }}
    shipyard.version: "{{ .Version.Version }}"
    shipyard.image-tag: "{{ .Version.ImageTag }}"
    shipyard.image-hash: "{{ .Version.ImageHash }}"
    shipyard.deployed-at: "{{ .Version.Timestamp.Format "2006-01-02T15-04-05Z07-00" }}"
    {{- if .Version.RollbackTo }}
    shipyard.rollback-from: "{{ .Version.RollbackTo }}"
    {{- end }}
    {{- end }}
spec:
  replicas: {{ .Scaling.Min }}
  selector:
    matchLabels:
      app: {{ .App.GetDNSName }}
  template:
    metadata:
      labels:
        app: {{ .App.GetDNSName }}
    spec:
      {{- if .ImagePullSecrets }}
      imagePullSecrets:
      {{- range .ImagePullSecrets }}
      - name: {{ . }}
      {{- end }}
      {{- end }}
      containers:
      - name: {{ .App.GetDNSName }}
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
            name: {{ .App.GetDNSName }}-secrets
        {{- end }}
        resources:
          requests:
            cpu: {{ .Resources.CPU }}
            memory: {{ .Resources.Memory }}
          limits:
            cpu: {{ .Resources.CPU }}
            memory: {{ .Resources.Memory }}
        {{- if .Health.Liveness.Path }}
        livenessProbe:
          httpGet:
            path: {{ .Health.Liveness.Path }}
            port: {{ if .Health.Liveness.Port }}{{ .Health.Liveness.Port }}{{ else }}{{ .App.Port }}{{ end }}
          initialDelaySeconds: {{ if .Health.Liveness.InitialDelaySeconds }}{{ .Health.Liveness.InitialDelaySeconds }}{{ else }}30{{ end }}
          periodSeconds: {{ if .Health.Liveness.PeriodSeconds }}{{ .Health.Liveness.PeriodSeconds }}{{ else }}10{{ end }}
        {{- else }}
        livenessProbe:
          httpGet:
            path: /
            port: {{ .App.Port }}
          initialDelaySeconds: 30
          periodSeconds: 10
        {{- end }}
        {{- if .Health.Readiness.Path }}
        readinessProbe:
          httpGet:
            path: {{ .Health.Readiness.Path }}
            port: {{ if .Health.Readiness.Port }}{{ .Health.Readiness.Port }}{{ else }}{{ .App.Port }}{{ end }}
          initialDelaySeconds: {{ if .Health.Readiness.InitialDelaySeconds }}{{ .Health.Readiness.InitialDelaySeconds }}{{ else }}5{{ end }}
          periodSeconds: {{ if .Health.Readiness.PeriodSeconds }}{{ .Health.Readiness.PeriodSeconds }}{{ else }}5{{ end }}
        {{- else }}
        readinessProbe:
          httpGet:
            path: /
            port: {{ .App.Port }}
          initialDelaySeconds: 5
          periodSeconds: 5
        {{- end }}
---
{{- if gt .Scaling.Max .Scaling.Min }}
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: {{ .App.GetDNSName }}-hpa
  namespace: {{ .App.GetNamespace }}
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: {{ .App.GetDNSName }}
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