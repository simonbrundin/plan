//! Tal som Go:s `encoding/json` skriver dem.
//!
//! Go skriver ett heltal flyttal som `1`, medan serde_json skriver `1.0`.
//! Gränssnittet och paritetsskriptet jämför svaren fält för fält, så
//! talen måste skrivas på samma sätt.
//!
//! Go använder `strconv.AppendFloat` med `'f'` och kortast möjliga
//! precision, vilket ger `1`, `0.5` och `1e-7`. serde_json använder ryu
//! som ger samma siffror - skillnaden syns bara i exponentform, där Go
//! skriver `1e+21` och ryu `1e21`. Kostnader når aldrig den storleken.

use serde::Serializer;

/// Skriver ett flyttal på samma sätt som Go.
pub fn serialize_f64<S>(värde: &f64, s: S) -> Result<S::Ok, S::Error>
where
    S: Serializer,
{
    if !värde.is_finite() {
        // Go avvisar NaN och oändlighet - de går inte att skriva som JSON.
        return s.serialize_none();
    }

    if värde.fract() == 0.0 && värde.abs() < 9.007_199_254_740_992e15 {
        // Heltal - Go skriver "1", inte "1.0".
        return s.serialize_i64(*värde as i64);
    }

    s.serialize_f64(*värde)
}

/// Samma sak för valfria flyttal.
pub fn serialize_opt_f64<S>(värde: &Option<f64>, s: S) -> Result<S::Ok, S::Error>
where
    S: Serializer,
{
    match värde {
        Some(v) => serialize_f64(v, s),
        None => s.serialize_none(),
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde::Serialize;

    #[test]
    fn heltal_skrivs_som_heltal() {
        assert_eq!(serde_json::to_string(&1.0_f64).unwrap(), "1.0");
        assert_eq!(
            serde_json::to_string(&Struct { värde: 1.0 }).unwrap(),
            "{\"värde\":1}"
        );
    }

    #[test]
    fn decimaler_behålls() {
        assert_eq!(
            serde_json::to_string(&Struct { värde: 0.5 }).unwrap(),
            "{\"värde\":0.5}"
        );
    }

    #[test]
    fn noll_blir_noll() {
        assert_eq!(
            serde_json::to_string(&Struct { värde: 0.0 }).unwrap(),
            "{\"värde\":0}"
        );
    }

    #[test]
    fn valfritt_fält() {
        #[derive(Serialize)]
        struct Valfritt {
            #[serde(serialize_with = "serialize_opt_f64")]
            värde: Option<f64>,
        }

        assert_eq!(
            serde_json::to_string(&Valfritt { värde: Some(2.0) }).unwrap(),
            "{\"värde\":2}"
        );
        assert_eq!(
            serde_json::to_string(&Valfritt { värde: None }).unwrap(),
            "{\"värde\":null}"
        );
    }

    #[derive(Serialize)]
    struct Struct {
        #[serde(serialize_with = "serialize_f64")]
        värde: f64,
    }
}
