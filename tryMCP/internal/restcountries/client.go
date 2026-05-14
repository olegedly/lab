package restcountries

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var (
	ErrCountryNotFound = errors.New("country not found")
	ErrUpstreamFailure = errors.New("upstream API failure")
)

type nameInfo struct {
	Common   string `json:"common"`
	Official string `json:"official"`
}

type currencyInfo struct {
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
}

type rawCountry struct {
	Name        nameInfo                `json:"name"`
	Capital     []string                `json:"capital"`
	Population  int64                   `json:"population"`
	Currencies  map[string]currencyInfo `json:"currencies"`
}

type Currency struct {
	Code   string  `json:"code"`
	Name   string  `json:"name"`
	Symbol *string `json:"symbol"`
}

type Summary struct {
	CommonName   string     `json:"common_name"`
	OfficialName string     `json:"official_name"`
	Capitals     []string   `json:"capitals"`
	Population   int64      `json:"population"`
	Currencies   []Currency `json:"currencies"`
}

func rawToSummary(r rawCountry) Summary {
	sum := Summary{
		CommonName:   r.Name.Common,
		OfficialName: r.Name.Official,
		Capitals:     r.Capital,
		Population:   r.Population,
	}
	for code, cur := range r.Currencies {
		sym := cur.Symbol
		sum.Currencies = append(sum.Currencies, Currency{
			Code:   code,
			Name:   cur.Name,
			Symbol: &sym,
		})
	}
	if sum.Currencies == nil {
		sum.Currencies = []Currency{}
	}
	if sum.Capitals == nil {
		sum.Capitals = []string{}
	}
	return sum
}

// Lookup fetches country info by name. On multiple matches, returns the first.
func Lookup(ctx context.Context, httpClient *http.Client, name string, fullText bool) (*Summary, error) {
	u := fmt.Sprintf("https://restcountries.com/v3.1/name/%s?fields=name,capital,currencies,population",
		name)
	if fullText {
		u += "&fullText=true"
	}

	resp, err := httpClient.Get(u)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUpstreamFailure, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrCountryNotFound
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: API returned status %d: %s",
			ErrUpstreamFailure, resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUpstreamFailure, err)
	}

	var raw []rawCountry
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUpstreamFailure, err)
	}

	if len(raw) == 0 {
		return nil, ErrCountryNotFound
	}

	sum := rawToSummary(raw[0])
	return &sum, nil
}

// All fetches summaries for all countries.
func All(ctx context.Context, httpClient *http.Client) ([]Summary, error) {
	resp, err := httpClient.Get("https://restcountries.com/v3.1/all?fields=name,capital,currencies,population")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUpstreamFailure, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: API returned status %d: %s",
			ErrUpstreamFailure, resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUpstreamFailure, err)
	}

	var raw []rawCountry
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUpstreamFailure, err)
	}

	summaries := make([]Summary, len(raw))
	for i, r := range raw {
		summaries[i] = rawToSummary(r)
	}
	return summaries, nil
}
