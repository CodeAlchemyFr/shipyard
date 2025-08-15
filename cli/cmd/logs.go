package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/shipyard/cli/pkg/k8s"
)

var (
	tail   bool
	follow bool
	since  string
)

var logsCmd = &cobra.Command{
	Use:   "logs [app-name]",
	Short: "Show application logs",
	Long:  `Display logs from your application pods with filtering and real-time options.`,
	Run: func(cmd *cobra.Command, args []string) {
		appName := ""
		if len(args) > 0 {
			appName = args[0]
		}
		
		if err := runLogs(appName); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}

func init() {
	logsCmd.Flags().BoolVar(&tail, "tail", false, "Follow logs in real-time")
	logsCmd.Flags().BoolVar(&follow, "follow", false, "Follow logs in real-time (alias for --tail)")
	logsCmd.Flags().StringVar(&since, "since", "", "Show logs since (e.g., 1h, 30m)")
}

func runLogs(appName string) error {
	fmt.Printf("📋 Fetching logs for app: %s\n", appName)
	
	client, err := k8s.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create k8s client: %w", err)
	}
	
	options := k8s.LogsOptions{
		Follow: tail || follow,
		Since:  since,
	}
	
	return client.GetLogs(appName, options)
}