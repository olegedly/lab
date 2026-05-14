package tools_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"trymcp/internal/restcountries"
	"trymcp/internal/tools"
)

func TestCountryOutputSerialization(t *testing.T) {
	sym := "€"
	output := tools.CountryOutput{
		Country: tools.CountryInfo{
			CommonName:   "France",
			OfficialName: "French Republic",
		},
		Capitals:   []string{"Paris"},
		Population: 67390000,
		Currencies: []restcountries.Currency{
			{Code: "EUR", Name: "Euro", Symbol: &sym},
		},
	}

	data, err := json.Marshal(output)
	require.NoError(t, err)

	var decoded tools.CountryOutput
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, "France", decoded.Country.CommonName)
	assert.Equal(t, "French Republic", decoded.Country.OfficialName)
	assert.Equal(t, []string{"Paris"}, decoded.Capitals)
	assert.Equal(t, int64(67390000), decoded.Population)
	require.Len(t, decoded.Currencies, 1)
	assert.Equal(t, "EUR", decoded.Currencies[0].Code)
	assert.Equal(t, "Euro", decoded.Currencies[0].Name)
	require.NotNil(t, decoded.Currencies[0].Symbol)
	assert.Equal(t, "€", *decoded.Currencies[0].Symbol)
}

func TestCountryInput(t *testing.T) {
	t.Run("with full text", func(t *testing.T) {
		input := tools.CountrySummaryInput{CountryName: "France", FullText: true}
		assert.Equal(t, "France", input.CountryName)
		assert.True(t, input.FullText)
	})

	t.Run("without full text", func(t *testing.T) {
		input := tools.CountrySummaryInput{CountryName: "France"}
		assert.Equal(t, "France", input.CountryName)
		assert.False(t, input.FullText)
	})
}

func TestCountryInfoString(t *testing.T) {
	info := tools.CountryInfo{CommonName: "Germany", OfficialName: "Federal Republic of Germany"}
	assert.Equal(t, "Germany", info.CommonName)
	assert.Equal(t, "Federal Republic of Germany", info.OfficialName)
}

func TestCountryOutput_NoCapitals(t *testing.T) {
	// Some territories may have no capital
	output := tools.CountryOutput{
		Country:    tools.CountryInfo{CommonName: "Somewhere", OfficialName: "Somewhere"},
		Capitals:   []string{},
		Population: 0,
		Currencies: []restcountries.Currency{},
	}

	data, err := json.Marshal(output)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"capitals":[]`)
}
