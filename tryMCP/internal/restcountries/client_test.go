package restcountries_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"trymcp/internal/restcountries"
)

func TestRawCountryUnmarshal(t *testing.T) {
	raw := `[{
		"name": {"common": "France", "official": "French Republic"},
		"capital": ["Paris"],
		"population": 67390000,
		"currencies": {"EUR": {"name": "Euro", "symbol": "€"}}
	}]`

	var countries []struct {
		Name struct {
			Common   string `json:"common"`
			Official string `json:"official"`
		} `json:"name"`
		Capital    []string `json:"capital"`
		Population int64    `json:"population"`
		Currencies map[string]struct {
			Name   string `json:"name"`
			Symbol string `json:"symbol"`
		} `json:"currencies"`
	}
	require.NoError(t, json.Unmarshal([]byte(raw), &countries))
	require.Len(t, countries, 1)
	assert.Equal(t, "France", countries[0].Name.Common)
	assert.Equal(t, "French Republic", countries[0].Name.Official)
	assert.Equal(t, []string{"Paris"}, countries[0].Capital)
	assert.Equal(t, int64(67390000), countries[0].Population)
	assert.Equal(t, "Euro", countries[0].Currencies["EUR"].Name)
}

func TestRawCountryUnmarshal_MissingCapitals(t *testing.T) {
	raw := `[{
		"name": {"common": "Test", "official": "Test"},
		"capital": [],
		"population": 1000,
		"currencies": {}
	}]`

	var countries []struct {
		Capital    []string `json:"capital"`
		Currencies map[string]any `json:"currencies"`
	}
	require.NoError(t, json.Unmarshal([]byte(raw), &countries))
	assert.Empty(t, countries[0].Capital)
	assert.Empty(t, countries[0].Currencies)
}

func TestSummary_EmptyFields(t *testing.T) {
	raw := `[{
		"name": {"common": "X", "official": "X"},
		"capital": [],
		"population": 0,
		"currencies": {}
	}]`

	var rawList []struct {
		Name struct {
			Common   string `json:"common"`
			Official string `json:"official"`
		} `json:"name"`
		Capital    []string `json:"capital"`
		Population int64    `json:"population"`
		Currencies map[string]struct {
			Name   string `json:"name"`
			Symbol string `json:"symbol"`
		} `json:"currencies"`
	}
	require.NoError(t, json.Unmarshal([]byte(raw), &rawList))
	require.Len(t, rawList, 1)
	assert.Empty(t, rawList[0].Capital)
	assert.Empty(t, rawList[0].Currencies)
}

func TestSentinelErrors(t *testing.T) {
	assert.ErrorIs(t, restcountries.ErrCountryNotFound, restcountries.ErrCountryNotFound)
	assert.ErrorIs(t, restcountries.ErrUpstreamFailure, restcountries.ErrUpstreamFailure)
}
