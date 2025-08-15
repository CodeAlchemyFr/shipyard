package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/shipyard/cli/pkg/k8s"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of deployed applications",
	Long:  `Display the current status of all deployed applications including pods, services, and ingress.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runStatus(); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}

func runStatus() error {
	fmt.Println("📊 Application Status:")
	
	client, err := k8s.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create k8s client: %w", err)
	}
	
	return client.ShowStatus()
}