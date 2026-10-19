package main

import (
	"database/sql"
	"fmt"
	"log"

	"BootDevClone/services/group-service/internal/api"
	"BootDevClone/services/group-service/internal/infrastructure/repository"
	"BootDevClone/services/group-service/internal/service"
	"BootDevClone/shared/env"
	"BootDevClone/shared/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	_ = godotenv.Load() // loads .env if present; ignored in production

	dbHost := env.GetString("POSTGRESQL_HOST", "localhost")
	dbPort := env.GetString("POSTGRESQL_PORT", "5432")
	dbName := env.GetString("POSTGRESQL_DB", "group_service")
	dbUser := env.GetString("POSTGRESQL_USERNAME", "postgres")
	dbPassword := env.GetString("POSTGRESQL_PASSWORD", "postgres")
	dbSSLMode := env.GetString("POSTGRESQL_SSLMODE", "disable")

	dsn := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		dbHost, dbPort, dbName, dbUser, dbPassword, dbSSLMode,
	)

	httpAddr := env.GetString("HTTP_ADDR", ":8085")

	// Connect to DB
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Run migrations
	if err := repository.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize components
	repo := repository.NewPostgresRepository(db)
	svc := service.NewGroupService(repo)
	handler := api.NewHandler(svc)

	// Setup Gin
	r := gin.Default()
	r.Use(middleware.GatewayAuthMiddleware())

	handler.RegisterRoutes(r)

	log.Printf("Group Service listening on %s", httpAddr)
	if err := r.Run(httpAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
