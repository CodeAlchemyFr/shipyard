package manifests

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/shipyard/cli/pkg/domains"
	"github.com/shipyard/cli/pkg/k8s"
)

const ingressTemplateDomain = `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: {{ .BaseDomain }}-ingress
  labels:
    managed-by: shipyard
    base-domain: {{ .BaseDomain }}
  annotations:
    # Traefik annotations (k3s default ingress controller)
    traefik.ingress.kubernetes.io/router.entrypoints: web,websecure
    traefik.ingress.kubernetes.io/router.tls: "true"
    # Cert-manager annotations (if cert-manager is installed)
    cert-manager.io/cluster-issuer: letsencrypt-prod
    # Nginx annotations (if nginx-ingress is used instead of Traefik)
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  {{- if .SSLEnabled }}
  tls:
  - hosts:
    {{- range .Domains }}
    - {{ .Hostname }}
    {{- end }}
    secretName: {{ .BaseDomain }}-tls
  {{- end }}
  rules:
  {{- range .Domains }}
  - host: {{ .Hostname }}
    http:
      paths:
      - path: {{ .Path }}
        pathType: Prefix
        backend:
          service:
            name: {{ .AppName }}-proxy
            port:
              number: {{ $.AppPort }}
  {{- end }}
---
{{- range .Domains }}
apiVersion: v1
kind: Service
metadata:
  name: {{ .AppName }}-proxy
  labels:
    managed-by: shipyard
    app: {{ .AppName }}
    proxy-for: {{ .AppName }}
spec:
  type: ClusterIP
  ports:
  - port: {{ $.AppPort }}
    targetPort: {{ $.AppPort }}
---
apiVersion: v1
kind: Endpoints
metadata:
  name: {{ .AppName }}-proxy
  labels:
    managed-by: shipyard
    app: {{ .AppName }}
    proxy-for: {{ .AppName }}
subsets:
- addresses:
  - ip: {{ .ServiceIP }}
  ports:
  - port: {{ $.AppPort }}
{{- end }}
`

// GenerateIngressFromDatabase generates ingress files based on domains in database
func (g *Generator) GenerateIngressFromDatabase() error {
	// Create domain manager
	domainManager, err := domains.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create domain manager: %w", err)
	}
	defer domainManager.Close()

	// Get all base domains
	baseDomains, err := domainManager.GetBaseDomains()
	if err != nil {
		return fmt.Errorf("failed to get base domains: %w", err)
	}

	if len(baseDomains) == 0 {
		fmt.Println("ℹ️  No domains found in database, skipping ingress generation")
		return nil
	}

	// Create shared directory
	sharedDir := filepath.Join(g.outputDir, "shared")
	if err := os.MkdirAll(sharedDir, 0755); err != nil {
		return fmt.Errorf("failed to create shared directory %s: %w", sharedDir, err)
	}

	// Generate ingress for each base domain
	for _, baseDomain := range baseDomains {
		domainsForBase, err := domainManager.GetDomainsByBaseDomain(baseDomain)
		if err != nil {
			return fmt.Errorf("failed to get domains for %s: %w", baseDomain, err)
		}

		ingressFile := filepath.Join(sharedDir, fmt.Sprintf("%s.yaml", baseDomain))
		
		if err := g.generateIngressFileFromDomains(ingressFile, baseDomain, domainsForBase); err != nil {
			return fmt.Errorf("failed to generate ingress for %s: %w", baseDomain, err)
		}

		fmt.Printf("🌐 Generated ingress: %s (%d domains)\n", ingressFile, len(domainsForBase))
	}

	return nil
}

// generateIngressFileFromDomains creates an ingress file from domain list
func (g *Generator) generateIngressFileFromDomains(ingressFile, baseDomain string, domainList []domains.Domain) error {
	// Check if any domain has SSL enabled
	sslEnabled := false
	for _, domain := range domainList {
		if domain.SSLEnabled {
			sslEnabled = true
			break
		}
	}

	// Get app port from config, default to 80 if not available
	appPort := 80
	if g.config != nil && g.config.App.Port > 0 {
		appPort = g.config.App.Port
	}

	// Enhance domain data with normalized names
	enhancedDomains := make([]struct {
		domains.Domain
		AppNamespace string
		ServiceIP    string
	}, len(domainList))
	
	for i, domain := range domainList {
		enhancedDomains[i] = struct {
			domains.Domain
			AppNamespace string
			ServiceIP    string
		}{
			Domain:       domain,
			AppNamespace: domain.AppName, // Use original name since we removed normalization
			ServiceIP:    "SERVICE_IP_PLACEHOLDER", // Will be updated after deployment
		}
	}

	// Prepare template data
	ingressData := struct {
		BaseDomain string
		Domains    []struct {
			domains.Domain
			AppNamespace string
			ServiceIP    string
		}
		SSLEnabled bool
		AppPort    int
	}{
		BaseDomain: baseDomain,
		Domains:    enhancedDomains,
		SSLEnabled: sslEnabled,
		AppPort:    appPort,
	}

	tmpl, err := template.New("ingress").Parse(ingressTemplateDomain)
	if err != nil {
		return fmt.Errorf("failed to parse ingress template: %w", err)
	}

	file, err := os.Create(ingressFile)
	if err != nil {
		return fmt.Errorf("failed to create ingress file %s: %w", ingressFile, err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, ingressData); err != nil {
		return fmt.Errorf("failed to execute ingress template: %w", err)
	}

	return nil
}


// UpdateIngressFromDatabase updates ingress files based on current database state
func (g *Generator) UpdateIngressFromDatabase(appName string) error {
	// Create domain manager
	domainManager, err := domains.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create domain manager: %w", err)
	}
	defer domainManager.Close()

	// Sync domains from config to database first
	if len(g.config.Domains) > 0 {
		fmt.Printf("🔄 Syncing %d domains from config to database...\n", len(g.config.Domains))
		if err := domainManager.SyncDomainsFromConfig(appName, g.config.Domains); err != nil {
			return fmt.Errorf("failed to sync domains from config: %w", err)
		}
	}

	// Generate all ingress files from database
	return g.GenerateIngressFromDatabase()
}

// UpdateIngressServiceIPs updates the ingress endpoints with actual service IPs after deployment
func (g *Generator) UpdateIngressServiceIPs() error {
	// Create k8s client to resolve service IPs
	client, err := k8s.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create k8s client: %w", err)
	}

	// Create domain manager
	domainManager, err := domains.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create domain manager: %w", err)
	}
	defer domainManager.Close()

	// Get all base domains
	baseDomains, err := domainManager.GetBaseDomains()
	if err != nil {
		return fmt.Errorf("failed to get base domains: %w", err)
	}

	// Update each base domain's ingress
	for _, baseDomain := range baseDomains {
		domainsForBase, err := domainManager.GetDomainsByBaseDomain(baseDomain)
		if err != nil {
			return fmt.Errorf("failed to get domains for %s: %w", baseDomain, err)
		}

		if len(domainsForBase) == 0 {
			continue
		}

		// Get the service IP for this domain
		domain := domainsForBase[0]
		serviceIP, err := client.GetServiceClusterIP(domain.AppName, domain.AppName)
		if err != nil {
			fmt.Printf("⚠️  Warning: Failed to get service IP for %s: %v\n", domain.AppName, err)
			continue
		}

		// Update the endpoints
		err = client.UpdateEndpointsIP(domain.AppName+"-proxy", "default", serviceIP, g.config.App.Port)
		if err != nil {
			fmt.Printf("⚠️  Warning: Failed to update endpoints for %s: %v\n", domain.AppName, err)
			continue
		}

		fmt.Printf("✅ Updated ingress endpoints for %s (IP: %s)\n", baseDomain, serviceIP)
	}

	return nil
}

// CleanupIngressFiles removes ingress files for base domains that no longer have domains
func (g *Generator) CleanupIngressFiles() error {
	// Create domain manager
	domainManager, err := domains.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create domain manager: %w", err)
	}
	defer domainManager.Close()

	// Get current base domains from database
	baseDomains, err := domainManager.GetBaseDomains()
	if err != nil {
		return fmt.Errorf("failed to get base domains: %w", err)
	}

	baseDomainsMap := make(map[string]bool)
	for _, baseDomain := range baseDomains {
		baseDomainsMap[baseDomain] = true
	}

	// Check shared directory for orphaned ingress files
	sharedDir := filepath.Join(g.outputDir, "shared")
	if _, err := os.Stat(sharedDir); os.IsNotExist(err) {
		return nil // No shared directory, nothing to clean
	}

	entries, err := os.ReadDir(sharedDir)
	if err != nil {
		return fmt.Errorf("failed to read shared directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".yaml" {
			// Extract base domain from filename
			baseDomain := entry.Name()[:len(entry.Name())-5] // Remove .yaml

			// If this base domain no longer exists in database, remove the file
			if !baseDomainsMap[baseDomain] {
				filePath := filepath.Join(sharedDir, entry.Name())
				if err := os.Remove(filePath); err != nil {
					fmt.Printf("⚠️  Warning: failed to remove orphaned ingress file %s: %v\n", filePath, err)
				} else {
					fmt.Printf("🗑️  Removed orphaned ingress: %s\n", filePath)
				}
			}
		}
	}

	return nil
}