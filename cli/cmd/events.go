package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/shipyard/cli/pkg/monitoring"
)

var eventsCmd = &cobra.Command{
	Use:   "events [app-name]",
	Short: "Show application and cluster events",
	Long: `Display recent events for applications and cluster resources.

Events include:
- Deployment updates
- Pod state changes
- Service modifications
- Error conditions
- Warning messages

Examples:
  shipyard events                     # Show all recent events
  shipyard events my-app              # Show events for specific app
  shipyard events --follow           # Stream events in real-time
  shipyard events --since 1h         # Show events from last hour
  shipyard events --type error       # Show only error events`,
	Run: func(cmd *cobra.Command, args []string) {
		var appName string
		if len(args) > 0 {
			appName = args[0]
		}

		follow, _ := cmd.Flags().GetBool("follow")
		since, _ := cmd.Flags().GetDuration("since")
		eventType, _ := cmd.Flags().GetString("type")

		if err := runEvents(appName, follow, since, eventType); err != nil {
			log.Fatalf("Events command failed: %v", err)
		}
	},
}

func init() {
	eventsCmd.Flags().BoolP("follow", "f", false, "Stream events in real-time")
	eventsCmd.Flags().DurationP("since", "s", time.Hour, "Show events since specified duration")
	eventsCmd.Flags().StringP("type", "t", "", "Filter by event type (normal, warning, error)")
}

// EventInfo represents an application or cluster event
type EventInfo struct {
	Timestamp    time.Time
	AppName      string
	Component    string
	Type         string
	Reason       string
	Message      string
	Source       string
	Count        int
	FirstSeen    time.Time
	LastSeen     time.Time
}

func runEvents(appName string, follow bool, since time.Duration, eventType string) error {
	// Initialize monitoring collector
	collector, err := monitoring.NewCollector()
	if err != nil {
		return fmt.Errorf("failed to initialize monitoring: %w", err)
	}
	defer collector.Close()

	if follow {
		return runEventsFollow(collector, appName, eventType)
	}

	events, err := getEvents(collector, appName, since, eventType)
	if err != nil {
		return fmt.Errorf("failed to get events: %w", err)
	}

	displayEventsTable(events, appName, since, eventType)
	return nil
}

func runEventsFollow(collector *monitoring.Collector, appName, eventType string) error {
	fmt.Printf("🔍 Streaming events (press Ctrl+C to stop)\n")
	if appName != "" {
		fmt.Printf("Filtering for app: %s\n", appName)
	}
	if eventType != "" {
		fmt.Printf("Filtering for type: %s\n", eventType)
	}
	fmt.Println()

	lastEventTime := time.Now()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Get new events since last check
			newEvents, err := getEventsSince(collector, appName, lastEventTime, eventType)
			if err != nil {
				fmt.Printf("Error getting events: %v\n", err)
				continue
			}

			// Display new events
			for _, event := range newEvents {
				displaySingleEvent(event)
				if event.Timestamp.After(lastEventTime) {
					lastEventTime = event.Timestamp
				}
			}
		}
	}
}

func getEvents(collector *monitoring.Collector, appName string, since time.Duration, eventType string) ([]EventInfo, error) {
	// This would normally query Kubernetes events and application events from the database
	// For now, return sample data
	
	now := time.Now()
	events := []EventInfo{
		{
			Timestamp: now.Add(-10 * time.Minute),
			AppName:   "web-app",
			Component: "deployment",
			Type:      "normal",
			Reason:    "ScalingReplicaSet",
			Message:   "Scaled up replica set web-app-7d4b9c8f5c to 3",
			Source:    "deployment-controller",
			Count:     1,
			FirstSeen: now.Add(-10 * time.Minute),
			LastSeen:  now.Add(-10 * time.Minute),
		},
		{
			Timestamp: now.Add(-15 * time.Minute),
			AppName:   "api-service",
			Component: "pod",
			Type:      "warning",
			Reason:    "BackOff",
			Message:   "Back-off restarting failed container api-service in pod api-service-6b8f4d9c7a-xyz12",
			Source:    "kubelet",
			Count:     3,
			FirstSeen: now.Add(-25 * time.Minute),
			LastSeen:  now.Add(-15 * time.Minute),
		},
		{
			Timestamp: now.Add(-20 * time.Minute),
			AppName:   "api-service",
			Component: "deployment",
			Type:      "normal",
			Reason:    "DeploymentRollout",
			Message:   "Deployment has minimum availability. Replica set api-service-6b8f4d9c7a has 2 available replicas",
			Source:    "deployment-controller",
			Count:     1,
			FirstSeen: now.Add(-20 * time.Minute),
			LastSeen:  now.Add(-20 * time.Minute),
		},
		{
			Timestamp: now.Add(-30 * time.Minute),
			AppName:   "worker",
			Component: "pod",
			Type:      "error",
			Reason:    "FailedMount",
			Message:   "Unable to attach or mount volumes: unmounted volumes=[config], unattached volumes=[config default-token]: timed out waiting for the condition",
			Source:    "kubelet",
			Count:     1,
			FirstSeen: now.Add(-30 * time.Minute),
			LastSeen:  now.Add(-30 * time.Minute),
		},
		{
			Timestamp: now.Add(-45 * time.Minute),
			AppName:   "web-app",
			Component: "service",
			Type:      "normal",
			Reason:    "ServiceCreated",
			Message:   "Service web-app created successfully",
			Source:    "service-controller",
			Count:     1,
			FirstSeen: now.Add(-45 * time.Minute),
			LastSeen:  now.Add(-45 * time.Minute),
		},
		{
			Timestamp: now.Add(-1 * time.Hour),
			AppName:   "database",
			Component: "pod",
			Type:      "warning",
			Reason:    "HighMemoryUsage",
			Message:   "Pod database-0 is using 95% of allocated memory",
			Source:    "monitoring-system",
			Count:     1,
			FirstSeen: now.Add(-1 * time.Hour),
			LastSeen:  now.Add(-1 * time.Hour),
		},
	}

	// Filter by app name
	if appName != "" {
		var filtered []EventInfo
		for _, event := range events {
			if event.AppName == appName {
				filtered = append(filtered, event)
			}
		}
		events = filtered
	}

	// Filter by time period
	cutoff := now.Add(-since)
	var timeFiltered []EventInfo
	for _, event := range events {
		if event.Timestamp.After(cutoff) {
			timeFiltered = append(timeFiltered, event)
		}
	}
	events = timeFiltered

	// Filter by event type
	if eventType != "" {
		var typeFiltered []EventInfo
		for _, event := range events {
			if event.Type == eventType {
				typeFiltered = append(typeFiltered, event)
			}
		}
		events = typeFiltered
	}

	// Sort by timestamp (most recent first)
	for i := 0; i < len(events)-1; i++ {
		for j := i + 1; j < len(events); j++ {
			if events[i].Timestamp.Before(events[j].Timestamp) {
				events[i], events[j] = events[j], events[i]
			}
		}
	}

	return events, nil
}

func getEventsSince(collector *monitoring.Collector, appName string, since time.Time, eventType string) ([]EventInfo, error) {
	// Get all events and filter by timestamp
	allEvents, err := getEvents(collector, appName, time.Since(since)+time.Minute, eventType)
	if err != nil {
		return nil, err
	}

	var newEvents []EventInfo
	for _, event := range allEvents {
		if event.Timestamp.After(since) {
			newEvents = append(newEvents, event)
		}
	}

	return newEvents, nil
}

func displayEventsTable(events []EventInfo, appName string, since time.Duration, eventType string) {
	if len(events) == 0 {
		fmt.Printf("📅 No events found")
		if appName != "" {
			fmt.Printf(" for %s", appName)
		}
		if eventType != "" {
			fmt.Printf(" of type %s", eventType)
		}
		fmt.Printf(" in the last %v\n", since)
		return
	}

	title := fmt.Sprintf("📅 Recent Events (Last %v)", since)
	if appName != "" {
		title = fmt.Sprintf("📅 Events for %s (Last %v)", appName, since)
	}
	if eventType != "" {
		title += fmt.Sprintf(" - Type: %s", eventType)
	}
	
	fmt.Println(title)
	fmt.Println("=" + fmt.Sprintf("%*s", len(title)-1, ""))
	fmt.Println()

	// Header
	fmt.Printf("┌%-20s┬%-12s┬%-12s┬%-10s┬%-18s┬%-40s┐\n",
		"────────────────────", "────────────", "────────────", "──────────", "──────────────────", "────────────────────────────────────────")
	fmt.Printf("│%-20s│%-12s│%-12s│%-10s│%-18s│%-40s│\n",
		"TIME", "APP", "COMPONENT", "TYPE", "REASON", "MESSAGE")
	fmt.Printf("├%-20s┼%-12s┼%-12s┼%-10s┼%-18s┼%-40s┤\n",
		"────────────────────", "────────────", "────────────", "──────────", "──────────────────", "────────────────────────────────────────")

	// Data rows
	for _, event := range events {
		typeIcon := "ℹ️"
		switch event.Type {
		case "warning":
			typeIcon = "⚠️"
		case "error":
			typeIcon = "❌"
		case "normal":
			typeIcon = "✅"
		}

		timeStr := event.Timestamp.Format("15:04:05 Jan 02")
		countStr := ""
		if event.Count > 1 {
			countStr = fmt.Sprintf(" (x%d)", event.Count)
		}

		fmt.Printf("│%-20s│%-12s│%-12s│%s %-8s│%-18s│%-40s│\n",
			timeStr,
			truncateString(event.AppName, 12),
			truncateString(event.Component, 12),
			typeIcon,
			truncateString(event.Type, 7),
			truncateString(event.Reason+countStr, 18),
			truncateString(event.Message, 40),
		)
	}

	fmt.Printf("└%-20s┴%-12s┴%-12s┴%-10s┴%-18s┴%-40s┘\n",
		"────────────────────", "────────────", "────────────", "──────────", "──────────────────", "────────────────────────────────────────")

	// Summary
	normalCount := 0
	warningCount := 0
	errorCount := 0
	for _, event := range events {
		switch event.Type {
		case "normal":
			normalCount++
		case "warning":
			warningCount++
		case "error":
			errorCount++
		}
	}

	fmt.Printf("\n📊 Summary: %d total events", len(events))
	if normalCount > 0 {
		fmt.Printf(", %d normal", normalCount)
	}
	if warningCount > 0 {
		fmt.Printf(", %d warnings", warningCount)
	}
	if errorCount > 0 {
		fmt.Printf(", %d errors", errorCount)
	}
	fmt.Println()
	
	fmt.Printf("💡 Tip: Use --follow to stream events in real-time\n")
}

func displaySingleEvent(event EventInfo) {
	typeIcon := "ℹ️"
	switch event.Type {
	case "warning":
		typeIcon = "⚠️"
	case "error":
		typeIcon = "❌"
	case "normal":
		typeIcon = "✅"
	}

	timeStr := event.Timestamp.Format("15:04:05")
	countStr := ""
	if event.Count > 1 {
		countStr = fmt.Sprintf(" (x%d)", event.Count)
	}

	fmt.Printf("[%s] %s [%s/%s] %s%s: %s\n",
		timeStr,
		typeIcon,
		event.AppName,
		event.Component,
		event.Reason,
		countStr,
		event.Message,
	)
}