package geebee

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"

	"geebee/internal/configuration"
	"geebee/internal/planespotter"

	"github.com/jftuga/geodist"
)

const baseURL = "https://api.adsb.one/v2"

func CalculateDistance(source geodist.Coord, destination geodist.Coord) int {
	_, kilometers := geodist.HaversineDistance(source, destination)

	return int(kilometers)
}

func checkAircraft() (aircraft []Aircraft, err error) {
	var flightData FlightData

	tailNumbersString := strings.Join(configuration.TailNumbers[:], ",")
	endpoint, err := url.JoinPath(baseURL, "reg", tailNumbersString)

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	err = json.Unmarshal(body, &flightData)
	if err != nil {
		return nil, err
	}

	return flightData.AC, nil
}

func newlySpotted(aircraft Aircraft, spottedAircraft []Aircraft) bool {
	return !containsAircraft(aircraft, spottedAircraft)
}

func containsAircraft(aircraft Aircraft, aircraftList []Aircraft) bool {
	for _, ac := range aircraftList {
		if ac.ICAO == aircraft.ICAO {
			return true
		}
	}

	return false
}

func updateSpottedAircraft(alreadySpottedAircraft, filteredAircraft []Aircraft) (aircraft []Aircraft) {
	for _, ac := range alreadySpottedAircraft {
		if containsAircraft(ac, filteredAircraft) {
			aircraft = append(aircraft, ac)
		}
	}

	return aircraft
}

func validateAircraft(allFilteredAircraft []Aircraft, alreadySpottedAircraft *[]Aircraft) (newlySpottedAircraft, updatedSpottedAircraft []Aircraft) {
	for _, ac := range allFilteredAircraft {
		if newlySpotted(ac, *alreadySpottedAircraft) {
			newlySpottedAircraft = append(newlySpottedAircraft, ac)
			*alreadySpottedAircraft = append(*alreadySpottedAircraft, ac)
		}
	}

	*alreadySpottedAircraft = updateSpottedAircraft(*alreadySpottedAircraft, allFilteredAircraft)

	return newlySpottedAircraft, *alreadySpottedAircraft
}

func HandleAircraft(alreadySpottedAircraft *[]Aircraft) (aircraft []AircraftOutput, err error) {
	var newlySpottedAircraft []Aircraft

	allAircraftInRange, err := checkAircraft()
	if err != nil {
		return nil, err
	}

	newlySpottedAircraft, *alreadySpottedAircraft = validateAircraft(allAircraftInRange, alreadySpottedAircraft)
	newlySpottedAircraftOutput, err := CreateAircraftOutput(newlySpottedAircraft)
	if err != nil {
		return nil, err
	}

	return newlySpottedAircraftOutput, nil
}

func isAircraftMilitary(aircraft Aircraft) bool {
	return aircraft.DbFlags == 1
}

func ConvertKnotsToKilometersPerHour(knots int) int {
	return int(float64(knots) * 1.852)
}

func ConvertFeetToMeters(feet float64) int {
	return int(feet * 0.3048)
}

func validateFields(aircraft Aircraft) Aircraft {
	if aircraft.Callsign == "" {
		aircraft.Callsign = "UNKNOWN"
	}

	if aircraft.AltBaro == "groundft" || aircraft.AltBaro == "ground" || aircraft.AltBaro == nil {
		aircraft.AltBaro = float64(0)
	}

	altitudeBarometricFloat := aircraft.AltBaro.(float64)
	if altitudeBarometricFloat < 0 {
		altitudeBarometricFloat = 0
		aircraft.AltBaro = altitudeBarometricFloat
	}

	return aircraft
}

func CreateAircraftOutput(aircraft []Aircraft) (acOutputs []AircraftOutput, err error) {
	var acOutput AircraftOutput

	for _, ac := range aircraft {
		ac = validateFields(ac)

		image := planespotter.GetImageFromAPI(ac.ICAO, ac.Registration)

		acOutput.Altitude = ac.AltBaro.(float64)
		acOutput.Callsign = ac.Callsign
		acOutput.Description = ac.Description
		acOutput.Heading = ac.Track
		acOutput.ICAO = ac.ICAO
		acOutput.OwnerOperator = ac.OwnerOperator
		acOutput.Registration = ac.Registration
		acOutput.Speed = int(ac.GS)
		acOutput.Squawk = ac.Squawk
		acOutput.Type = ac.PlaneType

		acOutput.TrackerURL = fmt.Sprintf(
			"https://globe.adsbexchange.com/?icao=%v&SiteLat=%f&SiteLon=%f&zoom=11&enableLabels&extendedLabels=1&noIsolation",
			ac.ICAO, ac.Lat, ac.Lon,
		)

		acOutput.ImageThumbnailURL = image.ThumbnailLarge.Src
		acOutput.ImageURL = image.Link
		acOutput.Military = isAircraftMilitary(ac)
		acOutputs = append(acOutputs, acOutput)
	}
	return acOutputs, nil
}

func CalculateBearing(source geodist.Coord, target geodist.Coord) float64 {
	y := math.Sin(toRadians(target.Lon-source.Lon)) * math.Cos(toRadians(target.Lat))
	x := math.Cos(toRadians(source.Lat))*math.Sin(toRadians(target.Lat)) - math.Sin(toRadians(source.Lat))*math.Cos(toRadians(target.Lat))*math.Cos(toRadians(target.Lon-source.Lon))

	bearing := math.Atan2(y, x)
	bearing = (toDegrees(bearing) + 360)

	if bearing >= 360 {
		bearing -= 360
	}

	return bearing
}

func toRadians(degrees float64) float64 {
	return degrees * (math.Pi / 180)
}

func toDegrees(rad float64) float64 {
	return rad * (180 / math.Pi)
}
