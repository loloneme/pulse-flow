package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/loloneme/pulse-flow/internal/infrastructure/db/postgres"
	"github.com/loloneme/pulse-flow/internal/infrastructure/messaging/in_memory"
	"github.com/loloneme/pulse-flow/internal/infrastructure/persistence/order"
	createOrderRPC "github.com/loloneme/pulse-flow/internal/rpc/create_order"
	createOrderUsecase "github.com/loloneme/pulse-flow/internal/usecase/create_order"
	"github.com/loloneme/pulse-flow/internal/workers/cancellation"
	"github.com/loloneme/pulse-flow/internal/workers/confirmation"
	"github.com/loloneme/pulse-flow/internal/workers/payment"
	paymentMocks "github.com/loloneme/pulse-flow/internal/workers/payment/mocks"
	"github.com/loloneme/pulse-flow/internal/workers/validation"
	validationMocks "github.com/loloneme/pulse-flow/internal/workers/validation/mocks"
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
	db, err := postgres.NewFromConfig(context.Background())
	if err != nil {
		panic(fmt.Errorf("failed to connect to postgres: %w", err))
	}

	orderRepo := order.NewRepository(db)
	eventBus := in_memory.New()

	// Setup external services
	externalServices := setupExternalServices()

	// Setup workers
	workers := setupWorkers(eventBus, orderRepo, externalServices)

	// Subscribe workers to events
	subscribeWorkers(eventBus, workers)

	// Start workers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startWorkers(ctx, workers)

	createOrderService := createOrderUsecase.New(orderRepo, eventBus)
	createOrderHandler := createOrderRPC.New(createOrderService)

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	api := e.Group("/api/v1")
	{
		orders := api.Group("/orders")
		{
			orders.POST("", createOrderHandler.CreateOrder)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	go func() {
		log.Printf("Starting reviewers-app HTTP server on %s\n", addr)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Printf("failed to start server: %v", err)
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	// Cancel workers context
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		e.Logger.Fatal(err)
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

func setupWorkers(eventBus *in_memory.Bus, orderRepo *order.Repository, services *ExternalServices) *Workers {
	return &Workers{
		ValidationWorker: validation.New(
			eventBus,
			orderRepo,
			services.WarehouseService,
			services.AntiFraudService,
			services.UserService,
		),
		PaymentWorker: payment.New(
			eventBus,
			orderRepo,
			services.PaymentService,
		),
		CancellationWorker: cancellation.New(
			eventBus,
			orderRepo,
		),
		ConfirmationWorker: confirmation.New(
			eventBus,
			orderRepo,
		),
	}
}

func subscribeWorkers(eventBus *in_memory.Bus, workers *Workers) {
	if err := eventBus.Subscribe("OrderCreated", workers.ValidationWorker); err != nil {
		log.Fatalf("Failed to subscribe ValidationWorker: %v", err)
	}
	if err := eventBus.Subscribe("OrderValidated", workers.PaymentWorker); err != nil {
		log.Fatalf("Failed to subscribe PaymentWorker: %v", err)
	}
	if err := eventBus.Subscribe("ValidationFailed", workers.CancellationWorker); err != nil {
		log.Fatalf("Failed to subscribe CancellationWorker to ValidationFailed: %v", err)
	}
	if err := eventBus.Subscribe("PaymentFailed", workers.CancellationWorker); err != nil {
		log.Fatalf("Failed to subscribe CancellationWorker to PaymentFailed: %v", err)
	}
	if err := eventBus.Subscribe("PaymentSucceeded", workers.ConfirmationWorker); err != nil {
		log.Fatalf("Failed to subscribe ConfirmationWorker: %v", err)
	}
	log.Println("All workers subscribed successfully")
}

func startWorkers(ctx context.Context, workers *Workers) {
	log.Println("Starting workers...")
	log.Println("Workers started and ready to process events")
}
