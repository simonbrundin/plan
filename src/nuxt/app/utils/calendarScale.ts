/**
 * Tidsaxeln i den vertikala kalendern.
 *
 * Axeln har fyra skalor: timmar, dagar, månader och år. Skalan bestäms av
 * hur mycket tid som syns, så att antalet markeringar som ska rymmas i
 * vyhöjden förblir hanterbart. En prick per dag i ett årfönster skulle kräva
 * 365 prickar, vilket inte går att läsa eller klicka på.
 */

export type ScaleLevel = "hour" | "day" | "month" | "year";

/**
 * Den skala som gäller för ett tidsfönster.
 *
 * Gränserna är satta efter presentation, inte efter kalender.
 */
export function resolveScaleLevel(duration: number): ScaleLevel {
  if (duration < 48 * 60 * 60 * 1000) return "hour";
  if (duration < 45 * 24 * 60 * 60 * 1000) return "day";
  if (duration < 2 * 365 * 24 * 60 * 60 * 1000) return "month";
  return "year";
}
