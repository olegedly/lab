package weather

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
)

var (
	ErrCityNotFound     = errors.New("city not found")
	ErrCityAmbiguous    = errors.New("city name is ambiguous")
	ErrUpstreamFailure  = errors.New("upstream API failure")
)

type GeoResult struct {
	Name        string  `json:"name"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Timezone    string  `json:"timezone"`
	Admin1      string  `json:"admin1"`
	Admin2      string  `json:"admin2"`
	Population  int64   `json:"population"`
}

type CurrentWeather struct {
	Time                string  `json:"time"`
	Temperature2m       float64 `json:"temperature_2m"`
	ApparentTemperature float64 `json:"apparent_temperature"`
	RelativeHumidity2m  float64 `json:"relative_humidity_2m"`
	WeatherCode         float64 `json:"weather_code"`
	WindSpeed10m        float64 `json:"wind_speed_10m"`
	IsDay               float64 `json:"is_day"`
}

type DailyForecast struct {
	Time             []string  `json:"time"`
	WeatherCode      []float64 `json:"weather_code"`
	Temperature2mMax []float64 `json:"temperature_2m_max"`
	Temperature2mMin []float64 `json:"temperature_2m_min"`
	PrecipitationSum []float64 `json:"precipitation_sum"`
}

type ForecastResponse struct {
	Latitude     float64            `json:"latitude"`
	Longitude    float64            `json:"longitude"`
	Timezone     string             `json:"timezone"`
	Current      CurrentWeather     `json:"current"`
	CurrentUnits json.RawMessage    `json:"current_units"`
	Daily        DailyForecast      `json:"daily"`
	DailyUnits   json.RawMessage    `json:"daily_units"`
}

type geocodeResponse struct {
	Results []GeoResult `json:"results"`
}

// Geocode searches for a city and returns the best matching result.
// If countryCode is non-empty, results are filtered to that country.
func Geocode(ctx context.Context, httpClient *http.Client, city, countryCode string) (*GeoResult, error) {
	u := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?name=%s&count=5&language=en&format=json",
		url.QueryEscape(city))

	resp, err := httpClient.Get(u)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUpstreamFailure, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: geocoding API returned status %d", ErrUpstreamFailure, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUpstreamFailure, err)
	}

	var geoResp geocodeResponse
	if err := json.Unmarshal(body, &geoResp); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUpstreamFailure, err)
	}

	if len(geoResp.Results) == 0 {
		return nil, ErrCityNotFound
	}

	results := geoResp.Results
	if countryCode != "" {
		cc := strings.ToLower(countryCode)
		filtered := slices.DeleteFunc(results, func(r GeoResult) bool {
			return strings.ToLower(r.CountryCode) != cc
		})
		if len(filtered) > 0 {
			results = filtered
		}
	}

	exactName := strings.ToLower(city)
	exactMatches := slices.DeleteFunc(results, func(r GeoResult) bool {
		return strings.ToLower(r.Name) != exactName
	})
	if len(exactMatches) > 0 {
		results = exactMatches
	}

	best := results[0]
	for _, r := range results[1:] {
		if r.Population > best.Population {
			best = r
		}
	}

	return &best, nil
}

// Forecast fetches weather data for the given coordinates.
func Forecast(ctx context.Context, httpClient *http.Client, lat, lon float64, forecastDays int) (*ForecastResponse, error) {
	u := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&current=temperature_2m,apparent_temperature,relative_humidity_2m,weather_code,wind_speed_10m,is_day&daily=weather_code,temperature_2m_max,temperature_2m_min,precipitation_sum&timezone=auto&forecast_days=%d&temperature_unit=celsius&wind_speed_unit=kmh",
		lat, lon, forecastDays,
	)

	resp, err := httpClient.Get(u)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUpstreamFailure, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: forecast API returned status %d: %s", ErrUpstreamFailure, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUpstreamFailure, err)
	}

	var forecast ForecastResponse
	if err := json.Unmarshal(body, &forecast); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUpstreamFailure, err)
	}

	return &forecast, nil
}
