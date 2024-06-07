package main

import (
	"context"
	"fmt"
	"log"
	"recommendation-service/config"
	"recommendation-service/internal/module/recommendation/handler"
	"recommendation-service/internal/module/recommendation/repositories"
	"recommendation-service/internal/module/recommendation/usecases"
	"recommendation-service/internal/pkg/database"
	"recommendation-service/internal/pkg/gorules"
	"recommendation-service/internal/pkg/http"
	"recommendation-service/internal/pkg/httpclient"
	log_internal "recommendation-service/internal/pkg/log"
	"recommendation-service/internal/pkg/messagestream"
	"recommendation-service/internal/pkg/middleware"
	"recommendation-service/internal/pkg/observability"
	"recommendation-service/internal/pkg/redis"
	router "recommendation-service/internal/route"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

func main() {
	cfg := config.InitConfig()

	app, messageRouters := initService(cfg)

	for _, router := range messageRouters {
		ctx := context.Background()
		go func(router *message.Router) {
			err := router.Run(ctx)
			if err != nil {
				log.Fatal(err)
			}
		}(router)
	}

	// start http server
	http.StartHttpServer(app, cfg.HttpServer.Port)
}

func initService(cfg *config.Config) (*fiber.App, []*message.Router) {
	db := database.GetConnection(&cfg.Database)
	redis := redis.SetupClient(&cfg.Redis)
	logger := log_internal.Setup()
	cb := httpclient.InitCircuitBreaker(&cfg.HttpClient, cfg.HttpClient.Type)
	httpClient := httpclient.InitHttpClient(&cfg.HttpClient, cb)

	// init business rules engine
	pathTicketDiscounted := "./assets/ticket-discounted.json"
	breTicketDiscounted, err := gorules.Init(pathTicketDiscounted)
	if err != nil {
		logger.Ctx(context.Background()).Error(fmt.Sprintf("Failed to init BRE: %v", err))
	}

	ctx := context.Background()
	// init message stream
	amqp := messagestream.NewAmpq(&cfg.MessageStream)

	// Init Subscriber
	subscriber, err := amqp.NewSubscriber()
	if err != nil {
		logger.Ctx(ctx).Fatal(fmt.Sprintf("Failed to create subscriber: %v", err))
	}

	// Init Publisher
	publisher, err := amqp.NewPublisher()
	if err != nil {
		logger.Ctx(ctx).Fatal(fmt.Sprintf("Failed to create publisher: %v", err))
	}

	recommendationRepo := repositories.New(db, logger, httpClient, redis, &cfg.UserService, &cfg.TicketService)
	recommendationUsecase := usecases.New(recommendationRepo, breTicketDiscounted)
	middleware := middleware.Middleware{
		Log:  logger,
		Repo: recommendationRepo,
	}

	validator := validator.New()
	recommendationHandler := handler.RecommendationHandler{
		Log:       logger,
		Validator: validator,
		Usecase:   recommendationUsecase,
		Publish:   publisher,
	}

	var messageRouters []*message.Router

	updateVenueStatus, err := messagestream.NewRouter(publisher, "update_venue_status_poisoned", "update_venue_status_handler", "update_venue_status", subscriber, recommendationHandler.UpdateVenueStatus)
	if err != nil {
		logger.Ctx(ctx).Error(fmt.Sprintf("Failed to create consume_booking_queue router: %v", err))
	}

	updateTicketSoldOut, err := messagestream.NewRouter(publisher, "update_ticket_sold_out_poisoned", "update_ticket_sold_out_handler", "update_ticket_sold_out", subscriber, recommendationHandler.UpdateTicketSoldOut)
	if err != nil {
		logger.Ctx(ctx).Error(fmt.Sprintf("Failed to create consume_booking_queue router: %v", err))
	}

	messageRouters = append(messageRouters, updateVenueStatus, updateTicketSoldOut)

	serverHttp := http.SetupHttpEngine()
	conn, serviceName, err := observability.InitConn(cfg)
	if err != nil {
		logger.Ctx(ctx).Fatal(fmt.Sprintf("Failed to create gRPC connection to collector: %v", err))
	}
	// setup log
	observability.InitLogOtel(cfg, serviceName)

	// setup tracer
	observability.InitTracer(conn, serviceName)

	// setup metric
	_, err = observability.InitMeterProvider(conn, serviceName)
	if err != nil {
		logger.Ctx(ctx).Fatal(fmt.Sprintf("Failed to create meter provider: %v", err))
	}

	r := router.Initialize(serverHttp, &recommendationHandler, &middleware)

	return r, messageRouters

}
