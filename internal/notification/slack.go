package notification

import (
	"fmt"
	"geebee/internal/configuration"
	"geebee/internal/geebee"
)

type SlackMessage struct {
	Blocks []Block `json:"blocks"`
}

type Block struct {
	Type     string  `json:"type,omitempty"`
	Fields   []Field `json:"fields,omitempty"`
	Title    *Title  `json:"title,omitempty"`
	ImageURL string  `json:"image_url,omitempty"`
	AltText  string  `json:"alt_text,omitempty"`
}

type Title struct {
	Type  string `json:"type,omitempty"`
	Text  string `json:"text,omitempty"`
	Emoji bool   `json:"emoji,omitempty"`
}

type Field struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`
}

func buildSlackMessage(aircraft []geebee.AircraftOutput) (SlackMessage, error) {
	var blocks []Block

	blocks = append(blocks, Block{
		Type: "section",
		Fields: []Field{
			{
				Type: "mrkdwn",
				Text: ":small_airplane: A plane has been spotted! :small_airplane:",
			},
		},
	})

	for _, ac := range aircraft {
		blocks = append(blocks, Block{
			Type: "section",
			Fields: []Field{
				{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*Callsign:* <%s|%s>", ac.TrackerURL, ac.Callsign),
				},
				{
					Type: "mrkdwn",
					Text: formatRegistration(ac, Slack),
				},
				{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*Speed:* %s", printSpeed(ac)),
				},
				{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*Altitude:* %s", printAltitude(ac)),
				},
				{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*Heading:* %s", printHeading(ac)),
				},
				{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*Type:* %s (%s)", ac.Description, ac.OwnerOperator),
				},
			},
		})

		imageURL := ac.ImageThumbnailURL
		if imageURL != "" {
			blocks = append(blocks,
				Block{
					Type: "image",
					Title: &Title{
						Type:  "plain_text",
						Text:  fmt.Sprintf("%s - %s", ac.Description, ac.Registration),
						Emoji: true,
					},
					ImageURL: imageURL,
					AltText:  fmt.Sprintf("%s with registration number %s", ac.Description, ac.Registration),
				})
		}

		blocks = append(blocks,
			Block{
				Type: "divider",
			},
		)
	}

	slackMessage := SlackMessage{Blocks: blocks}

	return slackMessage, nil
}

func SendSlackMessage(aircraft []geebee.AircraftOutput) error {
	message, err := buildSlackMessage(aircraft)
	if err != nil {
		return err
	}

	notification := Notification{
		Message: message,
		Type:    Slack,
		URL:     configuration.SlackWebhookUrl,
	}

	err = SendMessage(notification)
	if err != nil {
		return err
	}

	return nil
}
