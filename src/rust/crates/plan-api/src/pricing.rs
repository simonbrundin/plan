//! Prislista för agentkörningar.
//!
//! Port av `src/api/internal/pricing`. Priset sparas med varje körning, så
//! att en kostnad kan reproduceras och revideras senare. Saknas priset för
//! en modell används inte noll - det vore en falsk siffra - utan en
//! dokumenterad schablon, och det markeras i kostnadsraden.

/// Versionen av prislistan som gäller nu.
///
/// Versionsnumret följer kalenderdatumet, så att det går att se vilken
/// prislista som gällde vid en viss tidpunkt.
pub const CURRENT_VERSION: &str = "2026-11-24.1";

/// Valutan alla priser anges i.
pub const CURRENCY: &str = "USD";

/// Priset för en miljon tokens av en sort.
#[derive(Debug, Clone, Copy, PartialEq)]
pub struct Price {
    /// Priset per miljon indata-tokens.
    pub input_per_1m: f64,
    /// Priset per miljon utdata-tokens.
    pub output_per_1m: f64,
    /// Priset per miljon tokens som lästs från cachen.
    pub cache_read_per_1m: f64,
}

impl Price {
    /// Schablonen som används när modellen saknas i listan.
    ///
    /// Den är avsiktligt hög - en okänd modell ska inte kosta mindre än
    /// en känd.
    pub const FALLBACK: Price = Price {
        input_per_1m: 1.00,
        output_per_1m: 3.00,
        cache_read_per_1m: 0.10,
    };
}

/// Den användning som ska prissättas.
#[derive(Debug, Clone, Copy, Default, PartialEq)]
pub struct Tokens {
    /// Indata-tokens.
    pub input: i64,
    /// Utdata-tokens.
    pub output: i64,
    /// Tokens som lästs från modellens cache.
    pub cache_read: i64,
}

/// Källan till ett pris.
pub const SOURCE_PRICE_LIST: &str = "price_list";
/// Källan till ett pris när modellen saknas i listan.
pub const SOURCE_FALLBACK: &str = "fallback";

/// En prissatt användning.
#[derive(Debug, Clone, PartialEq)]
pub struct Cost {
    /// Den beräknade kostnaden.
    pub cost_usd: f64,
    /// Källan - "price_list" eller "fallback".
    pub source: String,
    /// Sant om modellen fanns i prislistan.
    pub known: bool,
    /// Versionen av prislistan som användes.
    pub price_list_version: String,
    /// Valutan kostnaden anges i.
    pub currency: String,
}

/// Prislistan.
#[derive(Debug, Clone)]
pub struct PriceList {
    /// Versionen av listan.
    pub version: String,
    /// Valutan priserna anges i.
    pub currency: String,
    /// Schablonen för okända modeller.
    pub fallback: Price,
    /// Priserna, med nyckeln "provider/model" i gemener.
    prices: Vec<(String, Price)>,
}

impl PriceList {
    /// Prislistan som gäller nu.
    pub fn current() -> PriceList {
        let priser = vec![
            (
                "minimax/MiniMax-M2.7",
                Price {
                    input_per_1m: 0.20,
                    output_per_1m: 0.40,
                    cache_read_per_1m: 0.02,
                },
            ),
            (
                "minimax/MiniMax-M1",
                Price {
                    input_per_1m: 0.10,
                    output_per_1m: 0.20,
                    cache_read_per_1m: 0.01,
                },
            ),
            (
                "openai/gpt-4o-mini",
                Price {
                    input_per_1m: 0.15,
                    output_per_1m: 0.60,
                    cache_read_per_1m: 0.075,
                },
            ),
            (
                "openai/gpt-4o",
                Price {
                    input_per_1m: 2.50,
                    output_per_1m: 10.00,
                    cache_read_per_1m: 1.25,
                },
            ),
            (
                "anthropic/claude-sonnet-4-7-25",
                Price {
                    input_per_1m: 3.00,
                    output_per_1m: 15.00,
                    cache_read_per_1m: 0.30,
                },
            ),
            (
                "anthropic/claude-opus-4-5",
                Price {
                    input_per_1m: 15.00,
                    output_per_1m: 75.00,
                    cache_read_per_1m: 1.50,
                },
            ),
            (
                "anthropic/claude-haiku-4-7",
                Price {
                    input_per_1m: 0.80,
                    output_per_1m: 4.00,
                    cache_read_per_1m: 0.08,
                },
            ),
        ];

        PriceList {
            version: CURRENT_VERSION.to_string(),
            currency: CURRENCY.to_string(),
            fallback: Price::FALLBACK,
            prices: priser
                .into_iter()
                .map(|(nyckel, pris)| (nyckel.to_lowercase(), pris))
                .collect(),
        }
    }

    /// Nyckeln för en provider och modell.
    ///
    /// Nyckeln skrivs i gemener, så att modellnamnet hittas oavsett hur det
    /// skrivs i konfigurationen.
    fn nyckel(provider: &str, model: &str) -> String {
        format!(
            "{}/{}",
            provider.trim().to_lowercase(),
            model.trim().to_lowercase()
        )
    }

    /// Priset för en provider och modell.
    ///
    /// `false` betyder att modellen saknas i listan.
    pub fn lookup(&self, provider: &str, model: &str) -> (Price, bool) {
        let nyckel = PriceList::nyckel(provider, model);

        for (kandidat, pris) in &self.prices {
            if *kandidat == nyckel {
                return (*pris, true);
            }
        }

        (self.fallback, false)
    }

    /// Prissätter en användning.
    ///
    /// Om modellen saknas i prislistan används schablonen, och källan blir
    /// "fallback" så att det går att se att kostnaden inte är exakt.
    pub fn cost(&self, provider: &str, model: &str, tokens: Tokens) -> Cost {
        let (pris, känd) = self.lookup(provider, model);

        let source = if känd {
            SOURCE_PRICE_LIST
        } else {
            SOURCE_FALLBACK
        };

        let cost_usd = tokens.input as f64 / 1_000_000.0 * pris.input_per_1m
            + tokens.output as f64 / 1_000_000.0 * pris.output_per_1m
            + tokens.cache_read as f64 / 1_000_000.0 * pris.cache_read_per_1m;

        Cost {
            cost_usd,
            source: source.to_string(),
            known: känd,
            price_list_version: self.version.clone(),
            currency: self.currency.clone(),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn en_kand_modell_prissatts_till_sitt_pris() {
        let lista = PriceList::current();
        let kostnad = lista.cost(
            "minimax",
            "MiniMax-M2.7",
            Tokens {
                input: 1_000_000,
                output: 1_000_000,
                cache_read: 0,
            },
        );

        assert!(kostnad.known);
        assert_eq!(kostnad.source, SOURCE_PRICE_LIST);
        assert!((kostnad.cost_usd - 0.60).abs() < 0.0001);
    }

    #[test]
    fn en_okand_modell_gar_pa_schablonen() {
        let lista = PriceList::current();
        let kostnad = lista.cost("okand", "modell", Tokens::default());

        assert!(!kostnad.known);
        assert_eq!(kostnad.source, SOURCE_FALLBACK);
        assert_eq!(kostnad.cost_usd, 0.0);
    }

    #[test]
    fn modellnamnet_hittas_oberoende_av_stavning() {
        let lista = PriceList::current();
        let (pris, känd) = lista.lookup("MINIMAX", " minimax-m2.7 ");

        assert!(känd);
        assert!((pris.input_per_1m - 0.20).abs() < 0.0001);
    }
}
