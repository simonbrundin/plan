package models

import (
	"errors"
	"testing"
)

func TestPlaceValidate(t *testing.T) {
	tests := []struct {
		name    string
		place   Place
		wantErr error
	}{
		{
			name:  "a complete place is valid",
			place: Place{Name: "Hem", Latitude: 59.3293, Longitude: 18.0686, RadiusMeters: 80},
		},
		{
			name:    "a place without a name is invalid",
			place:   Place{Latitude: 59.3293, Longitude: 18.0686, RadiusMeters: 80},
			wantErr: ErrPlaceNameRequired,
		},
		{
			name:    "a latitude above 90 is invalid",
			place:   Place{Name: "Hem", Latitude: 91, Longitude: 18.0686, RadiusMeters: 80},
			wantErr: ErrPlaceCoordinates,
		},
		{
			name:    "a latitude below -90 is invalid",
			place:   Place{Name: "Hem", Latitude: -91, Longitude: 18.0686, RadiusMeters: 80},
			wantErr: ErrPlaceCoordinates,
		},
		{
			name:    "a longitude above 180 is invalid",
			place:   Place{Name: "Hem", Latitude: 59.3293, Longitude: 181, RadiusMeters: 80},
			wantErr: ErrPlaceCoordinates,
		},
		{
			name:    "a longitude below -180 is invalid",
			place:   Place{Name: "Hem", Latitude: 59.3293, Longitude: -181, RadiusMeters: 80},
			wantErr: ErrPlaceCoordinates,
		},
		{
			name:    "a radius of zero is invalid",
			place:   Place{Name: "Hem", Latitude: 59.3293, Longitude: 18.0686, RadiusMeters: 0},
			wantErr: ErrPlaceRadius,
		},
		{
			name:    "a negative radius is invalid",
			place:   Place{Name: "Hem", Latitude: 59.3293, Longitude: 18.0686, RadiusMeters: -1},
			wantErr: ErrPlaceRadius,
		},
		{
			name:    "a radius above the maximum is invalid",
			place:   Place{Name: "Hem", Latitude: 59.3293, Longitude: 18.0686, RadiusMeters: MaxPlaceRadiusMeters + 1},
			wantErr: ErrPlaceRadius,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.place.Validate()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestPlaceContainsPosition(t *testing.T) {
	// En plats i mitten av Stockholm med 100 meters radie.
	place := Place{Name: "Hem", Latitude: 59.3293, Longitude: 18.0686, RadiusMeters: 100}

	tests := []struct {
		name      string
		latitude  float64
		longitude float64
		want      bool
	}{
		{
			name:      "the centre of the place is inside",
			latitude:  59.3293,
			longitude: 18.0686,
			want:      true,
		},
		{
			name:      "a position about 50 meters away is inside",
			latitude:  59.32975,
			longitude: 18.0686,
			want:      true,
		},
		{
			name:      "a position about 2 kilometers away is outside",
			latitude:  59.3473,
			longitude: 18.0686,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := place.ContainsPosition(tt.latitude, tt.longitude); got != tt.want {
				t.Errorf("ContainsPosition(%v, %v) = %v, want %v", tt.latitude, tt.longitude, got, tt.want)
			}
		})
	}
}

func TestPlaceDistanceMeters(t *testing.T) {
	place := Place{Latitude: 59.3293, Longitude: 18.0686}

	// Ett dygns promenad bort ska mäta i storleksordningen kilometer,
	// inte meter - kontrollerar att formeln använder jordradien.
	distance := place.DistanceMeters(59.3293, 18.0886)
	if distance < 1000 || distance > 1500 {
		t.Errorf("DistanceMeters() = %v, want between 1000 and 1500 meters", distance)
	}
}
