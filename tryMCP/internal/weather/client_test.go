package weather_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"trymcp/internal/weather"
)

func TestForecastUnmarshal(t *testing.T) {
	raw := `{
		"latitude": 51.5,
		"longitude": -0.1,
		"timezone": "Europe/London",
		"current": {
			"time": "2026-05-14T12:00",
			"temperature_2m": 15.2,
			"apparent_temperature": 13.1,
			"relative_humidity_2m": 65,
			"weather_code": 2,
			"wind_speed_10m": 12.5,
			"is_day": 1
		},
		"current_units": {"temperature_2m": "°C"},
		"daily": {
			"time": ["2026-05-14", "2026-05-15"],
			"weather_code": [2, 3],
			"temperature_2m_max": [18.0, 20.5],
			"temperature_2m_min": [10.0, 12.0],
			"precipitation_sum": [0.0, 2.5]
		},
		"daily_units": {"temperature_2m_max": "°C"}
	}`

	var f weather.ForecastResponse
	require.NoError(t, json.Unmarshal([]byte(raw), &f))

	assert.Equal(t, 51.5, f.Latitude)
	assert.Equal(t, -0.1, f.Longitude)
	assert.Equal(t, "Europe/London", f.Timezone)

	assert.Equal(t, 15.2, f.Current.Temperature2m)
	assert.Equal(t, float64(1), f.Current.IsDay)

	require.Len(t, f.Daily.Time, 2)
	assert.Equal(t, []string{"2026-05-14", "2026-05-15"}, f.Daily.Time)
	assert.Equal(t, []float64{18.0, 20.5}, f.Daily.Temperature2mMax)
}

func TestForecastUnmarshal_Empty(t *testing.T) {
	raw := `{}`
	var f weather.ForecastResponse
	require.NoError(t, json.Unmarshal([]byte(raw), &f))
	assert.Zero(t, f.Latitude)
}

func TestGeocodeUnmarshal(t *testing.T) {
	raw := `{
		"results": [
			{"name":"London","country":"United Kingdom","country_code":"GB","latitude":51.5,"longitude":-0.1,"timezone":"Europe/London","population":8982000},
			{"name":"London","country":"Canada","country_code":"CA","latitude":42.98,"longitude":-81.25,"timezone":"America/Toronto","population":590000}
		]
	}`

	var g struct {
		Results []weather.GeoResult `json:"results"`
	}
	require.NoError(t, json.Unmarshal([]byte(raw), &g))
	require.Len(t, g.Results, 2)
	assert.Equal(t, "London", g.Results[0].Name)
	assert.Equal(t, "GB", g.Results[0].CountryCode)
	assert.Equal(t, int64(8982000), g.Results[0].Population)
}

func TestGeocodeUnmarshal_NoResults(t *testing.T) {
	raw := `{}`
	var g struct {
		Results []weather.GeoResult `json:"results"`
	}
	require.NoError(t, json.Unmarshal([]byte(raw), &g))
	assert.Nil(t, g.Results)
}

func TestSentinelErrors(t *testing.T) {
	assert.ErrorIs(t, weather.ErrCityNotFound, weather.ErrCityNotFound)
	assert.ErrorIs(t, weather.ErrUpstreamFailure, weather.ErrUpstreamFailure)
}
