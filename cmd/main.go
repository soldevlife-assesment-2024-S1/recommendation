package main

import (
	"context"
	"log"
	"recommendation-service/config"
	"recommendation-service/internal/module/recommendation/handler"
	"recommendation-service/internal/module/recommendation/repositories"
	"recommendation-service/internal/module/recommendation/usecases"
	"recommendation-service/internal/pkg/database"
	"recommendation-service/internal/pkg/http"
	"recommendation-service/internal/pkg/httpclient"
	log_internal "recommendation-service/internal/pkg/log"
	"recommendation-service/internal/pkg/messagestream"
	"recommendation-service/internal/pkg/middleware"
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
	logZap := log_internal.SetupLogger()
	log_internal.Init(logZap)
	logger := log_internal.GetLogger()
	cb := httpclient.InitCircuitBreaker(&cfg.HttpClient, cfg.HttpClient.Type)
	httpClient := httpclient.InitHttpClient(&cfg.HttpClient, cb)
	ctx := context.Background()
	// init message stream
	amqp := messagestream.NewAmpq(&cfg.MessageStream)

	// Init Subscriber
	_, err := amqp.NewSubscriber()
	if err != nil {
		logger.Error(ctx, "Failed to create subscriber", err)
	}

	// Init Publisher
	publisher, err := amqp.NewPublisher()
	if err != nil {
		logger.Error(ctx, "Failed to create publisher", err)
	}

	recommendationRepo := repositories.New(db, logger, httpClient, redis)
	recommendationUsecase := usecases.New(recommendationRepo)
	middleware := middleware.Middleware{
		Repo: recommendationRepo,
	}

	validator := validator.New()
	ticketHandler := handler.TicketHandler{
		Log:       logger,
		Validator: validator,
		Usecase:   recommendationUsecase,
		Publish:   publisher,
	}

	var messageRouters []*message.Router

	// incrementTicketStock, err := messagestream.NewRouter(publisher, "increment_stock_ticket_poisoned", "increment_stock_ticket_handler", "increment_stock_ticket", subscriber, ticketHandler.IncrementTicketStock)
	// if err != nil {
	// 	logger.Error(ctx, "Failed to create consume_booking_queue router", err)
	// }

	// decrementTicketStock, err := messagestream.NewRouter(publisher, "decrement_stock_ticket_poisoned", "decrement_stock_ticket_handler", "decrement_stock_ticket", subscriber, ticketHandler.DecrementTicketStock)
	// if err != nil {
	// 	logger.Error(ctx, "Failed to create consume_booking_queue router", err)
	// }

	// messageRouters = append(messageRouters, incrementTicketStock, decrementTicketStock)

	serverHttp := http.SetupHttpEngine()

	r := router.Initialize(serverHttp, &ticketHandler, &middleware)

	return r, messageRouters

}
