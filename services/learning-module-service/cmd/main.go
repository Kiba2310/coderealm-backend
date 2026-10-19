package main

import (
	handler "BootDevClone/services/learning-module-service/internal/infrastructure/http"
	"BootDevClone/services/learning-module-service/internal/infrastructure/piston"
	"BootDevClone/services/learning-module-service/internal/infrastructure/repository"
	"BootDevClone/services/learning-module-service/internal/service"
	"BootDevClone/shared/cache"
	"BootDevClone/shared/env"
	"BootDevClone/shared/middleware"
	"BootDevClone/shared/rabbitmq"
	"database/sql"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var (
	httpAddr   = func() string { return env.GetString("HTTP_ADDR", ":8082") }
	dbHost     = func() string { return env.GetString("POSTGRESQL_HOST", "localhost") }
	dbPort     = func() string { return env.GetString("POSTGRESQL_PORT", "5432") }
	dbName     = func() string { return env.GetString("POSTGRESQL_DB", "learning_module") }
	dbUser     = func() string { return env.GetString("POSTGRESQL_USERNAME", "postgres") }
	dbPassword = func() string { return env.GetString("POSTGRESQL_PASSWORD", "postgres") }
	dbSSLMode  = func() string { return env.GetString("POSTGRESQL_SSLMODE", "disable") }
	rmqURL     = func() string { return env.GetString("RABBITMQ_URL", "amqp://user:password@localhost:5672/") }
	redisURL   = func() string { return env.GetString("REDIS_URL", "redis://localhost:6379/0") }
)

func main() {
	_ = godotenv.Load() // loads .env if present; ignored in production

	log.Println("Starting Learning Module Service")

	dsn := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		dbHost(), dbPort(), dbName(), dbUser(), dbPassword(), dbSSLMode(),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Connected to PostgreSQL")

	if err := repository.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Migrations applied")

	repo := repository.NewPostgresRepository(db)
	pistonClient := piston.NewClient()

	rmqClient, err := rabbitmq.Connect(rmqURL())
	if err != nil {
		log.Printf("Warning: Failed to connect to RabbitMQ: %v (Event publishing disabled)", err)
	} else {
		defer rmqClient.Close()
		log.Println("Connected to RabbitMQ")
	}

	redisClient, err := cache.Connect(redisURL())
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v (Caching disabled)", err)
	} else {
		defer redisClient.Close()
		log.Println("Connected to Redis")
	}

	svc := service.NewLearningService(repo, pistonClient, rmqClient, redisClient)
	h := handler.NewHandler(svc)

	r := gin.Default()
	r.Use(middleware.GatewayAuthMiddleware())
	h.RegisterRoutes(r)

	log.Printf("Listening on %s", httpAddr())
	if err := r.Run(httpAddr()); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
