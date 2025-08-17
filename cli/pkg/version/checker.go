package version

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Current holds the current version (set by main)
var Current = "dev"

// CheckInfo holds version check information
type CheckInfo struct {
	LastCheck    time.Time `json:"last_check"`
	LatestVersion string    `json:"latest_version"`
	CurrentVersion string   `json:"current_version"`
	Notified     bool      `json:"notified"`
}

// NotifyIfUpdateAvailable checks for updates and shows notification if needed
func NotifyIfUpdateAvailable(currentVersion string) {
	// Only check once per day maximum
	if !shouldCheckForUpdate() {
		return
	}

	latestVersion, err := getLatestVersionFromGitHub()
	if err != nil {
		// Silently fail for version checks
		return
	}

	// Update check info
	updateCheckInfo(currentVersion, latestVersion)

	// Show notification if update available and not already notified today
	if latestVersion != currentVersion && !wasNotifiedToday(currentVersion, latestVersion) {
		showUpdateNotification(currentVersion, latestVersion)
		markAsNotified(currentVersion, latestVersion)
	}
}

// shouldCheckForUpdate determines if we should check for updates
func shouldCheckForUpdate() bool {
	checkFile := getCheckFilePath()
	
	info, err := os.Stat(checkFile)
	if err != nil {
		// File doesn't exist, should check
		return true
	}

	// Check if more than 24 hours have passed
	return time.Since(info.ModTime()) > 24*time.Hour
}

// getLatestVersionFromGitHub fetches the latest version from GitHub releases
func getLatestVersionFromGitHub() (string, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/CodeAlchemyFr/shipyard/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("GitHub API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Parse JSON response
	var release struct {
		TagName string `json:"tag_name"`
	}

	if err := json.Unmarshal(body, &release); err != nil {
		// Fallback to simple string parsing if JSON fails
		content := string(body)
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

	return release.TagName, nil
}

// getCheckFilePath returns the path to the version check file
func getCheckFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/shipyard-version-check"
	}
	
	shipyardDir := filepath.Join(homeDir, ".shipyard")
	os.MkdirAll(shipyardDir, 0755)
	
	return filepath.Join(shipyardDir, "version-check.json")
}

// updateCheckInfo updates the version check information
func updateCheckInfo(currentVersion, latestVersion string) {
	checkFile := getCheckFilePath()
	
	info := CheckInfo{
		LastCheck:      time.Now(),
		LatestVersion:  latestVersion,
		CurrentVersion: currentVersion,
		Notified:       false,
	}

	// Try to read existing info to preserve notification status
	if data, err := os.ReadFile(checkFile); err == nil {
		var existingInfo CheckInfo
		if json.Unmarshal(data, &existingInfo) == nil {
			// Preserve notification status if versions haven't changed
			if existingInfo.CurrentVersion == currentVersion && existingInfo.LatestVersion == latestVersion {
				info.Notified = existingInfo.Notified
			}
		}
	}

	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return
	}

	os.WriteFile(checkFile, data, 0644)
}

// wasNotifiedToday checks if user was already notified about this version today
func wasNotifiedToday(currentVersion, latestVersion string) bool {
	checkFile := getCheckFilePath()
	
	data, err := os.ReadFile(checkFile)
	if err != nil {
		return false
	}

	var info CheckInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return false
	}

	// Check if we already notified about this specific version combination today
	return info.Notified && 
		   info.CurrentVersion == currentVersion && 
		   info.LatestVersion == latestVersion &&
		   time.Since(info.LastCheck) < 24*time.Hour
}

// markAsNotified marks that the user has been notified about the available update
func markAsNotified(currentVersion, latestVersion string) {
	checkFile := getCheckFilePath()
	
	info := CheckInfo{
		LastCheck:      time.Now(),
		LatestVersion:  latestVersion,
		CurrentVersion: currentVersion,
		Notified:       true,
	}

	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return
	}

	os.WriteFile(checkFile, data, 0644)
}

// showUpdateNotification displays the update notification
func showUpdateNotification(currentVersion, latestVersion string) {
	fmt.Printf("\n🎉 \033[32mNew Shipyard version available!\033[0m\n")
	fmt.Printf("   Current: %s → Latest: \033[32m%s\033[0m\n", currentVersion, latestVersion)
	fmt.Printf("   Run \033[36mshipyard upgrade\033[0m to update\n")
	fmt.Printf("   Or: \033[36mcurl -sSL https://github.com/CodeAlchemyFr/shipyard/releases/latest/download/install.sh | bash\033[0m\n\n")
}