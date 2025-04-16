package weather

import (
	"errors"
)

// WeatherData represents weather information for a location
type WeatherData struct {
	Location    string  `json:"location"`
	Temperature float64 `json:"temperature"`
	Condition   string  `json:"condition"`
	Humidity    int     `json:"humidity"`
	WindSpeed   float64 `json:"wind_speed"`
	Units       string  `json:"units"` // "metric" or "imperial"
}

// GetWeather returns weather data for a given location
// Currently returns stub data, to be implemented with a real weather API
func GetWeather(params map[string]interface{}) (WeatherData, error) {
	// Extract location from parameters
	locationParam, ok := params["location"]
	if !ok {
		return WeatherData{}, errors.New("location parameter is required")
	}

	location, ok := locationParam.(string)
	if !ok || location == "" {
		return WeatherData{}, errors.New("location must be a non-empty string")
	}

	// Extract units from parameters (default to metric if not provided)
	units := "metric"
	if unitsParam, ok := params["units"]; ok {
		if unitsStr, ok := unitsParam.(string); ok {
			if unitsStr == "imperial" || unitsStr == "metric" {
				units = unitsStr
			} else {
				return WeatherData{}, errors.New("units must be either 'metric' or 'imperial'")
			}
		}
	}

	// TODO: Implement actual weather API call here
	// For now, return stub data
	return WeatherData{
		Location:    location,
		Temperature: 22.5, // °C or °F depending on units
		Condition:   "Partly Cloudy",
		Humidity:    65,   // %
		WindSpeed:   10.5, // km/h or mph depending on units
		Units:       units,
	}, nil
}
