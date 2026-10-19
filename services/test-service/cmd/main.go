package main

import (
	handler "BootDevClone/services/test-service/internal/infrastructure/http"
	"BootDevClone/services/test-service/internal/infrastructure/piston"
	"BootDevClone/services/test-service/internal/infrastructure/repository"
	"BootDevClone/services/test-service/internal/service"
	"BootDevClone/shared/env"
	"BootDevClone/shared/middleware"
	"database/sql"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var (
	httpAddr   = func() string { return env.GetString("TEST_HTTP_ADDR", ":8083") }
	dbHost     = func() string { return env.GetString("POSTGRESQL_HOST", "localhost") }
	dbPort     = func() string { return env.GetString("POSTGRESQL_PORT", "5432") }
	dbName     = func() string { return env.GetString("POSTGRESQL_DB", "test_service") }
	dbUser     = func() string { return env.GetString("POSTGRESQL_USERNAME", "postgres") }
	dbPassword = func() string { return env.GetString("POSTGRESQL_PASSWORD", "postgres") }
	dbSSLMode  = func() string { return env.GetString("POSTGRESQL_SSLMODE", "disable") }
)

func main() {
	_ = godotenv.Load() // loads .env if present; ignored in production

	log.Println("Starting Test Service")

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
	svc := service.NewTestService(repo, pistonClient)
	h := handler.NewHandler(svc)

	r := gin.Default()
	r.Use(middleware.GatewayAuthMiddleware())
	h.RegisterRoutes(r)

	log.Printf("Listening on %s", httpAddr())
	if err := r.Run(httpAddr()); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
