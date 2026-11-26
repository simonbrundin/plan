//! Återkommande serier (RRULE).
//!
//! Port av `src/api/internal/recurrence/recurrence.go`. Go använder
//! `rrule-go` för att räkna fram förekomster; här finns en egen motor för
//! den delmängd av RFC 5545 som appen använder:
//!
//! * `FREQ=DAILY|WEEKLY|MONTHLY|YEARLY`
//! * `INTERVAL`, `COUNT` och `UNTIL`
//! * `BYDAY` för veckor (måndag till söndag)
//!
//! Testvektorerna är desamma som i Go - `recurrence_test.go` - så en
//! förändring i någon av portarna syns direkt.

use chrono::{DateTime, Datelike, Duration, NaiveDate, NaiveDateTime, TimeZone, Utc, Weekday};

/// Hur nästa förekomst räknas fram för ett återkommande mål.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Läge {
    /// Håller serien fast relativt målets startdatum.
    ///
    /// Nästa förekomst är seriens nästa datum som inte passerat - förfallna
    /// tillfällen hoppas över.
    Schemalagd,
    /// Skjuter fram serien.
    ///
    /// Nästa förekomst räknas från det datum förekomsten faktiskt slutfördes.
    Push,
}

impl Läge {
    /// Tolkar ett lägesnamn.
    pub fn tolka(namn: &str) -> std::result::Result<Self, Fel> {
        match namn {
            "schedule" => Ok(Self::Schemalagd),
            "push" => Ok(Self::Push),
            annat => Err(Fel::OgiltigLäge(annat.to_string())),
        }
    }

    /// Lägesnamnet, som det sparas i databasen.
    pub fn namn(self) -> &'static str {
        match self {
            Self::Schemalagd => "schedule",
            Self::Push => "push",
        }
    }
}

/// Fel från återkommande hantering.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum Fel {
    /// Upprepningsregeln är inte giltig.
    OgiltigRegel(String),
    /// Regeln ger ingen förekomst.
    IngenFörekomst(String),
    /// Schemaläggningsstrategin är okänd.
    OgiltigLäge(String),
}

impl std::fmt::Display for Fel {
    fn fmt(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Fel::OgiltigRegel(detalj) => write!(formatter, "ogiltig upprepningsregel: {detalj}"),
            Fel::IngenFörekomst(detalj) => {
                write!(formatter, "regeln ger inga förekomster: {detalj}")
            }
            Fel::OgiltigLäge(läge) => {
                write!(formatter, "ogiltig schemaläggningsstrategi: \"{läge}\"")
            }
        }
    }
}

impl std::error::Error for Fel {}

/// Resultatet av en upprepningsberäkning.
pub type Result<T> = std::result::Result<T, Fel>;

/// Hur ofta serien återkommer.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Frekvens {
    /// Varje dag.
    Daglig,
    /// Varje vecka.
    Vecklig,
    /// Varje månad.
    Månadslig,
    /// Varje år.
    Årlig,
}

/// En tolkad upprepningsregel.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Regel {
    /// Hur ofta serien återkommer.
    pub frekvens: Frekvens,
    /// Antalet perioder mellan två förekomster.
    pub intervall: i64,
    /// Maximalt antal förekomster, om regeln anger det.
    pub count: Option<i64>,
    /// Sista tidpunkt som får ha en förekomst, om regeln anger den.
    pub until: Option<DateTime<Utc>>,
    /// Veckodagarna som förekommer, för veckliga serier.
    pub by_day: Vec<Weekday>,
}

impl Regel {
    /// Tolkar en RRULE-sträng.
    ///
    /// Prefixet "RRULE:" är valfritt, och regeln är okänslig för
    /// gemener och mellanrum.
    pub fn tolka(regel: &str) -> Result<Self> {
        let normaliserad = regel.trim().to_uppercase();
        if normaliserad.is_empty() {
            return Err(Fel::OgiltigRegel("regeln är tom".to_string()));
        }

        let normaliserad = normaliserad
            .strip_prefix("RRULE:")
            .unwrap_or(&normaliserad)
            .trim();

        let mut frekvens = None;
        let mut intervall: i64 = 1;
        let mut count = None;
        let mut until = None;
        let mut by_day: Vec<Weekday> = Vec::new();

        for del in normaliserad.split(';') {
            let Some((nyckel, värde)) = del.split_once('=') else {
                continue;
            };

            let nyckel = nyckel.trim();
            let värde = värde.trim();

            match nyckel {
                "FREQ" => {
                    frekvens = Some(match värde {
                        "DAILY" => Frekvens::Daglig,
                        "WEEKLY" => Frekvens::Vecklig,
                        "MONTHLY" => Frekvens::Månadslig,
                        "YEARLY" => Frekvens::Årlig,
                        annat => {
                            // Samma ordval som rrule-go, som Go-tjänsten
                            // använder.
                            return Err(Fel::OgiltigRegel(format!("undefined frequency: {annat}")));
                        }
                    });
                }
                "INTERVAL" => {
                    intervall = värde.parse::<i64>().map_err(|_| {
                        Fel::OgiltigRegel(format!("INTERVAL har ogiltigt värde {värde:?}"))
                    })?;
                    if intervall < 1 {
                        return Err(Fel::OgiltigRegel("INTERVAL måste vara minst 1".to_string()));
                    }
                }
                "COUNT" => {
                    let antal = värde.parse::<i64>().map_err(|_| {
                        Fel::OgiltigRegel(format!("COUNT har ogiltigt värde {värde:?}"))
                    })?;
                    if antal < 1 {
                        return Err(Fel::OgiltigRegel("COUNT måste vara minst 1".to_string()));
                    }
                    count = Some(antal);
                }
                "UNTIL" => {
                    until = Some(tolka_until(värde)?);
                }
                "BYDAY" => {
                    for dag in värde.split(',') {
                        let dag = dag.trim();
                        if dag.is_empty() {
                            continue;
                        }
                        by_day.push(tolka_vardag(dag)?);
                    }
                }
                _ => {}
            }
        }

        let Some(frekvens) = frekvens else {
            // Samma ordval som rrule-go.
            return Err(Fel::OgiltigRegel(
                "RRULE property FREQ is required".to_string(),
            ));
        };

        Ok(Regel {
            frekvens,
            intervall,
            count,
            until,
            by_day,
        })
    }

    /// Kontrollerar att regeln är giltig.
    pub fn validera(regel: &str) -> Result<()> {
        Regel::tolka(regel).map(|_| ())
    }

    /// Kontrollerar att en schemaläggningsstrategi är känd.
    pub fn validera_läge(läge: Läge) -> Result<()> {
        match läge {
            Läge::Schemalagd | Läge::Push => Ok(()),
        }
    }

    /// Den första förekomsten på eller efter ankaret.
    ///
    /// Om regeln inte ger någon förekomst returneras
    /// [`Fel::IngenFörekomst`].
    pub fn första_förekomst(&self, ankare: DateTime<Utc>) -> Result<DateTime<Utc>> {
        let start = början_på_dagen(ankare);

        // Förekomsten på ankaret räknas med - som i Go.
        self.efter(&start, start, true)
            .ok_or_else(|| Fel::IngenFörekomst("serien börjar efter ankaret".to_string()))
    }

    /// Nästa förekomst efter att en förekomst slutförts.
    ///
    /// `slutförd` är datumet för förekomsten som slutfördes och
    /// `slutförd_på` är det datum slutförandet skedde. `ankare` är målets
    /// startdatum och används i [`Läge::Schemalagd`], där serien ligger
    /// fast.
    ///
    /// I [`Läge::Schemalagd`] hoppas förfallna förekomster över - nästa
    /// förekomst är det första seriedatumet som inte passerat. I
    /// [`Läge::Push`] räknas serien från `slutförd_på`, så att nästa
    /// förekomst följer slutförandetiden.
    pub fn nästa_förekomst(
        &self,
        läge: Läge,
        ankare: DateTime<Utc>,
        slutförd: DateTime<Utc>,
        slutförd_på: DateTime<Utc>,
    ) -> Result<DateTime<Utc>> {
        Regel::validera_läge(läge)?;

        let slutförd_dag = början_på_dagen(slutförd);
        let slutförd_på_dag = början_på_dagen(slutförd_på);

        let seriens_början;
        let undre_gräns = match läge {
            Läge::Push => {
                seriens_början = max(slutförd_dag, slutförd_på_dag);
                seriens_början + Duration::days(1)
            }
            Läge::Schemalagd => {
                seriens_början = början_på_dagen(ankare);
                max(slutförd_dag + Duration::days(1), slutförd_på_dag)
            }
        };

        self.efter(&seriens_början, undre_gräns, true)
            .ok_or_else(|| Fel::IngenFörekomst("serien är slut".to_string()))
    }

    /// Alla förekomster i intervallet.
    ///
    /// Serien börjar vid `start` och förekomst före `start` räknas inte.
    /// `inklusive` anger om en förekomst exakt på `undre` ska räknas med.
    pub fn mellan(
        &self,
        start: DateTime<Utc>,
        undre: DateTime<Utc>,
        över: DateTime<Utc>,
        inklusive: bool,
    ) -> Vec<DateTime<Utc>> {
        let mut förekomster = Vec::new();
        let mut räknare: i64 = 0;
        let mut period = 0i64;
        let mut slutar = false;

        // Skydd mot en regel som aldrig slutar - en serie med INTERVAL=1
        // och tusentals år ska inte hänga servern.
        let max_perioder = 100_000;

        while !slutar && period < max_perioder {
            for kandidat in self.förekomster_i_period(start, period) {
                if kandidat < start {
                    continue;
                }

                räknare += 1;
                if let Some(count) = self.count {
                    if räknare > count {
                        slutar = true;
                        break;
                    }
                }

                if let Some(until) = self.until {
                    if kandidat > until {
                        slutar = true;
                        break;
                    }
                }

                if kandidat < undre || (!inklusive && kandidat == undre) {
                    continue;
                }

                if kandidat > över {
                    slutar = true;
                    break;
                }

                förekomster.push(kandidat);
            }

            period += 1;
        }

        förekomster
    }

    /// Den första förekomsten på eller efter `undre`.
    ///
    /// Serien börjar vid `start`. `inklusive` anger om en förekomst exakt
    /// på `undre` ska räknas med.
    fn efter(
        &self,
        start: &DateTime<Utc>,
        undre: DateTime<Utc>,
        inklusive: bool,
    ) -> Option<DateTime<Utc>> {
        let mut räknare: i64 = 0;
        let mut period = 0i64;

        while period < 100_000 {
            for kandidat in self.förekomster_i_period(*start, period) {
                if kandidat < *start {
                    continue;
                }

                räknare += 1;
                if let Some(count) = self.count {
                    if räknare > count {
                        return None;
                    }
                }

                if let Some(until) = self.until {
                    if kandidat > until {
                        return None;
                    }
                }

                if kandidat < undre {
                    continue;
                }

                if !inklusive && kandidat == undre {
                    continue;
                }

                return Some(kandidat);
            }

            period += 1;
        }

        None
    }

    /// Förekomsterna i en given period.
    ///
    /// Period noll är perioden som `start` ligger i.
    fn förekomster_i_period(&self, start: DateTime<Utc>, period: i64) -> Vec<DateTime<Utc>> {
        match self.frekvens {
            Frekvens::Daglig => {
                let dag = start.date_naive() + chrono::Duration::days(period * self.intervall);
                vec![datum_tid(start, dag)]
            }
            Frekvens::Vecklig => {
                let veckans_start = veckans_start(start.date_naive());
                let dagar = if self.by_day.is_empty() {
                    vec![start.weekday()]
                } else {
                    self.by_day.clone()
                };

                let mut veckor = dagar.clone();
                veckor.sort_by_key(|dag| dag.num_days_from_monday());

                veckor
                    .into_iter()
                    .map(|dag| {
                        let veckstart =
                            veckans_start + chrono::Duration::weeks(period * self.intervall);
                        let dagnummer = veckstart.weekday().num_days_from_monday() as i64;
                        let måndag = veckstart - chrono::Duration::days(dagnummer);
                        datum_tid(
                            start,
                            måndag + chrono::Duration::days(dag.num_days_from_monday() as i64),
                        )
                    })
                    .collect()
            }
            Frekvens::Månadslig => {
                let månader = (period * self.intervall).max(0) as u32;
                let månad = start
                    .date_naive()
                    .checked_add_months(chrono::Months::new(månader));

                let Some(månad) = månad else {
                    return Vec::new();
                };

                let Some(dag) = månad.with_day(start.day()) else {
                    // En månad som saknar dagen hoppas över - RFC 5545
                    // säger att ogiltiga datum utelämnas, inte flyttas.
                    return Vec::new();
                };

                vec![datum_tid(start, dag)]
            }
            Frekvens::Årlig => {
                let år = (start.year() as i64 + period * self.intervall) as i32;
                let Some(datum) = NaiveDate::from_ymd_opt(år, start.month(), start.day()) else {
                    return Vec::new();
                };

                vec![datum_tid(start, datum)]
            }
        }
    }
}

/// Tolkar ett UNTIL-värde.
///
/// Formaten är `20260101T000000Z` (eller utan tidsdel), som RRULE använder.
fn tolka_until(värde: &str) -> Result<DateTime<Utc>> {
    let format = if värde.contains('T') {
        "%Y%m%dT%H%M%SZ"
    } else {
        "%Y%m%d"
    };

    NaiveDateTime::parse_from_str(värde, format)
        .map(|tid| tid.and_utc())
        .or_else(|_| {
            NaiveDate::parse_from_str(värde, "%Y%m%d")
                .map(|datum| datum.and_hms_opt(0, 0, 0).unwrap().and_utc())
        })
        .map_err(|_| Fel::OgiltigRegel(format!("parsing time {värde:?} as \"20060102T150405Z\": cannot parse {värde:?} as \"2006\"")))
}

/// Tolkar en veckodag i RRULE-format ("MO", "TU" ...).
fn tolka_vardag(dag: &str) -> Result<Weekday> {
    match dag {
        "MO" => Ok(Weekday::Mon),
        "TU" => Ok(Weekday::Tue),
        "WE" => Ok(Weekday::Wed),
        "TH" => Ok(Weekday::Thu),
        "FR" => Ok(Weekday::Fri),
        "SA" => Ok(Weekday::Sat),
        "SU" => Ok(Weekday::Sun),
        // Samma ordval som rrule-go.
        annat => Err(Fel::OgiltigRegel(format!("undefined weekday: {annat}"))),
    }
}

/// Gör om ett klockslag till det datum det tillhör.
fn början_på_dagen(tid: DateTime<Utc>) -> DateTime<Utc> {
    let datum = tid.date_naive();
    Utc.from_utc_datetime(&datum.and_hms_opt(0, 0, 0).unwrap())
}

/// Bygger ett klockslag utifrån ett datum och en tid.
fn datum_tid(mall: DateTime<Utc>, datum: NaiveDate) -> DateTime<Utc> {
    let tid = mall.time();
    Utc.from_utc_datetime(&datum.and_time(tid))
}

/// Måndagen i datumets vecka.
fn veckans_start(datum: NaiveDate) -> NaiveDate {
    let dagnummer = datum.weekday().num_days_from_monday() as i64;
    datum - chrono::Duration::days(dagnummer)
}

/// Det större av två tidpunkter.
fn max(a: DateTime<Utc>, b: DateTime<Utc>) -> DateTime<Utc> {
    if a > b {
        a
    } else {
        b
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn datum(år: i32, månad: u32, dag: u32) -> DateTime<Utc> {
        Utc.with_ymd_and_hms(år, månad, dag, 0, 0, 0).unwrap()
    }

    #[test]
    fn giltiga_regler_parsas() {
        assert!(Regel::validera("FREQ=DAILY").is_ok());
        assert!(Regel::validera("FREQ=DAILY;INTERVAL=2").is_ok());
        assert!(Regel::validera("FREQ=WEEKLY;BYDAY=MO").is_ok());
        assert!(Regel::validera("RRULE:FREQ=DAILY").is_ok());
    }

    #[test]
    fn ogiltiga_regler_avvisas() {
        assert!(Regel::validera("").is_err());
        assert!(Regel::validera("   ").is_err());
        assert!(Regel::validera("FREQ=DAILY;INTERVAL=0").is_err());
        assert!(Regel::validera("FREQ=DAILY;INTERVAL=-2").is_err());
        assert!(Regel::validera("FREQ=DAILY;COUNT=0").is_err());
        assert!(Regel::validera("FREQ=BOGUS").is_err());
        assert!(Regel::validera("FREQ=DAILY;INTERVAL=abc").is_err());
    }

    #[test]
    fn forsta_forekomsten_stammer_pa_ankaret() {
        let regel = Regel::tolka("FREQ=DAILY").unwrap();
        assert_eq!(
            regel.första_förekomst(datum(2026, 1, 5)).unwrap(),
            datum(2026, 1, 5)
        );

        let regel = Regel::tolka("FREQ=DAILY;INTERVAL=2").unwrap();
        assert_eq!(
            regel.första_förekomst(datum(2026, 1, 5)).unwrap(),
            datum(2026, 1, 5)
        );

        let regel = Regel::tolka("FREQ=MONTHLY").unwrap();
        assert_eq!(
            regel.första_förekomst(datum(2026, 1, 5)).unwrap(),
            datum(2026, 1, 5)
        );
    }

    #[test]
    fn veckodag_hoppar_till_nasta_matchning() {
        let regel = Regel::tolka("FREQ=WEEKLY;BYDAY=MO").unwrap();
        assert_eq!(
            regel.första_förekomst(datum(2026, 1, 7)).unwrap(),
            datum(2026, 1, 12)
        );
    }

    #[test]
    fn en_regel_som_slutar_har_ingen_forekomst() {
        let regel = Regel::tolka("FREQ=DAILY;UNTIL=20200101T000000Z").unwrap();
        assert!(matches!(
            regel.första_förekomst(datum(2026, 1, 5)),
            Err(Fel::IngenFörekomst(_))
        ));
    }

    #[test]
    fn nasta_forekomst_i_schemalagt_lage() {
        let ankare = datum(2026, 1, 1);
        let regel = Regel::tolka("FREQ=DAILY;INTERVAL=2").unwrap();

        assert_eq!(
            regel
                .nästa_förekomst(
                    Läge::Schemalagd,
                    ankare,
                    datum(2026, 1, 5),
                    datum(2026, 1, 5)
                )
                .unwrap(),
            datum(2026, 1, 7)
        );

        // Sent slutförande - förfallna förekomster hoppas över.
        assert_eq!(
            regel
                .nästa_förekomst(
                    Läge::Schemalagd,
                    ankare,
                    datum(2026, 1, 5),
                    datum(2026, 1, 10)
                )
                .unwrap(),
            datum(2026, 1, 11)
        );

        // Tidigt slutförande - serien ligger fast.
        assert_eq!(
            regel
                .nästa_förekomst(
                    Läge::Schemalagd,
                    ankare,
                    datum(2026, 1, 7),
                    datum(2026, 1, 5)
                )
                .unwrap(),
            datum(2026, 1, 9)
        );
    }

    #[test]
    fn veckodag_halls_i_schemalagt_lage() {
        let ankare = datum(2026, 1, 1);
        let regel = Regel::tolka("FREQ=WEEKLY;BYDAY=MO").unwrap();

        assert_eq!(
            regel
                .nästa_förekomst(
                    Läge::Schemalagd,
                    ankare,
                    datum(2026, 1, 5),
                    datum(2026, 1, 5)
                )
                .unwrap(),
            datum(2026, 1, 12)
        );
    }

    #[test]
    fn nasta_forekomst_i_push_lage() {
        let ankare = datum(2026, 1, 1);
        let regel = Regel::tolka("FREQ=DAILY;INTERVAL=2").unwrap();

        assert_eq!(
            regel
                .nästa_förekomst(Läge::Push, ankare, datum(2026, 1, 5), datum(2026, 1, 5))
                .unwrap(),
            datum(2026, 1, 7)
        );

        // Sent slutförande - serien glider.
        assert_eq!(
            regel
                .nästa_förekomst(Läge::Push, ankare, datum(2026, 1, 5), datum(2026, 1, 10))
                .unwrap(),
            datum(2026, 1, 12)
        );

        // Tidigt slutförande - serien glider bakåt.
        assert_eq!(
            regel
                .nästa_förekomst(Läge::Push, ankare, datum(2026, 1, 7), datum(2026, 1, 5))
                .unwrap(),
            datum(2026, 1, 9)
        );
    }

    #[test]
    fn veckodag_foljer_veckodagen_i_push_lage() {
        let ankare = datum(2026, 1, 1);
        let regel = Regel::tolka("FREQ=WEEKLY;BYDAY=MO").unwrap();

        assert_eq!(
            regel
                .nästa_förekomst(Läge::Push, ankare, datum(2026, 1, 5), datum(2026, 1, 6))
                .unwrap(),
            datum(2026, 1, 12)
        );
    }

    #[test]
    fn lagen_tolkas() {
        assert_eq!(Läge::tolka("schedule").unwrap(), Läge::Schemalagd);
        assert_eq!(Läge::tolka("push").unwrap(), Läge::Push);
        assert!(Läge::tolka("nonsense").is_err());
    }

    #[test]
    fn en_slutserie_har_ingen_nasta_forekomst() {
        let regel = Regel::tolka("FREQ=DAILY;COUNT=2").unwrap();
        assert!(matches!(
            regel.nästa_förekomst(
                Läge::Schemalagd,
                datum(2026, 1, 1),
                datum(2026, 1, 2),
                datum(2026, 1, 2)
            ),
            Err(Fel::IngenFörekomst(_))
        ));
    }

    #[test]
    fn forekomster_raknas_i_ett_intervall() {
        let regel = Regel::tolka("FREQ=DAILY").unwrap();
        let förekomster = regel.mellan(
            datum(2026, 1, 1),
            datum(2026, 1, 1),
            datum(2026, 1, 4),
            true,
        );

        assert_eq!(
            förekomster,
            vec![
                datum(2026, 1, 1),
                datum(2026, 1, 2),
                datum(2026, 1, 3),
                datum(2026, 1, 4)
            ]
        );
    }

    #[test]
    fn veckor_foljer_veckans_dagar() {
        let regel = Regel::tolka("FREQ=WEEKLY;BYDAY=MO,WE").unwrap();
        let förekomster = regel.mellan(
            datum(2026, 1, 5),
            datum(2026, 1, 5),
            datum(2026, 1, 13),
            true,
        );

        assert_eq!(
            förekomster,
            vec![datum(2026, 1, 5), datum(2026, 1, 7), datum(2026, 1, 12),]
        );
    }

    #[test]
    fn intervallet_haller_serien_pa_plats() {
        let regel = Regel::tolka("FREQ=DAILY;INTERVAL=3").unwrap();
        let förekomster = regel.mellan(
            datum(2026, 1, 1),
            datum(2026, 1, 1),
            datum(2026, 1, 10),
            true,
        );

        assert_eq!(
            förekomster,
            vec![
                datum(2026, 1, 1),
                datum(2026, 1, 4),
                datum(2026, 1, 7),
                datum(2026, 1, 10),
            ]
        );
    }

    #[test]
    fn count_begransar_antalet_forekomster() {
        let regel = Regel::tolka("FREQ=DAILY;COUNT=3").unwrap();
        let förekomster = regel.mellan(
            datum(2026, 1, 1),
            datum(2026, 1, 1),
            datum(2026, 1, 10),
            true,
        );

        assert_eq!(
            förekomster,
            vec![datum(2026, 1, 1), datum(2026, 1, 2), datum(2026, 1, 3)]
        );
    }
}
