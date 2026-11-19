package models

import (
	"errors"
	"fmt"
	"math"
	"time"
)

// Fel som kan uppstå när en plats valideras.
var (
	// ErrPlaceNameRequired betyder att en plats måste ha ett namn.
	ErrPlaceNameRequired = errors.New("platsen måste ha ett namn")
	// ErrPlaceCoordinates betyder att koordinaten ligger utanför jordens gränser.
	ErrPlaceCoordinates = errors.New("koordinaten ligger utanför världen")
	// ErrPlaceRadius betyder att radien är ogiltig.
	ErrPlaceRadius = errors.New("radien måste vara mellan 1 och 100000 meter")
)

// DefaultPlaceRadiusMeters är radien en ny plats får om inget annat anges.
const DefaultPlaceRadiusMeters = 100

// MaxPlaceRadiusMeters är den största radien som tillåts - 100 km.
const MaxPlaceRadiusMeters = 100000

// Place är en fysisk plats där uppgifter kan slutföras.
//
// Platsen är ett område, inte en punkt: Latitud/Longitude är cirkelns
// mittpunkt och RadiusMeters är dess radie. Radien används för att avgöra
// om en GPS-position ligger inom platsen.
type Place struct {
	ID            int64     `json:"id" example:"1"`
	Name          string    `json:"name" example:"Hem"`
	Latitude      float64   `json:"latitude" example:"59.3293"`
	Longitude     float64   `json:"longitude" example:"18.0686"`
	RadiusMeters  int       `json:"radius_meters" example:"100"`
	Icon          string    `json:"icon,omitempty" example:"heroicons:map-pin"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// DistanceMeters är avståndet från en punkt till platsens mittpunkt i meter.
// Den används för att matcha en GPS-position mot platsens radie.
func (p Place) DistanceMeters(latitude, longitude float64) float64 {
	return haversineMeters(p.Latitude, p.Longitude, latitude, longitude)
}

// ContainsPosition anger om en position ligger inom platsens radie.
func (p Place) ContainsPosition(latitude, longitude float64) bool {
	return p.DistanceMeters(latitude, longitude) <= float64(p.RadiusMeters)
}

// Validate kontrollerar att platsen har giltiga värden.
func (p Place) Validate() error {
	if p.Name == "" {
		return ErrPlaceNameRequired
	}
	if p.Latitude < -90 || p.Latitude > 90 {
		return fmt.Errorf("%w: latitud %v", ErrPlaceCoordinates, p.Latitude)
	}
	if p.Longitude < -180 || p.Longitude > 180 {
		return fmt.Errorf("%w: longitud %v", ErrPlaceCoordinates, p.Longitude)
	}
	if p.RadiusMeters <= 0 || p.RadiusMeters > MaxPlaceRadiusMeters {
		return fmt.Errorf("%w: %d", ErrPlaceRadius, p.RadiusMeters)
	}
	return nil
}

// CreatePlaceRequest är request-bodyn för att skapa en plats.
type CreatePlaceRequest struct {
	Name         string   `json:"name" binding:"required" example:"Hem"`
	Latitude     float64  `json:"latitude" example:"59.3293"`
	Longitude    float64  `json:"longitude" example:"18.0686"`
	RadiusMeters *int     `json:"radius_meters" example:"100"`
	Icon         *string  `json:"icon" example:"heroicons:map-pin"`
}

// UpdatePlaceRequest är request-bodyn för att uppdatera en plats.
//
// Alla fält är pekare - ett fält som saknas lämnas oförändrat.
type UpdatePlaceRequest struct {
	Name         *string  `json:"name"`
	Latitude     *float64 `json:"latitude"`
	Longitude    *float64 `json:"longitude"`
	RadiusMeters *int     `json:"radius_meters"`
	Icon         *string  `json:"icon"`
}

// SetPlacesRequest ersätter platserna på ett mål eller en händelse.
//
// PlaceIDs ersätter hela uppsättningen - en tom lista tar bort alla platser.
type SetPlacesRequest struct {
	PlaceIDs []int64 `json:"place_ids" example:"[1,2]"`
}

// SetCurrentPlaceRequest sätter användarens "jag är här"-plats.
//
// PlaceID = nil (null) rensar den valda platsen.
type SetCurrentPlaceRequest struct {
	PlaceID *int64 `json:"place_id"`
}

// haversineMeters beräknar avståndet mellan två punkter på jordytan i meter.
func haversineMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusMeters = 6371000.0

	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	return 2 * earthRadiusMeters * math.Asin(math.Sqrt(a))
}
