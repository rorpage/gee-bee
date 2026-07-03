package main

import (
	"fmt"
	"geebee/internal/configuration"
	"geebee/internal/geebee"
	"geebee/internal/notification"
	"log"
	"os"
)

// version is set at build time via -ldflags "-X main.version=..."
var version = "dev"

func exitWithError(err error) {
	log.Fatalf("An error occurred: %v\n", err)
}

func sendNotifications(aircraft []geebee.AircraftOutput) error {
	if len(aircraft) == 0 {
		return nil
	}

	// Custom URL
	if configuration.CustomWebhookUrl != "" {
		err := notification.SendCustomUrlMessage(aircraft)
		if err != nil {
			return err
		}
	}

	// Discord
	if configuration.DiscordWebhookUrl != "" {
		err := notification.SendDiscordMessage(aircraft)
		if err != nil {
			return err
		}
	}

	// Slack
	if configuration.SlackWebhookUrl != "" {
		err := notification.SendSlackMessage(aircraft)
		if err != nil {
			return err
		}
	}

	// Terminal
	notification.SendTerminalMessage(aircraft)

	return nil
}

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Println(version)
		return
	}

	fmt.Printf("=== Gee Bee %s ===\n", version)

	configuration.GetConfig()
	log.Printf("Watching for the following tail numbers: %s", configuration.TailNumbers)

	alreadySpottedAircraft := geebee.LoadState(configuration.StateFilePath)

	aircraft, err := geebee.HandleAircraft(&alreadySpottedAircraft)
	if err != nil {
		exitWithError(err)
	}

	if err := sendNotifications(aircraft); err != nil {
		exitWithError(err)
	}

	if err := geebee.SaveState(configuration.StateFilePath, alreadySpottedAircraft); err != nil {
		log.Printf("failed to save state: %v", err)
		os.Exit(1)
	}

	fmt.Printf("=== Done. %d new notification(s) sent. ===\n", len(aircraft))
}
