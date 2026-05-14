package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"trymcp/internal/restcountries"
)

type CountrySummaryInput struct {
	CountryName string `json:"country_name" jsonschema:"the country name to look up,required"`
	FullText    bool   `json:"full_text,omitempty" jsonschema:"exact full-text match,default=false"`
}

type CountryOutput struct {
	Country    CountryInfo             `json:"country"`
	Capitals   []string                `json:"capitals"`
	Population int64                   `json:"population"`
	Currencies []restcountries.Currency `json:"currencies"`
}

type CountryInfo struct {
	CommonName   string `json:"common_name"`
	OfficialName string `json:"official_name"`
}

func HandleCountrySummary(ctx context.Context, _ *mcp.CallToolRequest, input CountrySummaryInput) (*mcp.CallToolResult, any, error) {
	sum, err := restcountries.Lookup(ctx, http.DefaultClient, input.CountryName, input.FullText)
	if err != nil {
		return errorResult("restcountries", err), nil, nil
	}

	output := CountryOutput{
		Country: CountryInfo{
			CommonName:   sum.CommonName,
			OfficialName: sum.OfficialName,
		},
		Capitals:   sum.Capitals,
		Population: sum.Population,
		Currencies: sum.Currencies,
	}

	summary, _ := json.Marshal(output)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(summary)},
		},
		StructuredContent: output,
	}, nil, nil
}

func HandleAllCountries(ctx context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	summaries, err := restcountries.All(ctx, http.DefaultClient)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("restcountries error: %s", err)}},
			IsError: true,
		}, nil
	}

	raw, _ := json.Marshal(summaries)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(raw)},
		},
		StructuredContent: summaries,
	}, nil
}
