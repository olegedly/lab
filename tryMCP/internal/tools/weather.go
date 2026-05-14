package tools

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"trymcp/internal/weather"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type WeatherInput struct {
	City         string `json:"city" jsonschema:"the city name to search for,required"`
	CountryCode  string `json:"country_code,omitempty" jsonschema:"optional ISO 3166-1 alpha-2 country code to narrow the search"`
	ForecastDays int    `json:"forecast_days,omitempty" jsonschema:"number of forecast days (1-3),default=3"`
}

type WeatherOutput struct {
	Location WeatherLocation `json:"location"`
	Current  WeatherCurrent  `json:"current"`
	Daily    []WeatherDay    `json:"daily"`
}

type WeatherLocation struct {
	City        string  `json:"city"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Timezone    string  `json:"timezone"`
}

type WeatherCurrent struct {
	Time                string  `json:"time"`
	Temperature2m       float64 `json:"temperature_2m"`
	ApparentTemperature float64 `json:"apparent_temperature"`
	RelativeHumidity2m  float64 `json:"relative_humidity_2m"`
	WeatherCode         float64 `json:"weather_code"`
	WindSpeed10m        float64 `json:"wind_speed_10m"`
	IsDay               float64 `json:"is_day"`
}

type WeatherDay struct {
	Date             string  `json:"date"`
	WeatherCode      float64 `json:"weather_code"`
	Temperature2mMax float64 `json:"temperature_2m_max"`
	Temperature2mMin float64 `json:"temperature_2m_min"`
	PrecipitationSum float64 `json:"precipitation_sum"`
}

func HandleWeather(ctx context.Context, _ *mcp.CallToolRequest, input WeatherInput) (*mcp.CallToolResult, any, error) {
	if input.ForecastDays < 1 {
		input.ForecastDays = 1
	} else if input.ForecastDays > 3 {
		input.ForecastDays = 3
	}

	geo, err := weather.Geocode(ctx, http.DefaultClient, input.City, input.CountryCode)
	if err != nil {
		return errorResult("weather", err), nil, nil
	}

	forecast, err := weather.Forecast(ctx, http.DefaultClient, geo.Latitude, geo.Longitude, input.ForecastDays)
	if err != nil {
		return errorResult("weather", err), nil, nil
	}

	output := WeatherOutput{
		Location: WeatherLocation{
			City:        geo.Name,
			Country:     geo.Country,
			CountryCode: geo.CountryCode,
			Latitude:    forecast.Latitude,
			Longitude:   forecast.Longitude,
			Timezone:    forecast.Timezone,
		},
		Current: WeatherCurrent{
			Time:                forecast.Current.Time,
			Temperature2m:       forecast.Current.Temperature2m,
			ApparentTemperature: forecast.Current.ApparentTemperature,
			RelativeHumidity2m:  forecast.Current.RelativeHumidity2m,
			WeatherCode:         forecast.Current.WeatherCode,
			WindSpeed10m:        forecast.Current.WindSpeed10m,
			IsDay:               forecast.Current.IsDay,
		},
		Daily: make([]WeatherDay, len(forecast.Daily.Time)),
	}

	for i := range forecast.Daily.Time {
		output.Daily[i] = WeatherDay{
			Date:             forecast.Daily.Time[i],
			WeatherCode:      forecast.Daily.WeatherCode[i],
			Temperature2mMax: forecast.Daily.Temperature2mMax[i],
			Temperature2mMin: forecast.Daily.Temperature2mMin[i],
			PrecipitationSum: forecast.Daily.PrecipitationSum[i],
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s, %s — %d-day forecast\n\n", geo.Name, geo.Country, input.ForecastDays)
	fmt.Fprintf(&b, "Current: %.1f°C (feels like %.1f°C), %s, humidity %.0f%%, wind %.1f km/h",
		forecast.Current.Temperature2m,
		forecast.Current.ApparentTemperature,
		wmoDescription(forecast.Current.WeatherCode),
		forecast.Current.RelativeHumidity2m,
		forecast.Current.WindSpeed10m,
	)

	for i := range forecast.Daily.Time {
		fmt.Fprintf(&b, "\n\n%s: High %.1f°C / Low %.1f°C, %s, %.1f mm rain",
			forecast.Daily.Time[i],
			forecast.Daily.Temperature2mMax[i],
			forecast.Daily.Temperature2mMin[i],
			wmoDescription(forecast.Daily.WeatherCode[i]),
			forecast.Daily.PrecipitationSum[i],
		)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: b.String()},
		},
		StructuredContent: output,
	}, nil, nil
}

func wmoDescription(code float64) string {
	switch int(code) {
	case 0:
		return "clear sky"
	case 1:
		return "mainly clear"
	case 2:
		return "partly cloudy"
	case 3:
		return "overcast"
	case 45:
		return "foggy"
	case 48:
		return "rime fog"
	case 51:
		return "light drizzle"
	case 53:
		return "moderate drizzle"
	case 55:
		return "dense drizzle"
	case 56:
		return "light freezing drizzle"
	case 57:
		return "dense freezing drizzle"
	case 61:
		return "slight rain"
	case 63:
		return "moderate rain"
	case 65:
		return "heavy rain"
	case 66:
		return "light freezing rain"
	case 67:
		return "heavy freezing rain"
	case 71:
		return "slight snow"
	case 73:
		return "moderate snow"
	case 75:
		return "heavy snow"
	case 77:
		return "snow grains"
	case 80:
		return "slight rain showers"
	case 81:
		return "moderate rain showers"
	case 82:
		return "violent rain showers"
	case 85:
		return "slight snow showers"
	case 86:
		return "heavy snow showers"
	case 95:
		return "thunderstorm"
	case 96:
		return "thunderstorm with slight hail"
	case 99:
		return "thunderstorm with heavy hail"
	default:
		return "unknown"
	}
}

func errorResult(domain string, err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("%s error: %s", domain, err)}},
		IsError: true,
	}
}
