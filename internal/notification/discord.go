package notification

import (
	"fmt"
	"geebee/internal/configuration"
	"geebee/internal/geebee"

	"github.com/bwmarrin/discordgo"
)

const (
	darkOrange  = 15755540
	lightOrange = 15829792
	darkYellow  = 15772952
	yellow      = 15055122
	lightGreen  = 10340365
	green       = 2278429
	greenBlue   = 1686636
	lightBlue   = 1292194
	darkBlue    = 2650083
	purple      = 10754265
	grey        = 3815994
)

func SendDiscordMessage(aircraft []geebee.AircraftOutput) error {
	message, err := buildDiscordMessage(aircraft)
	if err != nil {
		return err
	}

	notification := Notification{
		Message: message,
		Type:    Discord,
		URL:     configuration.DiscordWebhookUrl,
	}

	err = SendMessage(notification)
	if err != nil {
		return err
	}

	return nil
}

func getColorByAltitude(altitude int) int {
	switch {
	case altitude < 1000:
		return darkOrange
	case altitude >= 1000 && altitude < 2000:
		return lightOrange
	case altitude >= 2000 && altitude < 3000:
		return darkYellow
	case altitude >= 3000 && altitude < 5000:
		return yellow
	case altitude >= 5000 && altitude < 7000:
		return lightGreen
	case altitude >= 7000 && altitude < 10000:
		return green
	case altitude >= 10000 && altitude < 15000:
		return greenBlue
	case altitude >= 15000 && altitude < 20000:
		return lightBlue
	case altitude >= 20000 && altitude < 30000:
		return darkBlue
	case altitude >= 30000:
		return purple
	default:
		return grey
	}
}

func buildDiscordMessage(aircraft []geebee.AircraftOutput) (message discordgo.Message, err error) {
	message.Content = ":small_airplane: A plane has been spotted! :small_airplane:"
	var embeds []*discordgo.MessageEmbed
	for _, ac := range aircraft {
		embed := &discordgo.MessageEmbed{
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Callsign",
					Value:  fmt.Sprintf("[%s](%s)", ac.Callsign, ac.TrackerURL),
					Inline: true,
				},
				{
					Name:   "Registration",
					Value:  formatRegistration(ac, Markdown),
					Inline: true,
				},
				{
					Name:   "Speed",
					Value:  printSpeed(ac),
					Inline: true,
				},
				{
					Name:   "Altitude",
					Value:  printAltitude(ac),
					Inline: true,
				},
				{
					Name:   "Heading",
					Value:  printHeading(ac),
					Inline: true,
				},
				{
					Name:   "Type",
					Value:  fmt.Sprintf("%s (%s)", ac.Description, ac.OwnerOperator),
					Inline: true,
				},
			},
		}

		embed.Color = getColorByAltitude(int(ac.Altitude))

		imageURL := ac.ImageThumbnailURL
		if imageURL != "" {
			embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
				URL: imageURL,
			}
		}

		embeds = append(embeds, embed)
	}

	message.Embeds = embeds

	return message, nil
}
