package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var sslCmd = &cobra.Command{
	Use:   "ssl",
	Short: "SSL/TLS certificate management",
	Long: `Manage SSL/TLS certificates for your applications.
This command installs cert-manager and configures automatic HTTPS certificates.`,
}

var installSSLCmd = &cobra.Command{
	Use:   "install",
	Short: "Install cert-manager for automatic SSL certificates",
	Long: `Install cert-manager on your Kubernetes cluster to enable automatic 
SSL certificate generation from Let's Encrypt.

This will:
- Install cert-manager
- Create a Let's Encrypt ClusterIssuer
- Configure automatic HTTPS for your domains`,
	Run: func(cmd *cobra.Command, args []string) {
		runInstallSSL()
	},
}

func init() {
	rootCmd.AddCommand(sslCmd)
	sslCmd.AddCommand(installSSLCmd)
}

func runInstallSSL() {
	fmt.Println("🔐 Installing cert-manager for automatic SSL certificates...")

	// Check if kubectl is available
	if err := checkKubectl(); err != nil {
		fmt.Printf("❌ kubectl not found or cluster not accessible: %v\n", err)
		os.Exit(1)
	}

	// Check if cert-manager is already installed
	if isCertManagerInstalled() {
		fmt.Println("✅ cert-manager is already installed")
	} else {
		// Install cert-manager
		fmt.Println("📦 Installing cert-manager...")
		if err := installCertManager(); err != nil {
			fmt.Printf("❌ Failed to install cert-manager: %v\n", err)
			os.Exit(1)
		}

		// Wait for cert-manager to be ready
		fmt.Println("⏳ Waiting for cert-manager to be ready...")
		if err := waitForCertManager(); err != nil {
			fmt.Printf("❌ cert-manager failed to start: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✅ cert-manager installed successfully!")
	}

	// Check if ClusterIssuer exists
	if isClusterIssuerInstalled() {
		fmt.Println("✅ Let's Encrypt ClusterIssuer already exists")
	} else {
		// Create ClusterIssuer
		fmt.Println("📄 Creating Let's Encrypt ClusterIssuer...")
		if err := createClusterIssuer(); err != nil {
			fmt.Printf("❌ Failed to create ClusterIssuer: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ Let's Encrypt ClusterIssuer created successfully!")
	}

	fmt.Println("🎉 SSL setup complete! Your domains will now get automatic HTTPS certificates.")
	fmt.Println("📋 Run 'shipyard deploy' to apply SSL to your applications.")
}

func checkKubectl() error {
	cmd := exec.Command("kubectl", "cluster-info")
	return cmd.Run()
}

func isCertManagerInstalled() bool {
	cmd := exec.Command("kubectl", "get", "namespace", "cert-manager")
	return cmd.Run() == nil
}

func installCertManager() error {
	cmd := exec.Command("kubectl", "apply", "-f", "https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func waitForCertManager() error {
	// Wait for cert-manager pods to be ready
	commands := [][]string{
		{"kubectl", "wait", "--for=condition=ready", "pod", "-l", "app=cert-manager", "-n", "cert-manager", "--timeout=120s"},
		{"kubectl", "wait", "--for=condition=ready", "pod", "-l", "app=cainjector", "-n", "cert-manager", "--timeout=120s"},
		{"kubectl", "wait", "--for=condition=ready", "pod", "-l", "app=webhook", "-n", "cert-manager", "--timeout=120s"},
	}

	for _, cmdArgs := range commands {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		if err := cmd.Run(); err != nil {
			// Try a few times as cert-manager might take time to start
			time.Sleep(10 * time.Second)
			if err := cmd.Run(); err != nil {
				return err
			}
		}
	}

	return nil
}

func isClusterIssuerInstalled() bool {
	cmd := exec.Command("kubectl", "get", "clusterissuer", "letsencrypt-prod")
	return cmd.Run() == nil
}

func createClusterIssuer() error {
	clusterIssuerYAML := `apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@shipyard.local
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: traefik
`

	cmd := exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(clusterIssuerYAML)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}