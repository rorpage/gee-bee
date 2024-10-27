package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"geebee/internal/geebee"
	"log"
	"net/http"
)

type Notification struct {
	Message interface{}
	Type    string
	URL     string
}

const (
	CustomUrl = "CustomUrl"
	Discord   = "Discord"
	Markdown  = "Markdown"
	Slack     = "Slack"
)

func formatRegistration(ac geebee.AircraftOutput, notificationType string) string {
	if notificationType == Markdown {
		return fmt.Sprintf("[%s](%s)", ac.Registration, ac.ImageURL)
	}

	if notificationType == Slack {
		if ac.ImageURL == "" {
			return fmt.Sprintf("*Registration:* %s", ac.Registration)
		}

		return fmt.Sprintf("*Registration:* <%s|%s>", ac.ImageURL, ac.Registration)
	}

	if ac.ImageURL == "" {
		return ac.Registration
	}

	return ac.Registration
}

func SendMessage(notification Notification) error {
	data, err := json.Marshal(notification.Message)
	if err != nil {
		return err
	}

	resp, err := http.Post(notification.URL, "application/json",
		bytes.NewReader(data))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		log.Printf("%s\n", string(data))
		return fmt.Errorf(fmt.Sprintf("Received status code %v", resp.StatusCode))
	}

	log.Printf("A %s notification has been sent!\n", notification.Type)

	return nil
}

func printSpeed(ac geebee.AircraftOutput) string {
	return fmt.Sprintf("%dkn | %dkm/h", ac.Speed, geebee.ConvertKnotsToKilometersPerHour(ac.Speed))
}

func printAltitude(ac geebee.AircraftOutput) string {
	return fmt.Sprintf("%vft | %dm", ac.Altitude, geebee.ConvertFeetToMeters(ac.Altitude))
}

func printHeading(ac geebee.AircraftOutput) string {
	return fmt.Sprintf("%.0f°", ac.Heading)
}
