//! Tal som Go:s `encoding/json` skriver dem.
//!
//! Go skriver ett heltal flyttal som `1`, medan serde_json skriver `1.0`.
//! Paritetsskriptet jämför agentens svar med Go:s, så talen måste skrivas
//! på samma sätt.

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

#[cfg(test)]
mod tests {
    use super::*;
    use serde::Serialize;

    #[derive(Serialize)]
    struct Svar {
        #[serde(serialize_with = "serialize_f64")]
        percent: f64,
    }

    #[test]
    fn heltal_skrivs_som_heltal() {
        assert_eq!(
            serde_json::to_string(&Svar { percent: 50.0 }).unwrap(),
            "{\"percent\":50}"
        );
    }

    #[test]
    fn decimaler_behålls() {
        assert_eq!(
            serde_json::to_string(&Svar { percent: 33.3 }).unwrap(),
            "{\"percent\":33.3}"
        );
    }
}
