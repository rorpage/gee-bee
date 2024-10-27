package notification

import (
	"fmt"
	"geebee/internal/configuration"
	"geebee/internal/geebee"
	"log"
)

func FormatAircraft(aircraft geebee.AircraftOutput) string {
	return fmt.Sprintf("Callsign: %s\n"+
		"Description: %s\n"+
		"Type: %s\n"+
		"Owner/operator: %s\n"+
		"Tail number: %s\n"+
		"Altitude: %dft | %dm\n"+
		"Speed: %dkn | %dkm/h\n"+
		"Heading: %.0f°\n"+
		"TrackerURL: %s\n"+
		"ImageURL: %s\n",

		aircraft.Callsign, aircraft.Description, aircraft.Type, aircraft.OwnerOperator,
		aircraft.Registration, int(aircraft.Altitude), geebee.ConvertFeetToMeters(aircraft.Altitude),
		aircraft.Speed, geebee.ConvertKnotsToKilometersPerHour(aircraft.Speed),
		aircraft.Heading, aircraft.TrackerURL, aircraft.ImageURL)
}

func SendTerminalMessage(aircraft []geebee.AircraftOutput) {
	log.Println("🛩️ A plane has been spotted! 🛩️")

	if configuration.LogPlanesToConsole {
		for _, ac := range aircraft {
			fmt.Println(FormatAircraft(ac))
		}
	}
}
