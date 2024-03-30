package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"recommendation-service/config"
	"recommendation-service/internal/module/recommendation/models/entity"
	"recommendation-service/internal/module/recommendation/models/response"
	"recommendation-service/internal/pkg/errors"
	"recommendation-service/internal/pkg/log"

	"github.com/goccy/go-json"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	circuit "github.com/rubyist/circuitbreaker"
)

type repositories struct {
	db             *sqlx.DB
	log            log.Logger
	httpClient     *circuit.HTTPClient
	cfgUserService *config.UserService
	redisClient    *redis.Client
}

// UpsertVenue implements Repositories.
func (r *repositories) UpsertVenue(ctx context.Context, payload entity.Venues) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	// Check if the venue already exists
	var existingVenueID int64
	err = tx.QueryRowContext(ctx, "SELECT id FROM venues WHERE name = $1 FOR UPDATE", payload.Name).Scan(&existingVenueID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if err == sql.ErrNoRows {
		// Venue does not exist, insert a new record
		_, err = tx.ExecContext(ctx, "INSERT INTO venues (name, is_sold_out, is_first_sold_out) VALUES ($1, $2, $3)", payload.Name, payload.IsSoldOut, payload.IsFirstSoldOut)
		if err != nil {
			return err
		}
	} else {
		// Venue already exists, update the existing record
		_, err = tx.ExecContext(ctx, "UPDATE venues SET is_sold_out = $1, SET is_sold_out_first = $2 WHERE id = $3", payload.IsSoldOut, payload.IsFirstSoldOut, existingVenueID)
		if err != nil {
			return err
		}
	}

	return nil
}

type Repositories interface {
	// http
	ValidateToken(ctx context.Context, token string) (bool, error)
	// db
	UpsertVenue(ctx context.Context, payload entity.Venues) error
}

func New(db *sqlx.DB, log log.Logger, httpClient *circuit.HTTPClient, redisClient *redis.Client) Repositories {
	return &repositories{
		db:          db,
		log:         log,
		httpClient:  httpClient,
		redisClient: redisClient,
	}
}

func (r *repositories) ValidateToken(ctx context.Context, token string) (bool, error) {
	// http call to user service
	url := fmt.Sprintf("http://%s:%s/api/private/token/validate?token=%s", r.cfgUserService.Host, r.cfgUserService.Port, token)
	resp, err := r.httpClient.Get(url)
	if err != nil {
		return false, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		r.log.Error(ctx, "Invalid token", resp.StatusCode)
		return false, errors.BadRequest("Invalid token")
	}

	// parse response
	var respData response.UserServiceValidate

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&respData); err != nil {
		return false, err
	}

	if !respData.IsValid {
		r.log.Error(ctx, "Invalid token", resp.StatusCode)
		return false, errors.BadRequest("Invalid token")
	}

	// validate token
	return true, nil
}
