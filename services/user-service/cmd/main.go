package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"BootDevClone/services/user-service/internal/api"
	"BootDevClone/services/user-service/internal/infrastructure/repository"
	"BootDevClone/services/user-service/internal/service"
	"BootDevClone/shared/cache"
	"BootDevClone/shared/env"
	"BootDevClone/shared/middleware"
	"BootDevClone/shared/rabbitmq"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	_ = godotenv.Load() // loads .env if present; ignored in production

	dbHost := env.GetString("POSTGRESQL_HOST", "localhost")
	dbPort := env.GetString("POSTGRESQL_PORT", "5432")
	dbName := env.GetString("POSTGRESQL_DB", "coderealm")
	dbUser := env.GetString("POSTGRESQL_USERNAME", "admin")
	dbPassword := env.GetString("POSTGRESQL_PASSWORD", "admin")
	dbSSLMode := env.GetString("POSTGRESQL_SSLMODE", "disable")
	rmqURL := env.GetString("RABBITMQ_URL", "amqp://user:password@localhost:5672/")
	redisURL := env.GetString("REDIS_URL", "redis://localhost:6379/0")

	dsn := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		dbHost, dbPort, dbName, dbUser, dbPassword, dbSSLMode,
	)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "supersecretkey_change_in_production"
	}

	httpAddr := env.GetString("HTTP_ADDR", ":8084")

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
	repo := repository.NewUserRepository(db)

	// Connect to Redis
	redisClient, err := cache.Connect(redisURL)
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v (Caching disabled)", err)
	} else {
		defer redisClient.Close()
		log.Println("Connected to Redis")
	}

	svc := service.NewUserService(repo, jwtSecret, redisClient)
	handler := api.NewAuthHandler(svc, jwtSecret)
	// Connect to RabbitMQ and start consumer
	rmqClient, err := rabbitmq.Connect(rmqURL)
	if err != nil {
		log.Printf("Warning: Failed to connect to RabbitMQ: %v (Event-driven features will be disabled)", err)
	} else {
		defer rmqClient.Close()
		log.Println("Connected to RabbitMQ")
		
		// Start consuming lesson.completed events
		err = rmqClient.Subscribe("user_lesson_completed_queue", "lesson.completed", func(body []byte) error {
			var event struct {
				UserID     string `json:"user_id"`
				LessonID   string `json:"lesson_id"`
				LessonType string `json:"lesson_type"`
				HelpCount  int    `json:"help_count"`
			}
			if err := json.Unmarshal(body, &event); err != nil {
				return err
			}
			
			log.Printf("Received lesson.completed event for user %s, lesson %s", event.UserID, event.LessonID)
			points, _, _, err := svc.CompleteLesson(event.UserID, event.LessonID, event.LessonType, event.HelpCount)
			if err != nil {
				log.Printf("Failed to complete lesson for user %s: %v", event.UserID, err)
				return err // Returns error so it can be nacked/requeued if needed
			}
			log.Printf("Awarded %d points to user %s", points, event.UserID)
			return nil
		})
		
		if err != nil {
			log.Printf("Warning: Failed to subscribe to RabbitMQ: %v", err)
		}
	}

	// Setup Gin
	r := gin.Default()
	r.Use(middleware.GatewayAuthMiddleware())
	
	// CORS is handled by the API Gateway, so we don't need it here.

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "user-service"})
	})

	handler.RegisterRoutes(r)

	log.Printf("User Service listening on %s", httpAddr)
	if err := r.Run(httpAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
