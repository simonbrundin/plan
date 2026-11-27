//! Tidsstämplar som Go skriver dem.
//!
//! Go använder `time.Time`, som vid JSON-serialisering skriver
//! RFC3339Nano: nanosekunderna skrivs utan avslutande nollor, och
//! hela andelen utelämnas om den är noll. Chrono skriver däremot alltid
//! sex decimaler och en "Z" - "2026-11-26T22:30:41.093560Z" mot Go:s
//! "2026-11-26T22:30:41.09356Z".
//!
//! Skillnaden syns direkt i gränssnittet, där tiden visas, och den gör
//! att paritetsjämförelsen mellan Go och Rust faller isär. Därför
//! serialiseras tidsstämplar här på Go:s sätt.

use chrono::{DateTime, SecondsFormat, Utc};
use serde::Serializer;

/// Skriver en tid som Go:s `time.Time` gör.
pub fn serialize<S>(tid: &DateTime<Utc>, serializer: S) -> Result<S::Ok, S::Error>
where
    S: Serializer,
{
    serializer.serialize_str(&formatera(tid))
}

/// Skriver en tid som Go gör, eller `null` om tiden inte finns.
pub fn serialize_option<S>(tid: &Option<DateTime<Utc>>, serializer: S) -> Result<S::Ok, S::Error>
where
    S: Serializer,
{
    match tid {
        Some(tid) => serializer.serialize_str(&formatera(tid)),
        None => serializer.serialize_none(),
    }
}

/// Formaterar en tid som Go:s `time.Time` gör vid JSON-serialisering.
///
/// Go använder UTC om tiden är i UTC. Rust lagrar alltid UTC, så "Z"
/// är rätt - och det är samma som Go ger när servern kör med `TZ=UTC`.
pub fn formatera(tid: &DateTime<Utc>) -> String {
    let nanos = tid.timestamp_subsec_nanos();
    if nanos == 0 {
        // Hela sekunder skrivs utan andel - som Go gör.
        return tid.to_rfc3339_opts(SecondsFormat::Secs, true);
    }

    // Nanosekunderna skrivs som nio siffror, utan avslutande nollor.
    let andel = format!("{nanos:09}");
    let andel = andel.trim_end_matches('0');

    format!("{}.{}Z", tid.format("%Y-%m-%dT%H:%M:%S"), andel)
}

#[cfg(test)]
mod tests {
    use super::*;
    use chrono::TimeZone;

    #[test]
    fn nanosekunder_skrivs_utan_avslutande_nollor() {
        let tid = Utc.with_ymd_and_hms(2026, 11, 26, 22, 30, 41).unwrap()
            + chrono::Duration::nanoseconds(93_560_000);
        assert_eq!(formatera(&tid), "2026-11-26T22:30:41.09356Z");
    }

    #[test]
    fn hela_sekunder_skrivs_utan_andel() {
        let tid = Utc.with_ymd_and_hms(2026, 11, 26, 22, 30, 41).unwrap();
        assert_eq!(formatera(&tid), "2026-11-26T22:30:41Z");
    }

    #[test]
    fn en_nanosekund_skrivs_som_en_siffra() {
        let tid = Utc.with_ymd_and_hms(2026, 11, 26, 22, 30, 41).unwrap()
            + chrono::Duration::nanoseconds(1);
        assert_eq!(formatera(&tid), "2026-11-26T22:30:41.000000001Z");
    }
}
