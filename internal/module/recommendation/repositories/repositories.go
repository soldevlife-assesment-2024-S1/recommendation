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
	db               *sqlx.DB
	log              log.Logger
	httpClient       *circuit.HTTPClient
	cfgUserService   *config.UserServiceConfig
	cfgTicketService *config.TicketServiceConfig
	redisClient      *redis.Client
}

// FindVenues implements Repositories.
func (r *repositories) FindVenues(ctx context.Context) ([]entity.Venues, error) {
	var venues []entity.Venues
	err := r.db.GetContext(ctx, &venues, "SELECT * FROM venues")
	if err != nil {
		return nil, err
	}
	return venues, nil
}

// FindTicketByRegionName implements Repositories.
func (r *repositories) FindTicketByRegionName(ctx context.Context, regionName string) ([]response.Ticket, error) {
	// call http to ticket service
	url := fmt.Sprintf("http://%s:%s/api/private/ticket?region_name=%s", r.cfgTicketService.Host, r.cfgTicketService.Port, regionName)
	resp, err := r.httpClient.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		r.log.Error(ctx, "Failed to get ticket", resp.StatusCode)
		return nil, errors.BadRequest("Failed to get ticket")
	}

	// parse response
	// var respData []response.Ticket

	// dec := json.NewDecoder(resp.Body)
	// if err := dec.Decode(&respData); err != nil {
	// 	return nil, err
	// }

	var respBase response.BaseResponse

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&respBase); err != nil {
		return nil, err
	}

	respBase.Data = respBase.Data.(map[string]interface{})

	respData := make([]response.Ticket, 0)
	for _, v := range respBase.Data.(map[string]interface{})["data"].([]interface{}) {
		respData = append(respData, v.(response.Ticket))
	}

	return respData, nil
}

// FindUserProfile implements Repositories.
func (r *repositories) FindUserProfile(ctx context.Context, userID int64) (response.UserProfile, error) {
	// http call to user service
	url := fmt.Sprintf("http://%s:%s/api/private/user/profile?user_id=%d", r.cfgUserService.Host, r.cfgUserService.Port, userID)
	resp, err := r.httpClient.Get(url)
	if err != nil {
		return response.UserProfile{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		r.log.Error(ctx, "Failed to get user profile", resp.StatusCode)
		return response.UserProfile{}, errors.BadRequest("Failed to get user profile")
	}

	// parse response
	// var respData response.UserProfile

	// dec := json.NewDecoder(resp.Body)
	// if err := dec.Decode(&respData); err != nil {
	// 	return response.UserProfile{}, err
	// }

	var respBase response.BaseResponse

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&respBase); err != nil {
		return response.UserProfile{}, err
	}

	respBase.Data = respBase.Data.(map[string]interface{})
	respData := response.UserProfile{
		UserID:   int(respBase.Data.(map[string]interface{})["user_id"].(float64)),
		Username: respBase.Data.(map[string]interface{})["username"].(string),
		Email:    respBase.Data.(map[string]interface{})["email"].(string),
		Region:   respBase.Data.(map[string]interface{})["region"].(string),
	}

	return respData, nil
}

// FindVenueByName implements Repositories.
func (r *repositories) FindVenueByName(ctx context.Context, name string) (entity.Venues, error) {
	var venue entity.Venues
	err := r.db.GetContext(ctx, &venue, "SELECT * FROM venues WHERE name = $1", name)
	if err != nil {
		return entity.Venues{}, err
	}
	return venue, nil
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
	ValidateToken(ctx context.Context, token string) (response.UserServiceValidate, error)
	FindUserProfile(ctx context.Context, userID int64) (response.UserProfile, error)
	FindTicketByRegionName(ctx context.Context, regionName string) ([]response.Ticket, error)
	// db
	UpsertVenue(ctx context.Context, payload entity.Venues) error
	FindVenueByName(ctx context.Context, name string) (entity.Venues, error)
	FindVenues(ctx context.Context) ([]entity.Venues, error)
}

func New(db *sqlx.DB, log log.Logger, httpClient *circuit.HTTPClient, redisClient *redis.Client, userService *config.UserServiceConfig, ticketService *config.TicketServiceConfig) Repositories {
	return &repositories{
		db:               db,
		log:              log,
		httpClient:       httpClient,
		redisClient:      redisClient,
		cfgUserService:   userService,
		cfgTicketService: ticketService,
	}
}

func (r *repositories) ValidateToken(ctx context.Context, token string) (response.UserServiceValidate, error) {
	// http call to user service
	url := fmt.Sprintf("http://%s:%s/api/private/token/validate?token=%s", r.cfgUserService.Host, r.cfgUserService.Port, token)
	resp, err := r.httpClient.Get(url)
	if err != nil {
		return response.UserServiceValidate{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		r.log.Error(ctx, "Invalid token", resp.StatusCode)
		return response.UserServiceValidate{}, errors.BadRequest("Invalid token")
	}

	// parse response
	var respBase response.BaseResponse

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&respBase); err != nil {
		return response.UserServiceValidate{
			IsValid: false,
			UserID:  0,
		}, err
	}

	respBase.Data = respBase.Data.(map[string]interface{})
	respData := response.UserServiceValidate{
		IsValid:   respBase.Data.(map[string]interface{})["is_valid"].(bool),
		UserID:    int64(respBase.Data.(map[string]interface{})["user_id"].(float64)),
		EmailUser: respBase.Data.(map[string]interface{})["email_user"].(string),
	}

	if !respData.IsValid {
		r.log.Error(ctx, "Invalid token", resp.StatusCode)
		return response.UserServiceValidate{
			IsValid: false,
			UserID:  0,
		}, errors.BadRequest("Invalid token")
	}

	// validate token
	return response.UserServiceValidate{
		IsValid:   respData.IsValid,
		UserID:    respData.UserID,
		EmailUser: respData.EmailUser,
	}, nil
}
