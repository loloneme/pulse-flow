package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/loloneme/pulse-flow/internal/config"
	"github.com/loloneme/pulse-flow/internal/infrastructure/db/postgres"
	"github.com/loloneme/pulse-flow/internal/infrastructure/logger"
	"github.com/loloneme/pulse-flow/internal/infrastructure/messaging"
	"github.com/loloneme/pulse-flow/internal/infrastructure/messaging/in_memory"
	"github.com/loloneme/pulse-flow/internal/infrastructure/persistence/order"
	"github.com/loloneme/pulse-flow/internal/middleware"
	createOrderRPC "github.com/loloneme/pulse-flow/internal/rpc/create_order"
	createOrderUsecase "github.com/loloneme/pulse-flow/internal/usecase/create_order"
	"github.com/loloneme/pulse-flow/internal/workers/cancellation"
	"github.com/loloneme/pulse-flow/internal/workers/confirmation"
	"github.com/loloneme/pulse-flow/internal/workers/payment"
	paymentMocks "github.com/loloneme/pulse-flow/internal/workers/payment/mocks"
	"github.com/loloneme/pulse-flow/internal/workers/validation"
	validationMocks "github.com/loloneme/pulse-flow/internal/workers/validation/mocks"
	"go.uber.org/zap"
)

type ExternalServices struct {
	AntiFraudService *validationMocks.MockAntiFraudService
	WarehouseService *validationMocks.MockWarehouseService
	UserService      *validationMocks.MockUserService
	PaymentService   *paymentMocks.MockPaymentService
}

type Workers struct {
	ValidationWorker   *validation.Worker
	PaymentWorker      *payment.Worker
	CancellationWorker *cancellation.Worker
	ConfirmationWorker *confirmation.Worker
}

func main() {
	if err := logger.Init(); err != nil {
		panic(fmt.Errorf("failed to init logger: %w", err))
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}

	db, err := postgres.NewFromConfig(context.Background())
	if err != nil {
		panic(fmt.Errorf("failed to connect to postgres: %w", err))
	}

	orderRepo := order.NewRepository(db)
	eventBus := in_memory.New()

	externalServices := setupExternalServices()

	workers := setupWorkers(eventBus, orderRepo, externalServices, cfg.WorkerConfig)

	subscribeWorkers(eventBus, workers)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startWorkers(ctx, workers)

	createOrderService := createOrderUsecase.New(orderRepo, eventBus)
	createOrderHandler := createOrderRPC.New(createOrderService)

	e := echo.New()
	e.Use(echomw.Recover())
	e.Use(echomw.CORS())
	e.Use(middleware.LoggingMiddleware())

	api := e.Group("/api/v1")
	{
		orders := api.Group("/orders")
		{
			orders.POST("", createOrderHandler.CreateOrder)
		}
	}

	addr := ":" + cfg.Port

	go func() {
		logger.Log.Info("Starting HTTP server", zap.String("addr", addr))
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			logger.Log.Error("Failed to start server", zap.Error(err))
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		logger.Log.Fatal("Shutdown failed", zap.Error(err))
	}
}

func setupExternalServices() *ExternalServices {
	return &ExternalServices{
		AntiFraudService: validationMocks.NewMockAntiFraudService(),
		WarehouseService: validationMocks.NewMockWarehouseService(),
		UserService:      validationMocks.NewMockUserService(),
		PaymentService:   paymentMocks.NewMockPaymentService(),
	}
}

func setupWorkers(eventBus *in_memory.Bus, orderRepo *order.Repository, services *ExternalServices, cfg *config.WorkerConfig) *Workers {
	return &Workers{
		ValidationWorker: validation.New(
			cfg,
			logger.NewWorkerLogger("validation"),
			eventBus,
			orderRepo,
			validation.Services{
				WarehouseService: services.WarehouseService,
				AntiFraudService: services.AntiFraudService,
				UserService:      services.UserService,
			},
		),
		PaymentWorker: payment.New(
			cfg,
			logger.NewWorkerLogger("payment"),
			eventBus,
			orderRepo,
			services.PaymentService,
		),
		CancellationWorker: cancellation.New(
			logger.NewWorkerLogger("cancellation"),
			eventBus,
			orderRepo,
		),
		ConfirmationWorker: confirmation.New(
			logger.NewWorkerLogger("confirmation"),
			eventBus,
			orderRepo,
		),
	}
}

func subscribeWorkers(eventBus *in_memory.Bus, workers *Workers) {
	subscriptions := map[messaging.EventType]messaging.Subscriber{
		messaging.OrderCreated:     workers.ValidationWorker,
		messaging.OrderValidated:   workers.PaymentWorker,
		messaging.ValidationFailed: workers.CancellationWorker,
		messaging.PaymentFailed:    workers.CancellationWorker,
		messaging.PaymentSucceeded: workers.ConfirmationWorker,
	}

	for event, handler := range subscriptions {
		if err := eventBus.Subscribe(event, handler); err != nil {
			logger.Log.Fatal("Failed to subscribe", zap.String("event", string(event)), zap.Error(err))
		}
	}

	logger.Log.Info("All workers subscribed successfully")
}

func startWorkers(ctx context.Context, workers *Workers) {
	logger.Log.Info("Starting workers...")
	logger.Log.Info("Workers started and ready to process events")
}
