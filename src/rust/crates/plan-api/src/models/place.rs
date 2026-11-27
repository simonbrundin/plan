//! Platser.
//!
//! Port av `src/api/internal/models/place.go`. Platsen är ett område - en
//! cirkel med en mittpunkt och en radie - inte en punkt.

use serde::{Deserialize, Serialize};
use utoipa::ToSchema;

/// Radien en ny plats får om inget annat anges.
pub const DEFAULT_PLACE_RADIUS_METERS: i32 = 100;
/// Största tillåtna radie - 100 km.
pub const MAX_PLACE_RADIUS_METERS: i32 = 100_000;

/// En fysisk plats där uppgifter kan slutföras.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, ToSchema)]
pub struct Place {
    /// Platsens id.
    pub id: i64,
    /// Platsens namn.
    pub name: String,
    /// Cirkelns mittpunkt i latitud.
    #[serde(serialize_with = "crate::models::go_num::serialize_f64")]
    pub latitude: f64,
    /// Cirkelns mittpunkt i longitud.
    #[serde(serialize_with = "crate::models::go_num::serialize_f64")]
    pub longitude: f64,
    /// Cirkelns radie i meter.
    pub radius_meters: i32,
    /// Ikonen som visas - tom sträng om ingen ikon satts.
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub icon: String,
    /// När platsen skapades.
    #[serde(serialize_with = "crate::models::go_time::serialize")]
    pub created_at: chrono::DateTime<chrono::Utc>,
    /// När platsen senast ändrades.
    #[serde(serialize_with = "crate::models::go_time::serialize")]
    pub updated_at: chrono::DateTime<chrono::Utc>,
}

/// Avståndet från en punkt till platsens mittpunkt, i meter.
pub fn haversine_meters(lat1: f64, lon1: f64, lat2: f64, lon2: f64) -> f64 {
    const JORD_RADIUS: f64 = 6_371_000.0;

    let rad = |grad: f64| grad.to_radians();

    let dlat = rad(lat2 - lat1);
    let dlon = rad(lon2 - lon1);

    let a =
        (dlat / 2.0).sin().powi(2) + rad(lat1).cos() * rad(lat2).cos() * (dlon / 2.0).sin().powi(2);

    2.0 * JORD_RADIUS * a.sqrt().asin()
}

impl Place {
    /// Avståndet till en position.
    pub fn distance_meters(&self, latitude: f64, longitude: f64) -> f64 {
        haversine_meters(self.latitude, self.longitude, latitude, longitude)
    }

    /// Ligger positionen inom platsens radie?
    pub fn contains_position(&self, latitude: f64, longitude: f64) -> bool {
        self.distance_meters(latitude, longitude) <= f64::from(self.radius_meters)
    }

    /// Kontrollerar att platsen har giltiga värden.
    ///
    /// Felmeddelandena är desamma som i Go, så gränssnittet visar samma
    /// sak.
    pub fn validate(&self) -> Result<(), String> {
        if self.name.is_empty() {
            return Err("platsen måste ha ett namn".to_string());
        }
        if !(-90.0..=90.0).contains(&self.latitude) {
            return Err(format!(
                "koordinaten ligger utanför världen: latitud {}",
                self.latitude
            ));
        }
        if !(-180.0..=180.0).contains(&self.longitude) {
            return Err(format!(
                "koordinaten ligger utanför världen: longitud {}",
                self.longitude
            ));
        }
        if self.radius_meters <= 0 || self.radius_meters > MAX_PLACE_RADIUS_METERS {
            return Err(format!(
                "radien måste vara mellan 1 och 100000 meter: {}",
                self.radius_meters
            ));
        }
        Ok(())
    }
}

/// Request-body för att skapa en plats.
#[derive(Debug, Clone, Deserialize, ToSchema)]
pub struct CreatePlaceRequest {
    /// Platsens namn - krävs.
    pub name: String,
    /// Mittpunktens latitud.
    #[serde(default)]
    pub latitude: f64,
    /// Mittpunktens longitud.
    #[serde(default)]
    pub longitude: f64,
    /// Radien i meter - standard är 100.
    #[serde(default)]
    pub radius_meters: Option<i32>,
    /// Ikonen som visas.
    #[serde(default)]
    pub icon: Option<String>,
}

/// Request-body för att uppdatera en plats.
///
/// Alla fält är valfria - ett fält som saknas lämnas oförändrat.
#[derive(Debug, Clone, Deserialize, ToSchema)]
pub struct UpdatePlaceRequest {
    /// Nytt namn.
    #[serde(default)]
    pub name: Option<String>,
    /// Ny latitud.
    #[serde(default)]
    pub latitude: Option<f64>,
    /// Ny longitud.
    #[serde(default)]
    pub longitude: Option<f64>,
    /// Ny radie i meter.
    #[serde(default)]
    pub radius_meters: Option<i32>,
    /// Ny ikon.
    #[serde(default)]
    pub icon: Option<String>,
}

/// Request-body som ersätter platserna på ett mål eller en händelse.
///
/// `place_ids` ersätter hela uppsättningen - en tom lista tar bort alla
/// platser.
#[derive(Debug, Clone, Deserialize, ToSchema)]
pub struct SetPlacesRequest {
    /// Id:n på platserna.
    #[serde(default)]
    pub place_ids: Vec<i64>,
}

/// Request-body som sätter användarens "jag är här"-plats.
///
/// `place_id` = null rensar den valda platsen.
#[derive(Debug, Clone, Deserialize, ToSchema)]
pub struct SetCurrentPlaceRequest {
    /// Id:t på platsen, eller inget för att rensa.
    #[serde(default)]
    pub place_id: Option<i64>,
}

#[cfg(test)]
mod tests {
    use super::*;

    fn exempelplats() -> Place {
        Place {
            id: 1,
            name: "Hem".to_string(),
            latitude: 59.3293,
            longitude: 18.0686,
            radius_meters: 100,
            icon: String::new(),
            created_at: chrono::Utc::now(),
            updated_at: chrono::Utc::now(),
        }
    }

    #[test]
    fn avstandet_till_samma_punkt_ar_noll() {
        let plats = exempelplats();
        assert!(plats.distance_meters(59.3293, 18.0686) < 0.001);
    }

    #[test]
    fn positionen_inne_i_radien() {
        let plats = exempelplats();
        assert!(plats.contains_position(59.3293, 18.0686));
        // ~1 km bort - långt utanför radien.
        assert!(!plats.contains_position(59.3393, 18.0686));
    }

    #[test]
    fn ogiltiga_platser_avvisas() {
        let mut plats = exempelplats();
        plats.name = String::new();
        assert_eq!(plats.validate().unwrap_err(), "platsen måste ha ett namn");

        let mut plats = exempelplats();
        plats.name = "Hem".to_string();
        plats.latitude = 91.0;
        assert!(plats.validate().is_err());

        let mut plats = exempelplats();
        plats.latitude = 59.0;
        plats.longitude = 181.0;
        assert!(plats.validate().is_err());

        let mut plats = exempelplats();
        plats.longitude = 18.0;
        plats.radius_meters = 0;
        assert!(plats.validate().is_err());
    }
}
