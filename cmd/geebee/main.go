package main

import (
	"geebee/internal/configuration"
	"geebee/internal/geebee"
	"geebee/internal/notification"
	"log"
	"time"
)

func exitWithError(err error) {
	log.Fatalf("An error occurred: %v\n", err)
}

func geebeeHandler(spotted map[string]bool) {
	aircraft, err := geebee.HandleAircraft(spotted)
	if err != nil {
		exitWithError(err)
	}

	err = sendNotifications(aircraft)
	if err != nil {
		exitWithError(err)
	}
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

func HandleGeeBee() {
	log.Printf("Watching for the following tail numbers: %s", configuration.TailNumbers)

	spotted := map[string]bool{}

	for {
		geebeeHandler(spotted)

		time.Sleep(time.Duration(configuration.FetchInterval) * time.Second)
	}
}

func main() {
	configuration.GetConfig()

	HandleGeeBee()
}
