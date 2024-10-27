package notification

import (
	"fmt"
	"geebee/internal/configuration"
	"geebee/internal/geebee"
	"strings"
)

type CustomUrlMessage struct {
	Altitude      string `json:"altitude"`
	Callsign      string `json:"callsign"`
	Description   string `json:"description"`
	Heading       string `json:"heading"`
	ImageUrl      string `json:"imageUrl"`
	OwnerOperator string `json:"ownerOperator"`
	Speed         string `json:"speed"`
	Squawk        string `json:"squawk"`
	TailNumber    string `json:"tailNumber"`
	Type          string `json:"type"`
	Url           string `json:"url"`
}

func SendCustomUrlMessage(aircraft []geebee.AircraftOutput) error {
	notification := Notification{
		Message: buildCustomMessage(aircraft),
		Type:    CustomUrl,
		URL:     configuration.CustomWebhookUrl,
	}

	err := SendMessage(notification)
	if err != nil {
		return err
	}

	return nil
}

func buildCustomMessage(aircraft []geebee.AircraftOutput) CustomUrlMessage {
	var msg CustomUrlMessage
	ac := aircraft[0]

	msg.Altitude = printAltitude(ac)
	msg.Callsign = strings.TrimSpace(ac.Callsign)
	msg.Description = ac.Description
	msg.Heading = fmt.Sprintf("%.0f", ac.Heading)
	msg.ImageUrl = ac.ImageURL
	msg.OwnerOperator = strings.TrimSpace(ac.OwnerOperator)
	msg.Speed = printSpeed(ac)
	msg.Squawk = ac.Squawk
	msg.TailNumber = strings.TrimSpace(ac.Registration)
	msg.Type = strings.TrimSpace(ac.Type)
	msg.Url = ac.TrackerURL

	return msg
}
