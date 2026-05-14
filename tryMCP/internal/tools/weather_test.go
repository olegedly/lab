package tools_test

import (
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"trymcp/internal/tools"
)

func TestWeatherOutputSerialization(t *testing.T) {
	output := tools.WeatherOutput{
		Location: tools.WeatherLocation{
			City:        "London",
			Country:     "United Kingdom",
			CountryCode: "GB",
			Latitude:    51.5,
			Longitude:   -0.1,
			Timezone:    "Europe/London",
		},
		Current: tools.WeatherCurrent{
			Time:                "2026-05-14T12:00",
			Temperature2m:       15.2,
			ApparentTemperature: 13.1,
			RelativeHumidity2m:  65,
			WeatherCode:         2,
			WindSpeed10m:        12.5,
			IsDay:               1,
		},
		Daily: []tools.WeatherDay{
			{Date: "2026-05-14", WeatherCode: 2, Temperature2mMax: 18.0, Temperature2mMin: 10.0, PrecipitationSum: 0.0},
			{Date: "2026-05-15", WeatherCode: 3, Temperature2mMax: 20.5, Temperature2mMin: 12.0, PrecipitationSum: 2.5},
		},
	}

	data, err := json.Marshal(output)
	require.NoError(t, err)

	var decoded tools.WeatherOutput
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, "London", decoded.Location.City)
	assert.Equal(t, "GB", decoded.Location.CountryCode)
	assert.InDelta(t, 51.5, decoded.Location.Latitude, 0.01)
	assert.Equal(t, 15.2, decoded.Current.Temperature2m)
	require.Len(t, decoded.Daily, 2)
	assert.Equal(t, "2026-05-14", decoded.Daily[0].Date)
	assert.Equal(t, 20.5, decoded.Daily[1].Temperature2mMax)
}

func TestWeatherInput_NoCountryCode(t *testing.T) {
	input := tools.WeatherInput{City: "Paris", ForecastDays: 2}
	assert.Equal(t, "Paris", input.City)
	assert.Empty(t, input.CountryCode)
	assert.Equal(t, 2, input.ForecastDays)
}

func TestWeatherInput_ZeroForecastDays(t *testing.T) {
	input := tools.WeatherInput{City: "Paris", ForecastDays: 0}
	assert.Equal(t, 0, input.ForecastDays)
}

func TestWeatherCallToolResult_Error(t *testing.T) {
	// The errorResult helper is unexported, so test indirectly via json tags
	raw := `{"content":[{"text":"weather error: city not found","type":"text"}],"isError":true}`
	var result mcp.CallToolResult
	require.NoError(t, json.Unmarshal([]byte(raw), &result))
	assert.True(t, result.IsError)
	require.Len(t, result.Content, 1)
	tc, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Equal(t, "weather error: city not found", tc.Text)
}
