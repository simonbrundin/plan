// Package pricing innehåller en versionsstyrd prislista för agentkörningar.
//
// Priset sparas med varje körning, så att en kostnad kan reproduceras
// och revideras senare. Saknas priset för en modell används inte noll -
// det vore en falsk siffra - utan en dokumenterad schablon, och det
// markeras i kostnadsraden.
package pricing

import (
	"fmt"
	"strings"
)

// CurrentVersion är versionen av prislistan som gäller nu.
//
// Versionsnumret följer kalenderdatumet, så att det går att se vilken
// prislista som gällde vid en viss tidpunkt.
const CurrentVersion = "2026-11-24.1"

// Currency är valutan alla priser anges i.
const Currency = "USD"

// Price är priset för en miljon tokens av en sort.
type Price struct {
	// InputPer1M är priset per miljon indata-tokens.
	InputPer1M float64 `json:"input_per_1m"`
	// OutputPer1M är priset per miljon utdata-tokens.
	OutputPer1M float64 `json:"output_per_1m"`
	// CacheReadPer1M är priset per miljon tokens som lästs från cachen.
	CacheReadPer1M float64 `json:"cache_read_per_1m"`
}

// Tokens är den användning som ska prissättas.
type Tokens struct {
	Input     int `json:"input"`
	Output    int `json:"output"`
	CacheRead int `json:"cache_read"`
}

// Source visar varifrån ett pris kommer.
const (
	// SourcePriceList betyder att priset kom från prislistan.
	SourcePriceList = "price_list"
	// SourceFallback betyder att modellen saknades i prislistan och att
	// en schablon användes.
	SourceFallback = "fallback"
)

// Result är en prissatt användning.
type Result struct {
	CostUSD float64 `json:"cost_usd"`
	// Source visar om priset kom från prislistan eller från schablonen.
	Source string `json:"source"`
	// Known är false om modellen saknades i prislistan.
	Known bool `json:"known"`
	// PriceListVersion är versionen av prislistan som användes.
	PriceListVersion string `json:"price_list_version"`
	// Currency är valutan kostnaden anges i.
	Currency string `json:"currency"`
}

// PriceList är en version av prislistan.
type PriceList struct {
	// Version identifierar prislistan.
	Version string
	// Currency är valutan priserna anges i.
	Currency string
	// Fallback används när en modell saknas i listan. Den är avsiktligt
	// hög - en okänd modell ska inte kosta mindre än en känd.
	Fallback Price
	prices   map[string]Price
}

// Current returnerar den prislista som gäller nu.
//
// Nycklarna lagras i gemener, så att modellnamnet hittas oavsett hur det
// skrivs.
func Current() PriceList {
	list := PriceList{
		Version:  CurrentVersion,
		Currency: Currency,
		Fallback: Price{
			InputPer1M:     1.00,
			OutputPer1M:    3.00,
			CacheReadPer1M: 0.10,
		},
		prices: map[string]Price{},
	}

	for providerModel, price := range map[string]Price{
		"minimax/MiniMax-M2.7":           {InputPer1M: 0.20, OutputPer1M: 0.40, CacheReadPer1M: 0.02},
		"minimax/MiniMax-M1":             {InputPer1M: 0.10, OutputPer1M: 0.20, CacheReadPer1M: 0.01},
		"openai/gpt-4o-mini":             {InputPer1M: 0.15, OutputPer1M: 0.60, CacheReadPer1M: 0.075},
		"openai/gpt-4o":                  {InputPer1M: 2.50, OutputPer1M: 10.00, CacheReadPer1M: 1.25},
		"anthropic/claude-sonnet-4-7-25": {InputPer1M: 3.00, OutputPer1M: 15.00, CacheReadPer1M: 0.30},
		"anthropic/claude-opus-4-5":      {InputPer1M: 15.00, OutputPer1M: 75.00, CacheReadPer1M: 1.50},
		"anthropic/claude-haiku-4-7":     {InputPer1M: 0.80, OutputPer1M: 4.00, CacheReadPer1M: 0.08},
	} {
		list.prices[strings.ToLower(strings.TrimSpace(providerModel))] = price
	}

	return list
}

// key bygger nyckeln för en provider och modell.
//
// Nyckeln skrivs i gemener, så att modellnamnet hittas oavsett hur det
// skrivs i konfigurationen.
func key(provider, model string) string {
	return strings.ToLower(strings.TrimSpace(provider)) + "/" + strings.ToLower(strings.TrimSpace(model))
}

// Lookup returnerar priset för en provider och modell.
//
// Den andra returvärdet är false om modellen saknas i listan.
func (p PriceList) Lookup(provider, model string) (Price, bool) {
	price, ok := p.prices[key(provider, model)]
	return price, ok
}

// Cost prissätter en användning.
//
// Om modellen saknas i prislistan används schablonen, och källan blir
// "fallback" så att det går att se att kostnaden inte är exakt.
func (p PriceList) Cost(provider, model string, tokens Tokens) Result {
	result := Result{
		PriceListVersion: p.Version,
		Currency:         p.Currency,
	}

	price, known := p.Lookup(provider, model)
	if known {
		result.Source = SourcePriceList
	} else {
		price = p.Fallback
		result.Source = SourceFallback
	}
	result.Known = known

	result.CostUSD = float64(tokens.Input)/1_000_000*price.InputPer1M +
		float64(tokens.Output)/1_000_000*price.OutputPer1M +
		float64(tokens.CacheRead)/1_000_000*price.CacheReadPer1M

	return result
}

// String beskriver en prislista - används i loggar.
func (p PriceList) String() string {
	return fmt.Sprintf("prislista %s (%s, %d modeller)", p.Version, p.Currency, len(p.prices))
}
