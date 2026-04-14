package main

import (
	"geebee/internal/configuration"
	"geebee/internal/geebee"
	"geebee/internal/metrics"
	"geebee/internal/notification"
	"log"
	"time"
)

func exitWithError(err error) {
	log.Fatalf("An error occurred: %v\n", err)
}

func geebeeHandler(alreadySpottedAircraft *[]geebee.Aircraft) {
	aircraft, err := geebee.HandleAircraft(alreadySpottedAircraft)
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

	for _, ac := range aircraft {
		metrics.AircraftSpottedTotal.WithLabelValues(ac.Registration).Inc()
	}

	// Custom URL
	if configuration.CustomWebhookUrl != "" {
		err := notification.SendCustomUrlMessage(aircraft)
		if err != nil {
			return err
		}
		metrics.NotificationsSentTotal.WithLabelValues("custom_url").Inc()
	}

	// Discord
	if configuration.DiscordWebhookUrl != "" {
		err := notification.SendDiscordMessage(aircraft)
		if err != nil {
			return err
		}
		metrics.NotificationsSentTotal.WithLabelValues("discord").Inc()
	}

	// Slack
	if configuration.SlackWebhookUrl != "" {
		err := notification.SendSlackMessage(aircraft)
		if err != nil {
			return err
		}
		metrics.NotificationsSentTotal.WithLabelValues("slack").Inc()
	}

	// Terminal
	notification.SendTerminalMessage(aircraft)
	metrics.NotificationsSentTotal.WithLabelValues("terminal").Inc()

	return nil
}

func HandleGeeBee() {
	log.Printf("Watching for the following tail numbers: %s", configuration.TailNumbers)

	var alreadySpottedAircraft []geebee.Aircraft

	for {
		geebeeHandler(&alreadySpottedAircraft)

		time.Sleep(time.Duration(configuration.FetchInterval) * time.Second)
	}
}

func main() {
	configuration.GetConfig()

	metrics.StartServer(configuration.MetricsPort)

	HandleGeeBee()
}
