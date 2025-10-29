// cmd/server/main.go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"url-shortener/internal/config"
	"url-shortener/internal/handlers"
	"url-shortener/internal/repository/cache"
	"url-shortener/internal/repository/postgres"
	"url-shortener/internal/service"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	"github.com/gorilla/mux"
)

func main() {
	// 1. Загрузка конфигурации
	cfg := config.LoadConfig()
	log.Println("Configuration loaded")

	// 2. Инициализация PostgreSQL
	db, err := initPostgres(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()
	log.Println("PostgreSQL connected")

	// 3. Инициализация Redis
	redisClient := initRedis(cfg)
	defer redisClient.Close()
	log.Println("Redis connected")

	// 4. Инициализация репозиториев
	urlRepo := postgres.NewPostgresURLRepo(db)
	clickRepo := postgres.NewPostgresClickRepo(db)
	cacheRepo := cache.NewCacheRepository(redisClient)

	// 5. Инициализация сервисов
	urlService := service.NewURLService(urlRepo, cacheRepo)
	analyticsService := service.NewAnalyticsService(clickRepo, urlRepo)
	workerService := service.NewWorkerService(analyticsService, 5) // 5 воркеров

	// 6. Инициализация хендлеров
	urlHandler := handlers.NewURLHandler(urlService, workerService)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)

	// 7. Настройка роутинга с middleware
	router := mux.NewRouter()
	router.Use(handlers.LoggingMiddleware)
	router.Use(handlers.RecoveryMiddleware)
	router.Use(handlers.CORSMiddleware)

	// API endpoints
	router.HandleFunc("/api/v1/urls", urlHandler.CreateShortURL).Methods("POST")
	router.HandleFunc("/api/v1/urls/{shortCode}", urlHandler.GetURLInfo).Methods("GET")
	router.HandleFunc("/api/v1/analytics/{shortCode}", analyticsHandler.GetAnalytics).Methods("GET")

	// Redirect endpoint
	router.HandleFunc("/{shortCode}", urlHandler.Redirect)

	// 8. Настройка HTTP сервера
	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 9. Graceful shutdown
	go func() {
		log.Printf("Server starting on port %s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Ожидание сигнала для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown воркеров
	workerService.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func initPostgres(cfg *config.Config) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

func initRedis(cfg *config.Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: "",
		DB:       0,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	return client
}
