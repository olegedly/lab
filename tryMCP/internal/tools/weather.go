package tools

import (
	"context"
	"fmt"
	"net/http"

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

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Weather in %s: %.1f°C, feels like %.1f°C",
				geo.Name, forecast.Current.Temperature2m, forecast.Current.ApparentTemperature)},
		},
		StructuredContent: output,
	}, nil, nil
}

func errorResult(domain string, err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("%s error: %s", domain, err)}},
		IsError: true,
	}
}
