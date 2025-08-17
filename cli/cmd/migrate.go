package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/shipyard/cli/pkg/config"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate manifests from local to global directory",
	Long: `Migrate existing manifests from the current directory to the global Shipyard directory.

This command moves manifests from ./manifests/ to ~/.shipyard/manifests/ to avoid
polluting your application directories with generated files.`,
	Run: func(cmd *cobra.Command, args []string) {
		runMigrate()
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}

func runMigrate() {
	fmt.Println("🔄 Migrating manifests to global directory...")

	// Check if local manifests directory exists
	localManifests := "manifests"
	if _, err := os.Stat(localManifests); os.IsNotExist(err) {
		fmt.Println("ℹ️  No local manifests directory found - nothing to migrate")
		return
	}

	// Get global manifests directory
	globalManifests, err := config.GetManifestsDir()
	if err != nil {
		fmt.Printf("❌ Failed to get global manifests directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📂 Moving from: %s\n", localManifests)
	fmt.Printf("📂 Moving to: %s\n", globalManifests)

	// Check if global directory already has content
	if hasContent(globalManifests) {
		fmt.Print("⚠️  Global manifests directory already has content. Merge? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "yes" {
			fmt.Println("❌ Migration cancelled")
			return
		}
	}

	// Migrate the content
	if err := migrateDirectory(localManifests, globalManifests); err != nil {
		fmt.Printf("❌ Migration failed: %v\n", err)
		os.Exit(1)
	}

	// Remove the local manifests directory
	fmt.Printf("🗑️  Removing local manifests directory...\n")
	if err := os.RemoveAll(localManifests); err != nil {
		fmt.Printf("⚠️  Warning: Failed to remove local directory: %v\n", err)
	}

	fmt.Println("✅ Migration completed successfully!")
	fmt.Printf("📁 Manifests are now in: %s\n", globalManifests)
}

// hasContent checks if a directory exists and has content
func hasContent(dir string) bool {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return false
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	return len(entries) > 0
}

// migrateDirectory moves content from source to destination
func migrateDirectory(src, dst string) error {
	// Create destination directory
	if err := os.MkdirAll(dst, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	// Move each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		fmt.Printf("📦 Moving: %s -> %s\n", srcPath, dstPath)

		if entry.IsDir() {
			// For directories, we need to handle potential merging
			if err := migrateDirectory(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to migrate directory %s: %w", srcPath, err)
			}
		} else {
			// For files, move directly (overwrite if exists)
			if err := moveFile(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to move file %s: %w", srcPath, err)
			}
		}
	}

	return nil
}

// moveFile moves a file from source to destination
func moveFile(src, dst string) error {
	// Create destination directory if needed
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}

	// Copy file
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

	// Copy content
	_, err = dstFile.ReadFrom(srcFile)
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