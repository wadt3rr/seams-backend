package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"seams-backend/internal/config"
	listCategories "seams-backend/internal/http-server/handlers/categories/list"
	getCustomerByID "seams-backend/internal/http-server/handlers/customers/get"
	listCustomers "seams-backend/internal/http-server/handlers/customers/list"
	getInvoiceByOrderID "seams-backend/internal/http-server/handlers/invoices/get"
	listInvoices "seams-backend/internal/http-server/handlers/invoices/list"
	updateInvoice "seams-backend/internal/http-server/handlers/invoices/update"
	listOrdersByEmail "seams-backend/internal/http-server/handlers/orders/byEmail"
	getOrderByID "seams-backend/internal/http-server/handlers/orders/get"
	listOrders "seams-backend/internal/http-server/handlers/orders/list"
	orderSave "seams-backend/internal/http-server/handlers/orders/save"
	getProductsByCategory "seams-backend/internal/http-server/handlers/products/byCategory"
	getProductBySku "seams-backend/internal/http-server/handlers/products/bySku"
	listProducts "seams-backend/internal/http-server/handlers/products/list"
	saveProduct "seams-backend/internal/http-server/handlers/products/save"
	saveRequest "seams-backend/internal/http-server/handlers/requests/save"
	mwLogger "seams-backend/internal/http-server/middleware/logger"
	"seams-backend/internal/lib/logger/sl"
	orderService "seams-backend/internal/services/orderService"
	productService "seams-backend/internal/services/productService"
	requestservice "seams-backend/internal/services/requestService"

	"seams-backend/internal/storage/postgres"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func init() {
	// Загружаем .env только в режиме local/dev
	_ = godotenv.Load(".env")
}

func main() {
	// Load config
	cfg := config.MustLoad()

	os.MkdirAll("./uploads", os.ModePerm)

	// Setup logger
	log := setupLogger(cfg.Env)

	ctx := context.Background()
	// dsn := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
	// 	cfg.Database.User, cfg.Database.Pass, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name, cfg.Database.SSLMode)
	// Connect to database
	log.Info("Connecting to database",
		slog.String("host", cfg.Database.Host),
		slog.Int("port", cfg.Database.Port),
	)

	storage, err := postgres.NewStorage(ctx, cfg.Database.Dsn, cfg.Database.MigrationsPath)
	if err != nil {
		log.Error("Failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer storage.Close()

	// Setup router
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000", "http://seam-s.shop", "https://seam-s.shop"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
	}))

	productService := productService.NewProductService(storage)
	router.Route("/products", func(r chi.Router) {
		r.Get("/", listProducts.New(log, productService))
		//r.Get("/{id}", getProductByID.New(log, storage))
		r.Get("/{sku}", getProductBySku.New(log, storage))
		r.Post("/new", saveProduct.New(log, storage))
	})

	orderService := orderService.NewOrderService(storage)

	router.Route("/admin", func(r chi.Router) {
		r.Get("/orders", listOrders.New(log, storage))
		r.Get("/orders/{id}", getOrderByID.New(log, orderService))
	})

	router.Route("/orders", func(r chi.Router) {
		r.Post("/new", orderSave.New(log, orderService))

	})

	router.Route("/track", func(r chi.Router) {
		r.Get("/", listOrdersByEmail.New(log, orderService))
		r.Get("/{id}", getOrderByID.New(log, orderService))
	})

	router.Route("/categories", func(r chi.Router) {
		r.Get("/", listCategories.New(log, storage))
		r.Get("/{slug}/products", getProductsByCategory.New(log, productService))
	})

	router.Route("/customers", func(r chi.Router) {
		r.Get("/", listCustomers.New(log, storage))
		r.Get("/{id}", getCustomerByID.New(log, storage))
	})

	router.Route("/invoices", func(r chi.Router) {
		r.Get("/", listInvoices.New(log, storage))
		r.Get("/{id}", getInvoiceByOrderID.New(log, storage))
		r.Patch("/update/{id}", updateInvoice.New(log, storage))
	})

	requestsService := requestservice.New(storage)
	router.Route("/requests", func(r chi.Router) {
		r.Post("/new", saveRequest.New(log, requestsService))
	})

	// Start server
	log.Info("starting server", slog.String("address", cfg.HTTPServer.Address))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error("failed to start server")
		}
	}()

	log.Info("server started")

	// Wait for interrupt signal to gracefully shutdown the server
	<-done
	log.Info("stopping server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", sl.Err(err))

		return
	}

	// TODO: close storage

	log.Info("server stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)

	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)

	default: //If env config is missing or invalid, use production logger
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
