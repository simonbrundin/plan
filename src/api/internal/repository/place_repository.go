package repository

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"plan-api/internal/database"
	"plan-api/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrPlaceNotFound betyder att den valda platsen inte finns.
var ErrPlaceNotFound = errors.New("place not found")

type PlaceRepository struct {
	pool *pgxpool.Pool
}

func NewPlaceRepository() *PlaceRepository {
	return &PlaceRepository{
		pool: database.GetPool(),
	}
}

// placeColumns är de kolumner som läses för en plats.
const placeColumns = `p.id, p.name, p.latitude, p.longitude, p.radius_meters,
	COALESCE(p.icon, ''), p.created_at, p.updated_at`

// scanPlace läser en platsrad.
func scanPlace(row pgx.Row) (*models.Place, error) {
	var p models.Place
	if err := row.Scan(&p.ID, &p.Name, &p.Latitude, &p.Longitude, &p.RadiusMeters, &p.Icon, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, fmt.Errorf("failed to scan place: %w", err)
	}
	return &p, nil
}

// List returnerar alla platser för en användare.
func (r *PlaceRepository) List(ctx context.Context, userID int64) ([]models.Place, error) {
	query := `SELECT ` + placeColumns + `
		FROM places p
		WHERE p.user_id = $1
		ORDER BY p.name`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query places: %w", err)
	}
	defer rows.Close()

	places := []models.Place{}
	for rows.Next() {
		p, err := scanPlace(rows)
		if err != nil {
			return nil, err
		}
		places = append(places, *p)
	}

	return places, rows.Err()
}

// Get returnerar en enskild plats. En plats som inte tillhör användaren
// returneras som nil (inte som ett fel) - samma mönster som övriga repos.
func (r *PlaceRepository) Get(ctx context.Context, userID, placeID int64) (*models.Place, error) {
	query := `SELECT ` + placeColumns + `
		FROM places p
		WHERE p.user_id = $1 AND p.id = $2`

	p, err := scanPlace(r.pool.QueryRow(ctx, query, userID, placeID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return p, nil
}

// Create skapar en ny plats.
func (r *PlaceRepository) Create(ctx context.Context, userID int64, req models.CreatePlaceRequest) (*models.Place, error) {
	radius := models.DefaultPlaceRadiusMeters
	if req.RadiusMeters != nil {
		radius = *req.RadiusMeters
	}
	icon := ""
	if req.Icon != nil {
		icon = *req.Icon
	}

	query := `INSERT INTO places (user_id, name, latitude, longitude, radius_meters, icon)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, latitude, longitude, radius_meters, COALESCE(icon, ''), created_at, updated_at`

	p, err := scanPlace(r.pool.QueryRow(ctx, query, userID, req.Name, req.Latitude, req.Longitude, radius, icon))
	if err != nil {
		return nil, fmt.Errorf("failed to create place: %w", err)
	}
	return p, nil
}

// Update uppdaterar en plats. Endast fält som sätts i requesten ändras.
func (r *PlaceRepository) Update(ctx context.Context, userID, placeID int64, req models.UpdatePlaceRequest) (*models.Place, error) {
	query := `UPDATE places SET
		name = COALESCE($3, name),
		latitude = COALESCE($4, latitude),
		longitude = COALESCE($5, longitude),
		radius_meters = COALESCE($6, radius_meters),
		icon = COALESCE($7, icon),
		updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $1 AND id = $2
		RETURNING id, name, latitude, longitude, radius_meters, COALESCE(icon, ''), created_at, updated_at`

	p, err := scanPlace(r.pool.QueryRow(ctx, query,
		userID, placeID,
		req.Name, req.Latitude, req.Longitude, req.RadiusMeters, req.Icon))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to update place: %w", err)
	}
	return p, nil
}

// Delete tar bort en plats. Kopplingar till mål och händelser försvinner
// med platsen (kaskad i databasen).
func (r *PlaceRepository) Delete(ctx context.Context, userID, placeID int64) (bool, error) {
	tag, err := r.pool.Exec(ctx, `DELETE FROM places WHERE user_id = $1 AND id = $2`, userID, placeID)
	if err != nil {
		return false, fmt.Errorf("failed to delete place: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

// GetGoalPlaces returnerar platserna kopplade till ett mål.
func (r *PlaceRepository) GetGoalPlaces(ctx context.Context, goalID int64) ([]models.Place, error) {
	query := `SELECT ` + placeColumns + `
		FROM places p
		INNER JOIN goal_places gp ON gp.place_id = p.id
		WHERE gp.goal_id = $1
		ORDER BY p.name`

	rows, err := r.pool.Query(ctx, query, goalID)
	if err != nil {
		return nil, fmt.Errorf("failed to query goal places: %w", err)
	}
	defer rows.Close()

	places := []models.Place{}
	for rows.Next() {
		p, err := scanPlace(rows)
		if err != nil {
			return nil, err
		}
		places = append(places, *p)
	}
	return places, rows.Err()
}

// SetGoalPlaces ersätter platserna på ett mål.
//
// Endast platser som tillhör användaren kopplas - en plats från någon
// annan ignoreras.
func (r *PlaceRepository) SetGoalPlaces(ctx context.Context, userID, goalID int64, placeIDs []int64) ([]models.Place, error) {
	return r.setPlaces(ctx, userID, goalID, placeIDs, "goal_places", "goal_id")
}

// GetEventPlaces returnerar platserna kopplade till en kalenderhändelse.
func (r *PlaceRepository) GetEventPlaces(ctx context.Context, eventID int64) ([]models.Place, error) {
	query := `SELECT ` + placeColumns + `
		FROM places p
		INNER JOIN calendar_event_places ep ON ep.place_id = p.id
		WHERE ep.event_id = $1
		ORDER BY p.name`

	rows, err := r.pool.Query(ctx, query, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to query event places: %w", err)
	}
	defer rows.Close()

	places := []models.Place{}
	for rows.Next() {
		p, err := scanPlace(rows)
		if err != nil {
			return nil, err
		}
		places = append(places, *p)
	}
	return places, rows.Err()
}

// SetEventPlaces ersätter platserna på en kalenderhändelse.
func (r *PlaceRepository) SetEventPlaces(ctx context.Context, userID, eventID int64, placeIDs []int64) ([]models.Place, error) {
	return r.setPlaces(ctx, userID, eventID, placeIDs, "calendar_event_places", "event_id")
}

// setPlaces ersätter kopplingarna mellan ett mål eller en händelse och
// en uppsättning platser.
func (r *PlaceRepository) setPlaces(ctx context.Context, userID, ownerID int64, placeIDs []int64, table, ownerColumn string) ([]models.Place, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, fmt.Sprintf(`DELETE FROM %s WHERE %s = $1`, table, ownerColumn), ownerID); err != nil {
		return nil, fmt.Errorf("failed to clear places: %w", err)
	}

	unique := uniqueIDs(placeIDs)
	for _, placeID := range unique {
		// Endast platser som tillhör användaren kopplas.
		query := fmt.Sprintf(`INSERT INTO %s (%s, place_id)
			SELECT $1, p.id FROM places p WHERE p.user_id = $2 AND p.id = $3
			ON CONFLICT DO NOTHING`, table, ownerColumn)
		if _, err := tx.Exec(ctx, query, ownerID, userID, placeID); err != nil {
			return nil, fmt.Errorf("failed to link place: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit places: %w", err)
	}

	return r.placesForOwner(ctx, table, ownerColumn, ownerID)
}

// placesForOwner läser tillbaka platserna kopplade till ett mål eller en
// händelse.
func (r *PlaceRepository) placesForOwner(ctx context.Context, table, ownerColumn string, ownerID int64) ([]models.Place, error) {
	query := fmt.Sprintf(`SELECT %s
		FROM places p
		INNER JOIN %s l ON l.place_id = p.id
		WHERE l.%s = $1
		ORDER BY p.name`, placeColumns, table, ownerColumn)

	rows, err := r.pool.Query(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query places: %w", err)
	}
	defer rows.Close()

	places := []models.Place{}
	for rows.Next() {
		p, err := scanPlace(rows)
		if err != nil {
			return nil, err
		}
		places = append(places, *p)
	}
	return places, rows.Err()
}

// GetPlacesForGoals returnerar platserna för flera mål samtidigt.
//
// Nyckeln är målets id. Mål utan platser finns inte i kartan.
func (r *PlaceRepository) GetPlacesForGoals(ctx context.Context, goalIDs []int64) (map[int64][]models.Place, error) {
	result := map[int64][]models.Place{}
	if len(goalIDs) == 0 {
		return result, nil
	}

	query := `SELECT gp.goal_id, ` + placeColumns + `
		FROM places p
		INNER JOIN goal_places gp ON gp.place_id = p.id
		WHERE gp.goal_id = ANY($1)
		ORDER BY p.name`

	rows, err := r.pool.Query(ctx, query, goalIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to query goal places: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var goalID int64
		var p models.Place
		if err := rows.Scan(&goalID, &p.ID, &p.Name, &p.Latitude, &p.Longitude, &p.RadiusMeters, &p.Icon, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan goal place: %w", err)
		}
		result[goalID] = append(result[goalID], p)
	}
	return result, rows.Err()
}

// GetPlacesForEvents returnerar platserna för flera kalenderhändelser.
func (r *PlaceRepository) GetPlacesForEvents(ctx context.Context, eventIDs []int64) (map[int64][]models.Place, error) {
	result := map[int64][]models.Place{}
	if len(eventIDs) == 0 {
		return result, nil
	}

	query := `SELECT ep.event_id, ` + placeColumns + `
		FROM places p
		INNER JOIN calendar_event_places ep ON ep.place_id = p.id
		WHERE ep.event_id = ANY($1)
		ORDER BY p.name`

	rows, err := r.pool.Query(ctx, query, eventIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to query event places: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var eventID int64
		var p models.Place
		if err := rows.Scan(&eventID, &p.ID, &p.Name, &p.Latitude, &p.Longitude, &p.RadiusMeters, &p.Icon, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan event place: %w", err)
		}
		result[eventID] = append(result[eventID], p)
	}
	return result, rows.Err()
}

// GetCurrentPlace returnerar användarens valda "jag är här"-plats.
//
// Nil betyder att ingen plats är vald.
func (r *PlaceRepository) GetCurrentPlace(ctx context.Context, userID int64) (*models.Place, error) {
	query := `SELECT ` + placeColumns + `
		FROM places p
		INNER JOIN users u ON u.current_place_id = p.id
		WHERE u.id = $1`

	p, err := scanPlace(r.pool.QueryRow(ctx, query, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query current place: %w", err)
	}
	return p, nil
}

// SetCurrentPlace sätter (eller rensar) användarens "jag är här"-plats.
func (r *PlaceRepository) SetCurrentPlace(ctx context.Context, userID int64, placeID *int64) error {
	if placeID == nil {
		_, err := r.pool.Exec(ctx, `UPDATE users SET current_place_id = NULL WHERE id = $1`, userID)
		if err != nil {
			return fmt.Errorf("failed to clear current place: %w", err)
		}
		return nil
	}

	tag, err := r.pool.Exec(ctx, `UPDATE users SET current_place_id = $1 WHERE id = $2`, *placeID, userID)
	if err != nil {
		return fmt.Errorf("failed to set current place: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrPlaceNotFound
	}
	return nil
}

// GetNearbyPlaces returnerar de platser vars radie innehåller positionen.
//
// Matchningen sker i Go med haversine-formeln, eftersom radien är per
// plats och inte kan uttryckas som en gemensam WHERE-villkor.
func (r *PlaceRepository) GetNearbyPlaces(ctx context.Context, userID int64, latitude, longitude float64) ([]models.Place, error) {
	places, err := r.List(ctx, userID)
	if err != nil {
		return nil, err
	}

	nearby := []models.Place{}
	for _, p := range places {
		if p.ContainsPosition(latitude, longitude) {
			nearby = append(nearby, p)
		}
	}

	sort.Slice(nearby, func(i, j int) bool {
		return nearby[i].DistanceMeters(latitude, longitude) < nearby[j].DistanceMeters(latitude, longitude)
	})

	return nearby, nil
}

// uniqueIDs tar bort dubbletter ur en lista med id:n.
func uniqueIDs(ids []int64) []int64 {
	seen := map[int64]bool{}
	unique := []int64{}
	for _, id := range ids {
		if id == 0 || seen[id] {
			continue
		}
		seen[id] = true
		unique = append(unique, id)
	}
	return unique
}
