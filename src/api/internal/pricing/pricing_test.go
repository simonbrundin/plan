package pricing

import "testing"

// TestCostMedKändModell kontrollerar att kostnaden räknas fram ur listan.
func TestCostMedKändModell(t *testing.T) {
	result := Current().Cost("minimax", "MiniMax-M2.7", Tokens{Input: 1_000_000, Output: 1_000_000})

	if !result.Known {
		t.Error("modellen ska vara känd")
	}
	if result.Source != SourcePriceList {
		t.Errorf("källa = %q, vill ha %q", result.Source, SourcePriceList)
	}
	// Tolerans för flyttalsavrundning.
	if diff := result.CostUSD - 0.60; diff > 0.000001 || diff < -0.000001 {
		t.Errorf("kostnad = %v, vill ha 0.60", result.CostUSD)
	}
	if result.PriceListVersion != CurrentVersion {
		t.Errorf("version = %q, vill ha %q", result.PriceListVersion, CurrentVersion)
	}
}

// TestCostMedOkändModell kontrollerar att en okänd modell inte kostar
// noll - det vore en falsk siffra.
func TestCostMedOkändModell(t *testing.T) {
	result := Current().Cost("okand", "modell-som-inte-finns", Tokens{Input: 1_000_000})

	if result.Known {
		t.Error("modellen ska inte vara känd")
	}
	if result.Source != SourceFallback {
		t.Errorf("källa = %q, vill ha %q", result.Source, SourceFallback)
	}
	if result.CostUSD <= 0 {
		t.Errorf("kostnad = %v - en okänd modell får inte kosta noll", result.CostUSD)
	}
}

// TestCacheTokensPrissattsSeparat kontrollerar att cachade tokens räknas
// med sitt eget pris.
func TestCacheTokensPrissattsSeparat(t *testing.T) {
	list := Current()

	withCache := list.Cost("minimax", "MiniMax-M2.7", Tokens{Input: 1_000_000, CacheRead: 1_000_000})
	withoutCache := list.Cost("minimax", "MiniMax-M2.7", Tokens{Input: 1_000_000})

	if withCache.CostUSD <= withoutCache.CostUSD {
		t.Errorf("cachade tokens ska kosta mer: %v <= %v", withCache.CostUSD, withoutCache.CostUSD)
	}
}

// TestLookupÄrOkänsligFörSkift kontrollerar att modellnamnet inte behöver
// matcha exakt.
func TestLookupÄrOkänsligFörSkift(t *testing.T) {
	if _, ok := Current().Lookup("MiniMax", "minimax-m2.7"); !ok {
		t.Error("modellen ska hittas oavsett skiftning")
	}
}
