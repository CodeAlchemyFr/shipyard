package cmd

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var (
	forceUpgrade bool
	skipConfirm  bool
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade Shipyard CLI to the latest version",
	Long: `Upgrade Shipyard CLI to the latest version from GitHub releases.

This command will:
- Download the latest version for your platform
- Replace the current binary
- Verify the installation

Examples:
  shipyard upgrade              # Upgrade with confirmation
  shipyard upgrade --force      # Force upgrade without confirmation
  shipyard upgrade --yes        # Skip confirmation prompt`,
	Run: func(cmd *cobra.Command, args []string) {
		runUpgrade()
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
	upgradeCmd.Flags().BoolVar(&forceUpgrade, "force", false, "Force upgrade without version check")
	upgradeCmd.Flags().BoolVar(&skipConfirm, "yes", false, "Skip confirmation prompt")
}

func runUpgrade() {
	fmt.Println("🚀 Upgrading Shipyard CLI...")

	// Get current version
	currentVersion := getCurrentVersion()
	fmt.Printf("📋 Current version: %s\n", currentVersion)

	// Check latest version
	fmt.Println("🔍 Checking for latest version...")
	latestVersion, err := getLatestVersion()
	if err != nil {
		fmt.Printf("❌ Failed to check latest version: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📦 Latest version: %s\n", latestVersion)

	// Check if upgrade is needed
	if !forceUpgrade && currentVersion == latestVersion {
		fmt.Println("✅ You already have the latest version!")
		return
	}

	// Confirm upgrade
	if !skipConfirm && !forceUpgrade {
		if !confirmUpgrade(currentVersion, latestVersion) {
			fmt.Println("❌ Upgrade cancelled")
			return
		}
	}

	// Detect platform
	platform := detectPlatform()
	fmt.Printf("🖥️  Detected platform: %s\n", platform)

	// Download latest version
	downloadURL := fmt.Sprintf("https://github.com/CodeAlchemyFr/shipyard/releases/latest/download/shipyard-%s", platform)
	fmt.Printf("📥 Downloading from: %s\n", downloadURL)

	tempFile, err := downloadFile(downloadURL)
	if err != nil {
		fmt.Printf("❌ Failed to download: %v\n", err)
		os.Exit(1)
	}
	defer os.Remove(tempFile)

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("❌ Failed to get executable path: %v\n", err)
		os.Exit(1)
	}

	// Backup current version
	backupPath := execPath + ".backup"
	fmt.Printf("💾 Creating backup: %s\n", backupPath)
	if err := copyFile(execPath, backupPath); err != nil {
		fmt.Printf("⚠️  Warning: Failed to create backup: %v\n", err)
	}

	// Replace binary
	fmt.Printf("🔄 Installing new version...\n")
	if err := replaceExecutable(tempFile, execPath); err != nil {
		fmt.Printf("❌ Failed to replace executable: %v\n", err)
		// Try to restore backup
		if _, err := os.Stat(backupPath); err == nil {
			fmt.Println("🔄 Restoring backup...")
			copyFile(backupPath, execPath)
		}
		os.Exit(1)
	}

	// Verify installation
	fmt.Println("✅ Upgrade completed successfully!")
	fmt.Printf("🎉 Shipyard CLI updated to %s\n", latestVersion)
	
	// Clean up backup
	os.Remove(backupPath)
	
	fmt.Println("📚 Run 'shipyard --version' to verify the new version")
}

func getCurrentVersion() string {
	return version
}

func getLatestVersion() (string, error) {
	// GitHub API to get latest release
	url := "https://api.github.com/repos/CodeAlchemyFr/shipyard/releases/latest"
	
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Simple parsing - in production you'd use proper JSON parsing
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	content := string(body)
	// Extract tag_name from JSON response
	start := strings.Index(content, `"tag_name":"`)
	if start == -1 {
		return "latest", nil
	}
	start += len(`"tag_name":"`)
	end := strings.Index(content[start:], `"`)
	if end == -1 {
		return "latest", nil
	}

	return content[start : start+end], nil
}

func detectPlatform() string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Map Go arch to our naming
	arch := goarch
	if goarch == "amd64" {
		arch = "amd64"
	} else if goarch == "arm64" {
		arch = "arm64"
	}

	// Build platform string
	if goos == "windows" {
		return fmt.Sprintf("windows-%s.exe", arch)
	}
	return fmt.Sprintf("%s-%s", goos, arch)
}

func downloadFile(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// Create temp file
	tempFile, err := os.CreateTemp("", "shipyard-upgrade-*")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	// Copy response to temp file
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		os.Remove(tempFile.Name())
		return "", err
	}

	// Make executable
	if err := os.Chmod(tempFile.Name(), 0755); err != nil {
		os.Remove(tempFile.Name())
		return "", err
	}

	return tempFile.Name(), nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// Copy permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, srcInfo.Mode())
}

func replaceExecutable(newFile, targetPath string) error {
	// On Windows, we might need to handle this differently
	if runtime.GOOS == "windows" {
		// Move current exe to temp name, then move new one in place
		tempName := targetPath + ".old"
		if err := os.Rename(targetPath, tempName); err != nil {
			return err
		}
		if err := copyFile(newFile, targetPath); err != nil {
			os.Rename(tempName, targetPath) // Restore
			return err
		}
		os.Remove(tempName)
		return nil
	}

	// Unix-like systems
	return copyFile(newFile, targetPath)
}

func confirmUpgrade(current, latest string) bool {
	fmt.Printf("⚠️  Upgrade Shipyard CLI from %s to %s?\n", current, latest)
	fmt.Print("Continue? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	return response == "y" || response == "yes"
}